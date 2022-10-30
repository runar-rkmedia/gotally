package tallylogic

import (
	"testing"

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

func mustCreateNewGameForTest(mode GameMode, template *GameTemplate, options ...NewGameOptions) Game {
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

func TestGame_Play(t *testing.T) {
	type playgame = func(game *Game, t *testing.T)
	tests := []struct {
		name          string
		fields        Game
		play          playgame
		expectedScore int64
	}{
		{
			"Play the first daily board",
			mustCreateNewGameForTest(GameModeTemplate, GetGameTemplateById("Ch:NotTheObviousPath")),
			func(g *Game, t *testing.T) {
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
					ok := g.instruct(v)
					// t.Logf("Performed isntruction %d %v \n%s", i, v, g.board.PrintBoard(h))
					if !ok {
						t.Errorf("Failed at instruction %d %#v\n%s", i, v, g.board.PrintBoard(h))
						return
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
			mustCreateNewGameForTest(GameModeDefault, nil, NewGameOptions{Seed: 123}),
			func(g *Game, t *testing.T) {
				instructions := []any{
					SwipeDirectionRight,
					SwipeDirectionUp,
					SwipeDirectionDown,
					SwipeDirectionLeft,
					SwipeDirectionDown,
					SwipeDirectionRight,
					SwipeDirectionDown,
					SwipeDirectionLeft,
					[]int{22, 21, 20},
					SwipeDirectionUp,
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

					ok = g.instruct(v)
					if !ok {
						t.Errorf("Original game failed at instruction %d %#v\n%s", i, v, g.board.PrintBoard(h))
						return
					}
					got := g.Print()
					ok = gCopy.instruct(v)
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
			g := &Game{
				board:         tt.fields.board,
				selectedCells: tt.fields.selectedCells,
				cellGenerator: tt.fields.cellGenerator,
				Rules:         tt.fields.Rules,
				score:         tt.fields.score,
			}
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
