package logwrap

import (
	"log"
	"net/http"
	"net/http/httputil"
)

var reqIDchan = make(chan requestID)

type requestID int64

func reqID() requestID {
	return <-reqIDchan
}

type rwWrap struct {
	http.ResponseWriter
	log *log.Logger
	r   *http.Request
	id  requestID
}

func init() {
	go func() {
		for i := int64(0); ; i++ {
			reqIDchan <- requestID(i)
		}
	}()
}

// WriteHeader is an implementation of http.ResponseWriter.
func (r rwWrap) WriteHeader(status int) {
	r.log.Printf("rsp#%d %d %s %s", r.id, status, r.r.Method, r.r.URL.String())
	r.ResponseWriter.WriteHeader(status)
}

type logger struct {
	h       http.Handler
	log     *log.Logger
	verbose bool
}

// ServeHTTP is implementation of net/http.Handler interface.
func (w logger) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	id := reqID()
	wrap := rwWrap{rw, w.log, r, id}
	if w.verbose {
		if dump, err := httputil.DumpRequest(r, false); err == nil {
			w.log.Printf("req#%d follows:\n%s", id, dump)
			w.h.ServeHTTP(wrap, r)
			return
		}
	}
	w.log.Printf("req#%d %s %s %s", id, r.Method, r.Proto, r.URL.String())
	w.h.ServeHTTP(wrap, r)
}

// Handler returns an http.Handler with a logging decorator.
func Handler(h http.Handler, l *log.Logger) http.Handler {
	return logger{h, l, false}
}

// VerboseHandler returns a http.Handler with verbose logging decorator.
func VerboseHandler(h http.Handler, l *log.Logger) http.Handler {
	return logger{h, l, true}
}
