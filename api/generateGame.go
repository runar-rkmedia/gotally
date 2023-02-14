package api

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	model "github.com/runar-rkmedia/gotally/gen/proto/tally/v1"
	"github.com/runar-rkmedia/gotally/tallylogic"
)

func (s *TallyServer) GenerateGame(
	ctx context.Context,
	req *connect.Request[model.GenerateGameRequest],
) (*connect.Response[model.GenerateGameResponse], error) {
	// session := ContextGetUserState(ctx)
	// TODO: Check that user is registered /admin etc.
	options := tallylogic.GameGeneratorOptions{
		Rows:                int(req.Msg.Rows),
		Columns:             int(req.Msg.Columns),
		GoalChecker:         nil,
		TargetCellValue:     uint64(req.Msg.TargetCellValue),
		MaxBricks:           int(req.Msg.MaxBricks),
		MinBricks:           int(req.Msg.MinBricks),
		MinMoves:            int(req.Msg.MinMoves),
		MaxMoves:            int(req.Msg.MaxMoves),
		MaxIterations:       int(req.Msg.MaxIterations),
		CellGenerator:       nil,
		Randomizer:          nil,
		MinGames:            1,
		GameSolutionChannel: make(chan tallylogic.SolvableGame, 8),
	}
	if options.TargetCellValue != 0 {
		options.GoalChecker = tallylogic.GoalCheckLargestCell{
			GoalCheck:       tallylogic.GoalCheck{},
			TargetCellValue: options.TargetCellValue,
		}
	}
	fmt.Println("\n\nboobobob")
	fmt.Println(devpretty(req.Msg))
	fmt.Printf("bobo %#v", options)

	generator, err := tallylogic.NewGameGenerator(options)
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
