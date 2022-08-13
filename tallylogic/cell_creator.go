package tallylogic

func cellCreator(vals ...int64) []Cell {
	cells := make([]Cell, len(vals))
	for i, v := range vals {
		cells[i] = NewCell(v, 0)
	}
	return cells
}
