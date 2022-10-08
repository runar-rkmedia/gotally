package types

import (
	"crypto/sha256"
	"encoding/base64"
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
	UpdatedAt *time.Time

	UserID       string
	InvalidAfter time.Time
	ActiveGame   *Game
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

type Game struct {
	ID          string
	CreatedAt   time.Time
	UpdatedAt   *time.Time
	UserID      string
	Seed, State uint64
	Score       uint64
	Moves       uint
	Cells       []cell.Cell
	PlayState
	Rules
}
type Rules struct {
	ID              string
	CreatedAt       time.Time
	UpdatedAt       *time.Time
	Description     string
	Mode            RuleMode
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
