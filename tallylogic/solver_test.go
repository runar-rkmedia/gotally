package tallylogic

import (
	"testing"
)

func Test_bruteSolver_SolveGame(t *testing.T) {
	tests := []struct {
		name              string
		g                 Game
		wantSolutionCount int
	}{

		{
			"Solve a simple game",
			mustCreateNewGame(GameModeTemplate, &ChallengeGames[0]),
			12,
		},
		{
			"Solve next game",
			mustCreateNewGame(GameModeTemplate, &ChallengeGames[1]),
			218,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBruteSolver()
			h := NewHintCalculator(tt.g.board, tt.g.board, tt.g.board)
			b.hinter = &h
			solutions, err := b.SolveGame(tt.g, 0)
			if err != nil {
				t.Error(err)
				return
			}

			if len(solutions) != tt.wantSolutionCount {
				t.Errorf("Found %d solutions, want %d", len(solutions), tt.wantSolutionCount)
				for i, solved := range solutions {
					t.Logf("Solution %d: solved in %d moves with a score of %d %#v", i, solved.Moves(), solved.Score(), solved.History)
				}
				t.Log(tt.g.board.String())
			}

		})
	}
}
