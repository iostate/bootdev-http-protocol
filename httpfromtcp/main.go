package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

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

	buffer := make([]byte, 8)

	currentLineContents := ""
	for {

		n, err := file.Read(buffer)

		if err != nil {
			if currentLineContents != "" {
				fmt.Printf("read: %s\n", currentLineContents)
				currentLineContents = ""
			}
			if errors.Is(err, io.EOF) {
				break
			}
			fmt.Printf("failed to read file: error: %v", err)
			break
		}
		string := string(buffer[:n])
		parts := strings.Split(string, "\n")
		for i := 0; i < len(parts)-1; i++ {
			fmt.Printf("read: %s%s\n", currentLineContents, parts[i])
			currentLineContents = ""
		}
		currentLineContents += parts[len(parts)-1]
	}
}
