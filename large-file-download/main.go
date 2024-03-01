package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"golang.org/x/net/html/charset"
)

func main() {
	url := "https://www.stats.govt.nz/assets/Uploads/Business-operations-survey/Business-operations-survey-2022/Download-data/business-operations-survey-2022-business-finance.csv" // Specify the URL of the file to download

	//brutedestFileName := "brutesample.csv" // Specify the name of the final file
	//bruteDownloadCsvName := "brutesample.csv"
	//chunkCsvName := "chunksample_chunks.csv"
	chunkCsvNameProgressBar := "chunksample_chunks_progress_bar.csv"

	// ExampleWriteAt()
	//BruteDownload(url, bruteDownloadCsvName)
	//DownloadFileInChunks(url, chunkCsvName)
	DownloadFileInChunksWithProgressBar(url, chunkCsvNameProgressBar, 10)
}

func ExecutionTime(name string) func() {
	startTime := time.Now()
	return func() {
		fmt.Printf("%s took %v\n", name, time.Since(startTime))
	}
}

func BruteDownload(url, destFileName string) {

	defer ExecutionTime("BruteDownload")()
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error downloading file:", err)
		return
	}
	defer resp.Body.Close()

	out, err := os.Create(destFileName)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer out.Close()

	// Determine the character encoding of the response body
	// and create a reader that converts to UTF-8 if necessary
	utf8Reader, err := charset.NewReader(resp.Body, resp.Header.Get("Content-Type"))
	if err != nil {
		fmt.Println("Error creating UTF-8 reader:", err)
		return
	}
	// Wrap the UTF-8 reader with a bufio.Reader for better efficiency
	reader := bufio.NewReader(utf8Reader)

	_, err = io.Copy(out, reader)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}

	fmt.Println("File downloaded successfully:", destFileName)
}
