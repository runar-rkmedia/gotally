package cellgenerator

import (
	"sort"

	"github.com/runar-rkmedia/gotally/tallylogic/cell"
)

type helpfulGenerator struct {
	r randomGenerator
}

type cellValueIndex struct {
	value int64
	index int
}

// Attempts to generate a helpful cell.
// Doing this too often can make the game very boring,
// as it will fill the board with only useful values,
// which removes all challenges
func (cg *helpfulGenerator) GenerateAt(index int, board BoardController) cell.Cell {
	cells := board.Cells()
	// Create a sorted list of cells by their value
	cellValues := make([]cellValueIndex, len(cells))
	withValue := 0
	for i, v := range cells {
		cellValues[i] = cellValueIndex{v.Value(), i}
		if !v.IsEmpty() {
			withValue++
		}
	}
	// empty board, use randomizer
	if withValue == 0 {
		return cg.r.GenerateAt(index, board)
	}
	sort.Slice(cellValues, func(i, j int) bool { return cellValues[i].value > cellValues[j].value })

	top := 3
	if withValue < top {
		top = withValue
	}
	pickedIndex := cg.r.r.Intn(top)
	pickedCellIndex := cellValues[pickedIndex].index
	pickedCell := cells[pickedCellIndex]

	base, maxToPow := pickedCell.Raw()
	if maxToPow == 0 {
		return cell.NewCell(2, 0)
	}
	pow := cg.r.r.Int63n(maxToPow)
	return cell.NewCell(base, int(pow))
}
func (cg *helpfulGenerator) GeneratePure() cell.Cell {
	return cg.r.GeneratePure()
}
func (cg *helpfulGenerator) Generate(board BoardController) (cell.Cell, int, bool) {
	cellIndex, ok := cg.r.PickRandomEmptyCell(board)
	if !ok {
		return cell.Cell{}, 0, false
	}
	return cg.GenerateAt(cellIndex, board), cellIndex, true
}
