package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
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
	w.WriteHeader(http.StatusNotFound)
	notFoundTmpl.Execute(w, r.URL.Path)
}

//
// State transitions:
//
// 1. asking for name
// 2. waiting others
// 3. in game: choosing numbers
// 4. end game
func main() {
	gm := game.NewGame()
	mux := http.NewServeMux()
	mux.HandleFunc("/join.html", func(w http.ResponseWriter, r *http.Request) {
		joinTmpl.Execute(w, game.Player{
			Id:   game.NewID(),
			Nick: r.FromValue("nickname"),
		})
	})
	mux.HandleFunc("/start.html", func(w http.ResponseWriter, r *http.Request) {
		p, err := gm.AddPlayer(game.NewPlayer(game.ID(r.FormValue("id")), r.FormValue("nickname")))
		if err != nil {
			failPage := struct {
				P   *Player
				Msg string
			}{
				P: p,
				Msg: err.Error(),
			}
			failedToJoinTmpl.Execute(w, failPage)
			return
		}
		hlog.Printf("game %v add -> %v, %v", gm, p, err)
		page := struct {
			Id       game.ID
			Nickname string
			Num      int
		}{
			Id:       id,
			Nickname: nickname,
			Num:      p.Num
		}
		startTmpl.Execute(w, p)
	})
	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		pageNotFound(w, r)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && r.URL.Path != "/index.html" {
			pageNotFound(w, r)
			return
		}
		http.Redirect(w, r, "/join.html", http.StatusFound)
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
