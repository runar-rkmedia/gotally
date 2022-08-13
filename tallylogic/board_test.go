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
				tb.cells[i].baseValue = int64(i)
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

func TestTableBoard_swipe(t *testing.T) {
	cellCreator := func(vals ...int64) []Cell {
		cells := make([]Cell, len(vals))
		for i, v := range vals {
			cells[i] = NewCell(v, 0)
		}
		return cells
	}
	type fields struct {
		cells   []Cell
		rows    int
		columns int
	}
	tests := []struct {
		name      string
		tb        fields
		direction SwipeDirection
		want      []Cell
	}{
		{
			"Swipe Left",
			fields{
				cellCreator(
					0, 0, 0, 0,
					1, 2, 0, 3,
					0, 0, 0, 0,
					4, 0, 5, 6,
				),
				4, 4,
			},
			SwipeDirectionLeft,
			cellCreator(
				0, 0, 0, 0,
				1, 2, 3, 0,
				0, 0, 0, 0,
				4, 5, 6, 0,
			),
		},
		{
			"Swipe Right",
			fields{
				cellCreator(
					0, 0, 0, 0,
					1, 2, 0, 3,
					0, 0, 0, 0,
					4, 0, 5, 6,
				),
				4, 4,
			},
			SwipeDirectionRight,
			cellCreator(
				0, 0, 0, 0,
				0, 1, 2, 3,
				0, 0, 0, 0,
				0, 4, 5, 6,
			),
		},
		{
			"Swipe Up",
			fields{
				cellCreator(
					0, 0, 0, 0,
					1, 2, 0, 3,
					0, 0, 0, 0,
					4, 0, 5, 6,
				),
				4, 4,
			},
			SwipeDirectionUp,
			cellCreator(
				1, 2, 5, 3,
				4, 0, 0, 6,
				0, 0, 0, 0,
				0, 0, 0, 0,
			),
		},
		{
			"Swipe Down",
			fields{
				cellCreator(
					0, 0, 0, 0,
					1, 2, 0, 3,
					0, 0, 0, 0,
					4, 0, 5, 6,
				),
				4, 4,
			},
			SwipeDirectionDown,
			cellCreator(
				0, 0, 0, 0,
				0, 0, 0, 0,
				1, 0, 0, 3,
				4, 2, 5, 6,
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tb := TableBoard{
				cells:   tt.tb.cells,
				rows:    tt.tb.rows,
				columns: tt.tb.columns,
			}
			got := tb.swipeDirection(tt.direction)
			if len(got) != len(tb.cells) {
				t.Errorf("lengths did not match got %d, expected %d", len(got), len(tb.cells))
			}
			tbGot := TableBoard{
				cells:   got,
				rows:    4,
				columns: 4,
			}.String()
			tbWant := TableBoard{
				cells:   tt.want,
				rows:    tt.tb.rows,
				columns: tt.tb.columns,
			}.String()
			// t.Error("vvv")

			if tbGot != tbWant {
				t.Errorf("TableBoard.swipeDirection(%v) \n from: %v\ngot: %v\nwant: %v", tt.direction, tb.String(), tbGot, tbWant)
			}
		})
	}
}
