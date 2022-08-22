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

func (tb *TableBoard) Copy() BoardController {
	board := TableBoard{
		cells:   make([]Cell, len(tb.cells)),
		rows:    tb.rows,
		columns: tb.columns,
	}
	for i, v := range tb.cells {
		board.cells[i] = NewCell(v.baseValue, v.power)
	}
	return &board

}
func (tb TableBoard) GetCellAtIndex(n int) *Cell {
	if tb.ValidCellIndex(n) {
		return &tb.cells[n]
	}
	return nil
}
func (tb TableBoard) FindCell(c Cell) (int, bool) {
	for i := 0; i < len(tb.cells); i++ {
		if tb.cells[i].ID == c.ID {
			return i, true
		}
	}
	return 0, false
}
func (tb TableBoard) String() string {
	return tb.PrintBoard(nil)
}
func (tb TableBoard) HighestValue() int64 {
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
	withColors := false
	h := int64(tb.HighestValue())
	longest := len(strconv.FormatInt(h, 10))
	headerPrinter := fmt.Sprintf
	if withColors {
		headerPrinter = color.Gray.Sprintf
	}

	s := "\n------ Table -------"
	s += "\n    "
	for j := 0; j < tb.columns; j++ {
		formatted := strconv.FormatInt(int64(j), 10)
		padLength := longest - len(formatted)
		padding := strings.Repeat(" ", padLength)
		valueStr := padding + formatted
		s += headerPrinter("  %s: ", valueStr)

	}
	for i := 0; i < tb.rows; i++ {
		s += "\n" + headerPrinter(" %d: ", i)
		for j := 0; j < tb.columns; j++ {
			index := ((i + 1) * tb.columns) - tb.columns + ((j + 1) + tb.rows) - tb.rows - 1
			value := tb.cells[index].Value()
			var formatted string
			if value > 0 {
				formatted = strconv.FormatInt(int64(value), 10)
			}
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

// Uniqely identifies the board by its value. ID's etc are ignored.
func (tb TableBoard) Hash() string {
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("%dx%d;", tb.columns, tb.rows))
	for _, v := range tb.cells {
		builder.WriteString(v.Hash() + ";")
	}
	return builder.String()
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
func (tb TableBoard) ValidCellIndex(index int) bool {
	if index < 0 {
		return false
	}
	if index >= len(tb.cells) {
		return false
	}
	return true
}
func (tb TableBoard) CoordToIndex(x, y int) (int, bool) {
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
	ErrResultInvalidCount     = errors.New("Too few items in the index")
	ErrResultIndexOverflow    = errors.New("Index overflow")
	ErrResultNoCell           = errors.New("No cell at index")
	ErrResultOvershot         = errors.New("The path evaluated to a higher value than the targetValue")
	ErrResultNoMatch          = errors.New("The path returned no result")
	ErrPathIndexDuplicate     = errors.New("The path includes duplicates")
	ErrPathTooLong            = errors.New("The path is too long")
	ErrPathTooShort           = errors.New("The path must consist of at least 2 items")
	ErrPathIndexOutsideBounds = errors.New("The path includes an item outside the current bounds")
	ErrPathIndexInvalidCell   = errors.New("The path-index pointed to an invalid cell")
	ErrPathIndexEmptyCell     = errors.New("The path-index pointed to an empty cell")

	ErrIndexInvalid               = errors.New("The index for the path is invalid")
	ErrCellAtIndexAlreadyHasValue = errors.New("The cell at the index already has a value")
)

func (tb TableBoard) ValidatePath(indexes []int) (err error, invalidIndex int) {
	nIndexes := len(indexes)
	if nIndexes <= 1 {
		return fmt.Errorf("%w at length %d", ErrPathTooShort, nIndexes), -1
	}
	nCells := len(tb.cells)
	if nIndexes > nCells {
		return fmt.Errorf("%w with length %d of maximum %d", ErrPathTooLong, nIndexes, nCells), -1
	}
	seen := map[int]int{}
	prevIndex := -1
	for i, index := range indexes {
		if duplicate, ok := seen[index]; ok {
			return fmt.Errorf("%w for index %d at position %d / %d", ErrPathIndexDuplicate, index, duplicate, i), i
		}
		if index > nCells {
			return fmt.Errorf("%w for index %d at position %d", ErrPathIndexOutsideBounds, index, i), i
		}
		if index < 0 {
			return fmt.Errorf("%w for index %d at position %d", ErrPathIndexOutsideBounds, index, i), i
		}
		c := tb.cells[index]
		if c.ID == "" {
			return fmt.Errorf("%w for index %d at position %d", ErrPathIndexInvalidCell, index, i), i
		}
		if c.Value() == 0 {
			return fmt.Errorf("%w for index %d at position %d", ErrPathIndexEmptyCell, index, i), i
		}
		if prevIndex >= 0 && !tb.AreNeighboursByIndex(index, prevIndex) {
			return fmt.Errorf("Not a neighbour %d %d", index, prevIndex), i

		}
		seen[index] = i
		prevIndex = index
	}
	return nil, 0
}

func (tb TableBoard) EvaluatesTo(indexes []int, commitResultToBoard bool, noValidate bool) (int64, EvalMethod, error) {
	// debug := len(indexes) == 4 && indexes[0] == 0 && indexes[1] == 1 && indexes[2] == 2 && indexes[3] == 5
	if !noValidate {
		err, _ := tb.ValidatePath(indexes)
		if err != nil {
			return 0, EvalMethodNil, err
		}
	}
	last := indexes[len(indexes)-1]
	rest := indexes[0 : len(indexes)-1]
	v, method, err := tb.SoftEvaluatesTo(rest, tb.cells[last].Value())
	if err != nil {
		return v, method, err
	}
	if method == EvalMethodNil {
		return v, method, err
	}
	if v == 0 || !commitResultToBoard {
		return v, method, err
	}
	tb.GetCellAtIndex(last).Double()
	for _, index := range rest {
		tb.cells[index] = NewEmptyCell()
	}
	return v, method, err
}

// Evaluates whether a path of indexes results in the targetValue.
// This method should ideally be as performant as possible, as it will be run in loops.
func (tb TableBoard) SoftEvaluatesTo(indexes []int, targetValue int64) (int64, EvalMethod, error) {
	cellCount := len(tb.cells)
	if len(indexes) == 0 {
		return 0, EvalMethodNil, ErrResultInvalidCount
	}
	if len(indexes) > cellCount {
		return 0, EvalMethodNil, ErrResultInvalidCount
	}
	// TODO: Check whether the path has duplicates
	// return 0, EvalResultInvalidNonUnique
	var sum int64
	var product int64 = 1
	for _, index := range indexes {
		if index > cellCount {
			return 0, EvalMethodNil, ErrResultIndexOverflow
		}
		cell := tb.cells[index]
		if cell.ID == "" {
			return 0, EvalMethodNil, ErrResultIndexOverflow
		}
		v := cell.Value()
		sum += v
		product *= v

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
				index, ok = tb.CoordToIndex(colIndex, rowIndex)
			} else {
				index, ok = tb.CoordToIndex(rowIndex, colIndex)
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

func (tb *TableBoard) SwipeDirection(direction SwipeDirection) bool {
	newCells := tb.SwipeDirectionPreview(direction)
	changed := false

	// Find if anything actually changed, ignoring empty cells
outer:
	for i := 0; i < len(tb.cells); i++ {
		if tb.cells[i].Value() == 0 {
			continue
		}
		for j := 0; j < len(newCells); j++ {
			if tb.cells[j].Value() == 0 {
				continue
			}
			if tb.cells[i].ID != newCells[i].ID {
				changed = true
				break outer
			}

		}

	}
	if !changed {
		return false
	}
	tb.cells = newCells
	return changed
}
func (tb *TableBoard) SwipeDirectionPreview(direction SwipeDirection) []Cell {
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
func (tb *TableBoard) Cells() []Cell {
	return tb.cells
}
func (tb *TableBoard) swipeHorizontal(positive bool) []Cell {
	rows := tb.getRows()
	tiles := make([]Cell, len(tb.cells))
	for r, row := range rows {
		sort.Slice(row, func(i, j int) bool {
			if positive {
				return row[i].Value() == 0
			}
			return row[j].Value() == 0
		})
		for i, cell := range row {
			tiles[i+tb.rows*r] = cell
		}
	}
	return tiles
}
func (tb *TableBoard) swipeVertical(positive bool) []Cell {
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

func (tb TableBoard) AreNeighboursByIndex(a, b int) bool {
	if a < 0 || b < 0 {
		return false
	}
	max := len(tb.cells)
	if a >= max || b >= max {
		return false
	}
	ac, ar := tb.indexToCord(a)
	bc, br := tb.indexToCord(b)

	diff := (ac - bc) + (ar - br)

	if diff != 1 && diff != -1 {
		return false
	}
	return true
}

// Returns the neighbours for a gives cell. Note that the cells might be empty
func (tb TableBoard) NeighboursForCellIndex(index int) ([]int, bool) {
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

func (tb *TableBoard) AddCellToBoard(c Cell, index int, overwrite bool) error {
	if !tb.ValidCellIndex(index) {
		return ErrIndexInvalid
	}
	if !overwrite && tb.cells[index].Value() > 0 {
		return ErrCellAtIndexAlreadyHasValue
	}
	tb.cells[index] = c
	return nil

}
