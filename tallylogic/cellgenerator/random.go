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
		// 1 is useful,
		// since it can be used to connect cells via multiplication (8 x 1 * 4 = 32),
		// and to combine odd-numbers, transforming unwanted numbers into highly usable numbers (7 + 1 = 8)
		Add(80, 1).
		// The most useful cell, but a bit boring
		Add(100, 2).
		// Offers both fun and a challenge
		Add(80, 3).
		// Is basically just a two, but not as good
		Add(50, 4).
		// Generally only makes the game harder, but it is also very easy to work with.
		// Most people are comfortable with multiplying fives
		Add(25, 5).
		Add(20, 6).
		// One of the higher primes available. Makes the game harder, but gives a nice challenge
		Add(10, 7).
		// Very useful, but many people have difficulties with multiplying it.
		Add(50, 8).
		Add(20, 9).
		Add(20, 10).
		Add(10, 11).
		// Highly composable, fun and gives a challenge
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
