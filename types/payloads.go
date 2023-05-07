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
type GetOriginalGamePayload struct {
	GameID string
}

type CreateUserSessionPayload struct {
	InvalidAfter time.Time
	SessionID    string
	UserID       string
	Username     string
	Game         Game
	TemplateID   string
}

func (p GetOriginalGamePayload) Validate() error {
	if p.GameID == "" {
		return fmt.Errorf("%w: Game.ID", ErrArgumentMissing)
	}
	return nil
}
func (p GetUserPayload) Validate() error {
	if p.ID == "" {
		return fmt.Errorf("%w: Game.ID", ErrArgumentMissing)
	}
	return nil
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
	if p.Game.Rules.Mode == RuleModeChallenge && p.Game.Rules.TargetCellValue == 0 {
		return fmt.Errorf("%w: Game.Rules.TargetCellValue for Game.Rules.Mode=%v (%#v)", ErrArgumentMissing, p.Game.Rules.Mode, p.Game.Rules)
	}
	if p.InvalidAfter.Before(time.Now()) {
		return fmt.Errorf("%w: InvalidAfter cannot be in the past", ErrArgumentInvalid)
	}

	return nil
}

type GetGameChallengePayload struct {
	// Optionally include stats for a user, from previous games
	StatsForUserID string
}
type CreateGameTemplatePayload struct {
	ID              string
	CreatedAt       time.Time
	UpdatedAt       *time.Time
	CreatedByID     string
	UpdatedBy       string
	Description     string
	ChallengeNumber *int
	IdealMoves      int
	IdealScore      int
	Name            string
	Cells           []cell.Cell
	Rules
}

func (p CreateGameTemplatePayload) Validate() error {
	if p.ID == "" {
		return fmt.Errorf("%w: Game.ID", ErrArgumentMissing)
	}
	if p.CreatedByID == "" {
		return fmt.Errorf("%w: CreatedByID", ErrArgumentMissing)
	}
	if p.Name == "" {
		return fmt.Errorf("%w: Name", ErrArgumentMissing)
	}
	if p.Rules.Mode == RuleModeChallenge && p.Rules.TargetCellValue == 0 {
		return fmt.Errorf("%w: Game.Rules.TargetCellValue for Game.Rules.Mode=%v (%#v)", ErrArgumentMissing, p.Rules.Mode, p.Rules)
	}
	if p.IdealMoves == 0 && p.IdealScore == 0 {
		return fmt.Errorf("%w: IdealMoves or IdealScore must be set", ErrArgumentMissing)
	}
	expectedCellCount := p.Rows * p.Columns
	if len(p.Cells) != int(expectedCellCount) {
		return fmt.Errorf("%w: Number of cells must match Rows*Columns", ErrArgumentInvalid)
	}
	if p.TargetCellValue == 0 {
		return fmt.Errorf("%w: Target cell value must be set", ErrArgumentInvalid)
	}

	return nil
}

type UpdateGamePayload struct {
	GameID string
	// Index for this move.
	Moves int
	// Score total in this game
	Score uint64
	// State of randomizer
	State uint64
	//Seed of randomizer
	Seed    uint64
	Cells   []cell.Cell
	History []byte
	PlayState
}

func (payload UpdateGamePayload) Validate() error {
	if payload.GameID == "" {
		return fmt.Errorf("%w: GameId", ErrArgumentMissing)
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
	if payload.Moves > 0 && len(payload.History) == 0 {
		return fmt.Errorf("%w: History", ErrArgumentMissing)
	}
	return nil
}

type NewGamePayload struct {
	Game       Game
	TemplateID string
}
type RestartGamePayload struct {
	UserID string
	GameID string
}

func (payload RestartGamePayload) Validate() error {
	if payload.GameID == "" {
		return fmt.Errorf("%w: GameId", ErrArgumentMissing)
	}
	if payload.UserID == "" {
		return fmt.Errorf("%w: UserID", ErrArgumentMissing)
	}

	return nil
}

var (
	ErrArgumentMissing = errors.New("missing argument")
	ErrArgumentInvalid = errors.New("invalid argument")
)
