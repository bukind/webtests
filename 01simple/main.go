package main

import (
	"fmt"
	"github.com/google/uuid"
	"html/template"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

var (
	hlog = log.New(os.Stdout, "", log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	reqIDchan = make(chan requestID)
	indexTmpl = template.Must(template.New("index").Parse(`<!DOCTYPE html>
<html>
<meta charset="UTF-8" />
<title>Initial page</title>
<body>
 <h2>Initial setup</h2>
 <p>Please enter your nickname below, then press Start button.</p>
 <form action="/start.html" method="POST">
  <label for="nickname">Nickname:</label>
  <input type="text" name="nickname" value="{{.Val "nickname"}}" />
  <input type="submit" value="Start" />
 </form>
</body>
</html>
`))
	startTmpl = template.Must(template.New("start").Parse(`<!DOCTYPE html>
<html>
<meta charset="UTF-8" />
<title>Waiting for other players...</title>
<body>
 <h2>Waiting for others</h2>
 <p>Hello, <b>{{.Val "nickname"}}</b>.  Your lucky number is <b>{{.Val "num"}}</b>.</p>
 <p>Meanwhile, we're waiting for other players...</p>
 <form action="/index.html" method="POST">
  <input type="hidden" name="id" value="{{.Val "id"}}" />
  <input type="hidden" name="nickname" value="{{.Val "nickname"}}" />
  <input type="submit" value="Go!" />
 </form>
</body>
</html>
`))
)

type Page struct {
	Req  *http.Request
	Vals map[string]string
}

func page(r *http.Request) *Page {
	return &Page{r, make(map[string]string)}
}

func (p *Page) Set(key, val string) *Page {
	p.Vals[key] = val
	return p
}

func (p *Page) Val(key string) string {
	if val, ok := p.Vals[key]; ok {
		return val
	}
	return p.Req.FormValue(key)
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
			indexTmpl.Execute(w, page(r))
			w.WriteHeader(http.StatusOK)
		case "/start.html":
			id := uuid.New().String()
			num := fmt.Sprint(rand.Intn(64)+1)
			startTmpl.Execute(w, page(r).Set("id", id).Set("num",num))
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
