package tallylogic

import "github.com/runar-rkmedia/gotally/tallylogic/cell"

func cellCreator(vals ...int64) []cell.Cell {
	cells := make([]cell.Cell, len(vals))
	for i, v := range vals {
		cells[i] = cell.NewCell(v, 0)
	}
	return cells
}
func cellCreatorUints(vals ...uint64) []cell.Cell {
	cells := make([]cell.Cell, len(vals))
	for i, v := range vals {
		cells[i] = cell.NewCell(int64(v), 0)
	}
	return cells
}
