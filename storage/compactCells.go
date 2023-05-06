package storage

import (
	"bytes"
	"compress/zlib"
	"context"
	"fmt"
	"io"

	protomodel "github.com/runar-rkmedia/gotally/gen/proto/tally/v1"
	"github.com/runar-rkmedia/gotally/tallylogic/cell"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
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

type Neighbourer interface {
	NeighboursForCellIndex(index int) ([]int, bool)
}

type direction = uint8

const (
	directionUp direction = iota
	directionRight
	directionDown
	directionLeft
)

func combinePathRelative(a, b int) direction {
	// b is to the right of a
	if b == a+1 {
		return directionRight
	}
	// b is to the left of a
	if b == a-1 {
		return directionLeft
	}
	// b is on top of a
	if b < a {
		return directionUp
	}
	// b is below a
	return directionDown
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

func AnnotateSpanError(span trace.Span, err error) {
	if err == nil {
		return
	}
	span.SetStatus(codes.Error, err.Error())
	span.RecordError(err, trace.WithStackTrace(true))
}

func MarshalInternalDataGame(ctx context.Context, seed, state uint64, cells []cell.Cell) ([]byte, error) {
	_, span := tracerMysql.Start(ctx, "MarshalInternalDataGame")
	defer span.End()
	packed := PackCells(cells)
	protocells := protomodel.InternalDataGame{
		Cells: packed,
		State: state,
		Seed:  seed,
	}
	b, err := proto.Marshal(&protocells)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal protocells: %s", err)
	}
	return compressProto(b)
}

func compressProto(b []byte) ([]byte, error) {
	w := bytes.Buffer{}
	z := zlib.NewWriter(&w)
	_, err := z.Write(b)
	if err != nil {
		return nil, fmt.Errorf("failed to compress protocells: %s", err)
	}
	z.Close()
	zb := w.Bytes()
	return zb, err
}

func UnmarshalInternalDataGame(ctx context.Context, b []byte) (cells []cell.Cell, seed uint64, state uint64, err error) {
	_, span := tracerMysql.Start(ctx, "UnmarshalInternalDataGame")
	defer span.End()
	var j protomodel.InternalDataGame
	err = unmarshalCompressedProto(b, &j)
	if err != nil {
		return []cell.Cell{}, 0, 0, err
	}
	return UnpackCells(j.Cells), j.Seed, j.State, nil
}
func unmarshalCompressedProto(b []byte, j proto.Message) error {
	rb := bytes.NewReader(b)
	r := bytes.Buffer{}
	z, err := zlib.NewReader(rb)
	if err != nil {
		return fmt.Errorf("failed to create zlib-reader: %w", err)
	}
	defer z.Close()
	_, err = io.Copy(&r, z)
	if err != nil {
		return fmt.Errorf("failed to copy zlib-reader: %w", err)
	}
	if err != nil {
		return fmt.Errorf("failed to read with zlib: %w", err)
	}
	zb := r.Bytes()

	err = proto.Unmarshal(zb, j)
	if err != nil {
		return fmt.Errorf("failed in protobuf-unmarshal: %w", err)
	}
	return nil
}
