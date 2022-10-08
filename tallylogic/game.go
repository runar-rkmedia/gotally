package tallylogic

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/runar-rkmedia/gotally/randomizer"
	"github.com/runar-rkmedia/gotally/tallylogic/cell"
	"github.com/runar-rkmedia/gotally/tallylogic/cellgenerator"
	"github.com/runar-rkmedia/gotally/types"
)

type CellGenerator interface {
	// Generates a cell, to an empty cell. ok is false if there is not possible to place a brick
	Generate(board cellgenerator.BoardController) (c cell.Cell, index int, ok bool)
	// Generates a cell for a certain position.
	GenerateAt(index int, board cellgenerator.BoardController) cell.Cell
	// Generates a cell
	GeneratePure() cell.Cell
	Intn(int) int
	IntRandomizer
}
type IntRandomizer interface {
	Seed() (uint64, uint64)
	SetSeed(seed uint64, state uint64) error
}

type Game struct {
	// Uniquely identifies this board
	ID            string
	board         BoardController
	selectedCells []int
	cellGenerator CellGenerator
	Rules         GameRules
	score         int64
	moves         int
	Description   string
	Name          string
	Hinter        hintCalculator
	GoalChecker   GoalChecker
	DefeatChecker GoalChecker
	History       Instruction
}

func (g Game) Seed() (uint64, uint64) {
	return g.cellGenerator.Seed()
}

type GameRules struct {
	GameMode
	SizeX int
	SizeY int
	// TODO: not implemented
	RecreateOnSwipe bool
	// TODO: not implemented
	WithSuperPowers bool
	StartingBricks  int
	// Whether to allow swipes in the same direction or not, for instance twice up.
	// This can have an effect if there are new items generated for each swipe
	NoReswipe bool
	Options   NewGameOptions
}

// Deprecated, use types.GameMode
type GameMode int

const (
	GameModeDefault GameMode = iota
	GameModeTemplate
	GameModeRandomChallenge
)

// Copies the game and all values to a new game
func (g Game) Copy() Game {
	// IMPORTANT: Do not copy the onDidChangeEvents here.
	seed, state := g.cellGenerator.Seed()
	r := randomizer.NewRandomizerFromSeed(seed, state)
	cg := cellgenerator.NewCellGenerator(r)
	game := Game{
		ID:            g.ID,
		board:         g.board.Copy(),
		selectedCells: g.selectedCells,
		cellGenerator: cg,
		Rules:         g.Rules,
		score:         g.score,
		moves:         g.moves,
		Name:          g.Name,
		Description:   g.Description,
		GoalChecker:   g.GoalChecker,
		DefeatChecker: g.DefeatChecker,
	}
	if g.Hinter.CellRetriever != nil {
		game.Hinter = NewHintCalculator(
			game.board, game.board, game.board,
		)
	}
	game.History = append(game.History, g.History...)
	return game

}

type NewGameOptions struct {
	TableBoardOptions
	Seed  uint64
	State uint64
}

func RestoreGame(g *types.Game) (Game, error) {
	var mode GameMode
	game := Game{
		ID:            g.ID,
		board:         nil,
		selectedCells: []int{},
		cellGenerator: nil,
		Rules: GameRules{
			GameMode:        mode,
			SizeX:           int(g.Rules.Columns),
			SizeY:           int(g.Rules.Rows),
			RecreateOnSwipe: g.Rules.RecreateOnSwipe,
			// WithSuperPowers: g.Rules.WithSuperPowers,
			// StartingBricks:  g.Rules.,
			NoReswipe: g.Rules.NoReSwipe,
			Options: NewGameOptions{
				TableBoardOptions: TableBoardOptions{
					EvaluateOptions: EvaluateOptions{
						NoMultiply: g.Rules.NoMultiply,
						NoAddition: g.Rules.NoAddition,
					},
				},
				Seed:  g.Seed,
				State: g.State,
			},
		},
		score:       int64(g.Score),
		moves:       int(g.Moves),
		Description: g.Description,
		// Name:        g.Name,
		GoalChecker:   nil,
		DefeatChecker: nil,
		History:       []any{},
	}

	switch game.Rules.GameMode {
	case GameModeDefault:
		game.DefeatChecker = DefeatCheckerNoMoreMoves{}
		game.GoalChecker = GoalCheck{"Game runs forever"}
	default:
		return game, fmt.Errorf("not implemented for this gamemode")
	}

	r := randomizer.NewRandomizerFromSeed(game.Rules.Options.Seed, game.Rules.Options.State)
	game.cellGenerator = cellgenerator.NewCellGenerator(r)
	board := RestoreTableBoard(game.Rules.SizeX, game.Rules.SizeY, g.Cells, game.Rules.Options.TableBoardOptions)
	game.board = &board
	game.Hinter = NewHintCalculator(game.board, game.board, game.board)

	return game, nil

}

