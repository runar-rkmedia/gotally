package tallylogic

import (
	"fmt"

	"github.com/runar-rkmedia/gotally/tallylogic/cell"
)

// These are a collection of constructed gamepositions that are intended to
// give a challenge

func NewDailyBoard() *TableBoard {
	return &TableBoard{
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
}

func GetGameTemplateById(ID string) *GameTemplate {
	for _, g := range TutorialGames {
		if g.ID == ID {
			return &g
		}
	}
	return nil
}

var (
	TutorialGames []GameTemplate = []GameTemplate{

		*NewGameTemplate(GameModeTutorial, "Sum&Product", "Sum & Product", "Get a brick to 36. Bricks can be added, or multiplied together. Try combining 5,4 into 9. What can you do with that 3 and 6?", 3, 3).
			SetStartingLayout(
				0, 0, 5,
				0, 0, 4,
				3, 6, 9,
			).
			SetGoalCheckerLargestValue(36).SetMaxMoves(8),
		*NewGameTemplate(GameModeTutorial, "TimesOne", "Times One", "Get a brick to 1000. Learning the usefulness of 1 times X", 3, 3).
			SetStartingLayout(
				500, 1, 0,
				1, 0, 100,
				0, 0, 5,
			).
			SetGoalCheckerLargestValue(1000).SetMaxMoves(5),
		*NewGameTemplate(GameModeTutorial, "AllLinedUp", "All Lined Up", "Get a brick to 512. Can you combine them all into one?", 4, 4).
			SetStartingLayout(
				4, 1, 1, 4,
				2, 16, 8, 4,
				8, 32, 4, 4,
				2, 8, 8, 1,
			).
			SetGoalCheckerLargestValue(512).SetMaxMoves(7),

		*NewGameTemplate(GameModeTutorial, "Ch:NotTheObviousPath", "Challenge: Not the obvious path", "Get a brick to 512. Multiplication is your friend.", 5, 5).
			SetStartingLayout(
				0, 2, 1, 0, 1,
				64, 4, 4, 1, 2,
				64, 8, 4, 1, 0,
				12, 3, 1, 0, 0,
				16, 0, 0, 0, 0,
			).
			SetGoalCheckerLargestValue(512).SetMaxMoves(10),
	}
)

type GoalChecker interface {
	Description() string
	Check(Game) bool
}

type GoalCheck struct {
	description string
}

type GoalCheckLargestCell struct {
	GoalCheck
	TargetCellValue uint64
}

type DefeatCheckerNoMoreMoves struct {
	GoalCheck
}
type GoalCheckerMaxMoves struct {
	GoalCheck
	MaxMoves int
}

func (g GoalCheck) Description() string {
	return g.description
}

func (g GoalCheck) Check(game Game) bool {
	return false
}
func (g DefeatCheckerNoMoreMoves) Check(game Game) bool {
	hints := game.GetHint()
	if len(hints) > 0 {
		return false
	}

	hash := game.board.Hash()

	// Swipe in all directions and see if we get new hints, or if that results in the same board
	var copy Game
	for _, dir := range []SwipeDirection{SwipeDirectionUp, SwipeDirectionRight, SwipeDirectionDown, SwipeDirectionLeft} {

		copy = game.Copy()
		copy.board.SwipeDirection(dir)
		hints = copy.GetHint()
		if len(hints) > 0 {
			return false
		}
	}

	// If we have swiped in all directions, not found any hints, and then the board is the same,
	// the user should be game over.
	// TODO: Verify that there are no edge-cases here, where a some other combination of swipes would make the game game-over
	return hash == copy.board.Hash()

}
func (g GoalCheckLargestCell) Description() string {
	return fmt.Sprintf("Get a cell to at least a value of %d", g.TargetCellValue)
}
func (g GoalCheckLargestCell) Check(game Game) bool {
	for _, c := range game.Cells() {
		value := c.Value()
		if uint64(value) >= g.TargetCellValue {
			return true
		}

	}
	return false
}
func (g GoalCheckerMaxMoves) Check(game Game) bool {
	return g.MaxMoves > game.Moves()
}

type GameTemplate struct {
	ID, Name, Description string
	Rows, Columns         int
	Rules                 GameRules
	Board                 TableBoard
	GoalChecker           GoalChecker
	DefeatChecker         GoalChecker
}

func DefaultGameRules(sizeX, sizeY int) GameRules {
	return GameRules{
		GameMode:        GameModeRandom,
		SizeX:           sizeX,
		SizeY:           sizeY,
		RecreateOnSwipe: true,
		WithSuperPowers: true,
		StartingBricks:  6,
	}
}
func DefaultChallengeGameRules(sizeX, sizeY int, mode GameMode) GameRules {
	rules := DefaultGameRules(sizeX, sizeY)
	rules.StartingBricks = 0
	rules.RecreateOnSwipe = false
	rules.GameMode = mode
	return rules
}

func NewGameTemplate(mode GameMode, id, name, description string, rows, columns int) *GameTemplate {
	return &GameTemplate{
		ID:          id,
		Name:        name,
		Description: description,
		Board:       NewTableBoard(rows, columns),
		Rows:        rows,
		Columns:     columns,
		Rules:       DefaultChallengeGameRules(rows, columns, mode),
	}
}

func (t *GameTemplate) SetGoalCheckerLargestValue(targetCellValue uint64) *GameTemplate {
	t.GoalChecker = GoalCheckLargestCell{
		TargetCellValue: targetCellValue,
	}
	t.Rules.TargetCellValue = targetCellValue
	return t

}
func (t *GameTemplate) SetMaxMoves(moves int) *GameTemplate {
	t.DefeatChecker = GoalCheckerMaxMoves{
		MaxMoves: moves,
	}
	t.Rules.MaxMoves = uint64(moves)
	return t

}

// Deprecated, use SetStartingLayoutUints
func (t *GameTemplate) SetStartingLayout(brickValue ...int64) *GameTemplate {
	t.Board.cells = cellCreator(brickValue...)
	return t
}
func (t *GameTemplate) SetStartingLayoutUints(brickValue ...uint64) *GameTemplate {
	t.Board.cells = cellCreatorUints(brickValue...)
	return t
}
func (t *GameTemplate) SetStartingCells(cells []cell.Cell) *GameTemplate {
	t.Board.cells = cells
	return t
}
func (t *GameTemplate) Create() GameTemplate {
	g := GameTemplate{
		Name:        t.Name,
		Description: t.Description,
		Rows:        t.Rows,
		Columns:     t.Columns,
		Rules: GameRules{
			GameMode:        t.Rules.GameMode,
			SizeX:           t.Rules.SizeX,
			SizeY:           t.Rules.SizeY,
			RecreateOnSwipe: t.Rules.RecreateOnSwipe,
			WithSuperPowers: t.Rules.WithSuperPowers,
			StartingBricks:  t.Rules.StartingBricks,
			MaxMoves:        t.Rules.MaxMoves,
			TargetCellValue: t.Rules.TargetCellValue,
			TargetScore:     t.Rules.TargetScore,
		},
		Board: TableBoard{
			rows:    t.Board.rows,
			columns: t.Board.columns,
			id:      t.ID,
		},
		GoalChecker:   t.GoalChecker,
		DefeatChecker: t.DefeatChecker,
	}

	g.Board.cells = append(g.Board.cells, t.Board.cells...)

	return g
}
