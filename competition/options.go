package competition

import "github.com/d-protocol/pokerlib"

type Options struct {
	GameType              string        `json:"game_type"`
	MaxTables             int           `json:"max_tables"`
	TableAllocationPeriod int           `json:"table_allocation_period"`
	Table                 *TableOptions `json:"table"`
}

type TableOptions struct {
	InitialPlayers int                   `json:"initial_players"`
	MinPlayers     int                   `json:"min_players"`
	MaxSeats       int                   `json:"max_seats"`
	Duration       int                   `json:"duration"`
	Interval       int                   `json:"interval"`
	ActionTime     int                   `json:"action_time"`
	Ante           int64                 `json:"ante"`
	Blind          pokerlib.BlindSetting `json:"blind"`
}

func NewOptions() *Options {
	return &Options{
		GameType:              "standard",
		MaxTables:             1,  // -1 or greater than 1 for dynamic table allocation
		TableAllocationPeriod: 10, // 10 seconds
		Table: &TableOptions{
			InitialPlayers: 2,
			MinPlayers:     2,
			MaxSeats:       9,
			Duration:       60 * 60, // one hour
			Interval:       0,       // 0 secs by default
			ActionTime:     10,      // 10 secs
			Ante:           0,
			Blind: pokerlib.BlindSetting{
				Dealer: 0,
				SB:     5,
				BB:     10,
			},
		},
	}
}
