package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

func getLinesFromChannel(file io.ReadCloser) <-chan string {

	results := make(chan string)
	buffer := make([]byte, 8)

	currentLineContents := ""

	go func() {
		for {
			n, err := file.Read(buffer)

			if err != nil {
				if currentLineContents != "" {
					// fmt.Printf("read: %s\n", currentLineContents)
					results <- currentLineContents
					currentLineContents = ""
				}
				if errors.Is(err, io.EOF) {
					close(results)
					file.Close()
					break
				}
				fmt.Printf("failed to read file: error: %v", err)
				break
			}
			string := string(buffer[:n])
			parts := strings.Split(string, "\n")
			for i := 0; i < len(parts)-1; i++ {
				results <- fmt.Sprintf("%s%s", currentLineContents, parts[i])
				currentLineContents = ""
			}
			currentLineContents += parts[len(parts)-1]
		}
	}()
	return results
}

func main() {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println("Unable to get working directory")
	}
	file, err := os.Open(wd + "/messages.txt")
	if err != nil {
		fmt.Printf("error opening file: error: %v", err)
	}
	defer file.Close()
	lines := getLinesFromChannel(file)
	for line := range lines {
		fmt.Printf("read: %s\n", line)
	}

}
