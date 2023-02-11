package tallylogic

import (
	"fmt"
	"testing"
	"time"

	"github.com/runar-rkmedia/gotally/types"
)

func createGame(vals ...int64) Game {

	g, err := RestoreGame(
		&types.Game{
			Rules: types.Rules{
				Mode:            types.RuleModeChallenge,
				Rows:            3,
				Columns:         3,
				TargetCellValue: 80,
			},
			Cells: cellCreator(vals...),
		},
	)
	if err != nil {
		panic(err)
	}

	return g
}

// Note that these requirements are likely to produce flaky tests
type solveStatsRequirements struct {
	maxTime      time.Duration
	maxDepth     int
	maxSeenGames int
}

func (s solveStatsRequirements) passesRequirements(t *testing.T, stats SolveStatistics) {
	if s.maxTime > 0 {
		if stats.Duration > s.maxTime {
			t.Errorf("Failed statisticsRequirements for maxTime within %s, was %s\n stats %#v", s.maxTime, stats.Duration, stats)
		}
	}
	if s.maxDepth > 0 && stats.Depth > s.maxDepth {
		t.Errorf("Failed statisticsRequirements for depth within %d, was %d\n stats %#v", s.maxDepth, stats.Depth, stats)
	}
	if s.maxSeenGames > 0 && stats.SeenGames > s.maxSeenGames {
		t.Errorf("Failed statisticsRequirements for number of seen games within %d, was %d\n stats %#v", s.maxSeenGames, stats.Depth, stats)
	}
	// return fmt.Errorf("Doesn not pass statisticsRequirements: %#v", stats)
}

