package api

import (
	"fmt"
	"testing"

	"github.com/MarvinJWendt/testza"
	"github.com/pelletier/go-toml/v2"
	tallyv1 "github.com/runar-rkmedia/gotally/gen/proto/tally/v1"
	"github.com/runar-rkmedia/gotally/generated"
	"github.com/runar-rkmedia/gotally/tallylogic"
)

func TestApi_Challange_Solving(t *testing.T) {
	runs := 3
	generated.ReadGeneratedBoardsFromDisk(generated.Options{MaxItems: 3})
	for i := 1; i <= runs; i++ {

		t.Run(fmt.Sprintf("Should be able to solve challenges %d/%d", i, runs), func(t *testing.T) {

			ts := newTestApi(t)

			newGame := ts.NewGame(tallyv1.GameMode_GAME_MODE_RANDOM_CHALLENGE)
			t.Logf("Attempting to solve the game '%s' '%s'\n%s",
				newGame.Msg.Board.Name,
				newGame.Msg.Description,
				ts.Game().Print())

			res := ts.SolveGameWithHints(3)
			testza.AssertTrue(t, res.Msg.DidWin, "expected game to be won (solved)")
			if res.Msg.DidWin == false {
				template := getTemplate(newGame.Msg.Board.Name)
				if template == nil {
					panic("not found: " + newGame.Msg.Board.Name)
				}
				b, err := toml.Marshal(template)
				if err != nil {
					panic(err)
				}
				t.Fatal("Solution", string(b))

			}
		})
	}
}
func TestApi_Challange_Restart(t *testing.T) {
	generated.ReadGeneratedBoardsFromDisk(generated.Options{MaxItems: 3})
	t.Run("Should be able to restart", func(t *testing.T) {
		ts := newTestApi(t)
		ts.NewGame(tallyv1.GameMode_GAME_MODE_RANDOM_CHALLENGE)

		{
			// This direction is a bit flaky. I don' actually care about swiping, I just need to perform any move.
			// for now, this just works because of a fluke. If it ever fails here, we should probably
			// have the api return a hint first, and then just perform thatn hint
			res := ts.SwipeDown()
			testza.AssertGreater(t, res.Msg.Moves, int64(0), "Moves should be 1")
		}
		{
			res := ts.RestartGame()
			testza.AssertEqual(t, res.Msg.Moves, int64(0), "Moves should be 0")
		}
	})
}
func getTemplate(s string) *tallylogic.GameTemplate {
	for i := 0; i < len(generated.GeneratedTemplates); i++ {
		if generated.GeneratedTemplates[i].Name == s {
			return &generated.GeneratedTemplates[i]

		}
	}
	return nil
}
