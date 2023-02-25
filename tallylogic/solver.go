package tallylogic

import (
	"time"
)

type Hinter interface {
	GetHints() map[string]Hint
}
type Solver interface {
	SolveGame(g Game) ([]Game, error)
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

type GameSolverFactoryOptions struct {
	SolveOptions
	BreadthFirst bool
}

type SolverErr struct {
	error
	ShouldQuit bool
}

func NewSolverErr(err error, shouldQuit bool) SolverErr {
	return SolverErr{err, shouldQuit}
}

func GameSolverFactory(options GameSolverFactoryOptions) Solver {
	if options.BreadthFirst {
		s := NewBruteBreadthSolver(options.SolveOptions)
		return &s
	}
	s := NewBruteDepthSolver(options.SolveOptions)
	return &s
}
func SolveGame(options SolveOptions, game Game) ([]Game, error) {
	// The breadth-first is not very good at solving infinite games
	// so we use the depth-first for these games
	if game.Rules.GameMode == GameModeRandom {
		s := NewBruteDepthSolver(options)
		return s.SolveGame(game)
	}

	s := NewBruteBreadthSolver(options)
	return s.SolveGame(game)
}
