package pokerlib

import (
	"testing"

	"github.com/d-protocol/pokerlib"
	"github.com/stretchr/testify/assert"
)

func Test_ShortDeck_Player_Wager(t *testing.T) {

	pf := pokerlib.NewPokerFace()

	opts := pokerlib.NewShortDeckGameOptions()
	opts.Blind.SB = 0
	opts.Blind.BB = 0
	opts.Blind.Dealer = 100
	opts.Ante = 10

	// Preparing deck
	opts.Deck = pokerlib.NewShortDeckCards()

	// Preparing players
	players := []*pokerlib.PlayerSetting{
		&pokerlib.PlayerSetting{
			Bankroll:  10000,
			Positions: []string{"dealer"},
		},
		&pokerlib.PlayerSetting{
			Bankroll:  10000,
			Positions: []string{"sb"},
		},
		&pokerlib.PlayerSetting{
			Bankroll:  10000,
			Positions: []string{"bb"},
		},
	}
	opts.Players = append(opts.Players, players...)

	// Initializing game
	g := pf.NewGame(opts)
	assert.Nil(t, g.Start())

	// Waiting for ready
	for _, p := range g.GetPlayers() {
		assert.Equal(t, 0, len(p.State().HoleCards))
		assert.Equal(t, false, p.State().Fold)
		assert.Equal(t, int64(0), p.State().Wager)
		assert.Equal(t, int64(0), p.State().Pot)
		assert.Equal(t, int64(players[p.SeatIndex()].Bankroll), p.State().Bankroll)
		assert.Equal(t, int64(players[p.SeatIndex()].Bankroll), p.State().InitialStackSize)
		assert.Equal(t, int64(players[p.SeatIndex()].Bankroll), p.State().StackSize)

		// Position checks
		if p.SeatIndex() == 0 {
			assert.True(t, p.CheckPosition("dealer"))
		}
	}

	// Waiting for ready
	assert.Equal(t, "ReadyRequested", g.GetState().Status.CurrentEvent)
	assert.Nil(t, g.ReadyForAll())

	// ante
	assert.Equal(t, "AnteRequested", g.GetState().Status.CurrentEvent)

	for _, p := range g.GetPlayers() {
		assert.Equal(t, false, p.State().Acted)
		assert.Equal(t, 0, len(p.State().HoleCards))
		assert.Equal(t, false, p.State().Fold)
		assert.Equal(t, int64(0), p.State().Wager)
		assert.Equal(t, int64(0), p.State().Pot)
		assert.Equal(t, int64(players[p.SeatIndex()].Bankroll), p.State().Bankroll)
		assert.Equal(t, int64(players[p.SeatIndex()].Bankroll), p.State().InitialStackSize)
		assert.Equal(t, int64(players[p.SeatIndex()].Bankroll), p.State().StackSize)
	}

	assert.Nil(t, g.PayAnte())

	// Entering Preflop
	t.Log("Entering \"Prflop\" round")
	assert.Equal(t, "preflop", g.GetState().Status.Round)

	// Blinds
	assert.Equal(t, "BlindsRequested", g.GetState().Status.CurrentEvent)
	for _, p := range g.GetPlayers() {
		assert.Equal(t, false, p.State().Acted)
		assert.Equal(t, 2, len(p.State().HoleCards))
		assert.Equal(t, false, p.State().Fold)
		assert.Equal(t, int64(0), p.State().Wager)
		assert.Equal(t, int64(10), p.State().Pot)
	}

	assert.Nil(t, g.PayBlinds())

	// Check Player Wager Value
	for _, p := range g.GetPlayers() {
		if p.SeatIndex() == 0 {
			assert.Equal(t, p.State().Wager, opts.Blind.Dealer, "wager should be 100")
		} else {
			assert.Equal(t, p.State().Wager, int64(0), "wager should be 0")
		}
	}

	//g.PrintState()
}
