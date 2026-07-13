package server

import (
	"bytes"
	"fmt"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"log"
	"net"
	"strconv"
	"sync/atomic"
)

const bufferLength = 4096

type Server struct {
	listener net.Listener
	open     atomic.Bool
	handler  Handler
}

func (s *Server) Close() error {
	err := s.listener.Close()
	if err != nil {
		return err
	}
	s.open.Store(false)
	return nil
}

func (s *Server) listen() {
	for s.open.Load() {
		conn, err := s.listener.Accept()
		if err != nil {
			if !s.open.Load() {
				return // server closed intentionally
			}
			log.Printf("error accepting connection: %v", err)
			continue
		}

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	req, err := request.RequestFromReader(conn)
	if err != nil {
		fmt.Printf("failed to read request, error: %v", err)
		return
	}

	buf := bytes.Buffer{}
	if handlerErr := s.handler(&buf, req); handlerErr != nil {
		messageBytes := []byte(handlerErr.Message)
		if err := response.WriteStatusLine(conn, handlerErr.Code); err != nil {
			fmt.Printf("failed to write status line, error: %v", err)
			return
		}
		errHeaders := response.GetDefaultHeaders(len(messageBytes))
		if err := response.WriteHeaders(conn, errHeaders); err != nil {
			fmt.Printf("failed to write headers, error: %v", err)
			return
		}
		conn.Write([]byte("\r\n"))
		conn.Write(messageBytes)
		return
	}

	defaultHeaders := response.GetDefaultHeaders(buf.Len())

	if err := response.WriteStatusLine(conn, response.Ok); err != nil {
		fmt.Printf("failed to write status line, error: %v", err)
	}
	if err := response.WriteHeaders(conn, defaultHeaders); err != nil {
		fmt.Printf("failed to write headers, error: %v", err)
	}
	conn.Write([]byte("\r\n"))
	conn.Write(buf.Bytes())
}

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return nil, err
	}

	server := &Server{
		listener: listener,
		handler:  handler,
	}
	server.open.Store(true)
	go server.listen()
	return server, nil
}
