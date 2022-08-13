package tallylogic

import (
	"testing"

	"github.com/gookit/color"
)

func BoardHightlighter(g *Game) func(Cell, int, string) string {

	return func(c Cell, index int, padded string) string {
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

func TestGame_Play(t *testing.T) {
	type fields struct {
		board         BoardController
		selectedCells []int
		cellGenerator CellGenerator
		rules         GameRules
		score         int64
	}
	type playgame = func(game *Game, t *testing.T)
	tests := []struct {
		name          string
		fields        fields
		play          playgame
		expectedScore int64
		expectedBoard TableBoard
	}{
		{
			"Play the first daily board",
			fields{
				board: &FirstDailyBoard,
				rules: GameRules{
					BoardType:       0,
					GameMode:        GameModeDefault,
					SizeX:           FirstDailyBoard.columns,
					SizeY:           FirstDailyBoard.rows,
					RecreateOnSwipe: false,
					WithSuperPowers: false,
				},
			},
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
			480,
			FirstDailyBoard,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Game{
				board:         tt.fields.board,
				selectedCells: tt.fields.selectedCells,
				cellGenerator: tt.fields.cellGenerator,
				rules:         tt.fields.rules,
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
