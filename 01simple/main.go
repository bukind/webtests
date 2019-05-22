package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	hlog = log.New(os.Stdout, "", log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	reqIDchan = make(chan requestID)
	indexTmpl = template.Must(template.New("index").Parse(`<!DOCTYPE html>
<html>
<meta charset="UTF-8">
<title>Initial page</title>
<body>
 <form action="/start.html" method="POST">
  <input type="text" name="nickname" value="{{.Values.nickname}}"></value>
  <input type="submit" value="Start"></value>
 </form>
</body>
</html>
`))
)

type Page struct {
	Title  string
	Values map[string]string
}

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
		switch r.URL.Path {
		case "/", "/index.html", "/index.htm":
			indexTmpl.Execute(w, Page{Values: map[string]string{"nickname": "Dim"}})
			w.WriteHeader(http.StatusOK)
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
