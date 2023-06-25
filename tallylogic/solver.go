package tallylogic

import (
	"time"
)

type Hinter interface {
	GetHints() map[string]Hint
}
type Solver interface {
	SolveGame(g Game, quitCh chan struct{}) ([]Game, error)
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

type HintMethod = string

const (
	HintMethodBreadthFirst = "BreadthFirst"
	HintMethodDepthFirst   = "DepthFirst"
)

func SolveGame(options SolveOptions, game Game, quitCh chan struct{}) (HintMethod, []Game, error) {
	// The breadth-first is not very good at solving infinite games
	// so we use the depth-first for these games
	if game.Rules.GameMode == GameModeRandom {
		s := NewBruteDepthSolver(options)
		g, err := s.SolveGame(game, quitCh)
		return HintMethodDepthFirst, g, err
	}

	s := NewBruteBreadthSolver(options)
	g, err := s.SolveGame(game, quitCh)
	return HintMethodBreadthFirst, g, err
}
