package storage

import (
	"bytes"
	"compress/zlib"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"math/bits"

	protomodel "github.com/runar-rkmedia/gotally/gen/proto/tally/v1"
	"github.com/runar-rkmedia/gotally/tallylogic/cell"
	"go.opentelemetry.io/otel/attribute"
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

func packPath(boardSize int, path []int) (b []byte, err error) {
	if boardSize >= 255 {
		// It is unlikely that a boardSize of 16 by 16 will ever be needed (even 6 by 6 is stretching it)
		return []byte{}, fmt.Errorf("NOT implemented: packPath does not at this time support more than 255 cells")
	}
	bitsNeededForBoard := bits.Len(uint(boardSize))
	x, err := packRelativePath(path)
	if err != nil {
		return
	}
	bitsSpent := bits.Len(uint(x))
	totalBitsNeeded := bitsSpent + bitsNeededForBoard
	if totalBitsNeeded > 64 {
		err = fmt.Errorf("packPath overflow. Current implementation only holds 64 bits. This call requires %d bits. Boardsize: %d, path: %v", totalBitsNeeded, boardSize, path)
		return
	}
	totalBytes := int(math.Ceil(float64(bits.Len(uint(x))) / 8.0))
	xx := make([]byte, 8)

	binary.LittleEndian.PutUint64(xx, x)
	b = append([]byte{byte(path[0]), byte(bitsSpent)}, xx[:totalBytes]...)
	return

}

// Packs a path of indexes.
// Expects that the path is already validated, so that each subsequent element is a "neighbour" of the previous
func packRelativePath(path []int) (x uint64, err error) {
	if len(path) > 32 {
		// TODO: implement support for longer paths
		// For now, this is fine, but will probably bite me in the rear at a later time.
		err = fmt.Errorf("path is longer than expected. Current implementation only supports up to 32 elements")
		return
	}
	// x := uint64(0)
	for i := 1; i < len(path); i++ {
		n := combinePathRelative(path[i-1], path[i])
		z := uint64(0)
		j := (i - 1) * 2
		switch n {
		case directionUp:
		case directionRight:
			x |= 1 << j
			z |= 1 << j
		case directionDown:
			x |= 1 << (j + 1)
			z |= 1 << (j + 1)
		case directionLeft:
			x |= 1<<j | 1<<(j+1)
			z |= 1<<j | 1<<(j+1)
		}
	}
	return

}

func unpackPath(columns, rows int, slice []byte) ([]int, error) {
	if columns*rows >= 255 {
		return []int{}, fmt.Errorf("NOT implemented: unpackPath does not at this time support more than 255 cells")
	}
	startingPosition := int(slice[0])
	bitsUsed := int(slice[1])
	return unpackRelativePath(startingPosition, columns, slice[2:], bitsUsed)
}
func unpackRelativePath(startingPosition int, columns int, slice []byte, bitsNeeded int) ([]int, error) {
	path := make([]int, int(math.Ceil(float64(bitsNeeded)/2))+1)
	prev := startingPosition
	path[0] = prev

outer:
	for i := 0; i < len(slice); i++ {
		b := uint8(slice[i])
		for j := 0; j < 8; j += 2 {
			bitOffset := (i * 8) + j
			pathIndex := (i * 4) + (j / 2) + 1
			if bitOffset > bitsNeeded {
				break outer
			}
			A := b&(1<<(j+1)) != 0
			B := b&(1<<(j)) != 0
			switch {
			// directionLeft
			case A && B:
				prev += -1
			// directionRight
			case !A && B:
				prev += 1
			// directionDown
			case A && !B:
				prev += columns
			// directionUp
			case !A && !B:
				prev += -columns
			}
			path[pathIndex] = prev
		}
	}

	return path, nil

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

func MarshalInternalDataHistory(ctx context.Context, state uint64, cells []cell.Cell, instruction *protomodel.Instruction) (b []byte, err error) {
	_, span := tracerMysql.Start(ctx, "MarshalInternalDataHistory")
	defer func() {
		AnnotateSpanError(span, err)
		span.End()
	}()
	packed := PackCells(cells)
	protocells := protomodel.InternalDataHistory{
		Cells:       packed,
		State:       state,
		Instruction: instruction,
	}
	if c := instruction.GetCombine(); c != nil {
		if len(c.Index) == 0 {
			err := fmt.Errorf("instruction was of type combine, but there were no items in the instruction-set: %#v", instruction)
			return nil, err
		}
		p := make([]int, len(c.Index))
		for i := 0; i < len(c.Index); i++ {
			p[i] = int(c.Index[i])

		}
		span.SetAttributes(attribute.Int("arg.cells.length", len(cells)))
		fmt.Println(cells, p, instruction)
		b, err := packPath(len(cells), p)
		if err != nil {
			return nil, err
		}
		protocells.Instruction = &protomodel.Instruction{
			InstructionOneof: &protomodel.Instruction_Bytes{
				Bytes: b,
			},
		}
	}
	b, err = proto.Marshal(&protocells)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal protocells: %s", err)
	}
	return compressProto(b)
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

func UnmarshalInternalDataHistory(ctx context.Context, b []byte) ([]cell.Cell, uint64, error) {
	_, span := tracerMysql.Start(ctx, "UnmarshalInternalDataHistory")
	defer span.End()
	var j protomodel.InternalDataHistory
	err := unmarshalCompressedProto(b, &j)
	if err != nil {
		return []cell.Cell{}, 0, err
	}
	return UnpackCells(j.Cells), j.State, nil
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
