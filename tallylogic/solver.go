package tallylogic

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type bruteSolver struct {
	SolveOptions
}

type Hinter interface {
	GetHints() map[string]Hint
}

func NewBruteSolver(options SolveOptions) bruteSolver {
	if options.MaxDepth == 0 {
		options.MaxDepth = 1_000
	}
	if options.MaxVisits == 0 {
		options.MaxVisits = 10_000
	}
	if options.MaxTime == 0 {
		options.MaxTime = 10 * time.Second
	}
	if options.InfiniteGameMaxScoreIncrease == 0 {
		options.InfiniteGameMaxScoreIncrease = 1000
	}
	return bruteSolver{
		SolveOptions: options,
	}
}

type SolveOptions struct {
	MaxDepth                     int
	MaxVisits                    int
	MinMoves                     int
	MaxMoves                     int
	MaxSolutions                 int
	InfiniteGameMaxScoreIncrease int
	MaxTime                      time.Duration
	WithStatistics               bool
}
type SolveStatistics struct {
	SeenGames int
	Depth     int
	Duration  time.Duration
}
type Solutions struct {
	Games      []Game
	Statistics SolveStatistics
}

func (b *bruteSolver) SolveGame(g Game) (Solutions, error) {

	seen := map[string]struct{}{}
	game := g.Copy()
	game.History = Instruction{}
	solutionsChan := make(chan Game)
	var statsChan chan SolveStatistics
	if b.WithStatistics {
		statsChan = make(chan SolveStatistics)
	}
	ctx, cancel := context.WithTimeout(context.Background(), b.MaxTime)
	defer cancel()
	solutions := Solutions{
		Games: []Game{},
	}
	var err error
	startTime := time.Now()
	go func() {
		err = b.solveGame(ctx, game, g.moves, solutionsChan, statsChan, -1, &seen, &g)
		cancel()
	}()
	for {
		select {
		case solvedGame := <-solutionsChan:
			solutions.Games = append(solutions.Games, solvedGame)
			if b.MaxSolutions >= len(solutions.Games) {
				cancel()
				solutions.Statistics.Duration = time.Now().Sub(startTime)
				return solutions, err
			}
			if solvedGame.Rules.GameMode == GameModeRandom && len(solutions.Games) > 0 {
				if solvedGame.score-g.score > int64(b.InfiniteGameMaxScoreIncrease) {
					if err != nil && errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
						err = nil
					}
					cancel()
					solutions.Statistics.Duration = time.Now().Sub(startTime)
					return solutions, err
				}
			}
		case stats := <-statsChan:
			if stats.SeenGames > 0 {
				solutions.Statistics.SeenGames += stats.SeenGames
			}
			if stats.Depth > solutions.Statistics.Depth {
				solutions.Statistics.Depth = stats.Depth
			}

		case <-ctx.Done():
			err := ctx.Err()
			if err != nil && len(solutions.Games) > 0 && errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
				err = nil
			}
			solutions.Statistics.Duration = time.Now().Sub(startTime)
			return solutions, err
		}
	}
}

type SolverErr struct {
	error
	ShouldQuit bool
}

func NewSolverErr(err error, shouldQuit bool) SolverErr {
	return SolverErr{err, shouldQuit}
}

// TODO: Major performance-boost is very much within reach with refactoring into a breadth-first implementation
// See https://github.com/runar-rkmedia/gotally/issues/14
func (b *bruteSolver) solveGame(
	ctx context.Context,
	g Game,
	startingMoves int,
	solutions chan Game,
	statsChannel chan SolveStatistics,
	depth int,
	seen *map[string]struct{},
	originalGame *Game,
) error {
	depth++
	if statsChannel != nil {
		statsChannel <- SolveStatistics{
			Depth: depth,
		}
	}
	if err := ctx.Err(); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return NewSolverErr(fmt.Errorf("MaxTime threshold exceeded (%w)", err), true)
		}
		return NewSolverErr(fmt.Errorf("context: err %w", err), true)
	}
	if depth > b.MaxDepth {
		return NewSolverErr(fmt.Errorf("Game-depth overflow %d (seen: %d)", depth, len(*seen)), false)
	}
	if len(*seen) > b.MaxVisits {
		return NewSolverErr(fmt.Errorf("Game-seen overflow (seend %d)", len(*seen)), false)
	}

	if b.MaxMoves > 0 && b.MaxMoves < (g.Moves()-startingMoves) {
		return NewSolverErr(fmt.Errorf("Max-moves threshold triggered: %d, maxmoves %d", g.Moves(), b.MaxMoves), true)
	}
	hash := g.board.Hash()
	if _, ok := (*seen)[hash]; ok {
		return NewSolverErr(fmt.Errorf("Already seen"), false)
	}
	(*seen)[hash] = struct{}{}
	if statsChannel != nil {
		statsChannel <- SolveStatistics{
			SeenGames: 1,
		}
	}
	hints := g.GetHint()
	for _, h := range hints {
		gameCopy := g.Copy()
		ok := gameCopy.EvaluateForPath(h.Path)
		if !ok {
			return NewSolverErr(fmt.Errorf("Failed in game-solving for hint"), true)
		}
		if gameCopy.IsGameWon() {
			// if b.MinMoves > 0 && b.MinMoves > gameCopy.Moves() {
			// 	return solutions, fmt.Errorf("Game solved in less moves than required: %d moves wanted at least %d", gameCopy.Moves(), b.MinMoves)
			// }
			solutions <- gameCopy
			if b.MaxSolutions > 0 && len(solutions) >= b.MaxSolutions {
				// TODO: introduce solutions-counter?
				return nil
			}
			continue
		}
		if gameCopy.Rules.GameMode == GameModeRandom {
			solutions <- gameCopy
		}
		err := b.solveGame(ctx, gameCopy, startingMoves, solutions, statsChannel, depth, seen, originalGame)
		if err != nil {
			if s, ok := err.(SolverErr); ok {
				if s.ShouldQuit {
					return err
				}
			}
			continue
			// return solutions, err
		}
		// solutions = more
		if b.MaxSolutions > 0 && len(solutions) >= b.MaxSolutions {
			return nil
		}
	}
	for _, dir := range []SwipeDirection{SwipeDirectionUp, SwipeDirectionRight, SwipeDirectionDown, SwipeDirectionLeft} {
		if !originalGame.Rules.NoReswipe && len(g.History) > 0 {
			if equal, _ := CompareInstrictionAreEqual(dir, g.History[len(g.History)-1]); equal {
				continue
			}
		}
		gameCopy := g.Copy()
		changed := gameCopy.Swipe(dir)
		if changed {
			err := b.solveGame(ctx, gameCopy, startingMoves, solutions, statsChannel, depth, seen, originalGame)
			if s, ok := err.(SolverErr); ok {
				if s.ShouldQuit {
					return err
				}
			}
			if err != nil {
				continue
				// return solutions, err
			}
			// solutions = more
			if b.MaxSolutions > 0 && len(solutions) >= b.MaxSolutions {
				return nil
			}
		}

	}
	// if depth == 0 {
	// 	stats := SolveStatistics{
	// 		SeenGames: len(*seen),
	// 	}
	// 	fmt.Printf("Depth 0 stats %#v %v\n", stats, statsChannel)
	// 	if statsChannel != nil {
	// 		statsChannel <- stats
	// 	}
	// }
	return nil

}

// func (b *bruteSolver) MaximumTheoriticalResult(g Game) int64 {

// }
