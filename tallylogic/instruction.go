package tallylogic

import (
	"fmt"
)

func (g *Game) DescribeInstruction(instruction Instruction_) string {
	switch {
	case instruction.IsPath:
		return fmt.Sprintf("Combining path: %v", instruction.Path)
	case instruction.IsSwipe:
		switch instruction.Direction {
		case SwipeDirectionUp:
			return "Swiping up"
		case SwipeDirectionRight:
			return "Swiping Right"
		case SwipeDirectionDown:
			return "Swiping Down"
		case SwipeDirectionLeft:
			return "Swiping Left"
		}
	}
	panic(fmt.Sprintf("NotImplemented:Game:DescribeInstruction: %#v", instruction))
}

// This is used to Instruct the game using small data-values Not really sure
// how this will work in the end, But I am thinking to have a
// data-expander/compressor that will have a precalculated set of instructions
// per gameboard, where each instruction is simply a int16 or something
func (g *Game) Instruct(instruction Instruction_) bool {
	switch {
	case instruction.IsSwipe:
		switch instruction.Direction {
		case SwipeDirectionUp:
			return g.Swipe(SwipeDirectionUp)
		case SwipeDirectionRight:
			return g.Swipe(SwipeDirectionRight)
		case SwipeDirectionDown:
			return g.Swipe(SwipeDirectionDown)
		case SwipeDirectionLeft:
			return g.Swipe(SwipeDirectionLeft)
		default:
			panic(fmt.Sprintf("Unknown swipe-instruction %#v", instruction))
		}
	case instruction.IsPath:
		g.selectedCells = instruction.Path
		return g.EvaluateSelection()
	case instruction.IsHelper:
		switch instruction.Helper {
		case helperUndo:
			g.Undo()
		}
	default:
		panic(fmt.Sprintf("Unknown instruction %#v", instruction))
	}

	return false
}
