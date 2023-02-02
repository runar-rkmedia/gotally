package storage

import (
	"context"
	"math/rand"
	"reflect"
	"testing"

	tallyv1 "github.com/runar-rkmedia/gotally/gen/proto/tally/v1"
	"github.com/runar-rkmedia/gotally/tallylogic/cell"
)

func TestMarshalCells(t *testing.T) {
	tests := []struct {
		name     string
		cells    []cell.Cell
		wantSize int
		wantErr  bool
	}{
		{
			"test",
			[]cell.Cell{
				cell.NewCell(3, 4),
				cell.NewCell(3, 4),
				cell.NewCell(3, 4),
				cell.NewCell(3, 4),
				cell.NewCell(3, 4),
				cell.NewCell(3, 4),
				cell.NewCell(3, 4),
				cell.NewCell(3, 4),
				cell.NewCell(3, 4),
				cell.NewCell(3, 4),
				cell.NewCell(3, 4),
				cell.NewCell(3, 4),
				cell.NewCell(3, 4),
				cell.NewCell(3, 4),
				cell.NewCell(3, 4),
				cell.NewCell(3, 4),
			},
			24,
			false,
		},
		{
			"test randomized",
			[]cell.Cell{
				cell.NewCell(rand.Int63n(12+1), rand.Intn(10)),
				cell.NewCell(rand.Int63n(12+1), rand.Intn(10)),
				cell.NewCell(rand.Int63n(12+1), rand.Intn(10)),
				cell.NewCell(rand.Int63n(12+1), rand.Intn(10)),
				cell.NewCell(rand.Int63n(12+1), rand.Intn(10)),
				cell.NewCell(rand.Int63n(12+1), rand.Intn(10)),
				cell.NewCell(rand.Int63n(12+1), rand.Intn(10)),
				cell.NewCell(rand.Int63n(12+1), rand.Intn(10)),
				cell.NewCell(rand.Int63n(12+1), rand.Intn(10)),
				cell.NewCell(rand.Int63n(12+1), rand.Intn(10)),
				cell.NewCell(rand.Int63n(12+1), rand.Intn(10)),
				cell.NewCell(rand.Int63n(12+1), rand.Intn(10)),
				cell.NewCell(rand.Int63n(12+1), rand.Intn(10)),
				cell.NewCell(rand.Int63n(12+1), rand.Intn(10)),
				cell.NewCell(rand.Int63n(12+1), rand.Intn(10)),
				cell.NewCell(rand.Int63n(12+1), rand.Intn(10)),
				cell.NewCell(rand.Int63n(12+1), rand.Intn(10)),
				cell.NewCell(rand.Int63n(12+1), rand.Intn(10)),
				cell.NewCell(rand.Int63n(12+1), rand.Intn(10)),
				cell.NewCell(rand.Int63n(12+1), rand.Intn(10)),
				cell.NewCell(rand.Int63n(12+1), rand.Intn(10)),
				cell.NewCell(rand.Int63n(12+1), rand.Intn(10)),
				cell.NewCell(rand.Int63n(12+1), rand.Intn(10)),
				cell.NewCell(rand.Int63n(12+1), rand.Intn(10)),
				cell.NewCell(rand.Int63n(12+1), rand.Intn(10)),
				cell.NewCell(rand.Int63n(12+1), rand.Intn(10)),
			},
			-1,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var seed uint64 = 123
			var state uint64 = 456
			got, err := MarshalInternalDataGame(context.TODO(), seed, state, tt.cells)
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalInternalDataGame() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantSize != -1 && len(got) != tt.wantSize {
				t.Errorf("MarshalInternalDataGame() = %v, want %v (%v)", len(got), tt.wantSize, got)
			}
			unmarshalled, gotseed, gotstate, err := UnmarshalInternalDataGame(context.TODO(), got)
			if err != nil {
				t.Errorf("failed to unmarshal: %s", err)
			}
			if gotseed != seed {
				t.Errorf("UnmarshalInternalDataGame() seed is , want %v (%v)", gotseed, seed)
			}
			if gotstate != state {
				t.Errorf("UnmarshalInternalDataGame() state is , want %v (%v)", gotstate, state)
			}
			// The ids do not matter, so we just compare the values
			uValues := make([]int64, len(unmarshalled))
			for i := 0; i < len(unmarshalled); i++ {
				uValues[i] = unmarshalled[i].Value()
			}
			wValues := make([]int64, len(unmarshalled))
			for i := 0; i < len(tt.cells); i++ {
				wValues[i] = tt.cells[i].Value()
			}
			if !reflect.DeepEqual(uValues, wValues) {
				t.Errorf("UnmarshalCells() = %#v, want %v", uValues, wValues)
			}
		})
	}
}

// func TestPackCombin_eInstruction(t *testing.T) {
// 	type args struct {
// 		boardSize int
// 		path      []int
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want []byte
// 	}{
// 		{

// 			"common case: path-length 3",
// 			args{8 * 8, []int{24, 9, 10, 5, 4, 3, 8, 13, 14, 19, 24, 25}},
// 			[]byte{},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got, err := packRelativePath(tt.args.path); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("PackCombineInstruction() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

