package api

import (
	"context"
	"fmt"
	"time"

	"github.com/bufbuild/connect-go"
	model "github.com/runar-rkmedia/gotally/gen/proto/tally/v1"
	"github.com/runar-rkmedia/gotally/randomizer"
	"github.com/runar-rkmedia/gotally/tallylogic"
	gamegenerator_target_cell "github.com/runar-rkmedia/gotally/tallylogic/gameGeneratorTargetCell"
	"github.com/runar-rkmedia/gotally/tallylogic/gamestats"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
)

type gamegenerator interface {
	GenerateGame(ctx context.Context) (tallylogic.Game, []tallylogic.Game, error)
}

func (s *TallyServer) GenerateGame(
	ctx context.Context,
	req *connect.Request[model.GenerateGameRequest],
) (*connect.Response[model.GenerateGameResponse], error) {
	if !s.FeatureGameGeneration {
		err := connect.NewError(connect.CodeResourceExhausted, fmt.Errorf("generating games has been disabled"))
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()
	// session := ContextGetUserState(ctx)
	// TODO: Check that user is registered /admin etc.
	var generator gamegenerator
	var err error
	if req.Msg.Rows != req.Msg.Columns {

		err := connect.NewError(connect.CodeUnimplemented, fmt.Errorf("Rows must be equal to Columns. See issue #21"))
		detail := errdetails.ErrorInfo{
			Reason: "An unresolved issue with underlying implementation",
			Domain: "generator",
			Metadata: map[string]string{
				"issue-url": "https://github.com/runar-rkmedia/gotally/issues/21",
			},
		}
		if detail, detailErr := connect.NewErrorDetail(&detail); detailErr == nil {
			err.AddDetail(detail)
		}
		return nil, err
	}
	if req.Msg.Algorithm == model.GeneratorAlgorithm_GENERATOR_ALGORITHM_REVERSE {
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
		generator, err = gamegenerator_target_cell.NewGameGeneratorForTargetCell(options)
	} else {
		gameCh := make(chan tallylogic.SolvableGame)
		go func() {
			for {
				select {
				case sg := <-gameCh:
					fmt.Println("got a game", sg.Game.Print())
					return
				}
			}
		}()
		legacy_options := tallylogic.GameGeneratorOptions{
			Rows:    int(req.Msg.Rows),
			Columns: int(req.Msg.Columns),
			GoalChecker: tallylogic.GoalCheckLargestCell{
				GoalCheck:       tallylogic.GoalCheck{},
				TargetCellValue: req.Msg.TargetCellValue,
			},
			TargetCellValue:     req.Msg.TargetCellValue,
			MaxBricks:           int(req.Msg.MaxBricks),
			MinBricks:           0,
			MinMoves:            int(req.Msg.MinMoves),
			MaxMoves:            int(req.Msg.MaxMoves),
			MaxIterations:       100_000_000,
			Concurrency:         0,
			CellGenerator:       nil,
			Seed:                0,
			MinGames:            0,
			GameSolutionChannel: gameCh,
			Randomizer:          randomizer.NewSeededRandomizer(),
		}
		generator, err = tallylogic.NewGameGenerator(legacy_options)
	}
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to initialize game-generator: %w", err))
	}
	game, solutions, err := generator.GenerateGame(ctx)
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
	gameStats, solutionStats, err := calculateStats(game, solutions)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
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
		Stats: &model.GameStats{
			UniqueFactors:           gameStats.UniqueFactors,
			UniqueValues:            gameStats.UniqueValues,
			DuplicateFactors:        uint64(gameStats.DuplicateFactors),
			DuplicateValues:         uint64(gameStats.DuplicateValues),
			WithValueCount:          uint64(gameStats.WithValueCount),
			CellCount:               uint64(gameStats.CellCount),
			UniqueHints:             uint64(gameStats.UniqueHints),
			Hints:                   toModelHint(gameStats.Hints),
			IdealMovesSolutionIndex: uint32(solutionStats.IdealMovesSolutionIndex),
			MaxScoreSolutionIndex:   uint32(solutionStats.MaxScoreSolutionIndex),
			IdealMoves:              uint32(solutionStats.IdealMoves),
			ScoreOnIdeal:            solutionStats.ScoreOnIdeal,
			MaxScore:                solutionStats.MaxScore,
			SolutionStats:           make([]*model.SolutionStat, len(solutionStats.Stats)),
		},
	}
	for i, stat := range solutionStats.Stats {
		response.Stats.SolutionStats[i] = &model.SolutionStat{
			Moves:          uint32(stat.Moves),
			Score:          stat.Score,
			InstructionTag: make([]*model.InstructionTag, len(stat.InstructionTags)),
		}
		for j, t := range stat.InstructionTags {
			response.Stats.SolutionStats[i].InstructionTag[j] = &model.InstructionTag{
				Ok:               t.Ok,
				IsMultiplication: t.IsMultiplication,
				IsAddition:       t.IsAddition,
				IsSwipe:          t.IsSwipe,
				TwoPow:           t.TwoPow,
			}
		}
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

func calculateStats(game tallylogic.Game, solutions []tallylogic.Game) (gamestats.GameStats, gamestats.SolutionStats, error) {
	stats, err := gamestats.NewGameStats(game)
	if err != nil {
		return stats, gamestats.SolutionStats{}, fmt.Errorf("failed to calculate stats for game %w", err)
	}

	solutionStats, err := gamestats.NewSolutionsStats(game, solutions)
	if err != nil {
		return stats, solutionStats, fmt.Errorf("failed to calculate stats for solutions %w", err)
	}

	fmt.Printf("Game can ideally be solved in %d moves, for %d points (solution index %d)\n", solutionStats.IdealMoves, solutionStats.ScoreOnIdeal, solutionStats.IdealMovesSolutionIndex)
	fmt.Printf("The best score found was %d (solution index %d)\n", solutionStats.MaxScore, solutionStats.MaxScoreSolutionIndex)

	fmt.Printf("game has %d unique factors with %d unique values\n", len(stats.UniqueFactors), len(stats.UniqueValues))
	fmt.Printf("game has %d duplicate factors with %d duplicate values\n", stats.DuplicateFactors, stats.DuplicateValues)
	fmt.Printf("game has %d hints available (%d unique)\n", len(stats.Hints), stats.UniqueHints)
	fmt.Printf("game has %d non-empty cells %2f%%\n", stats.WithValueCount, float64(stats.WithValueCount)/float64(stats.CellCount)*100)
	return stats, solutionStats, nil
}
