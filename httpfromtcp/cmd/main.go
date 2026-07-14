package main

import (
	"crypto/sha256"
	"fmt"
	"httpfromtcp/internal/headers"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
)

const port = 42069

func httpbinProxyHandler(w *response.Writer, req *request.Request) {
	path := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin")
	upstream, err := http.Get("https://httpbin.org" + path)
	if err != nil {
		fmt.Printf("error making request to url %s: %v\n", path, err)
		return
	}
	defer upstream.Body.Close()

	h := response.GetDefaultHeaders(0)
	delete(h, "Content-Length")
	h["Transfer-Encoding"] = "chunked"
	h["Content-Type"] = "text/plain"
	h["Trailer"] = "X-Content-SHA256, X-Content-Length"

	w.WriteStatusLine(response.StatusOk)
	w.WriteHeaders(h)

	var fullBody []byte
	buf := make([]byte, 1024)
	for {
		n, err := upstream.Body.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			fullBody = append(fullBody, chunk...)
			if _, writeErr := w.WriteChunkedBody(chunk); writeErr != nil {
				fmt.Printf("error writing chunked body: %v\n", writeErr)
				return
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("error reading upstream body: %v\n", err)
			break
		}
	}

	w.WriteChunkedBodyDone()

	sum := sha256.Sum256(fullBody)
	trailers := headers.NewHeaders()
	trailers["X-Content-SHA256"] = fmt.Sprintf("%x", sum)
	trailers["X-Content-Length"] = fmt.Sprintf("%d", len(fullBody))
	if err := w.WriteTrailers(trailers); err != nil {
		fmt.Printf("error writing trailers: %v\n", err)
	}
}

func videoHandler(w *response.Writer, req *request.Request) {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("error getting working directory: %v\n", err)
		return
	}

	// repo root when `go run .` from httpfromtcp/cmd; else assets/ at cwd (bootdev)
	candidates := []string{
		filepath.Join(cwd, "assets", "vim.mp4"),
		filepath.Join(cwd, "..", "..", "assets", "vim.mp4"),
	}

	var file []byte
	for _, videoPath := range candidates {
		file, err = os.ReadFile(videoPath)
		if err == nil {
			break
		}
	}
	if err != nil {
		fmt.Printf("error reading video file: %v\n", err)
		return
	}

	h := response.GetDefaultHeaders(len(file))
	h["Content-Type"] = "video/mp4"
	w.WriteStatusLine(response.StatusOk)
	w.WriteHeaders(h)
	w.WriteBody(file)
}

func handler(w *response.Writer, req *request.Request) {
	target := req.RequestLine.RequestTarget

	if strings.HasPrefix(target, "/httpbin") {
		httpbinProxyHandler(w, req)
		return
	}

	if strings.HasPrefix(target, "/video") {
		videoHandler(w, req)
		return
	}

	switch target {
	case "/yourproblem":
		body := []byte(`<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`)
		h := response.GetDefaultHeaders(len(body))
		h["Content-Type"] = "text/html"
		w.WriteStatusLine(response.StatusBadRequest)
		w.WriteHeaders(h)
		w.WriteBody(body)
	case "/myproblem":
		body := []byte(`<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`)
		h := response.GetDefaultHeaders(len(body))
		h["Content-Type"] = "text/html"
		w.WriteStatusLine(response.StatusInternalServerError)
		w.WriteHeaders(h)
		w.WriteBody(body)
	default:
		body := []byte(`<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`)
		h := response.GetDefaultHeaders(len(body))
		h["Content-Type"] = "text/html"
		w.WriteStatusLine(response.StatusOk)
		w.WriteHeaders(h)
		w.WriteBody(body)
	}
}

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
