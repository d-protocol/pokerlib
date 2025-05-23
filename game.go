package pokerlib

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/d-protocol/pokerlib/pot"
)

var (
	ErrNoDeck                      = errors.New("game: no deck")
	ErrNotEnoughBackroll           = errors.New("game: backroll is not enough")
	ErrNoDealer                    = errors.New("game: no dealer")
	ErrInsufficientNumberOfPlayers = errors.New("game: insufficient number of players")
	ErrUnknownRound                = errors.New("game: unknown round")
	ErrNotFoundDealer              = errors.New("game: not found dealer")
	ErrUnknownTask                 = errors.New("game: unknown task")
	ErrNotClosedRound              = errors.New("game: round is not closed")
)

type Game interface {
	ApplyOptions(opts *GameOptions) error
	Start() error
	Resume() error
	GetEvent() string
	GetState() *GameState
	GetStateJSON() ([]byte, error)
	LoadState(gs *GameState) error
	Player(idx int) Player
	Dealer() Player
	SmallBlind() Player
	BigBlind() Player
	Deal(count int) []string
	Burn(count int) error
	BecomeRaiser(Player) error
	ResetActedPlayers() error
	ResetAllPlayerStatus() error
	StartAtDealer() (Player, error)
	GetPlayerCount() int
	GetPlayers() []Player
	SetCurrentPlayer(Player) error
	GetCurrentPlayer() Player
	GetAllowedActions(Player) []string
	GetAvailableActions(Player) []string
	GetAlivePlayerCount() int
	GetMovablePlayerCount() int
	UpdateLastAction(source int, ptype string, value int64) error
	EmitEvent(event GameEvent) error
	PrintState() error
	PrintPots()

	// Operations
	Next() error
	ReadyForAll() error
	PayAnte() error
	PayBlinds() error

	// Actions
	Pass() error
	Pay(chips int64) error
	Fold() error
	Check() error
	Call() error
	Allin() error
	Bet(chips int64) error
	Raise(chipLevel int64) error
}

type game struct {
	gs         *GameState
	players    map[int]Player
	dealer     Player
	smallBlind Player
	bigBlind   Player
}

func NewGame(opts *GameOptions) *game {
	g := &game{
		players: make(map[int]Player),
	}
	g.ApplyOptions(opts)
	return g
}

func NewGameFromState(gs *GameState) *game {
	g := &game{
		players: make(map[int]Player),
	}
	g.LoadState(gs)
	return g
}

func (g *game) onBreakPoint() {
	g.gs.UpdatedAt = time.Now().UnixNano()
	//atomic.AddInt64(&g.gs.UpdatedAt, 1)
}

func (g *game) GetState() *GameState {
	return g.gs
}

func (g *game) GetStateJSON() ([]byte, error) {
	return json.Marshal(g.gs)
}

func (g *game) LoadState(gs *GameState) error {
	g.gs = gs

	// Initializing players
	for _, ps := range g.gs.Players {
		g.addPlayer(ps)
	}

	return nil
}

func (g *game) Resume() error {

	// emit event if state has event
	if len(g.gs.Status.CurrentEvent) > 0 {
		event := GameEventBySymbol[g.gs.Status.CurrentEvent]

		//fmt.Printf("Resume: %s\n", g.gs.Status.CurrentEvent.Name)

		// Activate by the last event
		return g.EmitEvent(event)
	}

	return nil
}

func (g *game) ApplyOptions(opts *GameOptions) error {

	g.gs = &GameState{
		Players: make([]*PlayerState, 0),
		Meta: Meta{
			Ante:                   opts.Ante,
			Blind:                  opts.Blind,
			Limit:                  opts.Limit,
			HoleCardsCount:         opts.HoleCardsCount,
			RequiredHoleCardsCount: opts.RequiredHoleCardsCount,
			CombinationPowers:      opts.CombinationPowers,
			Deck:                   opts.Deck,
			BurnCount:              opts.BurnCount,
		},
	}

	// Loading players
	for idx, p := range opts.Players {
		g.AddPlayer(idx, p)
	}

	return nil
}

func (g *game) addPlayer(state *PlayerState) error {

	// Create player instance
	p := &player{
		idx:   state.Idx,
		game:  g,
		state: state,
	}

	if p.CheckPosition("dealer") {
		g.dealer = p
	}

	if p.CheckPosition("sb") {
		g.smallBlind = p
	} else if p.CheckPosition("bb") {
		g.bigBlind = p
	}

	g.players[state.Idx] = p

	return nil
}

