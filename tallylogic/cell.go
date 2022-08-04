package tallylogic

import (
	"fmt"
	"math"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

type Cell struct {
	id        string
	baseValue int
	power     int
}

func NewCell(baseValue, power int) Cell {
	return Cell{
		id:        gonanoid.Must(),
		baseValue: baseValue,
		power:     power,
	}
}
func (c Cell) Doubled() Cell {
	return NewCell(c.baseValue, c.power+1)
}

func (c Cell) Value() int {
	if c.power == 0 {
		return c.baseValue
	}
	return c.baseValue * (int(math.Pow(2, float64(c.power))))
}
func (c Cell) String() string {
	return fmt.Sprintf("%d (%d**(2^%d))", c.Value(), c.baseValue, c.power)
}