func Test_packRelativePath(t *testing.T) {
	tests := []struct {
		name        string
		columns     int
		rows        int
		path        []int
		bytesNeeded int
		wantErr     bool
	}{
		// TODO: Add test cases.
		{
			"should unpack and return original path",
			5,
			5,
			[]int{
				9,
				// RIGHT
				10,
				// UP
				5,
				// LEFT
				4,
				// LEFT
				3,
				// DOWN
				8,
				// DOWN
				13,
				// RIGHT
				14,
				// DOWN
				19,
				// DOWN
				24,
				// RIGHT
				25},
			5,
			false,
		},
		{
			"should work with short paths",
			5, 5,
			[]int{1, 2},
			3,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := packPath(tt.columns*tt.rows, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("packRelativePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(b) != tt.bytesNeeded {
				t.Errorf("packRelativePath() bytesNeeded = %v, want %v", len(b), tt.bytesNeeded)
			}
			got, err := unpackPath(tt.columns, tt.rows, b)
			if (err != nil) != tt.wantErr {
				t.Errorf("unpackRelativePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.path) {
				t.Errorf("unpackRelativePath() = %v, want %v", got, tt.path)
			}
		})
	}
}

func TestMarshalHistory(t *testing.T) {
	tests := []struct {
		name     string
		cells    []cell.Cell
		wantSize int
		wantErr  bool
	}{
		{
			"test simple",
			[]cell.Cell{
				cell.NewCell(3, 4),
				cell.NewCell(3, 4),
				cell.NewCell(3, 4),
				cell.NewCell(3, 4),
				cell.NewCell(3, 4),
				cell.NewCell(3, 4),
				cell.NewCell(3, 4),
				cell.NewCell(3, 4),
				cell.NewCell(3, 4),
				cell.NewCell(3, 4),
				cell.NewCell(3, 4),
				cell.NewCell(3, 4),
				cell.NewCell(3, 4),
				cell.NewCell(3, 4),
				cell.NewCell(3, 4),
				cell.NewCell(3, 4),
			},
			31,
			false,
		},
		{
			"test randomized",
			[]cell.Cell{
				cell.NewCell(rand.Int63n(12+1), rand.Intn(10)),
				cell.NewCell(rand.Int63n(12+1), rand.Intn(10)),
				cell.NewCell(rand.Int63n(12+1), rand.Intn(10)),
				cell.NewCell(rand.Int63n(12+1), rand.Intn(10)),
				cell.NewCell(rand.Int63n(12+1), rand.Intn(10)),
				cell.NewCell(rand.Int63n(12+1), rand.Intn(10)),
				cell.NewCell(rand.Int63n(12+1), rand.Intn(10)),
				cell.NewCell(rand.Int63n(12+1), rand.Intn(10)),
				cell.NewCell(rand.Int63n(12+1), rand.Intn(10)),
				cell.NewCell(rand.Int63n(12+1), rand.Intn(10)),
				cell.NewCell(rand.Int63n(12+1), rand.Intn(10)),
				cell.NewCell(rand.Int63n(12+1), rand.Intn(10)),
				cell.NewCell(rand.Int63n(12+1), rand.Intn(10)),
				cell.NewCell(rand.Int63n(12+1), rand.Intn(10)),
				cell.NewCell(rand.Int63n(12+1), rand.Intn(10)),
				cell.NewCell(rand.Int63n(12+1), rand.Intn(10)),
				cell.NewCell(rand.Int63n(12+1), rand.Intn(10)),
				cell.NewCell(rand.Int63n(12+1), rand.Intn(10)),
				cell.NewCell(rand.Int63n(12+1), rand.Intn(10)),
				cell.NewCell(rand.Int63n(12+1), rand.Intn(10)),
				cell.NewCell(rand.Int63n(12+1), rand.Intn(10)),
				cell.NewCell(rand.Int63n(12+1), rand.Intn(10)),
				cell.NewCell(rand.Int63n(12+1), rand.Intn(10)),
				cell.NewCell(rand.Int63n(12+1), rand.Intn(10)),
				cell.NewCell(rand.Int63n(12+1), rand.Intn(10)),
				cell.NewCell(rand.Int63n(12+1), rand.Intn(10)),
			},
			-1,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var state uint64 = 456
			got, err := MarshalInternalDataHistory(context.TODO(), state, tt.cells, &tallyv1.Instruction{
				InstructionOneof: &tallyv1.Instruction_Combine{
					Combine: &tallyv1.Indexes{
						Index: []uint32{1, 2, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14},
					},
				},
			})
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalInternalDataHistory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantSize != -1 && len(got) != tt.wantSize {
				t.Errorf("MarshalInternalDataHistory() = %d, want %d (%x)", len(got), tt.wantSize, got)
			}
			unmarshalled, gotstate, err := UnmarshalInternalDataHistory(context.TODO(), got)
			if err != nil {
				t.Errorf("failed to unmarshal: %s", err)
			}
			if gotstate != state {
				t.Errorf("UnmarshalInternalDataHistory() state is , want %v (%v)", gotstate, state)
			}
			// The ids do not matter, so we just compare the values
			uValues := make([]int64, len(unmarshalled))
			for i := 0; i < len(unmarshalled); i++ {
				uValues[i] = unmarshalled[i].Value()
			}
			wValues := make([]int64, len(unmarshalled))
			for i := 0; i < len(tt.cells); i++ {
				wValues[i] = tt.cells[i].Value()
			}
			if !reflect.DeepEqual(uValues, wValues) {
				t.Errorf("UnmarshalCells() = %#v, want %v", uValues, wValues)
			}
		})
	}
}
