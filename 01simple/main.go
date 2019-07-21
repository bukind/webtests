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
	"github.com/bukind/webtests/logwrap"
	"github.com/bukind/webtests/01simple/game"
)

var (
	baseHTML = `<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8" />
<title>{{block "title" .}}{{end}}</title>{{block "style" .}}{{end}}
</head>
<body>{{template "content" .}}
</body>{{block "js" .}}{{end}}
</html>
`
)

var (
	hlog         = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	notFoundTmpl = template.Must(template.New("index").Parse(baseHTML + `
{{- define "title"}}Page not found{{end -}}
{{- define "content"}}
<h2>Page not found</h2>
<p>Page "{{.}}" is not found.</p>
{{- end}}
`))
	joinTmpl = template.Must(template.New("join").Parse(baseHTML + `
{{- define "title"}}Initial page{{end -}}
{{- define "content"}}
<h2>Initial setup</h2>
<p>Please enter your nickname below, then press Start button.</p>
<form action="/start.html" method="POST">
 <input type="hidden" name="id" value="{{.Val "id"}}" />
 <label for="nickname">Nickname:</label>
 <input type="text" name="nickname" value="{{.Val "nickname"}}" />
 <input type="submit" value="Start" />
</form>
{{- end}}
`))
	failedToJoinTmpl = template.Must(template.New("failJoin").Parse(baseHTML + `
{{- define "title"}}Failed to join the game{{end -}}
{{- define "content"}}
<h2>Sorry, you've failed to join the game</h2>
<p>{{.Val "error"}}</p>
<p>You can try again...</p>
<form action="/index.html" method="POST">
 <input type="hidden" name="id" value="{{.Val "id"}} />
 <input type="hidden" name="nickname" value="{{.Val "nickname"}}" />
 <input type="submit" value="Try again" />
</form>
{{- end}}
`))
	startTmpl = template.Must(template.New("start").Parse(baseHTML + `
{{- define "title"}}Waiting for other players...{{end -}}
{{- define "content"}}
<h2>Waiting for others</h2>
<p>Hello, <b>{{.Val "nickname"}}</b>.  Your lucky number is <b>{{.Val "num"}}</b>.</p>
<p>Meanwhile, we're waiting for other players...</p>
<form action="/index.html" method="POST">
 <input type="hidden" name="id" value="{{.Val "id"}}" />
 <input type="hidden" name="nickname" value="{{.Val "nickname"}}" />
 <input type="submit" value="Go!" />
</form>
{{- end}}
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
