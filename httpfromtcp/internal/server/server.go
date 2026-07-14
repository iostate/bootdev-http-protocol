package server

import (
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

	w := response.NewWriter(conn)

	req, err := request.RequestFromReader(conn)
	if err != nil {
		handlerError := &HandlerError{
			Code:    response.StatusBadRequest,
			Message: err.Error(),
		}
		fmt.Printf("failed to read request, error: %v", err)
		handlerError.Write(w)
		return
	}

	s.handler(w, req)
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
