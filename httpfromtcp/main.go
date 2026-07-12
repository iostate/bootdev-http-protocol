package main

import (
	"fmt"
	"httpfromtcp/internal/request"
	"log"
	"net"
)

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
		req, err := request.RequestFromReader(conn)
		if err != nil {
			log.Println(err)
			conn.Close()
			continue
		}
		fmt.Print(req.String())
		fmt.Printf("Headers:\n")
		for key, val := range req.Headers {
			fmt.Printf("- %s: %s\n", key, val)
		}
		fmt.Printf("Body:\n%s\n", string(req.Body))

		conn.Close()
		fmt.Println("connection closed")
	}

}
