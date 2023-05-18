package tallylogic

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"strconv"

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
	ID           string
	board        BoardController
	boardAtStart BoardController
	// deprecated (I think atleast, I dont believe there is a need for this)
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
	History       CompactHistory
}

func (g Game) Seed() (uint64, uint64) {
	return g.cellGenerator.Seed()
}
func (g Game) Board() BoardController {
	return g.board
}

type GameRules struct {
	ID       string
	GameMode GameMode
	SizeX    int
	SizeY    int
	// TODO: not implemented
	TargetCellValue uint64
	// TODO: not implemented
	MaxMoves uint64
	// TODO: not implemented
	TargetScore uint64
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
	GameModeRandom GameMode = iota + 1
	GameModeTutorial
	GameModeRandomChallenge
)

func (mode GameMode) String() string {
	switch mode {
	case GameModeRandom:
		return fmt.Sprintf("Default (%d)", mode)
	case GameModeTutorial:
		return fmt.Sprintf("Tutorial (%d)", mode)
	case GameModeRandomChallenge:
		return fmt.Sprintf("Challenge (%d)", mode)
	}
	return fmt.Sprintf("err: Invalid mode: %d", mode)
}
func (mode GameMode) MarshalJSON() ([]byte, error) {
	// return []byte(`"banana"`), nil
	return json.Marshal(mode.String())
}

// Copies the game and all values to a new game
func (g Game) Copy() Game {
	// IMPORTANT: Do not copy the onDidChangeEvents here.
	if g.cellGenerator == nil {
		panic("no cellgenerator for game")
	}
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
	game.History = NewCompactHistoryFromGame(game)
	game.History.c = append(game.History.c, g.History.c...)
	return game

}

type NewGameOptions struct {
	TableBoardOptions
	Seed  uint64
	State uint64
}

