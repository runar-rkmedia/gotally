package tallylogic

import (
	"encoding/base64"
	"fmt"
	"math/rand"
)

type CellGenerator interface {
	Generate() Cell
}

type Game struct {
	board         BoardController
	selectedCells []int
	cellGenerator CellGenerator
	Rules         GameRules
	score         int64
	moves         int
	Description   string
	Hinter        hintCalculator
	GoalChecker   GoalChecker
	DefeatChecker GoalChecker
	History       []any
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
	StartingBricks  int
}

type GameMode int
type BoardType int

const (
	GameModeDefault GameMode = iota
	GameModeTemplate
	GameModeRandomChallenge
)

// Copies the game and all values to a new game
func (g Game) Copy() Game {
	game := Game{
		board:         g.board.Copy(),
		selectedCells: g.selectedCells,
		cellGenerator: g.cellGenerator,
		Rules:         g.Rules,
		score:         g.score,
		moves:         g.moves,
		Description:   g.Description,
		GoalChecker:   g.GoalChecker,
		DefeatChecker: g.DefeatChecker,
	}
	game.Hinter = NewHintCalculator(
		game.board, game.board, game.board,
	)
	game.History = append(game.History, g.History...)
	return game

}
func NewGame(mode GameMode, template *GameTemplate) (Game, error) {
	game := Game{
		// Default rules
		Rules: GameRules{
			SizeX:           5,
			SizeY:           5,
			RecreateOnSwipe: true,
			WithSuperPowers: true,
			StartingBricks:  5,
		},
		cellGenerator: NewCellGenerator(),
		History:       []any{},
	}
	switch mode {
	case GameModeDefault:
		board := NewTableBoard(5, 5)
		game.board = &board
		game.Description = "Default game, 5x5"
		game.DefeatChecker = DefeatCheckerNoMoreMoves{}
		game.GoalChecker = GoalCheck{"Game runs forever"}
		break
	case GameModeTemplate, GameModeRandomChallenge:
		if template != nil {
			t := template.Create()
			game.board = &t.Board
			game.Rules = t.Rules
			game.Description = t.Description
			game.DefeatChecker = t.DefeatChecker
			game.GoalChecker = t.GoalChecker

		} else {

			board := TableBoard{
				rows:    5,
				columns: 5,
				cells: cellCreator(
					0, 2, 1, 0, 1,
					64, 4, 4, 1, 2,
					64, 8, 4, 1, 0,
					12, 3, 1, 0, 0,
					16, 0, 0, 0, 0,
				),
			}
			game.board = &board
			game.Rules = GameRules{
				BoardType:       0,
				GameMode:        GameModeDefault,
				SizeX:           board.columns,
				SizeY:           board.rows,
				RecreateOnSwipe: false,
				WithSuperPowers: false,
			}
			game.Description = "Get to 512 points withing 10 moves"
		}
		break
	default:
		return game, fmt.Errorf("Invalid gamemode: %d", mode)
	}
	if len(game.Cells()) == 0 {
		for i := 0; i < game.Rules.StartingBricks; i++ {
			game.generateCellToEmptyCell()
		}
	}
	game.Hinter = NewHintCalculator(game.board, game.board, game.board)
	return game, nil
}

func (g *Game) GetHint() map[string]Hint {
	return g.Hinter.GetHints()
}
func (g *Game) generateCellToEmptyCell() bool {
	i := g.getRandomEmptyCell()
	if i == nil {
		return false
	}
	cell := g.cellGenerator.Generate()
	err := g.board.AddCellToBoard(cell, *i, false)
	return err == nil

}
func (g *Game) getRandomEmptyCell() *int {
	empty := g.getEmptyCellIndexes()
	if len(empty) == 0 {
		return nil
	}
	i := rand.Intn(len(empty))
	return &empty[i]

}
func (g *Game) getEmptyCellIndexes() []int {
	cells := g.board.Cells()
	var empty []int
	for i, v := range cells {
		if v.Value() == 0 {
			empty = append(empty, i)
		}
	}
	return empty
}

func (g *Game) inceaseMoveCount() {
	g.moves++
}
func (g *Game) increaseScore(points int64) {
	g.score += points
}

func (g *Game) Swipe(direction SwipeDirection) (changed bool) {
	changed = g.board.SwipeDirection(direction)
	g.ClearSelection()
	if g.Rules.RecreateOnSwipe {
		g.generateCellToEmptyCell()
	}
	if changed {
		g.inceaseMoveCount()
		g.History = append(g.History, direction)
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
	ok := g.EvaluateForPath(g.selectedCells)
	g.ClearSelection()
	return ok
}
func (g *Game) EvaluateForPath(path []int) bool {
	err, _ := g.board.ValidatePath(path)
	if err != nil {
		return false
	}
	points, _, err := g.board.EvaluatesTo(path, true, false)
	if err != nil {
		return false
	}
	g.increaseScore(points)
	g.inceaseMoveCount()
	g.History = append(g.History, path)
	return true
}

func (g *Game) Print() string {
	return g.board.String()
}
func (g *Game) ForTemplate() map[string]any {
	m := map[string]any{}
	m["cells"] = g.board.Cells()
	return m
}
func (g *Game) IsLastSelection(requested Cell) bool {
	if len(g.selectedCells) == 0 {
		return false
	}
	cells := g.board.Cells()
	last := g.selectedCells[len(g.selectedCells)-1]
	if cells[last].ID == requested.ID {
		return true
	}

	return false

}
func (g *Game) IsSelected(requested Cell) bool {
	if len(g.selectedCells) == 0 {
		return false
	}
	cells := g.board.Cells()
	for _, index := range g.selectedCells {
		if cells[index].ID == requested.ID {
			return true
		}

	}
	return false
}

func (g *Game) IsCellIndexPartOfFirstHint(index int, hints map[string]Hint) bool {
	if hints == nil {
		return false
	}
	for _, h := range hints {

		return g.IsCellIndexPartOfHint(index, h)

	}
	return false
}
func (g *Game) IsCellIndexPartOfHint(index int, hint Hint) bool {
	if len(hint.Path) == 0 {
		return false
	}
	for _, i := range hint.Path {
		if index == i {
			return true
		}
	}
	return false
}

func (g Game) Score() int64 {
	return g.score
}
func (g Game) Moves() int {
	return g.moves
}
func (g Game) SelectedCells() []int {
	return g.selectedCells
}
func (g Game) Cells() []Cell {
	return g.board.Cells()
}
func (g Game) NeighboursForCellIndex(index int) ([]int, bool) {
	return g.board.NeighboursForCellIndex(index)
}
func (g Game) EvaluatesTo(indexes []int, commit bool, noValidate bool) (int64, EvalMethod, error) {
	return g.board.EvaluatesTo(indexes, commit, noValidate)
}
func (g Game) SoftEvaluatesTo(indexes []int, targetValue int64) (int64, EvalMethod, error) {
	return g.board.SoftEvaluatesTo(indexes, targetValue)
}

func (g Game) IsGameWon() bool {
	return g.GoalChecker.Check(g)
}
func (g Game) IsGameOver() bool {
	return g.DefeatChecker.Check(g)
}
func (g Game) Hash() string {
	b := []byte(g.board.Hash())
	return base64.URLEncoding.EncodeToString(b)
}
