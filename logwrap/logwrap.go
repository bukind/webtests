package logwrap

import (
	"log"
	"net/http"
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
}

type logger struct {
	h   http.Handler
	log *log.Logger
}

// ServeHTTP is implementation of net/http.Handler interface.
func (w logger) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	id := reqID()
	w.log.Printf("req#%d %s %s %s", id, r.Method, r.Proto, r.URL.String())
	w.h.ServeHTTP(rwWrap{rw, w.log, r, id}, r)
}

// Handler return an http.Handler with a logging decorator.
func Handler(h http.Handler, l *log.Logger) http.Handler {
	return logger{h, l}
}
