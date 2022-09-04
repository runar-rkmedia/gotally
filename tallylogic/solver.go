package tallylogic

import "fmt"

type bruteSolver struct {
	hinter Hinter
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
	return bruteSolver{
		SolveOptions: options,
	}
}

type SolveOptions struct {
	MaxDepth     int
	MaxVisits    int
	MinMoves     int
	MaxMoves     int
	MaxSolutions int
}

func (b *bruteSolver) SolveGame(g Game) ([]Game, error) {
	seen := map[string]struct{}{}
	game := g.Copy()
	game.History = Instruction{}
	brr, err := b.solveGame(game, g.moves, []Game{}, -1, &seen)

	return brr, err
}
func (b *bruteSolver) solveGame(g Game, startingMoves int, solutions []Game, depth int, seen *map[string]struct{}) ([]Game, error) {
	depth++
	if depth > b.MaxDepth {
		return solutions, fmt.Errorf("Game-depth overflow %d (seen: %d)", depth, len(*seen))
	}
	if len(*seen) > b.MaxVisits {
		return solutions, fmt.Errorf("Game-seen overflow")
	}

	if b.MaxMoves > 0 && b.MaxMoves < (g.Moves()-startingMoves) {
		return solutions, fmt.Errorf("Max-moves threshold triggered: %d, maxmoves %d", g.Moves(), b.MaxMoves)
	}
	hash := g.board.Hash()
	if _, ok := (*seen)[hash]; ok {
		return solutions, fmt.Errorf("Already seen")
	}
	(*seen)[hash] = struct{}{}
	hints := g.GetHint()
	for _, h := range hints {
		gameCopy := g.Copy()
		ok := gameCopy.EvaluateForPath(h.Path)
		if !ok {
			return solutions, fmt.Errorf("Failed in game-solving for hint")
		}
		if gameCopy.IsGameWon() {
			// if b.MinMoves > 0 && b.MinMoves > gameCopy.Moves() {
			// 	return solutions, fmt.Errorf("Game solved in less moves than required: %d moves wanted at least %d", gameCopy.Moves(), b.MinMoves)
			// }
			solutions = append(solutions, gameCopy)
			if b.MaxSolutions > 0 && len(solutions) >= b.MaxSolutions {
				return solutions, nil
			}
		} else {
			more, err := b.solveGame(gameCopy, startingMoves, solutions, depth, seen)
			if err != nil {
				continue
				// return solutions, err
			}
			solutions = more
			if b.MaxSolutions > 0 && len(solutions) >= b.MaxSolutions {
				return solutions, nil
			}
		}
	}
	for _, dir := range []SwipeDirection{SwipeDirectionUp, SwipeDirectionRight, SwipeDirectionDown, SwipeDirectionLeft} {
		gameCopy := g.Copy()
		changed := gameCopy.Swipe(dir)
		if changed {
			more, err := b.solveGame(gameCopy, startingMoves, solutions, depth, seen)
			if err != nil {
				continue
				// return solutions, err
			}
			solutions = more
			if b.MaxSolutions > 0 && len(solutions) >= b.MaxSolutions {
				return solutions, nil
			}
		}

	}
	return solutions, nil

}

// func (b *bruteSolver) MaximumTheoriticalResult(g Game) int64 {

// }
