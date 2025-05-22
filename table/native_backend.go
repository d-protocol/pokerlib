package table

import (
	"encoding/json"

	"github.com/d-protocol/pokerlib"
)

type NativeBackend struct {
	engine pokerlib.PokerFace
}

func NewNativeBackend() *NativeBackend {
	return &NativeBackend{
		engine: pokerlib.NewPokerFace(),
	}
}

func cloneState(gs *pokerlib.GameState) *pokerlib.GameState {

	//Note: we must clone a new structure for preventing original data of game engine is modified outside.
	data, err := json.Marshal(gs)
	if err != nil {
		return nil
	}

	var state pokerlib.GameState
	err = json.Unmarshal([]byte(data), &state)
	if err != nil {
		return nil
	}

	return &state
}

func (nb *NativeBackend) getState(g pokerlib.Game) *pokerlib.GameState {
	return cloneState(g.GetState())
}

func (nb *NativeBackend) CreateGame(opts *pokerlib.GameOptions) (*pokerlib.GameState, error) {

	// Initializing game
	g := nb.engine.NewGame(opts)
	err := g.Start()
	if err != nil {
		return nil, err
	}

	return nb.getState(g), nil
}

func (nb *NativeBackend) Next(gs *pokerlib.GameState) (*pokerlib.GameState, error) {

	g := nb.engine.NewGameFromState(cloneState(gs))
	err := g.Next()
	if err != nil {
		return nil, err
	}

	return nb.getState(g), nil
}

func (nb *NativeBackend) ReadyForAll(gs *pokerlib.GameState) (*pokerlib.GameState, error) {

	g := nb.engine.NewGameFromState(cloneState(gs))
	err := g.ReadyForAll()
	if err != nil {
		return nil, err
	}

	return nb.getState(g), nil
}

func (nb *NativeBackend) Pass(gs *pokerlib.GameState) (*pokerlib.GameState, error) {

	g := nb.engine.NewGameFromState(cloneState(gs))

	err := g.Pass()
	if err != nil {
		return nil, err
	}

	return nb.getState(g), nil
}

func (nb *NativeBackend) PayAnte(gs *pokerlib.GameState) (*pokerlib.GameState, error) {

	g := nb.engine.NewGameFromState(cloneState(gs))

	err := g.PayAnte()
	if err != nil {
		return nil, err
	}

	return nb.getState(g), nil
}

func (nb *NativeBackend) PayBlinds(gs *pokerlib.GameState) (*pokerlib.GameState, error) {

	g := nb.engine.NewGameFromState(cloneState(gs))

	err := g.PayBlinds()
	if err != nil {
		return nil, err
	}

	return nb.getState(g), nil
}

func (nb *NativeBackend) Pay(gs *pokerlib.GameState, chips int64) (*pokerlib.GameState, error) {

	g := nb.engine.NewGameFromState(cloneState(gs))

	err := g.Pay(chips)
	if err != nil {
		return nil, err
	}

	return nb.getState(g), nil
}

func (nb *NativeBackend) Fold(gs *pokerlib.GameState) (*pokerlib.GameState, error) {

	g := nb.engine.NewGameFromState(cloneState(gs))

	err := g.Fold()
	if err != nil {
		return nil, err
	}

	return nb.getState(g), nil
}

func (nb *NativeBackend) Check(gs *pokerlib.GameState) (*pokerlib.GameState, error) {

	g := nb.engine.NewGameFromState(cloneState(gs))

	err := g.Check()
	if err != nil {
		return nil, err
	}

	return nb.getState(g), nil
}

func (nb *NativeBackend) Call(gs *pokerlib.GameState) (*pokerlib.GameState, error) {

	g := nb.engine.NewGameFromState(cloneState(gs))

	err := g.Call()
	if err != nil {
		return nil, err
	}

	return nb.getState(g), nil
}

func (nb *NativeBackend) Allin(gs *pokerlib.GameState) (*pokerlib.GameState, error) {

	g := nb.engine.NewGameFromState(cloneState(gs))

	err := g.Allin()
	if err != nil {
		return nil, err
	}

	return nb.getState(g), nil
}

func (nb *NativeBackend) Bet(gs *pokerlib.GameState, chips int64) (*pokerlib.GameState, error) {

	g := nb.engine.NewGameFromState(cloneState(gs))

	err := g.Bet(chips)
	if err != nil {
		return nil, err
	}

	return nb.getState(g), nil
}

func (nb *NativeBackend) Raise(gs *pokerlib.GameState, chipLevel int64) (*pokerlib.GameState, error) {

	g := nb.engine.NewGameFromState(cloneState(gs))

	err := g.Raise(chipLevel)
	if err != nil {
		return nil, err
	}

	return nb.getState(g), nil
}
