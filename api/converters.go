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
func fromModalCells(cells []*model.Cell) []cell.Cell {
	c := make([]cell.Cell, len(cells))
	for i := 0; i < len(cells); i++ {
		c[i] = cell.NewCell(cells[i].Base, int(cells[i].Twopow))
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
		History:   Game.History.Bytes(),
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
			Mode:            toTypeMode(Game.Rules.GameMode),
		},
	}
	return g
}

func toTypeMode(mode logic.GameMode) types.RuleMode {
	switch mode {
	case tallylogic.GameModeRandom:
		return types.RuleModeInfiniteNormal
	case tallylogic.GameModeRandomChallenge:
		return types.RuleModeChallenge
	case tallylogic.GameModeTutorial:
		return types.RuleModeTutorial
	}
	return ""
}
func toModelInstruction(instructions logic.CompactHistory) ([]*model.Instruction, error) {
	all, err := instructions.FilterForUndo(false)
	if err != nil {
		return nil, fmt.Errorf("failed to map instructions: %w", err)
	}
	ins := make([]*model.Instruction, len(all))
	for i := 0; i < len(ins); i++ {
		h := all[i]
		switch {
		case h.IsSwipe:
			ins[i] = &model.Instruction{
				InstructionOneof: &model.Instruction_Swipe{
					Swipe: toModalDirection(h.Direction),
				},
			}
		case h.IsPath:
			ins[i] = &model.Instruction{
				InstructionOneof: &model.Instruction_Combine{
					Combine: &model.Indexes{
						Index: intsTouInt32s(h.Path),
					},
				},
			}
		default:
			return ins, fmt.Errorf("failed to resolve instruction %s %#v", instructions.Describe(), instructions)
		}
	}
	return ins, nil
}
func toModelHint(hints map[string]tallylogic.Hint) []*model.Instruction {
	ins := make([]*model.Instruction, len(hints))
	i := -1
	for _, h := range hints {
		i++
		ins[i] = &model.Instruction{
			InstructionOneof: &model.Instruction_Combine{
				Combine: &model.Indexes{Index: intsTouInt32s(h.Path)},
			},
		}
	}
	return ins
}
