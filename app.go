package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
)

// Port Constants
const (
	ServerPort = 8080
	SocketPort = 8081
)

var queue = make(chan string, 100)

// Collector handles POST requests to our web server and
// puts the data onto our queue channel.
func Collector(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Header().Set("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	msg := strings.Trim(r.URL.Path, "/")
	log.Printf("Adding '%s' to queue", msg)
	queue <- msg
}

// LogHandler Wrapper to log the http requests
func LogHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			log.Printf("[%s][%s] %s", r.RemoteAddr, r.Method, r.URL)
			handler.ServeHTTP(w, r)
		},
	)
}

// QueueHandler handles messages on the queue channel and sends them across
// the tcp socket
func QueueHandler() {
	port := fmt.Sprintf(":%v", SocketPort)

	// listen on all interfaces
	ln, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Waiting for tcp socket connection on %s", port)

	// Accept connection on port
	conn, err := ln.Accept()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Connected to %s", conn.RemoteAddr())

	// Read from queue and send to
	for msg := range queue {
		withNewline := fmt.Sprintf("%s\r\n", msg)
		conn.Write([]byte(withNewline))
	}
}

func main() {
	go QueueHandler()

	port := fmt.Sprintf(":%v", ServerPort)

	http.HandleFunc("/", Collector)

	log.Printf("Web server listing for POST on %s", port)
	log.Fatal(http.ListenAndServe(port, LogHandler(http.DefaultServeMux)))
}
