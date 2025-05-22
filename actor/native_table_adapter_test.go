package actor

import (
	"sync"
	"testing"
	"time"

	"github.com/d-protocol/pokerlib/table"
	"github.com/stretchr/testify/assert"
)

func Test_NativeTableAdapter_Basic(t *testing.T) {

	backend := table.NewNativeBackend()
	opts := table.NewOptions()

	nt := table.NewTable(opts, table.WithBackend(backend))
	nt.SetAnte(10)
	nt.SetBlinds(0, 5, 10)

	// Initializing bots
	players := map[string]int64{
		"player_1": 10000,
		"player_2": 10000,
		"player_3": 10000,
	}

	// Preparing actors
	actors := make([]Actor, 0)
	for id, bankroll := range players {

		sid, _ := nt.Join(-1, &table.PlayerInfo{
			ID:       id,
			Bankroll: bankroll,
		})
		nt.Activate(sid)

		// Create new actor
		a := NewActor()

		// Initializing table engine adapter to communicate with native table
		ta := NewNativeTableAdapter(nt)
		a.SetAdapter(ta)

		// Initializing bot runner
		bot := NewBotRunner(id)
		a.SetRunner(bot)

		actors = append(actors, a)
	}

	// Setup state handler
	var wg sync.WaitGroup
	wg.Add(1)
	nt.OnStateUpdated(func(s *table.State) {

		// Update table state via adapter
		go func() {
			for _, a := range actors {
				a.GetTable().(*NativeTableAdapter).UpdateNativeState(s)
			}
		}()

		if s.Status == "playing" && s.GameState.Status.CurrentEvent == "GameClosed" {
			t.Logf("GameClosed (id=%s, playable_players=%d)", s.GameState.GameID, nt.GetPlayablePlayerCount())
		}

		if s.Status == "closed" {
			t.Log("TableClosed")
			assert.Less(t, nt.GetPlayablePlayerCount(), nt.GetState().Options.MinPlayers)
			wg.Done()
		}
	})

	// Not allow new player to join table
	nt.SetJoinable(false)

	assert.Nil(t, nt.Start())

	wg.Wait()
}

func Test_NativeTableAdapter_Join_Slowly(t *testing.T) {

	backend := table.NewNativeBackend()
	opts := table.NewOptions()
	opts.EliminateMode = "leave"
	opts.Interval = 100

	nt := table.NewTable(opts, table.WithBackend(backend))
	nt.SetAnte(10)
	nt.SetBlinds(0, 5, 10)

	// Initializing bots
	players := map[string]int64{
		"player_1": 10000,
		"player_2": 10000,
		"player_3": 10000,
		"player_4": 10000,
		"player_5": 10000,
		"player_6": 10000,
		"player_7": 10000,
		"player_8": 10000,
		"player_9": 10000,
	}

	// Preparing actors
	actors := make([]Actor, 0)

	go func() {
		for id, bankroll := range players {

			sid, err := nt.Join(-1, &table.PlayerInfo{
				ID:       id,
				Bankroll: bankroll,
			})
			assert.Nil(t, err)
			nt.Activate(sid)

			// Create new actor
			a := NewActor()

			// Initializing table engine adapter to communicate with native table
			ta := NewNativeTableAdapter(nt)
			a.SetAdapter(ta)

			// Initializing bot runner
			bot := NewBotRunner(id)
			a.SetRunner(bot)

			actors = append(actors, a)

			time.Sleep(2 * time.Millisecond)
		}

		// Disallow new player to join table
		nt.SetJoinable(false)
		t.Log("Stopped registration")
	}()

	// Setup state handler
	var wg sync.WaitGroup
	wg.Add(1)
	var round string
	var gameID string
	nt.OnStateUpdated(func(s *table.State) {

		//t.Log("OnStateUpdated", s.Status)

		if s.GameState != nil && gameID != s.GameState.GameID {
			gameID = s.GameState.GameID
			t.Logf("NewGame (id=%s)", gameID)
		}

		if s.Status == "playing" {
			// Intentionally causing delays to prevent some new players from participating in the game
			time.Sleep(10 * time.Millisecond)

			if s.GameState.Status.CurrentEvent == "GameClosed" {
				t.Logf("GameClosed (id=%s, playable_players=%d)", s.GameState.GameID, nt.GetPlayablePlayerCount())
			} else if round != s.GameState.Status.Round {
				round = s.GameState.Status.Round
				if len(round) > 0 {
					t.Log(" |", s.GameState.Status.Round, gameID)
				}
			}
		}

		if s.Status == "closed" {

			t.Log("TableClosed")
			assert.Less(t, nt.GetPlayablePlayerCount(), nt.GetState().Options.MinPlayers)

			// Waiting for last GameClosed event
			time.Sleep(time.Second)

			wg.Done()
		}

		// Update table state via adapter
		go func() {
			for _, a := range actors {
				go a.GetTable().(*NativeTableAdapter).UpdateNativeState(s)
			}
		}()
	})

	assert.Nil(t, nt.Start())

	wg.Wait()
}
