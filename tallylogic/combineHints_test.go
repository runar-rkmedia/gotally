package tallylogic

import (
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/stretchr/testify/assert"
)

func TestGame_GetCombineHints(t *testing.T) {
	type args struct {
		g      Game
		quitCh chan struct{}
	}
	tests := []struct {
		name                       string
		g                          func() Game
		maxHints                   int
		wantedMultiplicationsCount int
		wantedAdditionsCount       int
		wantedMultiplications      [][]int
		wantedAdditions            [][]int
	}{
		// TODO: Add test cases.
		{
			"Should solve an intricate game fast",
			// TODO: copy the template here, since we are likely to remove that template in the future, but we should keep the test
			mustCreateNewGameForTest(GameModeTutorial, &TutorialGames[2]),
			0,
			20,
			39,
			nil,
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := tt.g()
			t.Log("Testing for game", g.Print())
			multiplications := [][]int{}
			additions := [][]int{}
			g.GetCombineHints(func(path []int, method EvalMethod) bool {
				switch method {
				case EvalMethodProduct:
					multiplications = append(multiplications, path)
				case EvalMethodSum:
					additions = append(additions, path)
				default:
					t.Fatalf("Unexpected EvalMethod: %v", method)
				}
				if tt.maxHints > 0 {
					return true
				}
				return false

			})
			assert.Equal(t, tt.wantedMultiplicationsCount, len(multiplications))
			assert.Equal(t, tt.wantedAdditionsCount, len(additions))
			if tt.wantedMultiplications != nil {
				assert.Equal(t, tt.wantedMultiplications, multiplications)
			}
			if tt.wantedAdditions != nil {
				assert.Equal(t, tt.wantedAdditions, additions)
			}
			if tt.wantedAdditions == nil || tt.wantedMultiplications == nil {
				cupaloy.SnapshotT(t, multiplications, additions)
			}
		})
	}
}