func (g *game) AddPlayer(idx int, setting *PlayerSetting) error {

	// Create player state
	ps := &PlayerState{
		Idx:              idx,
		Positions:        setting.Positions,
		Bankroll:         setting.Bankroll,
		InitialStackSize: setting.Bankroll,
		StackSize:        setting.Bankroll,
		Combination:      &CombinationInfo{},
	}

	g.gs.Players = append(g.gs.Players, ps)

	return g.addPlayer(ps)
}

func (g *game) Player(idx int) Player {

	if idx < 0 || idx >= g.GetPlayerCount() {
		return nil
	}

	return g.players[idx]
}

func (g *game) Dealer() Player {
	return g.dealer
}

func (g *game) SmallBlind() Player {
	return g.smallBlind
}

func (g *game) BigBlind() Player {
	return g.bigBlind
}

func (g *game) Deal(count int) []string {

	cards := make([]string, 0, count)

	finalPos := g.gs.Status.CurrentDeckPosition + count
	for i := g.gs.Status.CurrentDeckPosition; i < finalPos; i++ {
		cards = append(cards, g.gs.Meta.Deck[i])
		g.gs.Status.CurrentDeckPosition++
	}

	return cards
}

func (g *game) Burn(count int) error {
	g.gs.Status.Burned = append(g.gs.Status.Burned, g.Deal(count)...)
	return nil
}

func (g *game) ResetAllPlayerAllowedActions() error {
	for _, p := range g.GetPlayers() {
		p.Reset()
	}

	return nil
}

func (g *game) ResetAllPlayerStatus() error {
	for _, p := range g.GetPlayers() {
		ps := p.State()
		ps.AllowedActions = make([]string, 0)
		ps.Pot += ps.Wager
		ps.Wager = 0
		ps.InitialStackSize = ps.StackSize

		if ps.Fold {
			ps.DidAction = "fold"
		} else if ps.InitialStackSize == 0 {
			ps.DidAction = "allin"
		} else {
			ps.DidAction = ""
		}
	}

	return nil
}

func (g *game) ResetRoundStatus() error {
	g.gs.Status.PreviousRaiseSize = 0
	g.gs.Status.MaxWager = 0
	g.gs.Status.CurrentRoundPot = 0
	g.gs.Status.CurrentWager = 0
	g.gs.Status.CurrentRaiser = g.Dealer().State().Idx
	g.gs.Status.CurrentPlayer = g.gs.Status.CurrentRaiser
	return nil
}

func (g *game) StartAtDealer() (Player, error) {

	// Start at dealer
	dealer := g.Dealer()
	if dealer == nil {
		return nil, ErrNotFoundDealer
	}

	// Update status
	err := g.SetCurrentPlayer(dealer)
	if err != nil {
		return nil, err
	}

	return dealer, nil
}

func (g *game) GetCurrentPlayer() Player {
	return g.Player(g.gs.Status.CurrentPlayer)
}

func (g *game) NextPlayer() Player {

	cur := g.gs.Status.CurrentPlayer
	playerCount := g.GetPlayerCount()

	for i := 1; i < playerCount; i++ {

		// Find the next player
		cur++

		// The end of player list
		if cur == playerCount {
			cur = 0
		}

		p := g.gs.Players[cur]

		return g.Player(p.Idx)
	}

	return nil
}

func (g *game) GetPlayerCount() int {
	return len(g.gs.Players)
}

func (g *game) GetPlayers() []Player {

	players := make([]Player, 0)
	playerCount := g.GetPlayerCount()

	// Getting player list that dealer should be the first element of it
	cur := g.Dealer().SeatIndex()

	for i := 0; i < playerCount; i++ {

		players = append(players, g.players[cur])

		// Find the next player
		cur++

		// The end of player list
		if cur == playerCount {
			cur = 0
		}
	}

	return players
}

func (g *game) setCurrentPlayer(p Player) error {

	// Clear status
	if p == nil {
		g.gs.Status.CurrentPlayer = -1
		return nil
	}

	// Deside player who can move
	g.gs.Status.CurrentPlayer = p.SeatIndex()

	return nil
}

func (g *game) SetCurrentPlayer(p Player) error {

	if g.gs.Status.CurrentPlayer != -1 {
		// Clear allowed actions of current player
		g.GetCurrentPlayer().ResetAllowedActions()
	}

	err := g.setCurrentPlayer(p)
	if err != nil {
		return err
	}

	if p != nil {
		// Figure out actions that player can be allowed to take
		actions := g.GetAllowedActions(p)
		p.AllowActions(actions)
	}

	return nil
}

func (g *game) GetAlivePlayerCount() int {

	aliveCount := g.GetPlayerCount()

	for _, p := range g.gs.Players {
		if p.Fold {
			aliveCount--
		}
	}

	return aliveCount
}

