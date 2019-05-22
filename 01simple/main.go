package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	hlog = log.New(os.Stdout, "", log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	reqIDchan = make(chan requestID)
)

type requestID int

func reqID() requestID {
	return <-reqIDchan
}

func init() {
	go func() {
		for i := 0;; i++ {
			reqIDchan <- requestID(i)
		}
	}()
}

type rwWrap struct {
	http.ResponseWriter
	r  *http.Request
	id requestID
}

func (r rwWrap) WriteHeader(status int) {
	hlog.Printf("rsp#%d %d %s %s", r.id, status, r.r.Method, r.r.URL.String())
}

type logger struct {
	h http.Handler
}

func (w logger) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	id := reqID()
	hlog.Printf("req#%d %s %s %s", id, r.Method, r.Proto, r.URL.String())
	w.h.ServeHTTP(rwWrap{rw,r,id}, r)
}

func logWrapper(h http.Handler) http.Handler {
	return logger{h}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		hlog.Printf("req Path=%q Query=%q", r.URL.Path, r.URL.RawQuery)
		switch r.URL.Path {
		case "/", "/index.html", "/index.htm":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, `<http>
<head>
<title>Index</title>
</head>
<body>
  <a href="/date">[count]</a>
</body>
</http>`)
		case "/favicon.ico":
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
	mux.HandleFunc("/date", func(w http.ResponseWriter, _ *http.Request) {
		io.WriteString(w, time.Now().String())
	})

	server := &http.Server{
		Addr:           ":9999",
		Handler:        logWrapper(mux),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	defer server.Close()
	if err := server.ListenAndServe(); err != nil {
		fmt.Fprintln(os.Stderr, "failed to serve http:", err)
		os.Exit(1)
	}
}
