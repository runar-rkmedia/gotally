package tallylogic

import (
	"testing"

	"github.com/MarvinJWendt/testza"
	"github.com/go-test/deep"
	"github.com/gookit/color"
)

func BoardHightlighter(g *Game) func(CellValuer, int, string) string {

	return func(c CellValuer, index int, padded string) string {
		p := color.Yellow
		if c.Value() == 0 {
			p = color.Gray
		} else {

			nSelected := len(g.selectedCells)
			for i := 0; i < nSelected; i++ {
				if g.selectedCells[i] == index {
					if index == g.selectedCells[nSelected-1] {
						p = color.Cyan
					} else {
						p = color.LightBlue
					}
				}
			}
		}
		return p.Sprintf("[ %s ]", padded)
	}
}

func mustCreateNewGameForTest(mode GameMode, template *GameTemplate, options ...NewGameOptions) func() Game {
	return func() Game {

		opt := NewGameOptions{}
		if len(options) != 0 {
			opt = options[0]
		}
		if opt.Seed == 0 {
			opt.Seed = 1
		}
		if opt.State == 0 {
			opt.State = 1
		}
		game, err := NewGame(mode, template, opt)
		if err != nil {
			panic(err)
		}
		return game
	}
}

func TestGame_Undo(t *testing.T) {
	t.Run("Undoing a game should work", func(t *testing.T) {
		g := mustCreateNewGameForTest(GameModeTutorial, GetGameTemplateById("Ch:NotTheObviousPath"))()
		info := func() {
			t.Helper()
			// return
			t.Logf("[INFO] Moves: %d Score: %d HistoryLength: %d HistorySize: %d %s %s",
				g.Moves(), g.Score(), g.History.Length(), g.History.Size(), g.History.Describe(), g.Print())
		}
		gamesAtH := make([]Game, 6)
		gamesAtH[0] = g.Copy()

		// Checks that the games are equal, but ignores elements that would change due to undo.
		// For instance:
		// Moves-counter should increase whene doing an undo
		assertGameEquality := func(got, expected Game, expectedHistory string) {
			t.Helper()
			hasErr := false
			if expected.board == nil {
				t.Fatalf("The expected board was nil %#v", g)
			}
			if got.board == nil {
				t.Fatalf("The resulting board (got) was nil %#v", g)
			}
			if expected.Score() != got.Score() {
				t.Errorf("Score mismatch expected %d, got %d", expected.Score(), got.Score())
				hasErr = true
			}
			if expectedHistory != g.History.Describe() {
				t.Errorf("History mismatch expected %s, got %s", expectedHistory, g.History.Describe())
				hasErr = true
			}
			if expected.Print() != got.Print() {
				diffIndexes := getGameCellDiff(expected, g)
				hasErr = true
				t.Errorf(
					"The game-layout was not reset to the previous layout. Expected %s but got %s",
					expected.PrintForSelection(diffIndexes), got.PrintForSelection(diffIndexes))
			}
			if hasErr {
				t.FailNow()
			}
		}

		err := g.Undo()
		testza.AssertNotNil(t, err)
		testza.AssertContains(t, err.Error(), "Cannot undo at start")

		testza.AssertEqual(t, 0, g.Moves(), "Moves should be 0 at start")
		testza.AssertEqual(t, 0, g.History.Length(), "History-Length should be 0 at start")
		info()

		// Move 1 (1):
		t.Log("Swiping Up")
		changed := g.Swipe(SwipeDirectionUp)
		testza.AssertTrue(t, changed, "Expected board to change after swipe")
		testza.AssertEqual(t, 1, g.Moves(), "Moves should have increased")
		testza.AssertEqual(t, 1, g.History.Length(), "History-Length should be 1")
		gamesAtH[1] = g.Copy()
		info()

		// Move 2 (2):
		t.Log("Swiping Right")
		changed = g.Swipe(SwipeDirectionRight)
		testza.AssertTrue(t, changed, "Expected board to change after swipe")
		testza.AssertEqual(t, 2, g.Moves(), "Moves should have increased")
		testza.AssertEqual(t, 2, g.History.Length(), "History-Length should be 2")
		gamesAtH[2] = g.Copy()
		info()

		// Undo (1) (3):
		t.Log("Undoing")
		err = g.Undo()
		testza.AssertNoError(t, err, "Undo should not err")
		gamesAtH[3] = g.Copy()
		info()
		assertGameEquality(g, gamesAtH[1], "U;R;Z;")

		// Combine (2) (4):
		t.Log("Combining")
		ok := g.EvaluateForPath([]int{3, 2, 1, 6})
		testza.AssertTrue(t, ok, "Expecte EvaluateForPath to report ok")
		gamesAtH[4] = g.Copy()
		info()

		// Combine (3) (5):
		t.Log("Combining")
		ok = g.EvaluateForPath([]int{4, 9, 8, 7, 6})
		testza.AssertTrue(t, ok, "Expecte EvaluateForPath to report ok")
		gamesAtH[5] = g.Copy()
		info()

		// Swiping (4) (6):
		t.Log("Swiping Down")
		changed = g.Swipe(SwipeDirectionDown)
		testza.AssertTrue(t, changed, "Expected board to change after swipe")
		info()

		// Undo (3) (7):
		t.Log("Undoing")
		err = g.Undo()
		testza.AssertNoError(t, err, "Undo should not err")
		info()
		assertGameEquality(g, gamesAtH[5], "U;R;Z;3,2,1,6;4,9,8,7,6;D;Z;")

		// Undo (2) (8):
		t.Log("Undoing twice in a row")
		err = g.Undo()
		testza.AssertNoError(t, err, "Undo should not err")
		info()
		assertGameEquality(g, gamesAtH[4], "U;R;Z;3,2,1,6;4,9,8,7,6;D;Z;Z;")

		t.Log("Undoing a third time in a row")
		err = g.Undo()
		testza.AssertNoError(t, err, "Undo should not err")
		info()
		assertGameEquality(g, gamesAtH[1], "U;R;Z;3,2,1,6;4,9,8,7,6;D;Z;Z;Z;")

		t.Log("Undoing a fourth time in a row")
		err = g.Undo()
		testza.AssertNoError(t, err, "Undo should not err")
		info()
		assertGameEquality(g, gamesAtH[0], "U;R;Z;3,2,1,6;4,9,8,7,6;D;Z;Z;Z;Z;")
		for i := 0; i < len(gamesAtH); i++ {
			t.Logf("gamesAtH: index: %d, moves %d score %d %s", i, gamesAtH[i].Moves(), gamesAtH[i].Score(), gamesAtH[i].Print())

		}
		testza.AssertFalse(t, g.CanUndo())
		t.Log(g.CanUndo(), g.History.DescribeWithoutParams())
		// undo anyway, should not crash
		err = g.Undo()
		testza.AssertNotNil(t, err, "Expected to have error")

	})
}
func getGameCellDiff(a, b Game) []int {
	diffIndexes := []int{}
	for i := 0; i < a.BoardSize(); i++ {
		expected := a.board.GetCellAtIndex(i)
		got := b.board.GetCellAtIndex(i)
		if !expected.Equal(*got) {
			diffIndexes = append(diffIndexes, i)
		}

	}
	return diffIndexes
}

