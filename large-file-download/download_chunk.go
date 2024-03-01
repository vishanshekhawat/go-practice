package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"

	"golang.org/x/net/html/charset"
)

func DownloadFileInChunks(url, destFileName string) {
	defer ExecutionTime("DownloadFileInChunks")()
	numChunks := 10 // Specify the number of chunks to divide the file into

	resp, err := http.Head(url)
	if err != nil {
		fmt.Printf("Error getting file info: %s\n", err)
		return
	}

	contentLengthHeader := resp.Header.Get("Content-Length")
	contentLength, err := strconv.Atoi(contentLengthHeader)
	if err != nil {
		fmt.Printf("Error converting Content-Length to integer: %s\n", err)
		return
	}

	chunkSize := contentLength / numChunks

	var wg sync.WaitGroup

	for i := 0; i < numChunks; i++ {
		start := i * chunkSize
		end := start + chunkSize - 1
		if i == numChunks-1 {
			end = contentLength - 1
		}
		wg.Add(1)
		go DownloadChunks(url, i, start, end, &wg)
	}
	wg.Wait()
	MergeFiles(numChunks, destFileName)

}

func DownloadChunks(url string, chunkNum, start, end int, wg *sync.WaitGroup) {
	defer wg.Done()
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error While Creating Request")
		return
	}

	rangeHeaders := fmt.Sprintf("bytes=%d-%d", start, end)

	req.Header.Set("Range", rangeHeaders)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error While Calling request")
		return
	}

	defer resp.Body.Close()
	// Create a chunk file
	outFile, err := os.Create(fmt.Sprintf("chunk_%d.tmp", chunkNum))
	if err != nil {
		fmt.Printf("Error creating chunk file %d: %s\n", chunkNum, err)
		return
	}
	defer outFile.Close()

	// Determine the character encoding of the response body
	// and create a reader that converts to UTF-8 if necessary
	utf8Reader, err := charset.NewReader(resp.Body, resp.Header.Get("Content-Type"))
	if err != nil {
		fmt.Println("Error creating UTF-8 reader:", err)
		return
	}
	// Wrap the UTF-8 reader with a bufio.Reader for better efficiency
	reader := bufio.NewReader(utf8Reader)

	// Write the bytes to the chunk file
	_, err = io.Copy(outFile, reader)
	if err != nil {
		fmt.Printf("Error writing chunk %d: %s\n", chunkNum, err)
		return
	}

	fmt.Printf("Chunk %d downloaded successfully\n", chunkNum)

}

// MergeFiles merges all the chunk files into the final file.
func MergeFiles(totalChunks int, destFileName string) {
	destFile, err := os.Create(destFileName)
	if err != nil {
		fmt.Printf("Error creating final file: %s\n", err)
		return
	}
	defer destFile.Close()

	for i := 0; i < totalChunks; i++ {
		chunkFileName := fmt.Sprintf("chunk_%d.tmp", i)
		chunkFile, err := os.Open(chunkFileName)
		if err != nil {
			fmt.Printf("Error opening chunk %d: %s\n", i, err)
			return
		}

		_, err = io.Copy(destFile, chunkFile)
		chunkFile.Close()
		if err != nil {
			fmt.Printf("Error merging chunk %d: %s\n", i, err)
			return
		}

		// Delete chunk file after merging
		os.Remove(chunkFileName)
	}

	fmt.Println("All chunks merged successfully")
}
