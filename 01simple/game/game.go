package game

import (
	"fmt"
	"github.com/google/uuid"
	"math/rand"
	"sync"
)

const (
	StateInit = iota
	StatePlay
	StateStop
)

type GameState int

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

type ID string

func NewID() ID {
	return ID(uuid.New().String())
}

func (id ID) String() string {
	return string(id)
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
	mux     sync.Mutex
	Id      ID
	Players []*Player
	State   GameState
}

func (g *Game) String() string {
	return fmt.Sprintf("game(%q, %d players, %s)", g.Id, len(g.Players), g.State)
}

func NewGame() *Game {
	return &Game{
		Id:    NewID(),
		State: StateInit,
	}
}

func (g *Game) AddPlayer(player *Player) (*Player, error) {
	if player == nil {
		return player, errors.New("nil player")
	}
	if player.Id == "" {
		return player, errors.New("player ID is empty")
	}
	if player.Nick == "" || len(player.Nick) > 50 {
		return player, errors.New("invalid nickname len=%d", len(player.Nick))
	}
	g.mux.Lock()
	defer g.mux.Unlock()
	// Check if it is already too late to join.
	if g.State != StateInit {
		return player, errors.New("it is already too late, game has started")
	}
	// Check if player already exists.
	for _, p := range g.Players {
		if p.Id == player.Id {
			if p.Nick == player.Nick {
				// The same player just refreshed the page.
				// We return the reference to the existing player.
				return p, nil
			}
			return player, fmt.Errorf("player with Id=%s exists", player.Id)
		}
		if p.Nick == player.Nick {
			return player, fmt.Errorf("nick=%q is taken by someone else", player.Nick)
		}
	}
	g.Players = append(g.Players, player)
	return player, nil
}
