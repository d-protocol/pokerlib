package actor

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/d-protocol/pokerlib"
	"github.com/d-protocol/pokertable"
	"github.com/d-protocol/timebank"
)

type ActionProbability struct {
	Action string
	Weight float64
}

var (
	actionProbabilities = []ActionProbability{
		{Action: "check", Weight: 0.1},
		{Action: "call", Weight: 0.3},
		{Action: "fold", Weight: 0.15},
		{Action: "allin", Weight: 0.05},
		{Action: "raise", Weight: 0.3},
		{Action: "bet", Weight: 0.1},
	}
)

type BotRunner struct {
	actor             Actor
	actions           Actions
	playerID          string
	isHumanized       bool
	curGameID         string
	lastGameStateTime int64
	timebank          *timebank.TimeBank
	tableInfo         *pokertable.Table
}

func NewBotRunner(playerID string) *BotRunner {
	return &BotRunner{
		playerID: playerID,
		timebank: timebank.NewTimeBank(),
	}
}

func (br *BotRunner) SetActor(a Actor) {
	br.actor = a
	br.actions = NewActions(a, br.playerID)
}

func (br *BotRunner) Humanized(enabled bool) {
	br.isHumanized = enabled
}

func (br *BotRunner) UpdateTableState(table *pokertable.Table) error {

	gs := table.State.GameState
	//oldState := br.tableInfo.State
	br.tableInfo = table

	// Check if you have been eliminated
	isEliminated := true
	for _, ps := range table.State.PlayerStates {
		if ps.PlayerID == br.playerID {
			isEliminated = false
		}
	}

	if isEliminated {
		return nil
	}

	// The state remains unchanged or is outdated
	if gs != nil {

		// New game
		if gs.GameID != br.curGameID {
			br.curGameID = gs.GameID
		} else if br.lastGameStateTime >= gs.UpdatedAt {
			// Ignore if game state is too old
			//fmt.Println(br.playerID, table.ID)
			return nil
		}

		br.lastGameStateTime = gs.UpdatedAt
	}

	if table.State.Status == pokertable.TableStateStatus_TableGameStandby {
		return nil
	}

	// Getting player index in game
	gamePlayerIdx := table.GamePlayerIndex(br.playerID)

	// Somehow, this player is not in the game.
	// It probably has no chips already or just sat down and have not participated in the game yet
	if gamePlayerIdx == -1 {
		return nil
	}

	if table.State.Status != pokertable.TableStateStatus_TableGamePlaying {
		return nil
	}

	//fmt.Printf("Bot (player_id=%s, gameIdx=%d, event=%s)\n", br.playerID, gamePlayerIdx, gs.Status.CurrentEvent)

	// game is running so we have to check actions allowed
	player := gs.GetPlayer(gamePlayerIdx)
	if player == nil {
		return nil
	}

	if len(player.AllowedActions) > 0 {
		//fmt.Println(br.playerID, player.AllowedActions)
		err := br.requestMove(table.State.GameState, gamePlayerIdx)
		if err != nil {
			fmt.Printf("Bot got Failed (player_id=%s, gameIdx=%d, event=%s)\n", br.playerID, gamePlayerIdx, gs.Status.CurrentEvent)
			return err
		}
	}

	return nil
}

