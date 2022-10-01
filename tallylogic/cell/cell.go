package cell

import (
	"fmt"
	"math"
	"strconv"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

type Cell struct {
	ID        string
	baseValue int64
	power     int
}

func NewEmptyCell() Cell {
	return NewCell(0, 0)
}
func NewCellCopy(c Cell) Cell {
	return NewCell(c.baseValue, c.power)
}
func NewCell(baseValue int64, power int) Cell {
	return Cell{
		ID:        gonanoid.Must(),
		baseValue: baseValue,
		power:     power,
	}
}
func (c *Cell) Double() *Cell {
	c.power += 1
	return c
}
func (c *Cell) IsEmpty() bool {
	return c.baseValue == 0
}
func (c Cell) Doubled() Cell {
	return NewCell(c.baseValue, c.power+1)
}

func (c Cell) Raw() (int64, int64) {
	return c.baseValue, int64(c.power)
}
func (c Cell) Value() int64 {
	if c.power == 0 {
		return c.baseValue
	}
	return c.baseValue * (int64(math.Pow(2, float64(c.power))))
}
func (c Cell) String() string {
	return strconv.FormatInt(c.Value(), 10)
}
func (c Cell) Hash() string {
	return fmt.Sprintf("%d*%d", c.baseValue, c.power)
}
