package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

var (
	initialChunkSize int64 = 2 * 1024 * 1024
	// maxChunkSize            = 5 * 1024 * 1024 // 5 MB, maximum chunk size to avoid too large chunks
	//minChunkSize            = 512 * 1024      // 512 KB, minimum chunk size to maintain efficiency
	//progressUpdateSec       = 5
)

var totalDownloaded int64

func DownloadFileInChunksWithProgressBar(url string, destination string, numChunks int) {

	// Get total Size of File
	resp, err := http.Head(url)
	if err != nil {
		log.Fatal("Error: ", err)
	}
	defer resp.Body.Close()

	totalContentLength := resp.ContentLength
	// totalChunks := totalContentLength / int64(numChunks)
	chunkSize := totalContentLength / int64(numChunks)

	go func() {
		for {
			select {
			case <-time.After(5 * time.Second):
				progress := float64(totalDownloaded) / float64(totalContentLength) * 100
				fmt.Printf("\rDownload progress: %.2f%%", progress)
			}
		}
	}()

	var wg sync.WaitGroup
	for chunkNo := int64(0); chunkNo < int64(numChunks); chunkNo++ {
		start := chunkNo * chunkSize
		end := start + chunkSize - 1
		if chunkNo == int64(numChunks)-1 {
			end = totalContentLength - 1
		}
		wg.Add(1)
		go DownloadChunk(&wg, chunkNo, start, end, url)
	}
	wg.Wait()
	MergeFiles(numChunks, destination)
}

func DownloadChunk(wg *sync.WaitGroup, chunkNo, start, end int64, url string) {
	defer wg.Done()

	// Create a chunk file
	outFile, err := os.Create(fmt.Sprintf("chunk_%d.tmp", chunkNo))
	if err != nil {
		fmt.Printf("Error creating chunk file %d: %s\n", chunkNo, err)
		return
	}
	defer outFile.Close()

	var newWg sync.WaitGroup
	for start := int64(0); start < end; start += initialChunkSize {
		chunkSize := initialChunkSize
		if start+int64(chunkSize) > int64(end) {
			chunkSize = end - start
		}

		newWg.Add(1)
		go Download(&newWg, url, outFile, start, chunkSize, &totalDownloaded)
	}
	newWg.Wait()
}

func Download(wg *sync.WaitGroup, url string, destFile *os.File, start, chunkSize int64, totalDownloaded *int64) {
	defer wg.Done()
	end := start + chunkSize - 1
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error creating request: %s\n", err)
	}

	req.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", start, end))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Error downloading chunk: %s\n", err)
	}

	if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
		fmt.Printf("Unexpected status code: %d\n", resp.StatusCode)
		resp.Body.Close()
	}

	destFile.Seek(int64(start), 0)
	written, err := io.Copy(destFile, resp.Body)
	resp.Body.Close()
	if err != nil {
		fmt.Printf("Error writing chunk: %s\n", err)
	}

	atomic.AddInt64(totalDownloaded, written)

}
