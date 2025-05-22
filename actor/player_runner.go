package actor

import (
	"time"

	"github.com/d-protocol/pokerlib"
	"github.com/d-protocol/pokertable"
	"github.com/d-protocol/timebank"
)

type PlayerStatus int32

const (
	PlayerStatus_Running PlayerStatus = iota
	PlayerStatus_Idle
	PlayerStatus_Suspend
)

type PlayerRunner struct {
	actor               Actor
	actions             Actions
	playerID            string
	curGameID           string
	lastGameStateTime   int64
	tableInfo           *pokertable.Table
	timebank            *timebank.TimeBank
	onTableStateUpdated func(*pokertable.Table)

	// status
	status           PlayerStatus
	idleCount        int
	suspendThreshold int
}

func NewPlayerRunner(playerID string) *PlayerRunner {
	return &PlayerRunner{
		playerID:            playerID,
		timebank:            timebank.NewTimeBank(),
		status:              PlayerStatus_Running,
		suspendThreshold:    2,
		onTableStateUpdated: func(*pokertable.Table) {},
	}
}

func (pr *PlayerRunner) SetActor(a Actor) {
	pr.actor = a
	pr.actions = NewActions(a, pr.playerID)
}

func (pr *PlayerRunner) UpdateTableState(table *pokertable.Table) error {

	gs := table.State.GameState
	pr.tableInfo = table

	// The state remains unchanged or is outdated
	if gs != nil {

		// New game
		if gs.GameID != pr.curGameID {
			pr.curGameID = gs.GameID
		}

		//fmt.Println(br.lastGameStateTime, br.tableInfo.State.GameState.UpdatedAt)
		if pr.lastGameStateTime >= gs.UpdatedAt {
			//fmt.Println(br.playerID, table.ID)
			return nil
		}

		pr.lastGameStateTime = gs.UpdatedAt
	}

	// Check if you have been eliminated
	isEliminated := true
	for _, ps := range table.State.PlayerStates {
		if ps.PlayerID == pr.playerID {
			isEliminated = false
		}
	}

	if isEliminated {
		return nil
	}

	// Update seat index
	gamePlayerIdx := table.GamePlayerIndex(pr.playerID)

	// Emit event
	pr.onTableStateUpdated(table)

	// Game is running right now
	switch table.State.Status {
	case pokertable.TableStateStatus_TableGamePlaying:

		// Somehow, this player is not in the game.
		// It probably has no chips already.
		if gamePlayerIdx == -1 {
			return nil
		}

		// Filtering private information fpr player
		gs.AsPlayer(gamePlayerIdx)

		// We have actions allowed by game engine
		player := gs.GetPlayer(gamePlayerIdx)
		if len(player.AllowedActions) > 0 {
			pr.requestMove(gs, gamePlayerIdx)
		}
	}

	return nil
}

func (pr *PlayerRunner) OnTableStateUpdated(fn func(*pokertable.Table)) error {
	pr.onTableStateUpdated = fn
	return nil
}

func (pr *PlayerRunner) requestMove(gs *pokerlib.GameState, playerIdx int) error {

	// Do pass automatically
	if gs.HasAction(playerIdx, "pass") {
		return pr.actions.Pass()
	}

	// Player is suspended
	if pr.status == PlayerStatus_Suspend {
		return pr.automate(gs, playerIdx)
	}

	// Setup timebank to wait for player
	thinkingTime := time.Duration(pr.tableInfo.Meta.ActionTime) * time.Second
	return pr.timebank.NewTask(thinkingTime, func(isCancelled bool) {

		if isCancelled {
			return
		}

		// Stay idle already
		if pr.status == PlayerStatus_Idle {
			pr.Idle()
		}

		// Do default actions if player has no response
		pr.automate(gs, playerIdx)
	})
}

func (pr *PlayerRunner) automate(gs *pokerlib.GameState, playerIdx int) error {

	// Default actions for automation when player has no response
	if gs.HasAction(playerIdx, "ready") {
		return pr.actions.Ready()
	} else if gs.HasAction(playerIdx, "check") {
		return pr.actions.Check()
	} else if gs.HasAction(playerIdx, "fold") {
		return pr.actions.Fold()
	}

	// Pay for ante and blinds
	switch gs.Status.CurrentEvent {
	case pokerlib.GameEventSymbols[pokerlib.GameEvent_AnteRequested]:

		// Ante
		return pr.actions.Pay(gs.Meta.Ante)

	case pokerlib.GameEventSymbols[pokerlib.GameEvent_BlindsRequested]:

		// blinds
		if gs.HasPosition(playerIdx, "sb") {
			return pr.actions.Pay(gs.Meta.Blind.SB)
		} else if gs.HasPosition(playerIdx, "bb") {
			return pr.actions.Pay(gs.Meta.Blind.BB)
		}

		return pr.actions.Pay(gs.Meta.Blind.Dealer)
	}

	return nil
}

func (pr *PlayerRunner) SetSuspendThreshold(count int) {
	pr.suspendThreshold = count
}

func (pr *PlayerRunner) Resume() error {

	if pr.status == PlayerStatus_Running {
		return nil
	}

	pr.status = PlayerStatus_Running
	pr.idleCount = 0

	return nil
}

func (pr *PlayerRunner) Idle() error {
	if pr.status != PlayerStatus_Idle {
		pr.status = PlayerStatus_Idle
		pr.idleCount = 0
	} else {
		pr.idleCount++
	}

	if pr.idleCount == pr.suspendThreshold {
		return pr.Suspend()
	}

	return nil
}

func (pr *PlayerRunner) Suspend() error {
	pr.status = PlayerStatus_Suspend
	return nil
}

func (pr *PlayerRunner) Pass() error {

	err := pr.actions.Pass()
	if err != nil {
		return err
	}

	pr.idleCount = 0

	return nil
}

func (pr *PlayerRunner) Ready() error {

	err := pr.actions.Ready()
	if err != nil {
		return err
	}

	pr.idleCount = 0

	return nil
}

func (pr *PlayerRunner) Pay(chips int64) error {

	err := pr.actions.Pay(chips)
	if err != nil {
		return err
	}

	pr.idleCount = 0

	return nil
}

func (pr *PlayerRunner) Check() error {

	err := pr.actions.Check()
	if err != nil {
		return err
	}

	pr.idleCount = 0

	return nil
}

func (pr *PlayerRunner) Bet(chips int64) error {

	err := pr.actions.Bet(chips)
	if err != nil {
		return err
	}

	pr.idleCount = 0

	return nil
}

func (pr *PlayerRunner) Call() error {

	err := pr.actions.Call()
	if err != nil {
		return err
	}

	pr.idleCount = 0

	return nil
}

func (pr *PlayerRunner) Fold() error {
	pr.Resume()
	return pr.actions.Fold()
}

func (pr *PlayerRunner) Allin() error {

	err := pr.actions.Allin()
	if err != nil {
		return err
	}

	pr.idleCount = 0

	return nil
}

func (pr *PlayerRunner) Raise(chipLevel int64) error {

	err := pr.actions.Raise(chipLevel)
	if err != nil {
		return err
	}

	pr.idleCount = 0

	return nil
}
