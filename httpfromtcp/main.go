package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
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
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Connection has been accepted")
		lines := getLinesFromChannel(conn)
		for line := range lines {
			fmt.Println(line)
		}
		fmt.Println("connection closed")
	}

}
