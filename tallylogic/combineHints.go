package tallylogic

import (
	"sort"
	"strconv"

	"github.com/runar-rkmedia/gotally/tallylogic/cell"
)

// GetCombineHints finds all possible combined without performing any swipes used as a generator.
// Returning true from the process-function will stop any further work, returning early.
//
// The first processed hints are suppose to be the most likely to be helpful for the user,
// but does not guarantee that the hint will actually help solve the game.
//
// It is however a lot faster that the GameSolvers, since it does not look ahead, and it processes a lot fewer combinations.
//
// Currently, this is the preferred order (subject to change):
// 1. Multiplication, by targeting the highest values first.
// 2. Addition, by targeting odd values first, highest values first.
func (g Game) GetCombineHints(process func(path []int, method EvalMethod) bool) error {
	cells := g.Cells()
	indexCells := NewIndexedCells(cells)

	var p processor = func(p IndexedCells, tIndex int, method EvalMethod) bool {
		path := make([]int, len(p)+1)
		// Path must be reversed, since the underlying implementation works in reverse-order
		// and the rest of the API expects the path to be the other way around.
		var j int
		for i := len(p) - 1; i >= 0; i-- {
			path[j] = p[i].index
			j++
		}
		path[len(p)] = tIndex
		return process(path, method)
	}

	// Sort by highest value first, to start checking for possible multiplications in those places first.
	sort.Slice(indexCells, func(i, j int) bool {
		return indexCells[i].value > indexCells[j].value
	})
	for i := 0; i < len(indexCells); i++ {
		stop := FindAllValidMultiplications(p, indexCells, g.Rules.SizeY, g.Rules.SizeX, indexCells[i].value, indexCells[i].index)
		if stop {
			return nil
		}
	}
	// For addition, we prioritize removing values that are primenumbers, and higher primenumbers should be checked first.
	sort.Slice(indexCells, func(i, j int) bool {
		if indexCells[i].value%2 == 0 && indexCells[j].value%2 == 1 {
			return false // even number should come after odd number
		}
		if indexCells[i].value%2 == 1 && indexCells[j].value%2 == 0 {
			return true // odd number should come before even number
		}
		return indexCells[i].value > indexCells[j].value // sorting in ascending order for same parity
	})

	for i := 0; i < len(indexCells); i++ {
		stop := FindAllValidAdditions(p, indexCells, g.Rules.SizeY, g.Rules.SizeX, indexCells[i].value, indexCells[i].index)
		if stop {
			return nil
		}

	}
	return nil
}

type cellIndex struct {
	value int64
	index int
}

func (ci cellIndex) String() string {
	return strconv.FormatInt(int64(ci.value), 10) + ":" + strconv.FormatInt(int64(ci.index), 10)
}

type IndexedCells []cellIndex

func (ic IndexedCells) Print(target int, method EvalMethod) string {
	s := ""
	for i := 0; i < len(ic); i++ {
		if i > 0 {
			if method == EvalMethodSum {

				s += "+"
			} else if method == EvalMethodProduct {
				s += "*"
			}

		}
		s += strconv.FormatInt(ic[i].value, 10)
	}
	if target > 0 {
		s += "=" + strconv.FormatInt(int64(target), 10)
	}
	return s
}
func (ic IndexedCells) PrintIndexes(targetIndex int) string {

	s := ""
	if targetIndex > 0 {
		s = strconv.FormatInt(int64(targetIndex), 10) + "<-"
	}
	for i := 0; i < len(ic); i++ {
		if i > 0 {
			s += ","
		}
		s += strconv.FormatInt(int64(ic[i].index), 10)
	}
	return s
}

func NewIndexedCells(cells []cell.Cell) IndexedCells {
	values := IndexedCells{}
	for i := 0; i < len(cells); i++ {
		if cells[i].IsEmpty() {
			continue
		}
		values = append(values, cellIndex{value: cells[i].Value(), index: i})
	}
	return values

}
func FindAllValidMultiplications(p processor, A IndexedCells, rows, columns int, t int64, tIndex int) bool {
	if t == 1 {
		return false
	}
	return findAllValidMultiplications(p, A, rows, columns, t, tIndex, IndexedCells{})
}
func FindAllValidAdditions(p processor, A IndexedCells, rows, columns int, t int64, tIndex int) bool {
	return findAllValidAdditions(p, A, rows, columns, t, tIndex, IndexedCells{})
}

func findAllValidMultiplications(p processor, A IndexedCells, rows, columns int, t int64, tIndex int, currentCombination IndexedCells) bool {
	if t == 1 {
		return p(currentCombination, tIndex, EvalMethodProduct)
	}
	prev := tIndex
	if len(currentCombination) > 0 {
		prev = currentCombination[len(currentCombination)-1].index
	}
outer:
	for i := 0; i < len(A); i++ {
		index := A[i].index
		// Skip comparing to itself
		if A[i].index == tIndex {
			continue
		}
		// Skip if dit does not multiply
		if t%A[i].value != 0 {
			continue
		}
		// Skip if cell is already used
		for i := 0; i < len(currentCombination); i++ {
			if currentCombination[i].index == index {
				continue outer
			}
		}
		// Skip if not neighbours
		if !AreNeighboursByIndex(prev, A[i].index, rows, columns) {
			continue
		}
		if t%A[i].value == 0 {
			currentCombination = append(currentCombination, A[i])
			stop := findAllValidMultiplications(p, A, rows, columns, t/A[i].value, tIndex, currentCombination)
			if stop {
				return stop
			}
			currentCombination = currentCombination[:len(currentCombination)-1]
		}
	}
	return false
}

type processor func(cells IndexedCells, targetIndex int, method EvalMethod) bool

func findAllValidAdditions(p processor, A IndexedCells, rows, columns int, t int64, tIndex int, currentCombination IndexedCells) bool {
	if t == 0 {
		return p(currentCombination, tIndex, EvalMethodSum)
	}
	prev := tIndex
	if len(currentCombination) > 0 {
		prev = currentCombination[len(currentCombination)-1].index
	}
outer:
	for i := 0; i < len(A); i++ {
		index := A[i].index
		// Skip comparing to itself
		if index == tIndex {
			continue
		}
		value := A[i].value
		if t-value < 0 {
			continue
		}
		// Skip if cell is already used
		for i := 0; i < len(currentCombination); i++ {
			if currentCombination[i].index == index {
				continue outer
			}
		}
		// Skip if not neighbours
		if !AreNeighboursByIndex(prev, index, rows, columns) {
			continue
		}

		currentCombination = append(currentCombination, A[i])
		stop := findAllValidAdditions(p, A, rows, columns, t-value, tIndex, currentCombination)
		if stop {
			return stop
		}
		currentCombination = currentCombination[:len(currentCombination)-1]
	}
	return false
}
