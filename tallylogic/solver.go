package tallylogic

import "time"

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
		s := NewBruteDepthSolver(options.SolveOptions)
		return &s
	}
	s := NewBruteDepthSolver(options.SolveOptions)
	return &s
}
