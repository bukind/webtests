package main

import (
	"fmt"
	"github.com/google/uuid"
	"html/template"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
	"github.com/bukind/webtests/01simple/game"
	"github.com/bukind/webtests/filefinder"
	"github.com/bukind/webtests/logwrap"
)

var (
	ff = filefinder.New(os.ExpandEnv("${GOPATH}/src/github.com/bukind/webtests/01simple"), "01simple", ".")
	hlog         = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	notFoundTmpl = templateMust("templates/notfound.html")
	joinTmpl = templateMust("templates/join.html")
	failedToJoinTmpl = templateMust("templates/failed_to_join.html")
	startTmpl = templateMust("templates/start.html")
)

func templateMust(files ...string) *template.Template {
	return template.Must(template.ParseFiles(ff.Must(files...)...))
}

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

func pageNotFound(w http.ResponseWriter, r *http.Request) {
	notFoundTmpl.Execute(w, r.URL.Path)
	w.WriteHeader(http.StatusNotFound)
}

//
// State transitions:
//
// 1. asking for name
// 2. waiting others
// 3. in game: choosing numbers
// 4. end game
func main() {
	mux := http.NewServeMux()
	var mtx sync.Mutex
	var gm *game.Game
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && r.URL.Path != "/index.html" {
			pageNotFound(w, r)
			return
		}
		http.Redirect(w, r, "join.html", http.StatusFound)
	})
	mux.HandleFunc("/join.html", func(w http.ResponseWriter, r *http.Request) {
		joinTmpl.Execute(w, page(r).Set("id", uuid.New().String()))
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/start.html", func(w http.ResponseWriter, r *http.Request) {
		id := game.ID(r.FormValue("id"))
		nickname := r.FormValue("nickname")
		if len(id) == 0 {
			http.Redirect(w, r, "/index.html", http.StatusFound)
			return
		}
		if len(nickname) == 0 || len(nickname) > 50 {
			http.Redirect(w, r, "/index.html", http.StatusFound)
		}
		mtx.Lock()
		if gm == nil {
			gm = game.NewGame()
		}
		p, err := gm.AddPlayer(id, nickname)
		hlog.Printf("game %v add -> %v, %v", gm, p, err)
		mtx.Unlock()
		if err != nil {
			failedToJoinTmpl.Execute(w, page(r).Set("error", err.Error()))
			w.WriteHeader(http.StatusOK)
			return
		}
		startTmpl.Execute(w, page(r).Set("id", string(p.Id)).Set("num", fmt.Sprint(p.Num)))
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		pageNotFound(w, r)
	})
	static := ff.Must("static")[0]
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(static))))

	server := &http.Server{
		Addr:           ":9999",
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
