package cellgenerator

import (
	"github.com/runar-rkmedia/gotally/tallylogic/cell"
	"github.com/runar-rkmedia/gotally/weightmap"
)

type randomGenerator struct {
	r CellRandomizer
}

var (
	cellWeightMapEasy = weightmap.NewWeightMap().
		Add(1, 1).
		Add(100, 2).
		Add(80, 3).
		Add(50, 4).
		Add(25, 5).
		Add(20, 6).
		Add(10, 7).
		Add(50, 8).
		Add(20, 9).
		Add(20, 10).
		Add(10, 11).
		Add(50, 12)
)

func (cg *randomGenerator) GenerateAt(index int, board BoardController) cell.Cell {
	return cg.GeneratePure()
}
func (cg *randomGenerator) GeneratePure() cell.Cell {
	pow := 0
	c := cellWeightMapEasy.Get(int(cg.r.Int63()))
	return cell.NewCell(int64(c), pow)
}
func (cg *randomGenerator) Generate(board BoardController) (cell.Cell, int, bool) {
	cellIndex, ok := cg.PickRandomEmptyCell(board)
	if !ok {
		return cell.Cell{}, 0, false
	}
	return cg.GenerateAt(cellIndex, board), cellIndex, true
}
func (cg *randomGenerator) PickRandomEmptyCell(board BoardController) (int, bool) {
	empty := board.ListEmptyCells()
	if len(empty) == 0 {
		return 0, false
	}
	n := cg.r.Intn(len(empty))
	return empty[n], true
}
