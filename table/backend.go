package table

import "github.com/d-protocol/pokerlib"

type Backend interface {
	CreateGame(opts *pokerlib.GameOptions) (*pokerlib.GameState, error)
	Next(gs *pokerlib.GameState) (*pokerlib.GameState, error)
	ReadyForAll(gs *pokerlib.GameState) (*pokerlib.GameState, error)
	PayAnte(gs *pokerlib.GameState) (*pokerlib.GameState, error)
	PayBlinds(gs *pokerlib.GameState) (*pokerlib.GameState, error)

	// Actions
	Call(gs *pokerlib.GameState) (*pokerlib.GameState, error)
	Pass(gs *pokerlib.GameState) (*pokerlib.GameState, error)
	Fold(gs *pokerlib.GameState) (*pokerlib.GameState, error)
	Check(gs *pokerlib.GameState) (*pokerlib.GameState, error)
	Allin(gs *pokerlib.GameState) (*pokerlib.GameState, error)
	Bet(gs *pokerlib.GameState, chips int64) (*pokerlib.GameState, error)
	Raise(gs *pokerlib.GameState, chipLevel int64) (*pokerlib.GameState, error)
	Pay(gs *pokerlib.GameState, chips int64) (*pokerlib.GameState, error)
}
