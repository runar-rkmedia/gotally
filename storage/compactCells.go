package storage

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"

	protomodel "github.com/runar-rkmedia/gotally/gen/proto/tally/v1"
	"github.com/runar-rkmedia/gotally/tallylogic/cell"
	"google.golang.org/protobuf/proto"
)

// Packs the cells into a structure slice where the first half
// is for the cells base-value, while the last half is for their 'twoPow'-value
func PackCells(cells []cell.Cell) []int64 {
	length := len(cells)
	n := make([]int64, length*2)
	for i := 0; i < length; i++ {
		base, twoPow := cells[i].Raw()
		n[i] = base
		n[i+length] = twoPow
	}
	return n
}

// Unpacks a previously packed set of cells
func UnpackCells(m []int64) []cell.Cell {
	length := len(m)
	cellCount := length / 2
	cells := make([]cell.Cell, cellCount)

	for i := 0; i < cellCount; i++ {
		base := m[i]
		twoPow := m[i+cellCount]
		cells[i] = cell.NewCell(base, int(twoPow))
	}
	return cells
}

// Packs, marshals and compresses cellvalues.
// Note that this ignores other values of the cell, like the ID.
// The ID normally does not matter, and is only used by clients to track animation across changes.
func MarshalCellValues(cells []cell.Cell) ([]byte, error) {
	packed := PackCells(cells)
	protocells := protomodel.CompactCells{
		Cells: packed,
	}
	b, err := proto.Marshal(&protocells)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal protocells: %s", err)
	}

	w := bytes.Buffer{}
	z := zlib.NewWriter(&w)
	_, err = z.Write(b)
	if err != nil {
		return nil, fmt.Errorf("failed to compress protocells: %s", err)
	}
	z.Close()

	zb := w.Bytes()
	return zb, err
}

func UnmarshalCellValues(b []byte) ([]cell.Cell, error) {
	rb := bytes.NewReader(b)
	r := bytes.Buffer{}
	z, err := zlib.NewReader(rb)
	if err != nil {
		return []cell.Cell{}, fmt.Errorf("failed to create zlib-reader: %w", err)
	}
	defer z.Close()
	_, err = io.Copy(&r, z)
	if err != nil {
		return []cell.Cell{}, fmt.Errorf("failed to copy zlib-reader: %w", err)
	}
	// n, err := z.Read(b)
	if err != nil {
		return []cell.Cell{}, fmt.Errorf("failed to read with zlib: %w", err)
	}
	zb := r.Bytes()

	var m protomodel.CompactCells
	err = proto.Unmarshal(zb, &m)
	if err != nil {
		return []cell.Cell{}, fmt.Errorf("failed in protobuf-unmarshal: %w", err)
	}
	return UnpackCells(m.Cells), nil
}
