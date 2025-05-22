package pokerlib

import (
	"time"

	"github.com/google/uuid"
)

type PokerFace interface {
	NewGame(opts *GameOptions) Game
	NewGameFromState(gs *GameState) Game
}

type pokerlib struct {
}

func NewPokerFace() PokerFace {
	return &pokerlib{}
}

func (pf *pokerlib) NewGame(opts *GameOptions) Game {
	g := NewGame(opts)
	s := g.GetState()
	s.GameID = uuid.New().String()
	s.CreatedAt = time.Now().Unix()
	s.UpdatedAt = time.Now().UnixNano()
	//s.UpdatedAt = 0

	return g
}

func (pf *pokerlib) NewGameFromState(gs *GameState) Game {
	return NewGameFromState(gs)
}
