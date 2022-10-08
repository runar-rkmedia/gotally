package tallylogic

import (
	"reflect"
	"testing"

	"github.com/gookit/color"
	"github.com/runar-rkmedia/gotally/tallylogic/cell"
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
				tb.cells[i] = cell.NewCell(int64(i), 0) // int64(i)
			}
			got, got1 := tb.NeighboursForCellIndex(tt.args.index)
			t.Log(tb.PrintBoard(func(c CellValuer, index int, padded string) string {
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
	type fields struct {
		cells   []cell.Cell
		rows    int
		columns int
	}
	tests := []struct {
		name      string
		tb        fields
		direction SwipeDirection
		want      []cell.Cell
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
			got := tb.SwipeDirectionPreview(tt.direction)
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

func TestTableBoard_EvaluatesTo(t *testing.T) {
	type fields struct {
		cells   []cell.Cell
		rows    int
		columns int
	}
	type args struct {
		indexes     []int
		targetValue int64
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantNum    int64
		wantMethod EvalMethod
		wantErr    error
	}{
		{
			"Simple sum check",
			fields{
				cellCreator(
					0, 1, 3, 4,
					5, 6, 7, 8,
					8, 9, 10, 11,
					0, 0, 0, 0,
				),
				4, 4,
			},
			args{
				[]int{1, 2, 3},
				8,
			},
			8,
			EvalMethodSum,
			nil,
		},
		{
			"Simple product check",
			fields{
				cellCreator(
					0, 1, 3, 4,
					5, 6, 7, 8,
					8, 9, 10, 11,
					0, 0, 0, 0,
				),
				4, 4,
			},
			args{
				[]int{1, 2, 3},
				12,
			},
			12,
			EvalMethodProduct,
			nil,
		},
		{
			"Simple overshot check",
			fields{
				cellCreator(
					0, 1, 3, 4,
					5, 6, 7, 8,
					8, 9, 10, 11,
					0, 0, 0, 0,
				),
				4, 4,
			},
			args{
				[]int{1, 2, 3},
				-2,
			},
			0,
			EvalMethodNil,
			ErrResultOvershot,
		},
		{
			"Invalid path check (too few)",
			fields{
				cellCreator(
					0, 1, 3, 4,
					5, 6, 7, 8,
					8, 9, 10, 11,
					0, 0, 0, 0,
				),
				4, 4,
			},
			args{
				[]int{},
				2,
			},
			0,
			EvalMethodNil,
			ErrResultInvalidCount,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tb := TableBoard{
				cells:   tt.fields.cells,
				rows:    tt.fields.rows,
				columns: tt.fields.columns,
			}
			gotNum, gotMethod, err := tb.SoftEvaluatesTo(tt.args.indexes, tt.args.targetValue)
			if err != tt.wantErr {
				t.Errorf("TableBoard.EvaluatesTo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotNum != tt.wantNum {
				t.Errorf("TableBoard.EvaluatesTo() got = %v, want %v", gotNum, tt.wantNum)
			}
			if !reflect.DeepEqual(gotMethod, tt.wantMethod) {
				t.Errorf("TableBoard.EvaluatesTo() Method: got1 = %v, want %v", gotMethod, tt.wantMethod)
			}
		})
	}
}
func Benchmark_TableBoardHash(b *testing.B) {
	tb := NewTableBoard(5, 5, TableBoardOptions{
		Cells: cellCreator(
			1, 2, 3, 4, 5,
			1, 2, 3, 4, 5,
			1, 2, 3, 4, 5,
			1, 2, 3, 4, 5,
			1, 2, 3, 4, 5,
		),
	})
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			tb.Hash()
		}
	})
}
func Benchmark_TableBoardSwipeVertical(b *testing.B) {
	tb := NewTableBoard(5, 5, TableBoardOptions{
		Cells: cellCreator(
			0, 1, 3, 0, 0,
			0, 3, 4, 0, 0,
			0, 0, 0, 3, 0,
			1, 2, 3, 4, 5,
		),
	})
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			tb.swipeVertical(false)
		}
	})
}
func Benchmark_TableBoardSwipeHorizontal(b *testing.B) {
	tb := NewTableBoard(5, 5, TableBoardOptions{
		Cells: cellCreator(
			0, 1, 3, 0, 0,
			0, 3, 4, 0, 0,
			0, 0, 0, 3, 0,
			1, 2, 3, 4, 5,
		),
	})
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			tb.swipeHorizontal(true)
		}
	})
}