func (g *game) GetMovablePlayerCount() int {

	mCount := g.GetPlayerCount()

	for _, p := range g.gs.Players {
		// Fold or allin
		if p.Fold || p.StackSize == 0 {
			mCount--
		}
	}

	return mCount
}

func (g *game) BecomeRaiser(p Player) error {

	if p.State().Wager > 0 {
		p.State().VPIP = true
	}

	g.gs.Status.CurrentRaiser = p.SeatIndex()

	// Reset all player states except raiser
	g.ResetActedPlayers()
	p.State().Acted = true

	return nil
}

func (g *game) ResetActedPlayers() error {
	for _, ps := range g.gs.Players {
		ps.Acted = false
	}

	return nil
}

func (g *game) RequestPlayerAction() error {

	// only one player left
	if g.GetAlivePlayerCount() == 1 {
		return g.EmitEvent(GameEvent_RoundClosed)
	}

	// no player can move because everybody did all-in already for this game
	if g.GetMovablePlayerCount() == 0 {
		return g.EmitEvent(GameEvent_RoundClosed)
	}

	// next player
	p := g.NextPlayer()

	// Run around already, no one need to act
	if p.State().Acted {
		return g.EmitEvent(GameEvent_RoundClosed)
	}

	return g.SetCurrentPlayer(p)
}

func (g *game) UpdateLastAction(source int, aType string, value int64) error {

	if g.gs.Status.LastAction == nil {
		g.gs.Status.LastAction = &Action{
			Source: source,
			Type:   aType,
			Value:  value,
		}

		return nil
	}

	g.gs.Status.LastAction.Source = source
	g.gs.Status.LastAction.Type = aType
	g.gs.Status.LastAction.Value = value

	return nil
}

func (g *game) GetAllowedActions(p Player) []string {

	// player is movable for this round
	if g.gs.Status.CurrentPlayer == p.SeatIndex() {
		return g.GetAvailableActions(p)
	}

	return make([]string, 0)
}

func (g *game) GetAvailableActions(p Player) []string {

	actions := make([]string, 0)

	// Invalid
	if p == nil {
		return actions
	}

	ps := p.State()

	if ps.Fold {
		actions = append(actions, "pass")
		return actions
	}

	// chips left
	if ps.StackSize == 0 {
		actions = append(actions, "pass")
		return actions
	} else {
		actions = append(actions, "allin")
	}

	if ps.Wager < g.gs.Status.CurrentWager {
		actions = append(actions, "fold")

		// call
		if ps.InitialStackSize > g.gs.Status.CurrentWager {

			actions = append(actions, "call")

			// raise
			if ps.InitialStackSize > g.gs.Status.CurrentWager+g.gs.Status.PreviousRaiseSize {
				actions = append(actions, "raise")
			}
		}

	} else {
		actions = append(actions, "check")

		if ps.InitialStackSize >= g.gs.Status.MiniBet {
			if g.gs.Status.CurrentWager == 0 {
				actions = append(actions, "bet")
			} else {
				actions = append(actions, "raise")
			}
		}
	}

	return actions
}

func (g *game) Start() error {

	// Check the number of players
	if g.GetPlayerCount() < 2 {
		return ErrInsufficientNumberOfPlayers
	}

	// Require dealer
	if g.dealer == nil {
		return ErrNoDealer
	}

	// Check backroll
	for _, p := range g.gs.Players {

		if p.Bankroll <= 0 {
			return ErrNotEnoughBackroll
		}
	}

	// No desk was set
	if len(g.gs.Meta.Deck) == 0 {
		return ErrNoDeck
	}

	// Initializing game status
	g.gs.Status.Pots = make([]*pot.Pot, 0)
	g.gs.Status.Board = make([]string, 0)
	g.gs.Status.Burned = make([]string, 0)
	g.gs.Status.CurrentEvent = ""

	return g.EmitEvent(GameEvent_Started)
}

func (g *game) Initialize() error {

	// Shuffle cards
	g.gs.Meta.Deck = ShuffleCards(g.gs.Meta.Deck)

	// Initialize minimum bet
	if g.gs.Meta.Blind.Dealer > g.gs.Meta.Blind.BB {
		g.gs.Status.MiniBet = g.gs.Meta.Blind.Dealer
	} else {
		g.gs.Status.MiniBet = g.gs.Meta.Blind.BB
	}

	g.ResetRoundStatus()

	return g.EmitEvent(GameEvent_Initialized)
}

func (g *game) Prepare() error {
	return g.RequestReady()
}

func (g *game) RequestReady() error {

	// Clear all player allowed actions before request ready
	g.ResetAllPlayerAllowedActions()

	return g.EmitEvent(GameEvent_ReadyRequested)
}

