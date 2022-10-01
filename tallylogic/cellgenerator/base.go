package cellgenerator

import "github.com/runar-rkmedia/gotally/tallylogic/cell"

type cellGenerator struct {
	r          CellRandomizer
	strategies map[strategy]generator
}
type CellRandomizer interface {
	Int63n(n int64) int64
	Intn(n int) int
	Int63() int64
	IntRandomizer
}

type BoardController interface {
	Cells() []cell.Cell
	ListEmptyCells() []int
	HighestValue() (cell.Cell, int)
}
type IntRandomizer interface {
	Seed() (uint64, uint64)
	SetSeed(seed uint64, state uint64) error
}

func NewCellGenerator(r CellRandomizer) *cellGenerator {
	randomG := &randomGenerator{r}
	return &cellGenerator{
		r,
		map[strategy]generator{
			strategyRandom:  randomG,
			strategyHelpful: &helpfulGenerator{*randomG},
		},
	}
}

func (cg *cellGenerator) Seed() (uint64, uint64) {
	return cg.r.Seed()
}
func (cg *cellGenerator) Intn(n int) int {
	return cg.r.Intn(n)
}
func (cg *cellGenerator) SetSeed(seed, state uint64) error {
	return cg.r.SetSeed(seed, state)
}

type generator interface {
	GenerateAt(index int, board BoardController) cell.Cell
	GeneratePure() cell.Cell
	Generate(board BoardController) (cell.Cell, int, bool)
}