func TestGame_Play(t *testing.T) {
	type playgame = func(game *Game, t *testing.T)
	tests := []struct {
		name          string
		gamefactory   func() Game
		play          playgame
		expectedScore int64
	}{
		{
			"Play the first daily board",
			mustCreateNewGameForTest(GameModeTutorial, GetGameTemplateById("Ch:NotTheObviousPath")),
			func(g *Game, t *testing.T) {
				// TODO: update to use Instruction_
				instructions := []any{
					// Combine 4 into 4 (+) resulting in 8
					[2]int{2, 1},
					[2]int{2, 2},
					true,
					// Combine 8 into 8 (+) resulting in 16
					[2]int{2, 2},
					[2]int{1, 2},
					true,
					// Combine 4 and 16 into 64 (*) resulting in 128
					[2]int{1, 1},
					[2]int{1, 2},
					[2]int{0, 2},
					true,
					// Combine 1, 3, and 12 into 16 (+) resulting in 32
					[2]int{2, 3},
					[2]int{1, 3},
					[2]int{0, 3},
					[2]int{0, 4},
					true,
					SwipeDirectionDown,
					// Combine 1 into 1 (+) resulting in 2
					[2]int{3, 4},
					[2]int{2, 4},
					true,
					// Combine 2, 2 and 32 into 128 (+) resulting in 256
					[2]int{2, 4},
					[2]int{1, 4},
					[2]int{0, 4},
					[2]int{0, 3},
					true,
					// Combine 1 into 1 (+) resulting in 2
					[2]int{3, 3},
					[2]int{4, 3},
					true,
					// Combine 2 into 2 (+) resulting in 4
					[2]int{4, 4},
					[2]int{4, 3},
					true,
					SwipeDirectionUp,
					SwipeDirectionLeft,
					// Combine 4 and 64 into 256 (*) resulting in 512
					[2]int{1, 0},
					[2]int{0, 0},
					[2]int{0, 1},
					true,
				}
				h := BoardHightlighter(g)
				for i, v := range instructions {
					switch vt := v.(type) {
					case SwipeDirection:
						ok := g.Swipe(SwipeDirection(vt))
						if !ok {
							t.Errorf("Failed at instruction %d %#v\n%s", i, v, g.board.PrintBoard(h))
							return
						}
					case bool:
						ok := g.EvaluateSelection()
						if !ok {
							t.Errorf("Failed at instruction %d %#v\n%s", i, v, g.board.PrintBoard(h))
							return
						}
					case [2]int:
						g.selectCellCoord(vt[0], vt[1])
					default:
						t.Fatalf("unhandled type!!! %#v %v", vt, v)
					}

				}
			},
			960,
		},
		{
			"Game.History should reliably replay the game with the seeded randomizer",
			// This is important for at least these reasons:
			// 1. The game-solver should be able to play many moves ahead to look for good solutions
			// 2. A played game should be verifiable, for instace to detect some forms of cheating with highscores.
			// #. A played game should be replayable, in a UI.
			mustCreateNewGameForTest(GameModeRandom, nil, NewGameOptions{Seed: 123}),
			func(g *Game, t *testing.T) {
				instructions := []Instruction_{
					NewSwipeInstruction_(SwipeDirectionRight),
					NewSwipeInstruction_(SwipeDirectionUp),
					NewSwipeInstruction_(SwipeDirectionDown),
					NewSwipeInstruction_(SwipeDirectionLeft),
					NewSwipeInstruction_(SwipeDirectionDown),
					NewSwipeInstruction_(SwipeDirectionRight),
					NewSwipeInstruction_(SwipeDirectionDown),
					NewSwipeInstruction_(SwipeDirectionLeft),
					NewPathInstruction_([]int{22, 21, 20}),
					NewSwipeInstruction_(SwipeDirectionUp),
				}
				gCopy := g.Copy()
				h := BoardHightlighter(g)
				hCopy := BoardHightlighter(&gCopy)
				initialOriginalStr := gCopy.Print()
				initialCopyStr := gCopy.Print()
				if initialOriginalStr != initialCopyStr {
					t.Fatalf("The initial copy is not equal: got:\n %s\n want:\n %s", initialCopyStr, initialOriginalStr)

				}
				t.Log("originalGame", initialOriginalStr)
				for i, v := range instructions {
					desc := g.DescribeInstruction(v)
					descCopy := g.DescribeInstruction(v)
					var ok bool

					ok = g.Instruct(v)
					if !ok {
						t.Errorf("Original game failed at instruction %d %#v\n%s", i, v, g.board.PrintBoard(h))
						return
					}
					got := g.Print()
					ok = gCopy.Instruct(v)
					if !ok {
						t.Errorf("Copy failed at instruction %d %#v\n%s", i, v, gCopy.board.PrintBoard(hCopy))
						return
					}
					want := gCopy.Print()
					if got != want {
						if diff := deep.Equal(*g, gCopy); diff != nil {
							t.Errorf("DIFF! %#v", diff)
						}

						t.Fatalf("The copied game is not equal to the original after instruction %d (%s) (%s): got: \n %s \n want: \n %s", i, desc, descCopy, got, want)
					}
				}
			},
			16,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gg := tt.gamefactory()
			g := &Game{
				board:         gg.board,
				selectedCells: gg.selectedCells,
				cellGenerator: gg.cellGenerator,
				Rules:         gg.Rules,
				score:         gg.score,
			}
			g.History = NewCompactHistoryFromGame(*g)
			tt.play(g, t)
			if tt.expectedScore != g.score {
				t.Log("Selected", g.selectedCells)
				t.Log(g.board.PrintBoard(BoardHightlighter(g)))
				t.Logf("Moves %d", g.moves)
				t.Errorf("Expected score after play did not match, got %d, want %d", g.score, tt.expectedScore)
			}
		})
	}
}
