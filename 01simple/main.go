package main

import (
	"fmt"
	"github.com/google/uuid"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"
)

var (
	baseHTML = `<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8" />
<title>{{block "title" .}}{{end}}</title>{{block "style" .}}{{end}}
</head>
<body>{{template "content"}}
</body>{{block "js" .}}{{end}}
</html>
`
)

var (
	hlog = log.New(os.Stdout, "", log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	reqIDchan = make(chan requestID)
	notFoundTmpl = template.Must(template.New("index").Parse(baseHTML+`
{{- define "title"}}Page not found{{end -}}
{{- define "content"}}
<h2>Page not found</h2>
<p>Page "{{.Val "path"}}" is not found.</p>
{{- end}}
`))
	indexTmpl = template.Must(template.New("index").Parse(baseHTML+`
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
	failedToJoinTmpl = template.Must(template.New("failJoin").Parse(baseHTML+`
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
	startTmpl = template.Must(template.New("start").Parse(baseHTML+`
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

type ID string

type Player struct {
	Id   ID
	Nick string
	Min  int
	Max  int
	Num  int
}

func (p *Player) String() string {
	return fmt.Sprintf("player(%q,%q,%d,%d,%d)", p.Id, p.Nick, p.Min, p.Num, p.Max)
}

func NewPlayer(id ID, nick string) *Player {
	min := 1
	max := 64
	return &Player{
		Id:   id,
		Nick: nick,
		Min:  min, // >=
		Max:  max, // <=
		Num:  rand.Intn(max-min+1) + min,
	}
}

type GameState int

const (
	StateInit = iota
	StatePlay
	StateStop
)

func (s GameState) String() string {
	switch s {
	case StateInit: return "Init"
	case StatePlay: return "play"
	case StateStop: return "stop"
	}
	return "????"
}

type Game struct {
	Id      ID
	Players []*Player
	State   GameState
}

func (g *Game) String() string {
	return fmt.Sprintf("game(%q, %d players, %s)", g.Id, len(g.Players), g.State)
}

func NewGame() *Game {
	return &Game{
		Id:      ID(uuid.New().String()),
		State:   StateInit,
	}
}

func (g *Game) AddPlayer(id ID, nick string) (*Player, error) {
	// Check if it is already too late to join.
	if g.State != StateInit {
		return nil, fmt.Errorf("it is already too late, game has started")
	}
	if len(id) == 0 {
		return nil, fmt.Errorf("")
	}
	// Check if player already exists.
	for _, p := range g.Players {
		if p.Id == id {
			if p.Nick == nick {
				// The same player just refreshed the page.
				return p, nil
			}
			return nil, fmt.Errorf("player with id=%q exists", id)
		}
		if p.Nick == nick {
			return nil, fmt.Errorf("nick=%q is taken by someone else", nick)
		}
	}
	p := NewPlayer(id, nick)
	g.Players = append(g.Players, p)
	return p, nil
}

func pageNotFound(w http.ResponseWriter, r *http.Request) {
	notFoundTmpl.Execute(w, page(r).Set("path", r.URL.Path))
	w.WriteHeader(http.StatusNotFound)
}

func main() {
	mux := http.NewServeMux()
	var mtx sync.Mutex
	var game *Game
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && r.URL.Path != "/index.html" {
			pageNotFound(w, r)
			return
		}
		indexTmpl.Execute(w, page(r).Set("id", uuid.New().String()))
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/start.html", func(w http.ResponseWriter, r *http.Request) {
		id := ID(r.FormValue("id"))
		nickname := r.FormValue("nickname")
		if len(id) == 0 {
			http.Redirect(w, r, "/index.html", http.StatusFound)
			return
		}
		if len(nickname) == 0 || len(nickname) > 50 {
			http.Redirect(w, r, "/index.html", http.StatusFound)
		}
		mtx.Lock()
		if game == nil {
			game = NewGame()
		}
		p, err := game.AddPlayer(id, nickname)
		hlog.Printf("game %v add -> %v, %v", game, p, err)
		mtx.Unlock()
		if err != nil {
			failedToJoinTmpl.Execute(w, page(r).Set("error", err.Error()))
			w.WriteHeader(http.StatusOK)
			return
		}
		startTmpl.Execute(w, page(r).Set("id", string(p.Id)).Set("num",fmt.Sprint(p.Num)))
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		pageNotFound(w, r)
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
