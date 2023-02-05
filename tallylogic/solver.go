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
}

func (b *bruteSolver) SolveGame(g Game) ([]Game, error) {

	seen := map[string]struct{}{}
	game := g.Copy()
	game.History = Instruction{}
	solutionsChan := make(chan Game)
	ctx, cancel := context.WithTimeout(context.Background(), b.MaxTime)
	defer cancel()
	solutions := []Game{}
	var err error
	go func() {
		err = b.solveGame(ctx, game, g.moves, solutionsChan, -1, &seen, &g)
		cancel()
	}()
	for {
		select {
		case solvedGame := <-solutionsChan:
			solutions = append(solutions, solvedGame)
			if solvedGame.Rules.GameMode == GameModeRandom && len(solutions) > 0 {
				if solvedGame.score-g.score > int64(b.InfiniteGameMaxScoreIncrease) {
					if err != nil && errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
						err = nil
					}
					return solutions, err
				}
			}
		case <-ctx.Done():
			err := ctx.Err()
			if err != nil && len(solutions) > 0 && errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
				err = nil
			}
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
	depth int,
	seen *map[string]struct{},
	originalGame *Game,
) error {
	depth++
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
		err := b.solveGame(ctx, gameCopy, startingMoves, solutions, depth, seen, originalGame)
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
			err := b.solveGame(ctx, gameCopy, startingMoves, solutions, depth, seen, originalGame)
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
	return nil

}

// func (b *bruteSolver) MaximumTheoriticalResult(g Game) int64 {

// }
