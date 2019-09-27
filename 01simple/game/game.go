package game

import (
	"fmt"
	"github.com/google/uuid"
	"math/rand"
)

type ID string

type GameState int

const (
	StateInit = iota
	StatePlay
	StateStop
)

func (s GameState) String() string {
	switch s {
	case StateInit:
		return "Init"
	case StatePlay:
		return "play"
	case StateStop:
		return "stop"
	}
	return "????"
}

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
		Id:    ID(uuid.New().String()),
		State: StateInit,
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