func RestoreGame(g *types.Game) (Game, error) {
	var mode GameMode

	switch g.Rules.Mode {
	case types.RuleModeInfiniteEasy, types.RuleModeInfiniteNormal, types.RuleModeInfiniteHard:
		mode = GameModeRandom
	case types.RuleModeChallenge:
		mode = GameModeRandomChallenge
	case types.RuleModeTutorial:
		mode = GameModeTutorial
	default:
		return Game{}, fmt.Errorf("unsupported game-mode from rules: %v", g.Rules.Mode)
	}
	game := Game{
		ID:            g.ID,
		board:         nil,
		selectedCells: []int{},
		cellGenerator: nil,
		Rules: GameRules{
			ID:              g.Rules.ID,
			GameMode:        mode,
			SizeX:           int(g.Rules.Columns),
			SizeY:           int(g.Rules.Rows),
			RecreateOnSwipe: g.Rules.RecreateOnSwipe,
			TargetCellValue: g.Rules.TargetCellValue,
			TargetScore:     g.Rules.TargetScore,
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
		score:         int64(g.Score),
		moves:         int(g.Moves),
		Description:   g.Description,
		Name:          g.Name,
		GoalChecker:   nil,
		DefeatChecker: nil,
		History:       NewCompactHistoryFromBinary(int(g.Rules.Columns), int(g.Rules.Rows), g.History),
	}

	game.DefeatChecker = DefeatCheckerNoMoreMoves{}
	if game.Rules.TargetCellValue > 0 {

		game.GoalChecker = GoalCheckLargestCell{
			TargetCellValue: game.Rules.TargetCellValue,
		}
	} else if game.Rules.TargetScore > 0 {
		return game, fmt.Errorf("Not implemented: targetScore")
	} else {
		game.GoalChecker = GoalCheck{"Game runs forever (default)"}

	}
	switch game.Rules.GameMode {
	case GameModeRandom:
	case GameModeTutorial:
	case GameModeRandomChallenge:

		if game.Rules.TargetCellValue == 0 {
			return game, fmt.Errorf("TargetCellValue must be set for games of challenge-type")
		}
	default:
		return game, fmt.Errorf("not implemented for this gamemode: gameMode: %v typeGameMode: (%v)", game.Rules.GameMode, g.Rules.Mode)
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
	}
	for _, o := range options {
		game.Rules.Options = o
	}

	r := randomizer.NewRandomizerFromSeed(game.Rules.Options.Seed, game.Rules.Options.State)
	game.cellGenerator = cellgenerator.NewCellGenerator(r)
	switch mode {
	case GameModeRandom:
		board := NewTableBoard(5, 5, game.Rules.Options.TableBoardOptions)
		game.board = &board
		game.Description = "Default game, 5x5"
		game.DefeatChecker = DefeatCheckerNoMoreMoves{}
		game.GoalChecker = GoalCheck{"Game runs forever"}
	case GameModeTutorial, GameModeRandomChallenge:
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
		return game, fmt.Errorf("Invalid gamemode: %d %s", mode, string(debug.Stack()))
	}
	game.History = NewCompactHistory(game.Rules.SizeX, game.Rules.SizeY)
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
	game.boardAtStart = game.board.Copy()
	if len(game.board.Cells()) != (game.Rules.SizeX * game.Rules.SizeY) {
		return game, fmt.Errorf("Game has invalid size: %d cells, %dx%d, mode %v template %v", len(game.board.Cells()), game.Rules.SizeX, game.Rules.SizeY, mode, template)
	}
	return game, nil
}

func (g *Game) GetHint() map[string]Hint {
	return g.Hinter.GetHints()
}

// Returns hints in a consistant order, mostly useful for tests
func (g *Game) GetHintConsistantly(ctx context.Context, max int) []Hint {
	g2 := g.Copy()
	return g2.Hinter.GetNHintsConsistant(ctx, max)
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

func (g *Game) SoftSwipe(direction SwipeDirection) (wouldChange bool) {
	_, wouldChange = g.board.SwipeDirectionSoft(direction)
	return wouldChange

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
func (g *Game) ReplaceBasedOn(game Game) error {
	g.boardAtStart = game.board
	return nil
}
func (g *Game) CanUndo() bool {
	undos := 0
	others := 0
	g.History.Iterate(
		func(dir SwipeDirection, i int) error { others++; return nil },
		func(path []int, i int) error { others++; return nil },
		func(helper Helper, i int) error {
			if helper == helperUndo {
				undos++
			} else {
				others++
			}
			return nil
		},
	)
	undone := undos * 2
	return undone < others
}
func (g *Game) Undo() error {
	if g.moves == 0 {
		return fmt.Errorf("Cannot undo at start of game")
	}
	if g.boardAtStart == nil {
		return fmt.Errorf("Invalid state: Could not locate data for board-start.")
	}
	history, err := g.History.FilterForUndo(true)
	if err != nil {
		return fmt.Errorf("failed to undo game: %w", err)
	}
	hbytes := g.History.BytesCopy()
	g.board = g.boardAtStart.Copy()
	g.History = NewCompactHistory(g.Rules.SizeX, g.Rules.SizeY)
	// g.moves = 0
	g.score = 0
	for i := 0; i < len(history); i++ {
		ins := history[i]
		if ins.IsHelperUndo() {
			panic("Nested undo!")
		}
		ok := g.Instruct(ins)
		g.moves--
		// time.Sleep(1 * time.Millisecond)
		if !ok {
			return fmt.Errorf("failed to apply instruction  %d %#v", i, ins)
		}

	}

	g.History.c = hbytes
	g.History.AddUndo()
	g.moves++

	return nil
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

// Evaluates for the path, and mutates the game accordingly
//
// changes include:
// - Changes to cells
// - Score
// - Moves
// - History
func (g *Game) EvaluateForPath(path []int) bool {
	points, _, err := g.board.EvaluatesTo(path, true, false)
	if err != nil {
		return false
	}
	g.increaseScore(points * 2)
	g.inceaseMoveCount()
	g.History.AddPath(path)
	return true
}

func (g Game) Print() string {
	return g.board.String()
}
func (g Game) PrintForSelection(selection []int) string {
	return g.board.PrintForSelection(selection)
}
func (g Game) PrintForSelectionNoColor(selection []int) string {
	return g.board.PrintForSelectionNoColor(selection)
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
func (g Game) BoardSize() int {
	return g.Rules.SizeX * g.Rules.SizeY
}

// PathValues returns the values for the given path.
// Mostly used for diagnostics
func (g Game) PathValues(path []int) []int64 {
	cells := g.Cells()
	pathValues := make([]int64, len(path))
	for i, index := range path {
		pathValues[i] = cells[index].Value()
	}
	return pathValues
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
func (g Game) SoftEvaluatesForPath(path []int) (int64, EvalMethod, error) {
	if len(path) < 2 {
		return 0, EvalMethodNil, nil
	}
	indexes := path[0 : len(path)-1]
	targetIndex := path[len(path)-1]
	cell := g.Board().GetCellAtIndex(targetIndex)
	if cell == nil {
		return 0, EvalMethodNil, fmt.Errorf("cell at index %d returned nil", targetIndex)
	}
	targetValue := cell.Value()
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
	return s
}
