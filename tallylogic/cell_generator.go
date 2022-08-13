package tallylogic

import "math/rand"

type cellGenerator struct{}

func NewCellGenerator() cellGenerator {
	return cellGenerator{}
}

// TODO: Implement correctly, this is just a simple one
func (cg cellGenerator) Generate() Cell {
	pow := 0
	baseValue := rand.Int63n(11) + 1
	return NewCell(baseValue, pow)
}
