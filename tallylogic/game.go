package tallylogic

import (
	"fmt"
)

type CellGenerator interface {
	Generate() Cell
}

type Game struct {
	board         BoardController
	selectedCells []int
	cellGenerator CellGenerator
	rules         GameRules
	score         int64
	moves         int
}

type GameRules struct {
	BoardType
	GameMode
	SizeX int
	SizeY int
	// TODO: not implemented
	RecreateOnSwipe bool
	// TODO: not implemented
	WithSuperPowers bool
}

type GameMode int
type BoardType int

const (
	GameModeDefault GameMode = iota
)

func NewGame(mode GameMode) (Game, error) {
	game := Game{}
	switch mode {
	case GameModeDefault:
		game.rules.SizeX = 5
		game.rules.SizeY = 5
		game.rules.RecreateOnSwipe = true
		game.rules.WithSuperPowers = true
	default:
		return game, fmt.Errorf("Invalid gamemode: %d", mode)
	}
	return game, nil
}

func (g *Game) Swipe(direction SwipeDirection) (changed bool) {
	changed = g.board.SwipeDirection(direction)
	g.ClearSelection()
	if g.rules.RecreateOnSwipe {
		// TODO: pic a random empty cell and generate new here
	}
	if changed {
		g.moves++
	}

	return changed

}

// This is used to instruct the game using small data-values Not really sure
// how this will work in the end, But I am thinking to have a
// data-expander/compressor that will have a precalculated set of instructions
// per gameboard, where each instruction is simply a int16 or something
func (g *Game) instruct(instruction any) bool {
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

// Not all types of boards supports this, so this method will probably be removed
func (g *Game) selectCellCoord(x, y int) bool {
	n, ok := g.board.CoordToIndex(x, y)
	if !ok {
		return false
	}
	return g.SelectCell(n)
}
func (g *Game) SelectCell(n int) bool {
	ok := g.board.ValidCellIndex(n)
	if !ok {
		return false
	}
	c := g.board.GetCellAtIndex(n)
	if c == nil || c.Value() == 0 {
		g.ClearSelection()
		return false
	}
	if len(g.selectedCells) == 0 {
		g.selectedCells = append(g.selectedCells, n)
		return true
	}
	var next []int
	next = append(next, g.selectedCells...)
	next = append(next, n)
	err, _ := g.board.ValidatePath(next)
	if err != nil {
		g.ClearSelection()

		return false
	}
	g.selectedCells = next
	return true
}

func (g *Game) ClearSelection() {
	g.selectedCells = []int{}
}
func (g *Game) EvaluateSelection() bool {
	err, _ := g.board.ValidatePath(g.selectedCells)
	if err != nil {
		g.ClearSelection()
		return false
	}
	points, _, err := g.board.EvaluatesTo(g.selectedCells)
	if err != nil {
		return false
	}
	g.score += points
	g.moves++
	g.ClearSelection()
	return true
}

func (g Game) Score() int64 {
	return g.score
}
