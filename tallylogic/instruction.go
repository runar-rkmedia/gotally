package tallylogic

import (
	"fmt"
	"strconv"
	"strings"
)

type Instruction []any

func (i *Instruction) AddSwipe(dir SwipeDirection) {
	(*i) = append(*i, dir)
}
func (i *Instruction) AddSelectIndex(index int) {
	(*i) = append(*i, index)
}
func (i *Instruction) AddSelectCoord(column, row int) {
	(*i) = append(*i, [2]int{column, row})
}
func (i *Instruction) AddPath(indexes []int) {
	(*i) = append(*i, indexes)
}
func (i *Instruction) AddEvaluateSelection(indexes []int) {
	(*i) = append(*i, indexes)
}
func (i Instruction) Hash() string {
	return i.hash(36, false)
}
func (i Instruction) DescribeShort() string {
	return i.hash(10, true)
}
func (i Instruction) hash(base int, withPrefix bool) string {
	b := strings.Builder{}
	for _, ins := range i {
		switch ins {
		case true:
			b.WriteString("t")
		case SwipeDirectionUp:
			b.WriteString("U")
		case SwipeDirectionRight:
			b.WriteString("R")
		case SwipeDirectionDown:
			b.WriteString("D")
		case SwipeDirectionLeft:
			b.WriteString("L")
		default:
			switch t := ins.(type) {
			case int:
				if withPrefix {
					b.WriteString("index:")
				}
				b.WriteString(strconv.FormatInt(int64(t), base))
			case [2]int:
				if withPrefix {
					b.WriteString("coord:")
				}
				b.WriteString(strconv.FormatInt(int64(t[0]), base))
				b.WriteString("x")
				b.WriteString(strconv.FormatInt(int64(t[1]), base))
			case []int:
				if withPrefix {
					b.WriteString("indexes:")
				}
				for i := 0; i < len(t); i++ {
					b.WriteString(strconv.FormatInt(int64(t[i]), base))
					if withPrefix && i < len(t)-1 {
						b.WriteString(",")
					}

				}
			}
		}
		b.WriteString(";")
	}
	return b.String()
}

type InstructionType int

const (
	InstructionTypeUnknown InstructionType = iota
	InstructionTypeSwipe
	InstructionTypeCombinePath
	InstructionTypeSelectIndex
	InstructionTypeSelectCoord
)

func DescribeInstructionType(t InstructionType) string {
	switch t {
	case InstructionTypeUnknown:
		return "unknown"
	case InstructionTypeSwipe:
		return "Swipe"
	case InstructionTypeCombinePath:
		return "CombinePath"
	case InstructionTypeSelectCoord:
		return "Select:Coord"
	case InstructionTypeSelectIndex:
		return "Select:Index"
	}
	return "???"
}

func GetInstructionAsPath(ins any) ([]int, bool) {
	p, ok := ins.([]int)
	return p, ok
}
func GetInstructionAsSwipe(ins any) (SwipeDirection, bool) {
	switch ins {
	case SwipeDirectionUp:
		return SwipeDirectionUp, true
	case SwipeDirectionRight:
		return SwipeDirectionRight, true
	case SwipeDirectionLeft:
		return SwipeDirectionLeft, true
	case SwipeDirectionDown:
		return SwipeDirectionDown, true
	}
	return "???", false
}
func GetInstructionType(ins any) InstructionType {
	switch ins.(type) {
	case int:
		return InstructionTypeSelectIndex
	case SwipeDirection:
		return InstructionTypeSwipe
	case [2]int:
		return InstructionTypeSelectCoord
	case []int:
		return InstructionTypeCombinePath
	}
	return -1

}
func CompareInstrictionAreEqual(ins, ins2 any) (bool, InstructionType) {
	switch t := ins.(type) {
	case int:
		switch t2 := ins2.(type) {
		case int:
			return t == t2, InstructionTypeSelectIndex
		default:
			return false, InstructionTypeUnknown
		}
	case SwipeDirection:
		switch t2 := ins2.(type) {
		case SwipeDirection:
			return t == t2, InstructionTypeSwipe
		default:
			return false, InstructionTypeUnknown
		}

	}
	return fmt.Sprintf("%v", ins) == fmt.Sprintf("%#v", ins2), InstructionTypeCombinePath
}

// 8, 12, 8
func (g *Game) DescribePath(path []int) string {
	values := make([]int64, len(path))
	cells := g.Cells()
	sum := int64(0)
	product := int64(1)
	for i := 0; i < len(path); i++ {
		values[i] = cells[path[i]].Value()
		if i != len(path)-1 {
			sum += values[i]
			product *= values[i]
		}
	}
	targetValue := values[len(path)-1]
	sep := " + "
	if targetValue == product {
		sep = " * "
	}
	s := strconv.FormatInt(values[0], 10)
	for i := 1; i < len(path)-1; i++ {
		s += sep + strconv.FormatInt(values[i], 10)
	}
	s += " = " + strconv.FormatInt(targetValue, 10)
	fmt.Printf("Described %s, from values %v, for path %v and cells %s\n", s, values, path, cells)
	return s
}
func (g *Game) DescribeInstruction(instruction any) string {
	switch instruction {
	case true:
		return "Combinging selection"
	case SwipeDirectionUp:
		return "Swiping up"
	case SwipeDirectionRight:
		return "Swiping Right"
	case SwipeDirectionDown:
		return "Swiping Down"
	case SwipeDirectionLeft:
		return "Swiping Left"
	}
	switch t := instruction.(type) {
	case int:
		cell := g.board.GetCellAtIndex(t)
		x, y := g.board.IndexToCord(t)
		return fmt.Sprintf("Selecting %d at %dx%d", cell.Value(), x, y)
	case [2]int:
		index, ok := g.board.CoordToIndex(t[0], t[1])
		if !ok {
			return fmt.Sprintf("Attempted to select invalid cell at %dx%d", t[0], t[1])
		}
		cell := g.board.GetCellAtIndex(index)
		return fmt.Sprintf("Selecting %d at %dx%d", cell.Value(), t[0], t[1])
	case []int:
		coords := make([][2]int, len(t))
		for i := 0; i < len(t); i++ {
			index := t[i]
			x, y := g.board.IndexToCord(index)
			coords[i] = [2]int{x, y}
		}

		return fmt.Sprintf("Combining at coords (%v) -- original: {%v}", coords, t)
	case Instruction:
		s := make([]string, len(t))
		for i, v := range t {
			s[i] = g.DescribeInstruction(v)

		}
		return strings.Join(s, "\n\t")
	default:
		return fmt.Sprintf("unknown instruction, %#v", t)
	}
}

// This is used to Instruct the game using small data-values Not really sure
// how this will work in the end, But I am thinking to have a
// data-expander/compressor that will have a precalculated set of instructions
// per gameboard, where each instruction is simply a int16 or something
func (g *Game) Instruct(instruction any) bool {
	switch instruction {
	case true:
		return g.EvaluateSelection()
	case SwipeDirectionUp:
		return g.Swipe(SwipeDirectionUp)
	case SwipeDirectionRight:
		return g.Swipe(SwipeDirectionRight)
	case SwipeDirectionDown:
		return g.Swipe(SwipeDirectionDown)
	case SwipeDirectionLeft:
		return g.Swipe(SwipeDirectionLeft)
	}
	switch t := instruction.(type) {
	case int:
		return g.SelectCell(t)
	case [2]int:
		return g.selectCellCoord(t[0], t[1])
	case []int:
		g.selectedCells = t
		return g.EvaluateSelection()
	}

	return false
}
