package pokerlib

import (
	"testing"
)

// TestNextAfterRoundClosed tests if the Next() method works correctly
// after a round is closed
func TestNextAfterRoundClosed(t *testing.T) {
	// Create a game
	opts := &GameOptions{
		Blind: BlindSetting{
			SB: 1,
			BB: 2,
		},
		Limit:                  "no-limit",
		HoleCardsCount:         2,
		RequiredHoleCardsCount: 0,
		Deck:                   NewStandardDeckCards(),
	}
	
	// Add players
	opts.Players = make([]*PlayerSetting, 0)
	
	// Add dealer
	opts.Players = append(opts.Players, &PlayerSetting{
		Positions: []string{"dealer"},
		Bankroll:  100,
	})
	
	// Add small blind
	opts.Players = append(opts.Players, &PlayerSetting{
		Positions: []string{"sb"},
		Bankroll:  100,
	})
	
	// Add big blind
	opts.Players = append(opts.Players, &PlayerSetting{
		Positions: []string{"bb"},
		Bankroll:  100,
	})
	
	// Create game
	g := &game{
		players: make(map[int]Player),
	}
	g.ApplyOptions(opts)
	
	// Set up the game state for testing
	g.gs.Status.Round = "preflop"
	g.gs.Status.CurrentEvent = "RoundClosed"
	
	// Test the Next() method directly
	err := g.Next()
	if err != nil {
		t.Errorf("Next() failed with error: %v", err)
	}
	
	// Verify that we moved to the next round (flop)
	if g.gs.Status.Round != "flop" {
		t.Errorf("Expected round to be 'flop', got '%s'", g.gs.Status.Round)
	}
}