func (br *BotRunner) requestMove(gs *pokerlib.GameState, playerIdx int) error {

	//fmt.Println(br.tableInfo.State.GameState.Status.Round, br.gamePlayerIdx, gs.Players[br.gamePlayerIdx].AllowedActions)
	/*
		player := gs.Players[playerIdx]
		if len(player.AllowedActions) == 1 {
			fmt.Println(br.playerID, player.AllowedActions)
		}
	*/

	// Do ready() and pay() automatically
	if gs.HasAction(playerIdx, "ready") {
		return br.actions.Ready()
	} else if gs.HasAction(playerIdx, "pass") {
		return br.actions.Pass()
	} else if gs.HasAction(playerIdx, "pay") {

		// Pay for ante and blinds
		switch gs.Status.CurrentEvent {
		case pokerlib.GameEventSymbols[pokerlib.GameEvent_AnteRequested]:

			// Ante
			return br.actions.Pay(gs.Meta.Ante)

		case pokerlib.GameEventSymbols[pokerlib.GameEvent_BlindsRequested]:

			// blinds
			if gs.HasPosition(playerIdx, "sb") {
				return br.actions.Pay(gs.Meta.Blind.SB)
			} else if gs.HasPosition(playerIdx, "bb") {
				return br.actions.Pay(gs.Meta.Blind.BB)
			}

			return br.actions.Pay(gs.Meta.Blind.Dealer)
		}
	}

	if !br.isHumanized || br.tableInfo.Meta.ActionTime == 0 {
		return br.requestAI(gs, playerIdx)
	}

	// For simulating human-like behavior, to incorporate random delays when performing actions.
	thinkingTime := rand.Intn(br.tableInfo.Meta.ActionTime)
	if thinkingTime == 0 {
		return br.requestAI(gs, playerIdx)
	}

	return br.timebank.NewTask(time.Duration(thinkingTime)*time.Second, func(isCancelled bool) {

		if isCancelled {
			return
		}

		br.requestAI(gs, playerIdx)
	})
}

func (br *BotRunner) calcActionProbabilities(actions []string) map[string]float64 {

	probabilities := make(map[string]float64)
	totalWeight := 0.0
	for _, action := range actions {

		for _, p := range actionProbabilities {
			if action == p.Action {
				probabilities[action] = p.Weight
				totalWeight += p.Weight
				break
			}
		}
	}

	scaleRatio := 1.0 / totalWeight
	weightLevel := 0.0
	for action, weight := range probabilities {
		scaledWeight := weight * scaleRatio
		weightLevel += scaledWeight
		probabilities[action] = weightLevel
	}

	return probabilities
}

func (br *BotRunner) calcAction(actions []string) string {

	// Select action randomly
	rand.Seed(time.Now().UnixNano())

	probabilities := br.calcActionProbabilities(actions)
	randomNum := rand.Float64()

	for action, probability := range probabilities {
		if randomNum < probability {
			return action
		}
	}

	return actions[len(actions)-1]
}

func (br *BotRunner) requestAI(gs *pokerlib.GameState, playerIdx int) error {

	player := gs.Players[playerIdx]

	// None of actions is allowed
	if len(player.AllowedActions) == 0 {
		return nil
	}

	action := player.AllowedActions[0]

	if len(player.AllowedActions) > 1 {
		action = br.calcAction(player.AllowedActions)
	}

	// Calculate chips
	chips := int64(0)

	/*
		// Debugging messages
		defer func() {
			if chips > 0 {
				fmt.Printf("Action %s %v %s(%d)\n", br.playerID, player.AllowedActions, action, chips)
			} else {
				fmt.Printf("Action %s %v %s\n", br.playerID, player.AllowedActions, action)
			}
		}()
	*/

	switch action {
	case "bet":

		minBet := gs.Status.MiniBet

		if player.InitialStackSize <= minBet {
			return br.actions.Bet(player.InitialStackSize)
		}

		chips = rand.Int63n(player.InitialStackSize-minBet) + minBet

		return br.actions.Bet(chips)
	case "raise":

		maxChipLevel := player.InitialStackSize
		minChipLevel := gs.Status.CurrentWager + gs.Status.PreviousRaiseSize

		if maxChipLevel <= minChipLevel {
			return br.actions.Raise(maxChipLevel)
		}

		chips = rand.Int63n(maxChipLevel-minChipLevel) + minChipLevel

		return br.actions.Raise(chips)
	case "call":
		return br.actions.Call()
	case "check":
		return br.actions.Check()
	case "allin":
		return br.actions.Allin()
	}

	return br.actions.Fold()
}
