package api

import (
	"fmt"
	"time"

	model "github.com/runar-rkmedia/gotally/gen/proto/tally/v1"
	"github.com/runar-rkmedia/gotally/tallylogic"
	logic "github.com/runar-rkmedia/gotally/tallylogic"
	"github.com/runar-rkmedia/gotally/tallylogic/cell"
	"github.com/runar-rkmedia/gotally/types"
)

func toGameSwipeDirection(dir model.SwipeDirection) logic.SwipeDirection {
	switch dir {
	case model.SwipeDirection_SWIPE_DIRECTION_UP:
		return logic.SwipeDirectionUp
	case model.SwipeDirection_SWIPE_DIRECTION_RIGHT:
		return logic.SwipeDirectionRight
	case model.SwipeDirection_SWIPE_DIRECTION_DOWN:
		return logic.SwipeDirectionDown
	case model.SwipeDirection_SWIPE_DIRECTION_LEFT:
		return logic.SwipeDirectionLeft
	}
	return ""
}
func toModalDirection(dir logic.SwipeDirection) model.SwipeDirection {
	switch dir {
	case logic.SwipeDirectionUp:
		return model.SwipeDirection_SWIPE_DIRECTION_UP
	case logic.SwipeDirectionRight:
		return model.SwipeDirection_SWIPE_DIRECTION_RIGHT
	case logic.SwipeDirectionDown:
		return model.SwipeDirection_SWIPE_DIRECTION_DOWN
	case logic.SwipeDirectionLeft:
		return model.SwipeDirection_SWIPE_DIRECTION_LEFT
	}
	return model.SwipeDirection_SWIPE_DIRECTION_UNSPECIFIED
}

func toModalBoard(game *logic.Game) *model.Board {
	return &model.Board{
		Id:      game.ID,
		Cells:   toModalCells(game.Cells()),
		Columns: uint32(game.Rules.SizeX),
		Rows:    uint32(game.Rules.SizeX),
		Name:    game.Name,
	}
}

func toModalCells(cells []cell.Cell) []*model.Cell {
	c := make([]*model.Cell, len(cells))
	for i := 0; i < len(cells); i++ {
		base, twopow := cells[i].Raw()
		c[i] = &model.Cell{
			Base:   base,
			Twopow: twopow,
		}

	}
	return c
}

func toModelGameMode(mode tallylogic.GameMode) model.GameMode {
	switch mode {
	case tallylogic.GameModeRandomChallenge:
		return model.GameMode_GAME_MODE_RANDOM_CHALLENGE
	case tallylogic.GameModeRandom:
		return model.GameMode_GAME_MODE_RANDOM
	case tallylogic.GameModeTutorial:
		return model.GameMode_GAME_MODE_TUTORIAL
	}
	panic(fmt.Sprintf("Invalid game-mode %d", mode))
}
func toTypeGame(Game tallylogic.Game, userId string) types.Game {

	seed, state := Game.Seed()
	g := types.Game{
		ID:          Game.ID,
		CreatedAt:   time.Now(),
		UserID:      userId,
		Description: Game.Description,
		// The templates can have names, so why not in the database?
		Name:      Game.Name,
		Seed:      seed,
		State:     state,
		Score:     uint64(Game.Score()),
		Moves:     uint(Game.Moves()),
		Cells:     Game.Cells(),
		PlayState: types.PlayStateCurrent,
		Rules: types.Rules{
			ID:              Game.Rules.ID,
			CreatedAt:       time.Now(),
			Description:     "",
			Rows:            uint8(Game.Rules.SizeY),
			Columns:         uint8(Game.Rules.SizeX),
			RecreateOnSwipe: Game.Rules.RecreateOnSwipe,
			NoReSwipe:       Game.Rules.NoReswipe,
			NoMultiply:      Game.Rules.Options.NoMultiply,
			NoAddition:      Game.Rules.Options.NoAddition,
			TargetCellValue: Game.Rules.TargetCellValue,
			TargetScore:     Game.Rules.TargetScore,
			MaxMoves:        Game.Rules.MaxMoves,
		},
	}
	switch Game.Rules.GameMode {
	case tallylogic.GameModeRandom:
		g.Rules.Mode = types.RuleModeInfiniteNormal
	case tallylogic.GameModeRandomChallenge:
		g.Rules.Mode = types.RuleModeChallenge
	case tallylogic.GameModeTutorial:
		g.Rules.Mode = types.RuleModeTutorial
	}
	return g
}