func NewGame(mode GameMode, template *GameTemplate, options ...NewGameOptions) (Game, error) {
	game := Game{
		ID: gonanoid.Must(),
		// Default rules
		Rules: GameRules{
			SizeX:           5,
			SizeY:           5,
			RecreateOnSwipe: true,
			WithSuperPowers: true,
			StartingBricks:  5,
			GameMode:        mode,
		},
		History: []any{},
	}
	for _, o := range options {
		game.Rules.Options = o
	}

	r := randomizer.NewRandomizerFromSeed(game.Rules.Options.Seed, game.Rules.Options.State)
	game.cellGenerator = cellgenerator.NewCellGenerator(r)
	switch mode {
	case GameModeDefault:
		board := NewTableBoard(5, 5, game.Rules.Options.TableBoardOptions)
		game.board = &board
		game.Description = "Default game, 5x5"
		game.DefeatChecker = DefeatCheckerNoMoreMoves{}
		game.GoalChecker = GoalCheck{"Game runs forever"}
	case GameModeTemplate, GameModeRandomChallenge:
		if template != nil {
			t := template.Create()
			game.board = &t.Board
			game.Rules = t.Rules
			game.Description = t.Description
			game.Name = t.Name
			game.DefeatChecker = t.DefeatChecker
			game.GoalChecker = t.GoalChecker

			if game.Description == "" {
				game.Description = game.GoalChecker.Description()
			}

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
				GameMode:        mode,
				SizeX:           board.columns,
				SizeY:           board.rows,
				RecreateOnSwipe: false,
				WithSuperPowers: false,
			}
			game.Description = "Get to 512 points withing 10 moves"
		}
	default:
		return game, fmt.Errorf("Invalid gamemode: %d", mode)
	}
	allEmpty := true
	for _, c := range game.Cells() {
		if c.Value() > 0 {
			allEmpty = false
			break
		}

	}
	if allEmpty || len(game.Cells()) == 0 {
		for i := 0; i < game.Rules.StartingBricks; i++ {
			game.generateCellToEmptyCell()
		}
	}
	game.Hinter = NewHintCalculator(game.board, game.board, game.board)
	if len(game.board.Cells()) != (game.Rules.SizeX * game.Rules.SizeY) {
		return game, fmt.Errorf("Game has invalid size: %d cells, %dx%d, mode %v template %v", len(game.board.Cells()), game.Rules.SizeX, game.Rules.SizeY, mode, template)
	}
	return game, nil
}

func (g *Game) GetHint() map[string]Hint {
	return g.Hinter.GetHints()
}
func (g *Game) BoardID() string {
	return g.board.ID()
}
func (g *Game) generateCellToEmptyCell() bool {
	cell, index, ok := g.cellGenerator.Generate(g.board)
	if !ok {
		return false
	}
	err := g.board.AddCellToBoard(cell, index, false)
	return err == nil

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
		g.History.AddSwipe(direction)
	}
	return changed
}

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
				b.WriteString(strconv.FormatInt(int64(t), 36))
			case [2]int:
				b.WriteString(strconv.FormatInt(int64(t[0]), 36))
				b.WriteString("x")
				b.WriteString(strconv.FormatInt(int64(t[1]), 36))
			case []int:
				for i := 0; i < len(t); i++ {
					b.WriteString(strconv.FormatInt(int64(t[i]), 36))

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

func GetInstructionAsPath(ins any) ([]int, bool) {
	p, ok := ins.([]int)
	return p, ok
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
		values := make([]int64, len(t))
		coords := make([][2]int, len(t))
		for i := 0; i < len(t); i++ {
			cell := g.board.GetCellAtIndex(i)
			x, y := g.board.IndexToCord(i)
			coords[0] = [2]int{x, y}
			values[i] = cell.Value()
		}

		return fmt.Sprintf("Combining values %v (%v) {%v}", values, coords, t)
	}
	return "unknown instruction"
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
	g.increaseScore(points * int64(len(path)))
	g.inceaseMoveCount()
	g.History.AddPath(path)
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

func (g Game) Score() int64 {
	return g.score
}
func (g Game) ValidatePath(indexes []int) (error, int) {
	return g.board.ValidatePath(indexes)
}
func (g Game) HighestCellValue() int64 {
	c, _ := g.board.HighestValue()
	return c.Value()
}
func (g Game) Moves() int {
	return g.moves
}
func (g Game) SelectedCells() []int {
	return g.selectedCells
}
func (g Game) Cells() []cell.Cell {
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
