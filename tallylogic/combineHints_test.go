package tallylogic

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/MarvinJWendt/testza"
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
			mustCreateNewGameForTest(GameModeTutorial,
				NewGameTemplate(GameModeTutorial, "AllLinedUp", "All Lined Up", "Get a brick to 512. Can you combine them all into one?", 4, 4).
					SetStartingLayout(
						4, 1, 1, 4,
						2, 16, 8, 4,
						8, 32, 4, 4,
						2, 8, 8, 1,
					),
			),
			0,
			20,
			39,
			nil,
			nil,
		},
		{
			"Should not crash on zero-values t",
			mustCreateNewGameForTest(GameModeTutorial,
				NewGameTemplate(GameModeTutorial, "Sum&Product", "Sum & Product", "Get a brick to 36. Bricks can be added, or multiplied together. Try combining 5,4 into 9. What can you do with that 3 and 6?", 3, 3).
					SetStartingLayout(
						0, 0, 5,
						0, 0, 4,
						3, 6, 9,
					)),
			0,
			0,
			2,
			nil,
			nil,
		},
		{
			"Should not present a single value as a hint, like simply the cell 1.",
			mustCreateNewGameForTest(GameModeTutorial, NewGameTemplate(GameModeRandom, "test-no-hint-1", "no-1-hint", "foo", 5, 5).SetStartingCells(
				cellCreator(
					32, 0, 0, 0, 0,
					0, 0, 0, 3, 0,
					1, 0, 0, 0, 0,
					0, 0, 0, 0, 0,
					0, 0, 0, 0, 0,
				),
			)),
			0,
			0,
			0,
			[][]int{},
			[][]int{},
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
			if tt.wantedMultiplications != nil {
				testza.AssertEqual(t, tt.wantedMultiplications, multiplications)
			}
			if tt.wantedAdditions != nil {
				testza.AssertEqual(t, tt.wantedAdditions, additions, "wantedAdditions did not match")
			}
			if tt.wantedAdditions == nil || tt.wantedMultiplications == nil {
				cells := g.Cells()
				// The original is terrible to read, so I create a new one:
				var s = fmt.Sprintf("\nMultiplication (%d):\n", len(multiplications))
				for j, path := range multiplications {
					indexes := make([]string, len(path))
					values := make([]string, len(path))
					for i, index := range path {
						indexes[i] = strconv.FormatInt(int64(index), 10)
						values[i] = strconv.FormatInt(cells[index].Value(), 10)
					}
					s += fmt.Sprintf("% 3d: (% 6s)  Index: % 12s->% -2s Values: %s=%s ",
						j,
						values[len(values)-1],
						strings.Join(indexes[:len(indexes)-1], ","),
						indexes[len(indexes)-1],
						strings.Join(values[:len(values)-1], "*"),
						values[len(values)-1],
					)
					s += "\n"
				}
				s += fmt.Sprintf("Additions (%d):\n", len(additions))
				for j, path := range additions {
					indexes := make([]string, len(path))
					values := make([]string, len(path))
					for i, index := range path {
						indexes[i] = strconv.FormatInt(int64(index), 10)
						values[i] = strconv.FormatInt(cells[index].Value(), 10)
					}
					s += fmt.Sprintf("% 3d: (% 6s)  Index: % 12s->% -2s Values: %s=%s ",
						j,
						values[len(values)-1],
						strings.Join(indexes[:len(indexes)-1], ","),
						indexes[len(indexes)-1],
						strings.Join(values[:len(values)-1], "+"),
						values[len(values)-1],
					)
					s += "\n"
				}
				testza.SnapshotCreateOrValidate(t, t.Name()+"_multiplications", s)
			}
		})
	}
}
