package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

const (
	filename   = "sample1.txt"
	singleLine = "1234567890123456789012345678901234567890\n"
)

var mu sync.Mutex

func printContents() {
	data, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))
}

func ExampleWriteAt() {

	outFile, err := os.Create("sample1.txt")
	if err != nil {
		fmt.Printf("Error creating chunk file %s\n", err)
		return
	}
	defer outFile.Close()

	start := 0

	var wg sync.WaitGroup
	for i := 0; i < 9; i++ {
		go WriteToFile(&wg, start, outFile)
		start = start + len(singleLine)
		wg.Add(1)
	}
	wg.Wait()
	// if _, err := f.WriteAt([]byte("A"), 15); err != nil {
	// 	panic(err)
	// }

	printContents()
}

func WriteToFile(wg *sync.WaitGroup, start int, outFile *os.File) {
	defer wg.Done()
	fmt.Println(start)
	str := strings.NewReader(singleLine)
	mu.Lock()
	outFile.Seek(int64(start), 0) // Seek from the beginning of the file
	written, err := io.Copy(outFile, str)
	mu.Unlock()
	fmt.Println(written, err)
}
