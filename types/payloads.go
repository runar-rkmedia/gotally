package types

import (
	"errors"
	"fmt"
	"time"

	"github.com/runar-rkmedia/gotally/tallylogic/cell"
)

type GetUserPayload struct {
	ID string
}
type CreateUserSessionPayload struct {
	InvalidAfter time.Time
	SessionID    string
	Username     string
	Game         Game
}

func (p CreateUserSessionPayload) Validate() error {
	if p.Game.ID == "" {
		return fmt.Errorf("%w: Game.ID", ErrArgumentMissing)
	}
	if p.Username == "" {
		return fmt.Errorf("%w: Username", ErrArgumentMissing)
	}
	if p.SessionID == "" {
		return fmt.Errorf("%w: SessionID", ErrArgumentMissing)
	}
	if p.InvalidAfter.Before(time.Now()) {
		return fmt.Errorf("%w: InvalidAfter cannot be in the past", ErrArgumentInvalid)
	}
	return nil
}

type SwipePayload struct {
	GameID         string
	SwipeDirection SwipeDirection
	// Index for this move.
	Moves int
	// State of any randomizer
	State uint64
	Cells []cell.Cell
}
type CombinePathPayload struct {
	GameID string
	// Index for this move.
	Moves int
	// Points achieved for this move
	Points int
	// Score total in this game
	Score uint64
	// State of any randomizer
	State uint64
	Cells []cell.Cell
}
type NewGamePayload struct {
	Game Game
}

var (
	ErrArgumentMissing = errors.New("missing argument")
	ErrArgumentInvalid = errors.New("invalid argument")
)

type SwipeDirection string

const (
	SwipeDirectionUp    SwipeDirection = "Up"
	SwipeDirectionRight SwipeDirection = "Right"
	SwipeDirectionDown  SwipeDirection = "Down"
	SwipeDirectionLeft  SwipeDirection = "Left"
)
