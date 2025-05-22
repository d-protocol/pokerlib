package pokerlib

import (
	"testing"

	"github.com/d-protocol/pokerlib"
	"github.com/stretchr/testify/assert"
)

func Test_PoorPlayer_DealerEnoughForAnteOnly(t *testing.T) {

	pf := pokerlib.NewPokerFace()

	// Options
	opts := pokerlib.NewStardardGameOptions()
	opts.Ante = 10

	// Preparing deck
	opts.Deck = pokerlib.NewStandardDeckCards()

	// Preparing players
	players := []*pokerlib.PlayerSetting{
		&pokerlib.PlayerSetting{
			Bankroll:  10,
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
		&pokerlib.PlayerSetting{
			Bankroll: 10000,
		},
	}
	opts.Players = append(opts.Players, players...)

	// Initializing game
	g := pf.NewGame(opts)

	// Start the game
	assert.Nil(t, g.Start())

	// Waiting for initial ready
	assert.Nil(t, g.ReadyForAll())

	// Ante
	assert.Nil(t, g.PayAnte())

	// Blinds
	assert.Nil(t, g.PayBlinds())

	// Round: Preflop
	assert.Nil(t, g.ReadyForAll()) // ready for the round
	assert.Equal(t, false, g.GetCurrentPlayer().CheckPosition("dealer"))
	assert.Equal(t, false, g.GetCurrentPlayer().CheckPosition("sb"))
	assert.Equal(t, false, g.GetCurrentPlayer().CheckPosition("bb"))
	assert.Nil(t, g.Call())
	assert.Nil(t, g.Pass())  // Dealer
	assert.Nil(t, g.Call())  // SB
	assert.Nil(t, g.Check()) // BB

	// Round: Flop
	assert.Nil(t, g.Next())
	assert.Nil(t, g.ReadyForAll()) // ready for the round
	assert.Equal(t, true, g.GetCurrentPlayer().CheckPosition("sb"))
	assert.Nil(t, g.Check()) // SB
	assert.Nil(t, g.Check()) // BB
	assert.Nil(t, g.Bet(100))
	assert.Nil(t, g.Pass()) // Dealer
	assert.Nil(t, g.Call()) // SB
	assert.Nil(t, g.Call()) // BB

	// Round: Turn
	assert.Nil(t, g.Next())
	assert.Nil(t, g.ReadyForAll()) // ready for the round
	assert.Equal(t, true, g.GetCurrentPlayer().CheckPosition("sb"))
	assert.Nil(t, g.Check())  // SB
	assert.Nil(t, g.Bet(100)) // BB
	assert.Nil(t, g.Raise(200))
	assert.Nil(t, g.Pass()) // Dealer
	assert.Nil(t, g.Call()) // SB
	assert.Nil(t, g.Call()) // BB

	// Round: River
	assert.Nil(t, g.Next())
	assert.Nil(t, g.ReadyForAll()) // ready for the round
	assert.Equal(t, true, g.GetCurrentPlayer().CheckPosition("sb"))
	assert.Nil(t, g.Check()) // SB
	assert.Nil(t, g.Check()) // BB
	assert.Nil(t, g.Check())
	assert.Nil(t, g.Pass()) // Dealer

	// Game closed
	assert.Nil(t, g.Next())
}

func Test_PoorPlayer_DealerEnoughForAnteAndBlindOnly(t *testing.T) {

	pf := pokerlib.NewPokerFace()

	// Options
	opts := pokerlib.NewStardardGameOptions()
	opts.Ante = 10

	// Preparing deck
	opts.Deck = pokerlib.NewStandardDeckCards()

	// Preparing players
	players := []*pokerlib.PlayerSetting{
		&pokerlib.PlayerSetting{
			Bankroll:  10000,
			Positions: []string{"dealer"},
		},
		&pokerlib.PlayerSetting{
			Bankroll:  15,
			Positions: []string{"sb"},
		},
		&pokerlib.PlayerSetting{
			Bankroll:  10000,
			Positions: []string{"bb"},
		},
		&pokerlib.PlayerSetting{
			Bankroll: 10000,
		},
	}
	opts.Players = append(opts.Players, players...)

	// Initializing game
	g := pf.NewGame(opts)

	// Start the game
	assert.Nil(t, g.Start())

	// Waiting for initial ready
	assert.Nil(t, g.ReadyForAll())

	// Ante
	assert.Nil(t, g.PayAnte())

	// Blinds
	assert.Nil(t, g.PayBlinds())

	// Round: Preflop
	assert.Nil(t, g.ReadyForAll()) // ready for the round
	assert.Equal(t, false, g.GetCurrentPlayer().CheckPosition("dealer"))
	assert.Equal(t, false, g.GetCurrentPlayer().CheckPosition("sb"))
	assert.Equal(t, false, g.GetCurrentPlayer().CheckPosition("bb"))
	assert.Nil(t, g.Call())
	assert.Nil(t, g.Call()) // Dealer

	// SB has no enough stack
	assert.Equal(t, true, g.GetCurrentPlayer().CheckPosition("sb"))
	assert.Equal(t, int64(0), g.GetCurrentPlayer().State().StackSize)
	assert.Nil(t, g.Pass()) // SB

	assert.Nil(t, g.Check()) // BB

	// Round: Flop
	assert.Nil(t, g.Next())
	assert.Nil(t, g.ReadyForAll()) // ready for the round
	assert.Equal(t, true, g.GetCurrentPlayer().CheckPosition("sb"))
	assert.Nil(t, g.Pass())  // SB
	assert.Nil(t, g.Check()) // BB
	assert.Nil(t, g.Bet(100))
	assert.Nil(t, g.Call()) // Dealer
	assert.Nil(t, g.Pass()) // SB
	assert.Nil(t, g.Call()) // BB

	// Round: Turn
	assert.Nil(t, g.Next())
	assert.Nil(t, g.ReadyForAll()) // ready for the round
	assert.Equal(t, true, g.GetCurrentPlayer().CheckPosition("sb"))
	assert.Nil(t, g.Pass())   // SB
	assert.Nil(t, g.Bet(100)) // BB
	assert.Nil(t, g.Raise(200))
	assert.Nil(t, g.Call()) // Dealer
	assert.Nil(t, g.Pass()) // SB
	assert.Nil(t, g.Call()) // BB

	// Round: River
	assert.Nil(t, g.Next())
	assert.Nil(t, g.ReadyForAll()) // ready for the round
	assert.Equal(t, true, g.GetCurrentPlayer().CheckPosition("sb"))
	assert.Nil(t, g.Pass())  // SB
	assert.Nil(t, g.Check()) // BB
	assert.Nil(t, g.Check())
	assert.Nil(t, g.Check()) // Dealer

	// Game closed
	assert.Nil(t, g.Next())
}
