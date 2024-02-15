package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func main() {
	url := "https://files.icyflamestudio.com/512MB.zip" // Specify the URL of the file to download
	destFileName := "downloaded_file.zip"               // Specify the name of the final file

	// BruteDownload(url, destFileName)
	DownloadFileInChunks(url, destFileName)
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

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}

	fmt.Println("File downloaded successfully:", destFileName)
}
