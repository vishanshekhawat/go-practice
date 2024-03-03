package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/net/html/charset"
	"golang.org/x/sync/errgroup"
)

var (
	initialChunkSize int64 = 2 * 1024 * 1024
	// maxChunkSize            = 5 * 1024 * 1024 // 5 MB, maximum chunk size to avoid too large chunks
	//minChunkSize            = 512 * 1024      // 512 KB, minimum chunk size to maintain efficiency
	//progressUpdateSec       = 5
)

var totalDownloaded int64

func DownloadFileInChunksWithProgressBar(url string, destination string, numChunks int) {

	chunkSizeArr, totalContentLength := getChunkSizes(url, numChunks)

	go func() {
		for {
			time.Sleep(5 * time.Millisecond)
			progress := float64(totalDownloaded) / float64(totalContentLength) * 100
			fmt.Printf("\rDownload progress: %.2f%%", progress)
		}
	}()

	var eg errgroup.Group

	for key, val := range chunkSizeArr {
		key, val1, val2 := key, val[0], val[1] // capture loop variables

		eg.Go(func() error {
			fmt.Println(key, val1, val2)
			return DownloadChunk(key, val1, val2, url)
		})
	}

	if err := eg.Wait(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	MergeFiles(numChunks, destination)
}

func DownloadChunk(chunkNo int, inital, end int64, url string) error {

	// Create a chunk file
	outFile, err := os.Create(fmt.Sprintf("chunk_%d.tmp", chunkNo))
	if err != nil {
		return fmt.Errorf("error creating chunk file %d: %s\n", chunkNo, err)
	}
	defer outFile.Close()

	var eg errgroup.Group

	var mu sync.Mutex

	fileStart := int64(0)
	for start := inital; start < end; start += initialChunkSize {
		chunkSize := initialChunkSize
		if start+chunkSize > end {
			chunkSize = end - start
		}
		eg.Go(func() error {
			return Download(&mu, url, outFile, fileStart, start, chunkSize, &totalDownloaded)
		})

		// Download each chunk

		fileStart += chunkSize
	}

	// Wait for all chunks to be downloaded
	if err := eg.Wait(); err != nil {
		return fmt.Errorf("error downloading chunks: %w", err)
	}

	return nil
}

func Download(mu *sync.Mutex, url string, destFile *os.File, fileStart, start, chunkSize int64, totalDownloaded *int64) error {

	end := start + chunkSize - 1
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("Error creating request: %s\n", err)
	}

	req.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", start, end))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("Error downloading chunk: %s\n", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Unexpected status code: %d\n", resp.StatusCode)
	}

	// Determine the character encoding of the response body
	// and create a reader that converts to UTF-8 if necessary
	utf8Reader, err := charset.NewReader(resp.Body, resp.Header.Get("Content-Type"))
	if err != nil {

		return fmt.Errorf("Error creating UTF-8 reader:", err)
	}
	// Wrap the UTF-8 reader with a bufio.Reader for better efficiency
	reader := bufio.NewReader(utf8Reader)

	mu.Lock()
	destFile.Seek(fileStart, 0)
	written, err := io.Copy(destFile, reader)
	if err != nil {
		return fmt.Errorf("Error writing chunk: %s\n", err)
	}
	mu.Unlock()
	atomic.AddInt64(totalDownloaded, written)
	return nil

}

func getChunkSizes(url string, numChunks int) ([][]int64, int64) {
	res := make([][]int64, numChunks)

	// Get total Size of File
	resp, err := http.Head(url)
	if err != nil {
		log.Fatal("Error: ", err)
	}
	defer resp.Body.Close()

	totalContentLength := resp.ContentLength
	// totalChunks := totalContentLength / int64(numChunks)
	chunkSize := totalContentLength / int64(numChunks)

	for chunkNo := int64(0); chunkNo < int64(numChunks); chunkNo++ {
		start := chunkNo * chunkSize
		end := start + chunkSize - 1
		if chunkNo == int64(numChunks)-1 {
			end = totalContentLength - 1
		}
		res[chunkNo] = []int64{start, end}
	}

	return res, totalContentLength
}
