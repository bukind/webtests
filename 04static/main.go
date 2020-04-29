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
	"regexp"
	"strconv"
	"time"
)

const defaultPort = 9988
var portRe = regexp.MustCompile(".*-([0-9]+)$")

func main() {
	var port int
	flag.IntVar(&port, "port", 0, "port to listen to")
	verbose := flag.Bool("verbose", false, "verbose logger")
	flag.Parse()

	if port == 0 {
		if m := portRe.FindStringSubmatch(os.Args[0]); len(m) > 1 {
			var err error
			port, err = strconv.Atoi(m[1])
			if err != nil {
				fmt.Fprintln(os.Stderr, "bad port specified:", err)
				os.Exit(1)
			}
		}
		if port == 0 {
			port = defaultPort
		}
	}

	hlog := log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("")))
	handler := logwrap.Handler(mux, hlog)
	if *verbose {
		handler = logwrap.VerboseHandler(mux, hlog)
	}
	server := &http.Server{
		Addr:           fmt.Sprintf(":%d", port),
		Handler:        handler,
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
