package tallylogic

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/gookit/color"
	"github.com/runar-rkmedia/gotally/tallylogic/cell"
)

type TableBoard struct {
	cells   []cell.Cell
	rows    int
	columns int
	// Only for saved boards
	id string
	TableBoardOptions
	neighboursForCellIndexCache *[][]int
}

type TableBoardOptions struct {
	EvaluateOptions
	Cells []cell.Cell
}

type EvaluateOptions struct {
	// If set, will not attempt to use multiplication for evauluation
	NoMultiply bool
	// If set, will not attempt to use addition for evauluation
	NoAddition bool
}

// TODO: return error here
func RestoreTableBoard(columns, rows int, cells []cell.Cell, options ...TableBoardOptions) TableBoard {

	tb := TableBoard{
		rows:    rows,
		columns: columns,
	}
	for _, v := range options {
		tb.TableBoardOptions = v
	}
	if tb.NoMultiply && tb.NoAddition {
		panic(fmt.Errorf("NoAddition && NoMultiply cannot both be set (need at least one evauluation-method)"))
	}
	tb.cells = cells
	tb.precalculateNeighboursForCellIndex()
	return tb
}
func NewTableBoard(columns, rows int, options ...TableBoardOptions) TableBoard {
	cellCount := columns * rows
	cells := make([]cell.Cell, cellCount)
	for i := 0; i < cellCount; i++ {
		cells[i] = cell.NewCell(0, 0)
	}
	return RestoreTableBoard(columns, rows, cells, options...)
}

func (tb *TableBoard) Copy() BoardController {
	board := TableBoard{
		cells:                       make([]cell.Cell, len(tb.cells)),
		neighboursForCellIndexCache: tb.neighboursForCellIndexCache,
		rows:                        tb.rows,
		columns:                     tb.columns,
		TableBoardOptions:           tb.TableBoardOptions,
	}
	for i, c := range tb.cells {
		board.cells[i] = cell.NewCellCopy(c)
	}
	return &board

}
func (tb TableBoard) GetCellAtIndex(n int) *cell.Cell {
	if tb.ValidCellIndex(n) {
		return &tb.cells[n]
	}
	return nil
}
func (tb TableBoard) String() string {
	return tb.PrintBoard(nil)
}
func (tb TableBoard) ID() string {
	return tb.id
}
func (tb TableBoard) HighestValue() (cell.Cell, int) {
	var v int64
	var index int

	for i := 0; i < len(tb.cells); i++ {
		val := tb.cells[i].Value()
		if val > v {
			v = val
			index = i
		}
	}
	return tb.cells[index], index
}

type Sprinter interface {
	Sprintf(format string, a ...any) string
}

type CellValuer interface {
	Value() int64
}

