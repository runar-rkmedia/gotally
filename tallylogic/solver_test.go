package tallylogic

import (
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/MarvinJWendt/testza"
	"github.com/runar-rkmedia/gotally/types"
)

func init() {
	testza.SetShowStartupMessage(false)

}

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

func Test_bruteSolver_SolveGame(t *testing.T) {
	tests := []struct {
		name string
		GameSolverFactoryOptions
		g                                 func() Game
		wantSolutionCountGte              int
		wantOneOfSolutionShortDescription []string
	}{

		{
			"Solve a simple game",
			GameSolverFactoryOptions{BreadthFirst: false},
			mustCreateNewGameForTest(GameModeTutorial, &TutorialGames[0]),
			12,
			nil,
		},
		{
			"Solve a simple game",
			GameSolverFactoryOptions{BreadthFirst: true},
			mustCreateNewGameForTest(GameModeTutorial, &TutorialGames[0]),
			12,
			nil,
		},
		{
			"Solve next game",
			GameSolverFactoryOptions{BreadthFirst: false},
			mustCreateNewGameForTest(GameModeTutorial, &TutorialGames[1]),
			28,
			nil,
		},
		{
			"Solve next game",
			GameSolverFactoryOptions{BreadthFirst: true},
			mustCreateNewGameForTest(GameModeTutorial, &TutorialGames[1]),
			// I have not confirmed why there is a difference in the count of soltuions produced between breadth-first and depth-first
			// but I dont have the time to figure it out.
			// Could it be that when they decide that they have seen a game before, they are visiting vastly different nodes?
			211,
			nil,
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
			GameSolverFactoryOptions{BreadthFirst: true, SolveOptions: SolveOptions{MaxSolutions: 1}},
			func() Game {
				return createGame(
					4, 10, 12,
					2, 10, 8,
					1, 7, 0,
				)
			},
			1,
			[]string{
				"7,6,3,4;2,5,4;0,1,4;",
				"3,6,7,4;2,5,4;0,1,4;",
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
			GameSolverFactoryOptions{BreadthFirst: false},
			mustCreateNewGameForTest(GameModeRandom, nil, NewGameOptions{Seed: 1238}),
			// Not sure what to make of this value
			-1,
			nil,
		},
	}

	for _, tt := range tests {
		tt.SolveOptions.MaxTime = 10 * time.Second
		prefix := ""
		if tt.BreadthFirst {
			prefix = "breadth "
		} else {
			prefix = "depth "
		}
		t.Run(prefix+tt.name, func(t *testing.T) {
			gg := tt.g()
			if tt.BreadthFirst && gg.Rules.GameMode == GameModeRandom {
				t.Log("Skipping infinite game-test for BreadthFirst since it should not be used for that ")
				return
			}
			// err := tt.g.cellGenerator.SetSeed(uint64(123))
			// if err != nil {
			// 	t.Fatalf("failed to set seed %s", err)
			// }
			originalSeed, originalState := gg.cellGenerator.Seed()
			b := GameSolverFactory(tt.GameSolverFactoryOptions)
			t.Logf("BruteSolver: %#v", b)
			start := time.Now()
			timeTaken := time.Now().Sub(start)
			solutions, err := b.SolveGame(gg, nil)
			t.Logf("Found %d solutions in %s", len(solutions), timeTaken)
			if err != nil {
				t.Error(err)
				return
			}
			if tt.wantOneOfSolutionShortDescription != nil && len(tt.wantOneOfSolutionShortDescription) > 0 {
				found := false
				s := make([]string, len(solutions))
			outer:
				for _, want := range tt.wantOneOfSolutionShortDescription {
					for i, solved := range solutions {
						s[i] = solved.History.Describe()
					}
					sort.Slice(s, func(i, j int) bool {
						li := len(s[i])
						lj := len(s[j])
						if li == lj {
							return s[i] < s[j]
						}
						return li < lj
					})
					for _, v := range s {

						if want == v {
							found = true
							break outer
						}
					}
				}
				if !found {
					t.Fatalf("Expected to find one of these solutions, but none of them were included.\nOne of: %v\nGot: %v", tt.wantOneOfSolutionShortDescription, s)
				}
			}

			if tt.wantSolutionCountGte >= 0 && len(solutions) < tt.wantSolutionCountGte {
				t.Errorf("Found %d solutions, want at least %d", len(solutions), tt.wantSolutionCountGte)
				for i, solved := range solutions {
					t.Logf("Solution %d: solved - %d moves with a score of %d %s", i, solved.Moves(), solved.Score(), solved.History.Describe())
				}
				// intentionally no prefix for testname here, since we want to compare them
				t.Log(gg.board.String())
			}

			// intentionally no prefix for testname here, since we want to compare them
			t.Log(gg.board.String())
			if gg.Rules.NoReswipe {
				for _, solution := range solutions {
					history, err := solution.History.All()
					if err != nil {
						t.Fatalf("Failed to retrieve all history-items: %v", err)
					}
					for i := 1; i < len(history); i++ {
						prev := history[i-1]
						curr := history[i]
						if !prev.IsPath || !curr.IsPath {
							continue
						}
						if prev.Equal(curr) {
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
			if seed, state := gg.cellGenerator.Seed(); seed != originalSeed || state != originalState {
				t.Fatal("seed changed")
			}

			// Check that we can replay the instructions from the solutions (first one at least)
			// This is mostly to check that randomizers have been reset, and any other state is not changed
			solution := solutions[0]
			history, err := solution.History.All()
			for i := 0; i < len(history); i++ {
				instr := history[i]
				combine := func(vs ...any) string {
					return fmt.Sprintf("%v", vs)
				}
				t.Logf("seed: %d-%d, %s %s", originalSeed, originalState, combine(gg.cellGenerator.Seed()), combine(solution.cellGenerator.Seed()))
				desc := gg.DescribeInstruction(instr)
				switch {
				case instr.IsPath:
					for _, v := range instr.Path {
						gg.SelectCell(v)
					}
				}
				t.Log(desc, gg.board.PrintBoard(BoardHightlighter(&gg)))
				ok := gg.Instruct(instr)
				if !ok {
					t.Errorf("Failed to run instruction for game at positon %d %s", i, desc)
					return
				}

			}
		})
	}
}

func Benchmark_Solver(b *testing.B) {
	game := mustCreateNewGameForTest(GameModeTutorial, &TutorialGames[1])()
	options := GameSolverFactoryOptions{
		SolveOptions: SolveOptions{
			MaxTime: 10000 * time.Millisecond,
		},
	}
	brute := GameSolverFactory(options)
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			s, err := brute.SolveGame(game, nil)
			if err != nil {
				b.Error(err)
			}
			if len(s) == 0 {
				b.Errorf("Found no solutions")
			}
		}
	})
}
