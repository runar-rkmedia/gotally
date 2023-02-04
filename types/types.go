package types

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"runtime/debug"
	"time"

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
}

type Game struct {
	ID          string
	CreatedAt   time.Time
	UpdatedAt   *time.Time
	UserID      string
	Description string
	Name        string
	Seed, State uint64
	Score       uint64
	Moves       uint
	Cells       []cell.Cell
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
		return fmt.Errorf("%w: Game.Rules.Mode %s", ErrArgumentMissing, debug.Stack())
	}

	return nil

}

type Rules struct {
	ID              string
	CreatedAt       time.Time
	UpdatedAt       *time.Time
	Description     string
	Mode            RuleMode
	TargetCellValue uint64
	TargetScore     uint64
	MaxMoves        uint64
	Rows            uint8
	Columns         uint8
	RecreateOnSwipe bool
	NoReSwipe       bool
	NoMultiply      bool
	NoAddition      bool
}

func (r Rules) Hash() string {
	h := sha256.New()
	h.Write([]byte(r.Description))
	h.Write([]byte(r.Mode))
	h.Write([]byte{r.Rows, r.Columns})
	h.Write(boolsToBytes(r.RecreateOnSwipe, r.NoReSwipe, r.NoMultiply, r.NoAddition))
	b := h.Sum(nil)
	return base64.URLEncoding.EncodeToString(b)
}

func boolsToBytes(t ...bool) []byte {
	b := make([]byte, (len(t)+7)/8)
	for i, x := range t {
		if x {
			b[i/8] |= 0x80 >> uint(i%8)
		}
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
	// Size of the data-column in the history (counting only combines), represented as a standard-deviation
	CombineDataStdDev float64
	// Size of the data-column in the history (counting only combines), represented as average
	CombineDataAvg float64
	// Size of the data-column in the history (counting only combines), represented as max
	CombineDataMax uint64
	// Size of the data-column in the history (counting only combines), represented as min
	CombineDataMin uint64
	// Size of the data-column in the history (counting only combines), represented as total
	CombineDataTotal uint64
}
