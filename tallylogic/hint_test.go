package tallylogic

import (
	"reflect"
	"testing"
)

func Test_hintCalculator_GetHints(t *testing.T) {
	tests := []struct {
		name  string
		board TableBoard
		want  []Hint
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
					Path:   []int{2, 5},
				},
				{
					Value:  6,
					Method: EvalMethodSum,
					Path:   []int{3, 4, 5},
				},
			},
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
			// There may be more hints, but there are at least one.
			[]Hint{
				{
					Value:  20,
					Method: EvalMethodSum,
					// The bottom row , 10, 4, 6,
					Path: []int{22, 21, 20},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &hintCalculator{
				CellRetriever:      &tt.board,
				NeighbourRetriever: tt.board,
				Evaluator:          tt.board,
			}
			if got := g.GetHints(); !reflect.DeepEqual(got, tt.want) {
				for i, v := range got {
					t.Logf("hint %d %v", i, v)

				}
				t.Errorf("hintCalculator.GetHints() = (count %d wanted %d) %v, want %v", len(got), len(tt.want), got, tt.want)
			}
		})
	}
}

func Test_hintCalculator_getHints(t *testing.T) {
	type fields struct {
		CellRetriever      CellRetriever
		NeighbourRetriever NeighbourRetriever
		Evaluator          Evaluator
	}
	type args struct {
		valueForIndexMap map[int]int64
		path             []int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []Hint
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &hintCalculator{
				CellRetriever:      tt.fields.CellRetriever,
				NeighbourRetriever: tt.fields.NeighbourRetriever,
				Evaluator:          tt.fields.Evaluator,
			}
			if got := g.getHints(tt.args.valueForIndexMap, tt.args.path); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("hintCalculator.getHints() = %v, want %v", got, tt.want)
			}
		})
	}
}
