package tallylogic

import (
	"fmt"
	"testing"
	"time"
)

func Test_bruteSolver_SolveGame(t *testing.T) {
	tests := []struct {
		name              string
		g                 Game
		wantSolutionCount int
	}{

		{
			"Solve a simple game",
			mustCreateNewGameForTest(GameModeTemplate, &ChallengeGames[0]),
			12,
		},
		{
			"Solve next game",
			mustCreateNewGameForTest(GameModeTemplate, &ChallengeGames[1]),
			218,
		},
		{
			// Infinite games cannot be solved, but it should calculate the "best" moves that it can make
			// to get the best points, or the boards complexity is reduced.
			// TOOD: create a measure of complexity.
			// - An empty board is low (zero) amount of complex
			// - A mix of root-number-types makes it more complex. This has to take into account the
			//   fraction of the composite.
			"'Solve' an infinite game",
			mustCreateNewGameForTest(GameModeDefault, nil, NewGameOptions{Seed: 1238}),
			// Not sure what to make of this value
			-1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// err := tt.g.cellGenerator.SetSeed(uint64(123))
			// if err != nil {
			// 	t.Fatalf("failed to set seed %s", err)
			// }
			originalSeed, originalState := tt.g.cellGenerator.Seed()
			b := NewBruteSolver(SolveOptions{MaxTime: 10000 * time.Millisecond})
			h := NewHintCalculator(tt.g.board, tt.g.board, tt.g.board)
			b.hinter = &h
			solutions, err := b.SolveGame(tt.g)
			t.Logf("Found %d solutions", len(solutions))
			if err != nil {
				t.Error(err)
				return
			}

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
	game := mustCreateNewGameForTest(GameModeTemplate, &ChallengeGames[1])
	brute := NewBruteSolver(SolveOptions{MaxTime: 10000 * time.Millisecond})
	hinter := NewHintCalculator(game.board, game.board, game.board)
	brute.hinter = &hinter
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			s, err := brute.SolveGame(game)
			if err != nil {
				b.Error(err)
			}
			if len(s) == 0 {
				b.Errorf("Found no solutions")
			}
		}
	})
}
