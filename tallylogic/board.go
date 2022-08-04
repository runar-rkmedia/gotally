package tallylogic

import (
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
func (tb TableBoard) highestValue() int {
	var v int
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

			var c Sprinter
			if highlighter != nil {
				s += highlighter(tb.cells[index], index, valueStr)
				continue
			}
			s += c.Sprintf("[ %s ]", valueStr)
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

	return y*tb.columns + x, false

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
