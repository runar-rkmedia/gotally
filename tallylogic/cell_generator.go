package tallylogic

type cellGenerator struct {
	r CellRandomizer
}
type CellRandomizer interface {
	Int63n(n int64) int64
	Intn(n int) int
	IntRandomizer
}

func NewCellGenerator(r CellRandomizer) *cellGenerator {
	return &cellGenerator{
		r,
	}
}

// TODO: Implement correctly, this is just a simple one
func (cg *cellGenerator) Generate() Cell {
	pow := 0
	baseValue := cg.r.Int63n(11) + 1
	return NewCell(baseValue, pow)
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
