package api

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	model "github.com/runar-rkmedia/gotally/gen/proto/tally/v1"
	"github.com/runar-rkmedia/gotally/randomizer"
	"github.com/runar-rkmedia/gotally/tallylogic/gameGeneratorTargetCell"
)

func (s *TallyServer) GenerateGame(
	ctx context.Context,
	req *connect.Request[model.GenerateGameRequest],
) (*connect.Response[model.GenerateGameResponse], error) {
	// session := ContextGetUserState(ctx)
	// TODO: Check that user is registered /admin etc.
	options := gamegenerator_target_cell.GameGeneratorTargetCellOptions{
		TargetCell:         req.Msg.TargetCellValue,
		MinCellValue:       0,
		MaxCellValue:       12,
		RandomCellChance:   -1,
		MaxCells:           int(req.Msg.MaxBricks),
		MaxAdditionalCells: int(req.Msg.MaxAdditionalCells),
		Rows:               int(req.Msg.Rows),
		Columns:            int(req.Msg.Columns),
		MaxMoves:           int(req.Msg.MaxMoves),
		MinMoves:           int(req.Msg.MinMoves),
		// Seed:               req.Msg.Seed,
		Randomizer: randomizer.NewRandomizerFromSeed(req.Msg.Seed, req.Msg.Salt),
	}
	generator, err := gamegenerator_target_cell.NewGameGeneratorForTargetCell(options)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to initialize game-generator: %w", err))
	}
	game, solutions, err := generator.GenerateGame()
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to generate game: %w", err))
	}
	var ideal int
	var score int
	var maxScore int64
	for _, s := range solutions {
		if s.Moves() < ideal {
			ideal = s.Moves()
			score = int(s.Score())
		}
		if s.Score() > maxScore {
			maxScore = s.Score()
		}
	}
	response := &model.GenerateGameResponse{
		Game: &model.Game{
			Board:       toModalBoard(&game),
			Score:       game.Score(),
			Moves:       int64(game.Moves()),
			Description: game.Description,
			Mode:        toModelGameMode(game.Rules.GameMode),
		},
		IdealMoves:   uint32(ideal),
		IdealScore:   uint64(score),
		HighestScore: uint64(maxScore),
	}
	if req.Msg.WithSolutions {
		max := len(solutions)
		response.Solutions = make([]*model.Game, max)
		for i, s := range solutions {
			if i >= max {
				break
			}
			response.Solutions[i] = &model.Game{
				Board:       toModalBoard(&s),
				Score:       s.Score(),
				Moves:       int64(s.Moves()),
				Description: s.Description,
				Mode:        toModelGameMode(s.Rules.GameMode),
			}
		}
	}
	res := connect.NewResponse(response)
	return res, nil
}
