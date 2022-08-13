package tallylogic

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/gookit/color"
)

type TableBoard struct {
	cells   []Cell
	rows    int
	columns int
}

func NewTableBoard(columns, rows int) TableBoard {
	tb := TableBoard{
		rows:    rows,
		columns: columns,
	}
	cellCount := columns * rows
	tb.cells = make([]Cell, cellCount)
	for i := 0; i < cellCount; i++ {
		tb.cells[i] = NewCell(0, 0)
	}
	return tb
}

func (tb TableBoard) FindCell(c Cell) (int, bool) {
	for i := 0; i < len(tb.cells); i++ {
		if tb.cells[i].id == c.id {
			return i, true
		}
	}
	return 0, false
}
func (tb TableBoard) String() string {
	return tb.PrintBoard(nil)
}
func (tb TableBoard) highestValue() int64 {
	var v int64
	for i := 0; i < len(tb.cells); i++ {
		val := tb.cells[i].Value()
		if val > v {
			v = val
		}
	}
	return v
}

type Sprinter interface {
	Sprintf(format string, a ...any) string
}

func (tb TableBoard) PrintBoard(highlighter func(c Cell, index int, padded string) string) string {
	h := int64(tb.highestValue())
	longest := len(strconv.FormatInt(h, 10))

	s := "\n------ Table -------"
	s += "\n    "
	for j := 0; j < tb.columns; j++ {
		formatted := strconv.FormatInt(int64(j), 10)
		padLength := longest - len(formatted)
		padding := strings.Repeat(" ", padLength)
		valueStr := padding + formatted
		s += color.Gray.Sprintf("  %s: ", valueStr)

	}
	for i := 0; i < tb.rows; i++ {
		s += "\n" + color.Gray.Sprintf(" %d: ", i)
		for j := 0; j < tb.columns; j++ {
			index := ((i + 1) * tb.columns) - tb.columns + ((j + 1) + tb.rows) - tb.rows - 1
			value := tb.cells[index].Value()
			formatted := strconv.FormatInt(int64(value), 10)
			padLength := longest - len(formatted)

			padding := strings.Repeat(" ", padLength)
			valueStr := padding + formatted

			// var c Sprinter
			if highlighter != nil {
				s += highlighter(tb.cells[index], index, valueStr)
				continue
			}
			s += fmt.Sprintf("[ %s ]", valueStr)
		}
	}
	s += "\n---- End Table -----"
	return s
}

func (tb TableBoard) cellRow(i int) int {
	return i / tb.columns
}
func (tb TableBoard) cellColumn(i int) int {
	return i % tb.columns
}
func (tb TableBoard) indexToCord(i int) (column int, row int) {
	return tb.cellColumn(i), tb.cellRow(i)
}
func (tb TableBoard) coordToIndex(x, y int) (int, bool) {
	if x < 0 {
		return 0, false
	}
	if y < 0 {
		return 0, false
	}
	if y > tb.rows {
		return 0, false
	}
	if x > tb.columns {
		return 0, false
	}

	return y*tb.columns + x, true

}

type EvalMethod = int

const (
	EvalMethodNil EvalMethod = iota
	EvalMethodSum
	EvalMethodProduct
)

var (
	ErrResultInvalidCount  = errors.New("Too few items in the index")
	ErrResultIndexOverflow = errors.New("Index overflow")
	ErrResultNoCell        = errors.New("No cell at index")
	ErrResultOvershot      = errors.New("The path evaluated to a higher value than the targetValue")
	ErrResultNoMatch       = errors.New("The path evaluated to a higher value than the targetValue")
)

// Evaluates whether a path of indexes results in the targetValue.
// This method should ideally be as performant as possible, as it will be run in loops.
func (tb TableBoard) EvaluatesTo(indexes []int, targetValue int64) (int64, EvalMethod, error) {
	cellCount := len(tb.cells)
	if len(indexes) < 2 {
		return 0, EvalMethodNil, ErrResultInvalidCount
	}
	if len(indexes) > cellCount {
		return 0, EvalMethodNil, ErrResultInvalidCount
	}
	// TODO: Check whether the path has duplicates
	// return 0, EvalResultInvalidNonUnique
	var sum int64
	var product int64
	for _, index := range indexes {
		if index > cellCount {
			return 0, EvalMethodNil, ErrResultIndexOverflow
		}
		cell := tb.cells[index]
		if cell.id == "" {
			return 0, EvalMethodNil, ErrResultIndexOverflow
		}
		sum += cell.Value()

		// return early if we have overshow the targetValue
		if sum > int64(targetValue) && product > targetValue {
			return 0, EvalMethodNil, ErrResultOvershot
		}
	}
	if sum == targetValue {
		return sum, EvalMethodSum, nil
	}
	if product == targetValue {
		return product, EvalMethodProduct, nil
	}

	return 0, EvalMethodNil, ErrResultNoMatch
}