func (g *game) RequestAnte() error {
	return g.EmitEvent(GameEvent_AnteRequested)
}

func (g *game) RequestBlinds() error {

	// No need to pay blinds
	if g.gs.Meta.Blind.Dealer == 0 && g.gs.Meta.Blind.SB == 0 && g.gs.Meta.Blind.BB > 0 {
		return g.EmitEvent(GameEvent_BlindsPaid)
	}

	return g.EmitEvent(GameEvent_BlindsRequested)
}

func (g *game) Next() error {

	g.UpdateLastAction(-1, "next", 0)

	switch g.gs.Status.Round {
	case "preflop":
		fallthrough
	case "flop":
		fallthrough
	case "turn":
		fallthrough
	case "river":
		return g.nextRound()
	}

	return nil
}

func (g *game) nextRound() error {

	g.ResetRoundStatus()
	g.ResetAllPlayerStatus()

	if g.GetAlivePlayerCount() == 1 {
		// Game is completed
		return g.EmitEvent(GameEvent_GameCompleted)
	}

	// Going to the next round
	switch g.gs.Status.Round {
	case "preflop":
		return g.EnterFlopRound()
	case "flop":
		return g.EnterTurnRound()
	case "turn":
		return g.EnterRiverRound()
	case "river":
		return g.EmitEvent(GameEvent_GameCompleted)
	}

	return ErrUnknownRound
}

func (g *game) EnterPreflopRound() error {
	g.gs.Status.Round = "preflop"
	return g.EmitEvent(GameEvent_PreflopRoundEntered)
}

func (g *game) EnterFlopRound() error {
	g.gs.Status.Round = "flop"
	return g.EmitEvent(GameEvent_FlopRoundEntered)
}

func (g *game) EnterTurnRound() error {
	g.gs.Status.Round = "turn"
	return g.EmitEvent(GameEvent_TurnRoundEntered)
}

func (g *game) EnterRiverRound() error {
	g.gs.Status.Round = "river"
	return g.EmitEvent(GameEvent_RiverRoundEntered)
}

func (g *game) InitializeRound() error {

	// Initializing for stages (Preflop, Flop, Turn and River)
	switch g.gs.Status.Round {
	case "preflop":

		// Deal cards to players
		for _, p := range g.gs.Players {
			p.HoleCards = g.Deal(g.gs.Meta.HoleCardsCount)
		}
	case "flop":

		g.Burn(1)

		// Deal 3 board cards
		g.gs.Status.Board = append(g.gs.Status.Board, g.Deal(3)...)

		// Start at dealer
		_, err := g.StartAtDealer()
		if err != nil {
			return err
		}

	case "turn":
		fallthrough
	case "river":

		g.Burn(1)

		// Deal board card
		g.gs.Status.Board = append(g.gs.Status.Board, g.Deal(1)...)

		// Start at dealer
		_, err := g.StartAtDealer()
		if err != nil {
			return err
		}
	}

	// Calculate power of the best combination for each player
	err := g.UpdateCombinationOfAllPlayers()
	if err != nil {
		return err
	}

	return g.EmitEvent(GameEvent_RoundInitialized)
}

func (g *game) PrepareRound() error {

	//fmt.Printf("Preparing round: %s\n", g.gs.Status.Round)

	if g.gs.Status.Round == "preflop" {
		return g.RequestReady()
	}

	// Everybody did all-in or one movable player left, no need to keep going with normal way
	if g.GetMovablePlayerCount() <= 1 {
		return g.EmitEvent(GameEvent_RoundClosed)
	}

	return g.RequestReady()
}

func (g *game) StartRound() error {

	g.ResetAllPlayerAllowedActions()

	if g.gs.Status.Round == "preflop" {

		// everyone did all-in, no need to keep going with normal way
		if g.GetMovablePlayerCount() == 0 {
			return g.EmitEvent(GameEvent_RoundClosed)
		}

		// Set Dealer to the first player
		g.SetCurrentPlayer(g.Dealer())

		for i := 0; i < g.GetPlayerCount(); i++ {
			p := g.NextPlayer()

			if p.CheckPosition("bb") {
				g.SetCurrentPlayer(g.NextPlayer())
				break
			}

			g.SetCurrentPlayer(p)
		}

	} else {

		_, err := g.StartAtDealer()
		if err != nil {
			return err
		}
	}

	return g.EmitEvent(GameEvent_RoundStarted)
}

func (g *game) PrintState() error {

	data, err := g.GetStateJSON()
	if err != nil {
		return err
	}

	fmt.Println(string(data))

	return nil
}
