package tallylogic

import "fmt"

type bruteSolver struct {
	hinter    Hinter
	seen      map[string]struct{}
	maxDepth  int
	maxVisits int
}

type Hinter interface {
	GetHints() map[string]Hint
}

func NewBruteSolver() bruteSolver {
	return bruteSolver{
		maxDepth:  1000,
		maxVisits: 10000,
	}
}

func (b *bruteSolver) SolveGame(g Game, maxSolutions int) ([]Game, error) {
	seen := map[string]struct{}{}
	return b.solveGame(g, []Game{}, maxSolutions, -1, &seen)
}
func (b *bruteSolver) solveGame(g Game, solutions []Game, maxSolutions int, depth int, seen *map[string]struct{}) ([]Game, error) {
	depth++
	if depth > b.maxDepth {
		return solutions, fmt.Errorf("Game-depth overflow %d (seen: %d)", depth, len(*seen))
	}
	if len(*seen) > b.maxVisits {
		return solutions, fmt.Errorf("Game-seen overflow")
	}
	// fmt.Println("Solving", depth, g.board.String(), len(seen))
	hash := g.board.Hash()
	if _, ok := (*seen)[hash]; ok {
		return solutions, nil
	}
	(*seen)[hash] = struct{}{}
	hints := g.GetHint()
	for _, h := range hints {
		gameCopy := g.Copy()
		ok := gameCopy.EvaluateForPath(h.Path)
		// fmt.Printf("Game won %#v, %d (%d)\n\n", gameCopy.GoalChecker, gameCopy.score, g.score)
		if !ok {
			return solutions, fmt.Errorf("Failed in game-solving for hint")
		}
		if gameCopy.IsGameWon() {
			solutions = append(solutions, gameCopy)
			if maxSolutions > 0 && len(solutions) >= maxSolutions {
				return solutions, nil
			}
		} else {
			more, err := b.solveGame(gameCopy, solutions, maxSolutions, depth, seen)
			if err != nil {
				return solutions, err
			}
			solutions = more
		}
	}
	for _, dir := range []SwipeDirection{SwipeDirectionUp, SwipeDirectionRight, SwipeDirectionDown, SwipeDirectionLeft} {
		gameCopy := g.Copy()
		// fmt.Println("Swiping", dir, gameCopy.board.String())
		changed := gameCopy.Swipe(dir)
		if changed {
			// fmt.Println("Swiped", dir, gameCopy.History, gameCopy.board.String()) //gameCopy.board.Hash(), seen)
			more, err := b.solveGame(gameCopy, solutions, maxSolutions, depth, seen)
			if err != nil {
				return solutions, err
			}
			solutions = more
		}

	}
	return solutions, nil

}

// func (b *bruteSolver) MaximumTheoriticalResult(g Game) int64 {

// }
