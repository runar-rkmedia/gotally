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
	UserID       string
	Username     string
	Game         Game
}

func (p CreateUserSessionPayload) Validate() error {
	if p.Game.ID == "" {
		return fmt.Errorf("%w: Game.ID", ErrArgumentMissing)
	}
	if p.UserID == "" {
		return fmt.Errorf("%w: User.ID", ErrArgumentMissing)
	}
	if p.Username == "" {
		return fmt.Errorf("%w: Username", ErrArgumentMissing)
	}
	if p.SessionID == "" {
		return fmt.Errorf("%w: SessionID", ErrArgumentMissing)
	}
	if p.Game.Rules.ID == "" {
		return fmt.Errorf("%w: Game.Rules.ID", ErrArgumentMissing)
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
	//Seed of randomizer
	Seed  uint64
	Cells []cell.Cell
}

func (p SwipePayload) Validate() error {

	if p.GameID == "" {
		return fmt.Errorf("%w: GameId", ErrArgumentMissing)
	}
	if p.Moves <= 0 {
		return fmt.Errorf("%w: Moves", ErrArgumentMissing)
	}
	if p.State <= 0 {
		return fmt.Errorf("%w: Seed", ErrArgumentMissing)
	}
	if p.Seed <= 0 {
		return fmt.Errorf("%w: Seed", ErrArgumentMissing)
	}
	if p.SwipeDirection == "" {
		return fmt.Errorf("%w: SwipeDirection", ErrArgumentMissing)
	}
	if len(p.Cells) == 0 {
		return fmt.Errorf("%w: Cells", ErrArgumentMissing)
	}
	return nil
}

type CombinePathPayload struct {
	GameID string
	// Index for this move.
	Moves int
	// Points achieved for this move
	Points int
	// Score total in this game
	Score uint64
	// State of randomizer
	State uint64
	//Seed of randomizer
	Seed  uint64
	Path  []uint32
	Cells []cell.Cell
}

func (payload CombinePathPayload) Validate() error {
	if payload.GameID == "" {
		return fmt.Errorf("%w: GameId", ErrArgumentMissing)
	}
	if payload.Moves <= 0 {
		return fmt.Errorf("%w: Moves", ErrArgumentMissing)
	}
	if payload.Points == 0 {
		return fmt.Errorf("%w: Points", ErrArgumentMissing)
	}
	if payload.Score == 0 {
		return fmt.Errorf("%w: Score", ErrArgumentMissing)
	}
	if payload.State == 0 {
		return fmt.Errorf("%w: State", ErrArgumentMissing)
	}
	if payload.Seed == 0 {
		return fmt.Errorf("%w: Seed", ErrArgumentMissing)
	}
	if len(payload.Cells) == 0 {
		return fmt.Errorf("%w: Cells", ErrArgumentMissing)
	}
	if len(payload.Path) == 0 {
		return fmt.Errorf("%w: Path", ErrArgumentMissing)
	}
	return nil
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
