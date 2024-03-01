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

	var wg sync.WaitGroup
	for key, val := range chunkSizeArr {
		wg.Add(1)
		DownloadChunk(&wg, key, val[0], val[1], url)
	}

	wg.Wait()
	MergeFiles(numChunks, destination)
}

func DownloadChunk(wg *sync.WaitGroup, chunkNo int, inital, end int64, url string) {
	defer wg.Done()

	// Create a chunk file
	outFile, err := os.Create(fmt.Sprintf("chunk_%d.tmp", chunkNo))
	if err != nil {
		fmt.Printf("Error creating chunk file %d: %s\n", chunkNo, err)
		return
	}
	defer outFile.Close()

	var newWg sync.WaitGroup
	var mu sync.Mutex

	fileStart := int64(0)
	for start := inital; start < end; start += initialChunkSize {
		chunkSize := initialChunkSize
		if start+chunkSize > end {
			chunkSize = end - start
		}
		newWg.Add(1)
		// Download each chunk
		Download(&newWg, &mu, url, outFile, fileStart, start, chunkSize, &totalDownloaded)
		fileStart += chunkSize
	}

	// Wait for all chunks to be downloaded
	newWg.Wait()
}

func Download(wg *sync.WaitGroup, mu *sync.Mutex, url string, destFile *os.File, fileStart, start, chunkSize int64, totalDownloaded *int64) {
	defer wg.Done()
	end := start + chunkSize - 1
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error creating request: %s\n", err)
		return
	}

	req.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", start, end))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Error downloading chunk: %s\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
		fmt.Printf("Unexpected status code: %d\n", resp.StatusCode)
		return
	}

	// Determine the character encoding of the response body
	// and create a reader that converts to UTF-8 if necessary
	utf8Reader, err := charset.NewReader(resp.Body, resp.Header.Get("Content-Type"))
	if err != nil {
		fmt.Println("Error creating UTF-8 reader:", err)
		return
	}
	// Wrap the UTF-8 reader with a bufio.Reader for better efficiency
	reader := bufio.NewReader(utf8Reader)

	mu.Lock()
	destFile.Seek(fileStart, 0)
	written, err := io.Copy(destFile, reader)
	if err != nil {
		fmt.Printf("Error writing chunk: %s\n", err)
	}
	mu.Unlock()

	atomic.AddInt64(totalDownloaded, written)

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
