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
		board     BoardController
		want      []Hint
		wantCount int
	}{
		// TODO: Add test cases.
		{
			"Get simple hints",
			&TableBoard{
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
					Path:   []int{2, 5},
				},
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
			3,
		},
		{
			"Test for challenge 0",
			mustCreateNewGame(GameModeTemplate, &ChallengeGames[0]).board,
			[]Hint{
				{
					Value:  18,
					Method: EvalMethodSum,
					// The bottom row , 10, 4, 6,
					Path: []int{2, 5, 8},
				},
				{
					Value:  18,
					Method: EvalMethodSum,
					// The bottom row , 10, 4, 6,
					Path: []int{6, 7, 8},
				},
			},
			2,
		},
		{
			"Test for challenge 1 after swipe",
			&TableBoard{
				cells: cellCreator(
					500, 1, 100,
					1, 0, 5,
					0, 0, 0,
				),
				rows:    3,
				columns: 3,
			},
			[]Hint{
				{
					Value:  1000,
					Method: EvalMethodProduct,
					Path:   []int{5, 2, 1, 0},
				},
			},
			1,
		},
		{
			"Test for bigger board",
			&TableBoard{
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
			&TableBoard{
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
			// This is not verified, but it seems reasonable.
			3_060_392,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &hintCalculator{
				CellRetriever:      tt.board,
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
					for _, hint := range gotHints {
						t.Logf("hint %v %d, %s", hint.Path, hint.Value, tt.board.PrintBoard(func(c CellValuer, index int, padded string) string {
							s := fmt.Sprintf("[ %s ]", padded)
							for _, v := range hint.Path {
								if v == index {
									return color.Red.Sprintf(s)
								}
							}
							return s
						}))
					}
					t.Errorf("hintCalculator.GetHints() = (count %d wanted %d) \ngot : %v, \nwant: %v", len(gotHints), len(tt.want), gotHints, wantMap)
				}
			}
			if tt.wantCount != len(gotHints) {
				for i, h := range gotHints {
					t.Logf("Hint %s  :path: %#v", i, h.Path)
				}
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

func BenchmarkHint_Hash(b *testing.B) {
	h := Hint{
		Path: []int{0, 1, 2, 3, 4, 5},
	}
	for i := 0; i < b.N; i++ {
		h.Hash()
	}
}
func BenchmarkHint_Hash_reverse(b *testing.B) {
	h := Hint{
		Path: []int{5, 4, 3, 2, 1, 0},
	}
	for i := 0; i < b.N; i++ {
		h.Hash()
	}
}

func TestHint_Hash(t *testing.T) {
	tests := []struct {
		name              string
		pathA             []int
		pathB             []int
		shouldBeEqual     bool
		shouldBeEqualHash bool
	}{
		{
			"should return equal for equal paths",
			[]int{1, 2, 3},
			[]int{1, 2, 3},
			true,
			true,
		},
		{
			"should not return equal for reversed paths",
			[]int{1, 2, 3},
			[]int{3, 2, 1},
			false,
			false,
		},
		{
			"should return inequal for differing paths",
			[]int{1, 2, 3},
			[]int{2, 3, 1},
			false,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ha := Hint{
				Path: tt.pathA,
			}
			hb := Hint{
				Path: tt.pathB,
			}

			ha.pathHash = ha.Hash()
			hb.pathHash = hb.Hash()
			if tt.shouldBeEqual && !ha.AreEqaul(hb) {
				t.Error("should be equal, but was not:", ha, hb, ha.pathHash, hb.pathHash)
			}
			if !tt.shouldBeEqual && ha.AreEqaul(hb) {
				t.Error("should not be equal, but was:", ha, hb, ha.pathHash, hb.pathHash)
			}
			if tt.shouldBeEqualHash && ha.pathHash != hb.pathHash {
				t.Error("hash should be equal, but was not:", ha.pathHash, hb.pathHash)
			}
			if !tt.shouldBeEqualHash && ha.pathHash == hb.pathHash {
				t.Error("hash should not be equal, but was:", ha.pathHash, hb.pathHash)
			}
		})
	}
}
