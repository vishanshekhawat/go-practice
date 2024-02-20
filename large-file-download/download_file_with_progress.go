package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
)

var (
	initialChunkSize int64 = 2 * 1024 * 1024
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
	// Create a file if not exist
	file, err := os.Create(url)
	if err != nil {
		log.Fatal("Error: ", err)
	}
	fmt.Println(chunkSize)
	fmt.Println(totalContentLength)
	var wg sync.WaitGroup
	for chunkNo := int64(0); chunkNo < int64(numChunks); chunkNo++ {
		start := chunkNo * chunkSize
		end := start + chunkSize - 1
		if chunkNo == int64(numChunks)-1 {
			end = totalContentLength - 1
		}
		wg.Add(1)
		go DownloadChunk(&wg, file, chunkNo, start, end, url)
	}
	wg.Wait()
	MergeFiles(numChunks, destination)
}

func DownloadChunk(wg *sync.WaitGroup, file *os.File, chunkNo, start, end int64, url string) {
	defer wg.Done()

	var newWg sync.WaitGroup
	for start := int64(0); start < end; start += initialChunkSize {
		chunkSize := initialChunkSize
		if start+int64(chunkSize) > int64(end) {
			chunkSize = end - start
		}

		newWg.Add(1)
		go Download(&newWg, url, chunkNo, start, chunkSize, &totalDownloaded)
	}
	newWg.Wait()
}

func Download(wg *sync.WaitGroup, url string, chunkNo, start, chunkSize int64, totalDownloaded *int64) {
	defer wg.Done()

}
