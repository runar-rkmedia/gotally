package cellgenerator

import (
	"strconv"

	"github.com/runar-rkmedia/gotally/tallylogic/cell"
	"github.com/runar-rkmedia/gotally/weightmap"
)

type strategy int

const (
	strategyRandom strategy = iota
	strategyHelpful
)

var (
	weighMapStrategy = weightmap.NewWeightMap().
		Add(100, int(strategyRandom)).
		Add(4, int(strategyHelpful))
)

func (s strategy) String() string {
	switch s {
	case strategyRandom:
		return "Random strategy"
	case strategyHelpful:
		return "Helpful strategy"
	default:
		return "Unknown strategy: " + strconv.FormatInt(int64(s), 10)
	}
}

func (cg *cellGenerator) pickStrategy() strategy {
	// n := strategy(cg.r.Intn(len(cg.strategies)))
	nonce := int(cg.r.Int63())
	n := strategy(weighMapStrategy.Get(nonce))
	return n
}
func (cg *cellGenerator) getGenerator() generator {
	return cg.strategies[cg.pickStrategy()]
}

// Generates a cell for a index on the board. The index is used as a hint as to what index would require
func (cg *cellGenerator) GenerateAt(index int, board BoardController) cell.Cell {
	return cg.getGenerator().GenerateAt(index, board)
}
func (cg *cellGenerator) GeneratePure() cell.Cell {
	return cg.getGenerator().GeneratePure()
}
func (cg *cellGenerator) Generate(board BoardController) (cell.Cell, int, bool) {
	return cg.getGenerator().Generate(board)
}