func (tb TableBoard) PrintBoard(highlighter func(c CellValuer, index int, padded string) string) string {
	withColors := false
	h, _ := tb.HighestValue()
	longest := len(strconv.FormatInt(h.Value(), 10))
	headerPrinter := fmt.Sprintf
	if withColors {
		headerPrinter = color.Gray.Sprintf
	}

	s := "\n    "
	for j := 0; j < tb.columns; j++ {
		formatted := strconv.FormatInt(int64(j), 10)
		padLength := longest - len(formatted)
		padding := strings.Repeat(" ", padLength)
		valueStr := padding + formatted
		s += headerPrinter("  %s: ", valueStr)

	}
	for i := 0; i < tb.rows; i++ {
		s += "\n" + headerPrinter(" %02d: ", i*tb.rows)
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
	return s
}

// Uniqely identifies the board by its value. ID's etc are ignored.
func (tb TableBoard) Hash() string {
	builder := strings.Builder{}
	// builder.WriteString(fmt.Sprintf("%dx%d;", tb.columns, tb.rows))
	builder.WriteByte(byte(tb.columns))
	builder.WriteByte(byte(tb.rows))
	for i := 0; i < len(tb.cells); i++ {
		builder.WriteByte(byte(tb.cells[i].Value()))
	}
	return builder.String()
}

func (tb TableBoard) cellRow(i int) int {
	return i / tb.columns
}
func (tb TableBoard) cellColumn(i int) int {
	return i % tb.columns
}
func (tb TableBoard) IndexToCord(i int) (column int, row int) {
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

func PrintEvalMethod(e EvalMethod) string {
	switch e {
	case EvalMethodNil:
		return "EvalMethodNil"
	case EvalMethodProduct:
		return "EvalMethodProduct"
	case EvalMethodSum:
		return "EvalMethodSum"
	default:
		return "EvalMethodUnknown"

	}
}

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

func (tb TableBoard) EvaluatesTo(indexes []int, commit, noValidate bool) (int64, EvalMethod, error) {
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
	if v == 0 || !commit {
		return v, method, err
	}
	tb.GetCellAtIndex(last).Double()
	for _, index := range rest {
		tb.cells[index] = cell.NewEmptyCell()
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
		if cell.IsEmpty() {
			return 0, EvalMethodNil, ErrResultIndexOverflow
		}
		v := cell.Value()
		sum += v
		product *= v

		// return early if we have overshow the targetValue
		overshotMultiply := !tb.NoMultiply && product > targetValue
		overshotAddition := !tb.NoAddition && sum > targetValue
		if overshotAddition && overshotMultiply {
			return 0, EvalMethodNil, ErrResultOvershot
		}
	}
	// In cases where both addition and multipcation can be used,
	// prefer multipcation.
	if !tb.NoMultiply && product == targetValue {
		return product, EvalMethodProduct, nil
	}
	if !tb.NoAddition && sum == targetValue {
		return sum, EvalMethodSum, nil
	}

	return 0, EvalMethodNil, ErrResultNoMatch
}

func (tb TableBoard) getRows() [][]*cell.Cell {
	return tb._getColumnsOrRows(true)
}
func (tb TableBoard) getColumns() [][]*cell.Cell {
	return tb._getColumnsOrRows(false)
}
func (tb TableBoard) _getColumnsOrRows(rows bool) [][]*cell.Cell {
	var columns = make([][]*cell.Cell, tb.rows)
	for rowIndex := 0; rowIndex < tb.rows; rowIndex++ {
		columns[rowIndex] = make([]*cell.Cell, tb.columns)
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
			columns[rowIndex][colIndex] = &tb.cells[index]
		}
	}
	return columns
}

type SwipeDirection string

const (
	SwipeDirectionUp    SwipeDirection = "Up"
	SwipeDirectionRight SwipeDirection = "Right"
	SwipeDirectionDown  SwipeDirection = "Down"
	SwipeDirectionLeft  SwipeDirection = "Left"
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
			if tb.cells[i].Hash() != newCells[i].Hash() {
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
func (tb *TableBoard) SwipeDirectionPreview(direction SwipeDirection) []cell.Cell {
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
func (tb *TableBoard) Cells() []cell.Cell {
	return tb.cells
}

// hotpath-method, therefore optimized
func (tb *TableBoard) swipeHorizontal(positive bool) []cell.Cell {
	rows := tb.getRows()
	tiles := make([]cell.Cell, len(tb.cells))
	for ri := 0; ri < len(rows); ri++ {
		if positive {
			sort.Sort(PosCellRange(rows[ri]))
		} else {
			sort.Sort(NegCellRange(rows[ri]))
		}
		for i := 0; i < len(rows[ri]); i++ {
			tiles[i+tb.rows*ri] = *rows[ri][i]
		}
	}
	return tiles
}

// hotpath-method, therefore optimized
func (tb *TableBoard) swipeVertical(positive bool) []cell.Cell {
	columns := tb.getColumns()
	tiles := make([]cell.Cell, len(tb.cells))
	for ci := 0; ci < len(columns); ci++ {

		if positive {
			sort.Sort(PosCellRange(columns[ci]))
		} else {
			sort.Sort(NegCellRange(columns[ci]))
		}
		for i := 0; i < len(columns[ci]); i++ {
			// for i, cell := range columns[ci] {
			tiles[i*tb.columns+ci] = *columns[ci][i]
		}
	}
	return tiles
}

type PosCellRange []*cell.Cell

func (pp PosCellRange) Less(i, j int) bool { return pp[i].IsEmpty() }
func (pp PosCellRange) Len() int           { return len(pp) }
func (pp PosCellRange) Swap(i, j int)      { pp[i], pp[j] = pp[j], pp[i] }

type NegCellRange []*cell.Cell

func (pp NegCellRange) Less(i, j int) bool { return pp[j].IsEmpty() }
func (pp NegCellRange) Len() int           { return len(pp) }
func (pp NegCellRange) Swap(i, j int)      { pp[i], pp[j] = pp[j], pp[i] }

func (tb TableBoard) AreNeighboursByIndex(a, b int) bool {
	if a == b {
		return false
	}
	if a < 0 || b < 0 {
		return false
	}
	max := len(tb.cells)
	if a >= max || b >= max {
		return false
	}
	ac, ar := tb.IndexToCord(a)
	bc, br := tb.IndexToCord(b)

	diffc := ac - bc
	diffr := ar - br

	switch {
	// The cells cannot both be on different columns and rows and still be neighbours
	case diffc != 0 && diffr != 0:
		return false
	// The cells cannot be the same
	case diffc == 0 && diffr == 0:
		return false
	case diffc == 1:
		return true
	case diffc == -1:
		return true
	case diffr == 1:
		return true
	case diffr == -1:
		return true
	}
	return false
}

func (tb *TableBoard) precalculateNeighboursForCellIndex() {

	cache := make([][]int, len(tb.cells))
	for i := 0; i < len(tb.cells); i++ {
		cache[i] = tb.neighboursForCellIndex(i)
	}
	tb.neighboursForCellIndexCache = &cache

}

// Returns the neighbours for a gives cell. Note that the cells might be empty
func (tb *TableBoard) NeighboursForCellIndex(index int) ([]int, bool) {
	if index < 0 {
		return []int{}, false
	}
	if index >= len(tb.cells) {
		return []int{}, false
	}
	if tb.neighboursForCellIndexCache == nil {
		tb.precalculateNeighboursForCellIndex()
	}
	return (*tb.neighboursForCellIndexCache)[index], true
	// return tb.neighboursForCellIndex(index), false
}
func (tb TableBoard) neighboursForCellIndex(index int) []int {
	var neighbours []int
	column, row := tb.IndexToCord(index)

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
	return neighbours
}

func (tb *TableBoard) AddCellToBoard(c cell.Cell, index int, overwrite bool) error {
	if !tb.ValidCellIndex(index) {
		return ErrIndexInvalid
	}
	if !overwrite && tb.cells[index].Value() > 0 {
		return ErrCellAtIndexAlreadyHasValue
	}
	tb.cells[index] = c
	return nil

}

func (cg *TableBoard) ListEmptyCells() []int {
	cells := cg.Cells()
	empty := []int{}
	for i := 0; i < len(cells); i++ {
		if cells[i].IsEmpty() {
			empty = append(empty, i)
		}
	}
	return empty
}
