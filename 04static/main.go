// Package 04static is to a small webserver for a static content only.
// It is supposed to be run in a directory you want to expose in web.
// Usage:
//
// $ cd DIRTOEXPOSE
// $ 04static [--port PORT]
package main

import (
	"flag"
	"fmt"
	"github.com/bukind/webtests/logwrap"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	port := flag.Int("port", 9988, "port to listen to")
	flag.Parse()

	hlog := log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("")))
	server := &http.Server{
		Addr:           fmt.Sprintf(":%d", *port),
		Handler:        logwrap.Handler(mux, hlog),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	defer server.Close()
	hlog.Printf("Starting the server to listen on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		fmt.Fprintln(os.Stderr, "failed to serve http:", err)
		os.Exit(1)
	}
}