func (tb TableBoard) getRows() [][]Cell {
	return tb._getColumnsOrRows(true)
}
func (tb TableBoard) getColumns() [][]Cell {
	return tb._getColumnsOrRows(false)
}
func (tb TableBoard) _getColumnsOrRows(rows bool) [][]Cell {
	var columns = make([][]Cell, tb.rows)
	for rowIndex := 0; rowIndex < tb.rows; rowIndex++ {
		columns[rowIndex] = make([]Cell, tb.columns)
		for colIndex := 0; colIndex < tb.columns; colIndex++ {
			var index int
			var ok bool
			if rows {
				index, ok = tb.coordToIndex(colIndex, rowIndex)
			} else {
				index, ok = tb.coordToIndex(rowIndex, colIndex)
			}
			if !ok {
				panic(fmt.Sprintf("overflow in getRows %d %d", rowIndex, colIndex))
			}
			columns[rowIndex][colIndex] = tb.cells[index]
		}
	}
	return columns
}

type SwipeDirection string

const (
	SwipeDirectionUp    SwipeDirection = "Up"
	SwipeDirectionRight                = "Right"
	SwipeDirectionDown                 = "Down"
	SwipeDirectionLeft                 = "Left"
)

func (tb TableBoard) swipeDirection(direction SwipeDirection) []Cell {
	switch direction {
	case SwipeDirectionUp:
		return tb.swipeVertical(false)
	case SwipeDirectionRight:
		return tb.swipeHorizontal(true)
	case SwipeDirectionDown:
		return tb.swipeVertical(true)
	case SwipeDirectionLeft:
		return tb.swipeHorizontal(false)
	}
	return tb.cells

}
func (tb TableBoard) swipeHorizontal(positive bool) []Cell {
	rows := tb.getRows()
	var tiles []Cell
	for _, row := range rows {
		sort.Slice(row, func(i, j int) bool {
			if positive {
				return row[i].Value() == 0
			}
			return row[j].Value() == 0
		})
		tiles = append(tiles, row...)
	}
	return tiles
}
func (tb TableBoard) swipeVertical(positive bool) []Cell {
	columns := tb.getColumns()
	tiles := make([]Cell, len(tb.cells))
	for c, column := range columns {
		sort.Slice(column, func(i, j int) bool {
			if positive {
				return column[i].Value() == 0
			}
			return column[j].Value() == 0
		})
		for i, cell := range column {
			tiles[i*tb.columns+c] = cell
		}
	}
	return tiles
}

func (tb TableBoard) neighboursForCellIndex(index int) ([]int, bool) {
	var neighbours []int

	if index < 0 {
		return neighbours, false
	}
	if index >= len(tb.cells) {
		return neighbours, false
	}

	column, row := tb.indexToCord(index)

	// Neighbour above
	if row > 0 {
		neighbours = append(neighbours, index-tb.columns)
	}
	// Neighbour to the left
	if column > 0 {
		neighbours = append(neighbours, index-1)
	}
	// Neighbour to the right
	if column < tb.columns-1 {
		neighbours = append(neighbours, index+1)
	}
	// Neighbour below
	if row < tb.rows-1 {
		neighbours = append(neighbours, index+tb.columns)
	}
	// This should now be sorted, becouse of the ordering above
	return neighbours, true
}
func (tb TableBoard) NeighboursForCell(c Cell) ([]Cell, bool) {
	index, ok := tb.FindCell(c)
	if !ok {
		return []Cell{}, false
	}
	indexes, ok := tb.neighboursForCellIndex(index)
	if !ok {
		return []Cell{}, false
	}
	neighbours := make([]Cell, len(indexes))
	for i := 0; i < len(indexes); i++ {
		neighbours[i] = tb.cells[i]

	}
	return neighbours, true
}