func Test_bruteSolver_SolveGame(t *testing.T) {
	tests := []struct {
		name              string
		g                 Game
		wantSolutionCount int
		SolveOptions
		solveStatsRequirements
	}{

		{
			"Solve a simple game",
			mustCreateNewGameForTest(GameModeTutorial, &TutorialGames[0]),
			12,
			SolveOptions{},
			solveStatsRequirements{},
		},
		{
			"Solve next game",
			mustCreateNewGameForTest(GameModeTutorial, &TutorialGames[1]),
			218,
			SolveOptions{},
			solveStatsRequirements{},
		},
		{
			// This game currently goes very deep
			// This can be mitigated with https://github.com/runar-rkmedia/gotally/issues/14
			// The shortest solution should be:
			// Combine 7+1+2=10 into 20
			// [7, 6, 3, 4]
			// Combine 12+8=20 into 40
			// [2, 5, 4]
			// Combine 4*10=40 into 80
			// [0, 1, 4]
			"Solve challenge game 0130-current-paul-robin",
			createGame(
				4, 10, 12,
				2, 10, 8,
				1, 7, 0,
			),
			32,
			SolveOptions{},
			solveStatsRequirements{},
		},
		{
			// Same game as above
			// TODO: make the game pass by implementing a breadth-first solver
			// UPDATE: Nah, there is no realy point in this. Just use MaxSolutions=1 for challenges.
			/// It is fast enough (68 microseconds if it finds it fast, but here up 40 milliseoncds for the slowest)
			// The shortest solution should be:
			// Combine 7+1+2=10 into 20
			// [7, 6, 3, 4]
			// Combine 12+8=20 into 40
			// [2, 5, 4]
			// Combine 4*10=40 into 80
			// [0, 1, 4]
			// There is a very short path available, so in theory, it should solve the game within:
			// maxdepth: 4
			// seenGames at most: 9^3 = 27
			"Solve challenge game 0130-current-paul-robin within strict requirements",
			createGame(
				4, 10, 12,
				2, 10, 8,
				1, 7, 0,
			),
			1,
			SolveOptions{MaxSolutions: 1, MaxTime: time.Second},
			solveStatsRequirements{
				// maxDepth:     3,
				// maxSeenGames: 100,
				maxTime: time.Millisecond * 10,
			},
		},
		{
			// Infinite games cannot be solved, but it should calculate the "best" moves that it can make
			// to get the best points, or the boards complexity is reduced.
			// TOOD: create a measure of complexity.
			// - An empty board is low (zero) amount of complex
			// - A mix of root-number-types makes it more complex. This has to take into account the
			//   fraction of the composite.
			"'Solve' an infinite game",
			mustCreateNewGameForTest(GameModeRandom, nil, NewGameOptions{Seed: 1238}),
			// Not sure what to make of this value
			-1,
			SolveOptions{},
			solveStatsRequirements{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// err := tt.g.cellGenerator.SetSeed(uint64(123))
			// if err != nil {
			// 	t.Fatalf("failed to set seed %s", err)
			// }
			originalSeed, originalState := tt.g.cellGenerator.Seed()
			opts := tt.SolveOptions
			opts.WithStatistics = true
			if opts.MaxTime == 0 {
				opts.MaxTime = 10 * time.Second
			}
			b := NewBruteSolver(opts)
			t.Logf("BruteSolver: %#v", b)
			solves, err := b.SolveGame(tt.g)
			solutions := solves.Games
			t.Logf("Found %d solutions", len(solutions))
			if err != nil {
				t.Error(err)
				return
			}
			t.Logf("\n\nstat %s %#v", solves.Statistics.Duration, solves.Statistics)
			tt.solveStatsRequirements.passesRequirements(t, solves.Statistics)

			if tt.wantSolutionCount >= 0 && len(solutions) != tt.wantSolutionCount {
				t.Errorf("Found %d solutions, want %d", len(solutions), tt.wantSolutionCount)
				for i, solved := range solutions {
					t.Logf("Solution %d: solved in %d moves with a score of %d %#v", i, solved.Moves(), solved.Score(), solved.History)
				}
				t.Log(tt.g.board.String())
			}

			if tt.g.Rules.NoReswipe {
				for _, solution := range solutions {
					for i := 1; i < len(solution.History); i++ {
						prev := solution.History[i-1]
						curr := solution.History[i]
						if equal, kind := CompareInstrictionAreEqual(prev, curr); equal && kind == InstructionTypeSwipe {
							// Duplicate paths are not allowed, see Rules.NoReswipe
							t.Errorf("duplicate swipe in path at %d: %v", i, solution.History)
							break
						}
					}
				}
			}
			if len(solutions) == 0 {
				return
			}
			if seed, state := tt.g.cellGenerator.Seed(); seed != originalSeed || state != originalState {
				t.Fatal("seed changed")
			}

			// Check that we can replay the instructions from the solutions (first one at least)
			// This is mostly to check that randomizers have been reset, and any other state is not changed
			solution := solutions[0]
			for i := 0; i < len(solution.History); i++ {
				instr := solution.History[i]
				combine := func(vs ...any) string {
					return fmt.Sprintf("%v", vs)
				}
				t.Logf("seed: %d-%d, %s %s", originalSeed, originalState, combine(tt.g.cellGenerator.Seed()), combine(solution.cellGenerator.Seed()))
				desc := tt.g.DescribeInstruction(instr)
				switch t := instr.(type) {
				case []int:
					for _, v := range t {
						tt.g.SelectCell(v)
					}
				}
				t.Log(desc, tt.g.board.PrintBoard(BoardHightlighter(&tt.g)))
				ok := tt.g.instruct(instr)
				if !ok {
					t.Errorf("Failed to run instruction for game at positon %d %s", i, desc)
					return
				}

			}
		})
	}
}

func Benchmark_Solver(b *testing.B) {
	game := mustCreateNewGameForTest(GameModeTutorial, &TutorialGames[1])
	brute := NewBruteSolver(SolveOptions{MaxTime: 10000 * time.Millisecond})
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			s, err := brute.SolveGame(game)
			if err != nil {
				b.Error(err)
			}
			if len(s.Games) == 0 {
				b.Errorf("Found no solutions")
			}
		}
	})
}
