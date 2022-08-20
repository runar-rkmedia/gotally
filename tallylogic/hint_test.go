package tallylogic

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/gookit/color"
)

func Test_hintCalculator_GetHints(t *testing.T) {
	tests := []struct {
		name      string
		board     TableBoard
		want      []Hint
		wantCount int
	}{
		// TODO: Add test cases.
		{
			"Get simple hints",
			TableBoard{
				cells: cellCreator(
					0, 0, 3,
					1, 2, 3,
					0, 0, 0,
				),
				rows:    3,
				columns: 3,
			},
			[]Hint{
				{
					Value:  6,
					Method: EvalMethodSum,
					Path:   []int{5, 2},
				},
				{
					Value:  6,
					Method: EvalMethodSum,
					Path:   []int{3, 4, 5},
				},
			},
			2,
		},
		{
			"Test for bigger board",
			TableBoard{
				cells: cellCreator(
					6, 7, 6, 11, 14,
					20, 18, 20, 8, 16,
					11, 4, 18, 1, 12,
					5, 11, 5, 10, 7,
					10, 4, 6, 18, 4,
				),
				rows:    5,
				columns: 5,
			},
			[]Hint{
				{
					Value:  36,
					Method: EvalMethodSum,
					// The bottom row , 10, 4, 6,
					Path: []int{19, 18, 13, 12},
				},
				{
					Value:  22,
					Method: EvalMethodSum,
					// The bottom row , 10, 4, 6,
					Path: []int{22, 17, 16},
				},
				{
					Value:  20,
					Method: EvalMethodSum,
					// The bottom row , 10, 4, 6,
					Path: []int{22, 21, 20},
				},
			},
			3,
		},
		{
			"Test for stupid amount of hints",
			TableBoard{
				cells: cellCreator(
					1, 1, 1, 1, 1,
					1, 1, 1, 1, 1,
					1, 1, 1, 1, 1,
					1, 1, 1, 1, 1,
					1, 1, 1, 1, 1,
				),
				rows:    5,
				columns: 5,
			},
			nil,
			// This is not verified, but it seems reasonable
			15448,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &hintCalculator{
				CellRetriever:      &tt.board,
				NeighbourRetriever: tt.board,
				Evaluator:          tt.board,
			}
			gotHints := g.GetHints()

			if tt.want != nil {
				wantMap := map[string]Hint{}
				for _, h := range tt.want {
					hash := h.Hash()
					h.pathHash = hash
					wantMap[h.pathHash] = h
				}
				if !reflect.DeepEqual(gotHints, wantMap) {
					for i, hint := range gotHints {
						t.Logf("hint %s %d, %s", i, hint.Value, tt.board.PrintBoard(func(c Cell, index int, padded string) string {
							s := fmt.Sprintf("[ %s ]", padded)
							for _, v := range hint.Path {
								if v == index {
									return color.Red.Sprintf(s)
								}
							}
							return s
						}))
					}
					t.Errorf("hintCalculator.GetHints() = (count %d wanted %d) \ngot : %v, \nwant: %#v", len(gotHints), len(tt.want), gotHints, tt.want)
				}
			}
			if tt.wantCount != len(gotHints) {
				t.Errorf("fail gotCount %d, wantCount %d", len(gotHints), tt.wantCount)
			}
		})
	}
}

func BenchmarkGetHints5x5ofOnes(b *testing.B) {
	board := TableBoard{
		cells: cellCreator(
			1, 1, 1, 1, 1,
			1, 1, 1, 1, 1,
			1, 1, 1, 1, 1,
			1, 1, 1, 1, 1,
			1, 1, 1, 1, 1,
		),
		rows:    5,
		columns: 5,
	}
	g := &hintCalculator{
		CellRetriever:      &board,
		NeighbourRetriever: board,
		Evaluator:          board,
	}

	for i := 0; i < b.N; i++ {
		g.GetHints()
	}

}
func BenchmarkGetHints3x3ofOnes(b *testing.B) {
	board := TableBoard{
		cells: cellCreator(
			1, 1, 1,
			1, 1, 1,
			1, 1, 1,
		),
		rows:    3,
		columns: 3,
	}
	g := &hintCalculator{
		CellRetriever:      &board,
		NeighbourRetriever: board,
		Evaluator:          board,
	}

	for i := 0; i < b.N; i++ {
		g.GetHints()
	}

}
