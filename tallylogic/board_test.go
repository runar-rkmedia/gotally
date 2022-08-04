package tallylogic

import (
	"reflect"
	"testing"

	"github.com/gookit/color"
)

func TestTableBoard_neighboursForCellIndex(t *testing.T) {
	type gridSize struct {
		rows    int
		columns int
	}
	type args struct {
		index int
	}
	tests := []struct {
		name   string
		fields gridSize
		args   args
		want   []int
		want1  bool
	}{
		{
			"Neighbours in corner upper-left",
			gridSize{
				3, 4,
			},
			args{
				0,
			},
			[]int{1, 4},
			true,
		},
		{
			"Neighbours in corner upper-right",
			gridSize{
				3, 4,
			},
			args{
				3,
			},
			[]int{2, 7},
			true,
		},
		{
			"Neighbours in corner bottom-right",
			gridSize{
				3, 4,
			},
			args{
				11,
			},
			[]int{7, 10},
			true,
		},
		{
			"Neighbours in corner bottom-left",
			gridSize{
				3, 4,
			},
			args{
				8,
			},
			[]int{4, 9},
			true,
		},
		{
			"Neighbours in corner the middle (6)",
			gridSize{
				3, 4,
			},
			args{
				6,
			},
			[]int{2, 5, 7, 10},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tb := NewTableBoard(tt.fields.columns, tt.fields.rows)
			for i := 0; i < len(tb.cells); i++ {
				tb.cells[i].baseValue = i
			}
			got, got1 := tb.neighboursForCellIndex(tt.args.index)
			t.Log(tb.PrintBoard(func(c Cell, index int, padded string) string {
				p := color.Yellow
				if index == tt.args.index {
					p = color.BgGray
				} else {
					for i := 0; i < len(tt.want); i++ {
						if tt.want[i] == index {
							p = color.BgRed
							for j := 0; j < len(got); j++ {
								if got[j] == index {
									p = color.Blue
								}
							}
						}

					}
				}
				return p.Sprintf("[ %s ]", padded)
			}))
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TableBoard.neighboursForCellIndex(%d) got = %v, want %v", tt.args.index, got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("TableBoard.neighboursForCellIndex() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
