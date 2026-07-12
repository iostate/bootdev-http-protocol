package server

import (
	"log"
	"net"
	"strconv"
	"sync/atomic"
)

type Server struct {
	listener net.Listener
	open     atomic.Bool
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

	response := "HTTP/1.1 200 OK\r\n" +
		"Content-Type: text/plain\r\n" +
		"Content-Length: 13\r\n" +
		"\r\n" +
		"Hello World!\n"
	conn.Write([]byte(response))
}

func Serve(port int) (*Server, error) {
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return nil, err
	}

	server := &Server{
		listener: listener,
	}
	server.open.Store(true)
	go server.listen()
	return server, nil
}
