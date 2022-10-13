package storage

import (
	"context"
	"math/rand"
	"reflect"
	"testing"

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
			"test",
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
			62,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var seed uint64 = 123
			var state uint64 = 456
			got, err := MarshalInternalDataGame(context.TODO(), seed, state, tt.cells)
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalInternalDataHistory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(len(got), tt.wantSize) {
				t.Errorf("MarshalInternalDataHistory() = %v, want %v (%v)", len(got), tt.wantSize, got)
			}
			unmarshalled, gotseed, gotstate, err := UnmarshalInternalDataGame(context.TODO(), got)
			if err != nil {
				t.Errorf("failed to unmarshal: %s", err)
			}
			if gotseed != seed {
				t.Errorf("UnmarshalInternalDataHistory() seed is , want %v (%v)", gotseed, seed)
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
