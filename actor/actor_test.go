package actor

import (
	"sync"
	"testing"

	pokertable "github.com/d-protocol/pokertable"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestActor_Basic(t *testing.T) {

	// Initializing table
	manager := pokertable.NewManager()
	tableEngineOptions := pokertable.NewTableEngineOptions()
	tableEngineCallbacks := pokertable.NewTableEngineCallbacks()
	tableEngineCallbacks.OnTableErrorUpdated = func(table *pokertable.Table, err error) {
		t.Logf("Table error: %v", err)
	}
	table, err := manager.CreateTable(tableEngineOptions, tableEngineCallbacks, pokertable.TableSetting{
		TableID: uuid.New().String(),
		Meta: pokertable.TableMeta{
			CompetitionID:       "1005c477-84b4-4d1b-9fca-3a6ad84e0fe7",
			Rule:                pokertable.CompetitionRule_Default,
			Mode:                pokertable.CompetitionMode_CT,
			MaxDuration:         3,
			TableMaxSeatCount:   9,
			TableMinPlayerCount: 2,
			MinChipUnit:         10,
			ActionTime:          10,
		},
	})
	assert.Nil(t, err)

	tableEngine, err := manager.GetTableEngine(table.ID)
	assert.Nil(t, err)

	// Initializing bot
	players := []pokertable.JoinPlayer{
		{PlayerID: "Jeffrey", RedeemChips: 3000},
		{PlayerID: "Chuck", RedeemChips: 3000},
		{PlayerID: "Fred", RedeemChips: 3000},
	}

	// Preparing actors
	actors := make([]Actor, 0)
	for _, p := range players {

		// Create new actor
		a := NewActor()

		// Initializing table engine adapter to communicate with table engine
		tc := NewTableEngineAdapter(tableEngine, table)
		a.SetAdapter(tc)

		// Initializing bot runner
		bot := NewBotRunner(p.PlayerID)
		a.SetRunner(bot)

		actors = append(actors, a)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	// Preparing table state updater
	tableEngine.OnTableUpdated(func(table *pokertable.Table) {

		if table.State.Status == pokertable.TableStateStatus_TableGameSettled {
			if table.State.GameState.Status.CurrentEvent == "GameClosed" {
				t.Log("GameClosed", table.State.GameState.GameID)

				if len(table.State.GameState.Players) == 1 {
					tableEngine.CloseTable()
					wg.Done()
					return
				}

				json, _ := table.GetJSON()
				t.Log(json)
			}
		}

		// Update table state via adapter
		go func() {
			for _, a := range actors {
				a.GetTable().UpdateTableState(table)
			}
		}()

	})

	// Add player to table
	for _, p := range players {
		tableEngine.PlayerReserve(p)
		err := tableEngine.PlayerJoin(p.PlayerID)
		assert.Nil(t, err)
	}

	// Start game
	err = tableEngine.StartTableGame()
	assert.Nil(t, err)

	wg.Wait()
}
