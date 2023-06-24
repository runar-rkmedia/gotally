package types

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/runar-rkmedia/gotally/dev"
	"github.com/runar-rkmedia/gotally/tallylogic/cell"
)

type Vote struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt *time.Time

	UserID  string
	FunVote int
}
type Session struct {
	ID        string
	CreatedAt time.Time

	UserID       string
	InvalidAfter time.Time
	// Deprecated: TBD. May be moved to Game
	// TODO: move this to the user to keep things logical
	ActiveGame *Game
}

type GameTemplate struct {
	ID              string
	CreatedAt       time.Time
	UpdatedAt       *time.Time
	ChallengeNumber *int
	IdealMoves      *int
	CreatedByID     string
	UpdatedBy       string
	Description     string
	Name            string
	Cells           []cell.Cell
	Rules
	Stats []PlayStats
}

type PlayStats struct {
	GameID   string
	UserID   string
	Username string
	Score    uint64
	Moves    uint64
}
type SessionUser struct {
	Session
	User
}

type User struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt *time.Time

	UserName string
}

type Dump struct {
	Games         any //[]Game
	GameHistories any //[]any
	Rules         any //[]Rules
	Sessions      any //[]Session
	Users         any //[]User
	Template      any //[]GameTemplate
}

type Game struct {
	ID                      string
	CreatedAt               time.Time
	UpdatedAt               *time.Time
	UserID                  string
	Description             string
	Name                    string
	Seed, State             uint64
	OptionSeed, OptionState uint64
	Score                   uint64
	Moves                   uint
	Cells                   []cell.Cell
	History                 []byte
	PlayState
	Rules
}

func (p Game) Validate() error {
	if p.ID == "" {
		return fmt.Errorf("%w: Game.ID", ErrArgumentMissing)
	}
	if p.UserID == "" {
		return fmt.Errorf("%w: Game.UserID", ErrArgumentMissing)
	}
	if p.Seed == 0 {
		return fmt.Errorf("%w: Game.Seed", ErrArgumentMissing)
	}
	if p.State == 0 {
		return fmt.Errorf("%w: Game.State", ErrArgumentMissing)
	}
	if len(p.Cells) == 0 {
		return fmt.Errorf("%w: Game.Cells", ErrArgumentMissing)
	}
	if len(p.Cells) != int(p.Rules.Rows)*int(p.Rules.Columns) {
		return fmt.Errorf("%w: Game.Cells should have matching lenght for board was %d for %dx%d board", ErrArgumentInvalid, len(p.Cells), p.Rules.Rows, p.Rules.Columns)
	}
	if p.Rules.Mode == "" {
		return fmt.Errorf("%w: Game.Rules.Mode %s", ErrArgumentMissing, dev.Stack())
	}
	if p.Moves > 0 && p.History == nil {
		return fmt.Errorf("%w: Game.Rules.History %s", ErrArgumentMissing, dev.Stack())
	}
	switch p.Mode {
	case RuleModeInfiniteEasy, RuleModeInfiniteNormal, RuleModeInfiniteHard:
		if p.OptionSeed == 0 {
			return fmt.Errorf("%w: Game.OptionSeed is required in mode %s \n%s", ErrArgumentMissing, p.Mode, dev.Stack())
		}
		if p.OptionState == 0 {
			return fmt.Errorf("%w: Game.OptionState is required in mode %s %s", ErrArgumentMissing, p.Mode, dev.Stack())
		}
	}

	err := p.Rules.Validate()
	if err != nil {
		return err
	}

	return nil

}

type Rules struct {
	ID          string
	CreatedAt   time.Time
	UpdatedAt   *time.Time
	Description string
	Mode        RuleMode
	// The player wins if a cell reaches this value. 0 means there is not limit.
	TargetCellValue uint64
	// The player wins if the score reaches this value. 0 means there is not limit.
	TargetScore uint64
	// The player loses if there are more moves than this. 0 means there is no limit
	MaxMoves uint64
	// The number of rows for the board
	Rows uint8
	// The number of columns for the board
	Columns uint8
	// How many cells that are generated at the start of the game.
	StartingCells uint8
	// If set, every time the player swipes, there will be a new cell.
	RecreateOnSwipe bool
	NoReSwipe       bool
	NoMultiply      bool
	NoAddition      bool
}

func (r Rules) Validate() error {
	switch r.Mode {
	case RuleModeInfiniteEasy, RuleModeInfiniteNormal, RuleModeInfiniteHard:
		if r.StartingCells == 0 {
			return fmt.Errorf("%w: Rules.StartingCells required for mode %s %s", ErrArgumentMissing, r.Mode, dev.Stack())
		}
	case RuleModeChallenge:
		if r.TargetCellValue == 0 {
			return fmt.Errorf("%w: Rules.TargetCellValue required for mode %s %s", ErrArgumentMissing, r.Mode, dev.Stack())
		}
	case RuleModeTutorial:
	default:
		return fmt.Errorf("%w: Rules.Mode %s", ErrArgumentInvalid, dev.Stack())
	}
	return nil

}
func (r Rules) Hash() string {
	h := sha256.New()
	if r.Description != "" {
		h.Write([]byte(r.Description))
	} else {
		h.Write([]byte{0x01})
	}
	if r.Mode != "" {
		h.Write([]byte(r.Mode))
	} else {
		h.Write([]byte{0x02})
	}
	h.Write([]byte{r.Rows, r.Columns, r.StartingCells})
	{
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(r.TargetCellValue))
		h.Write(b)
	}
	{
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(r.TargetScore))
		h.Write(b)
	}
	{
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(r.MaxMoves))
		h.Write(b)
	}
	h.Write(boolsToBytes(false, r.RecreateOnSwipe, r.NoReSwipe, r.NoMultiply, r.NoAddition))
	b := h.Sum(nil)
	return base64.URLEncoding.EncodeToString(b)
}

func boolsToBytes(debug bool, t ...bool) []byte {
	b := make([]byte, (len(t)+7)/8)
	for i, x := range t {
		if x {
			b[i/8] |= 0x80 >> uint(i%8)
		}
	}
	if debug {
		fmt.Println("bools", t, b)
	}
	return b
}

type PlayState = string
type RuleMode = string

const (
	PlayStateWon       PlayState = "won"
	PlayStateLost      PlayState = "lost"
	PlayStateAbandoned PlayState = "abandoned"
	PlayStateCurrent   PlayState = "current"

	RuleModeInfiniteEasy   RuleMode = "infinite-easy"
	RuleModeInfiniteNormal RuleMode = "infinite-normal"
	RuleModeInfiniteHard   RuleMode = "infinite-hard"
	RuleModeChallenge      RuleMode = "challenge"
	RuleModeTutorial       RuleMode = "tutorial"
)

type Statistics struct {
	// Totaly number of users
	Users int64
	// Totaly number of sessions
	Session int64
	// Totaly number of games
	Games int64
	// Totaly number of games marked as won
	GamesWon int64
	// Totaly number of games marked as lost
	GamesLost int64
	// Totaly number of games marked as abandoned
	GamesAbandoned int64
	// Totaly number of games marked as current
	GamesCurrent int64
	// Most moves recorded for a game
	LongestGame uint64
	// Highest score recorded for a game
	HighestScore uint64
	// Size of the history-column in the history, represented as a standard-deviation
	HistoryStdDev float64
	// Size of the history-column in the history, represented as average
	HistoryAvg float64
	// Size of the history-column in the history, represented as max
	HistoryMax uint64
	// Size of the history-column in the history, represented as min
	HistoryMin uint64
	// Size of the history-column in the history, represented as total
	HistoryTotal uint64
}
