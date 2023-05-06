package api

import (
	"fmt"
	"testing"

	"github.com/MarvinJWendt/testza"
	"github.com/bufbuild/connect-go"
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

			challenge := ts.CreateDefaultChallenge()
			newGame := ts.NewGameChallenge(challenge.Msg.Id)
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

		// ------------------------------------------------------------
		ts.LogMark("Creating an initial challenge for testing")
		// ------------------------------------------------------------
		payload := tallyv1.CreateGameChallengeRequest{
			ChallengeNumber: 100,
			IdealMoves:      5,
			TargetCellValue: 5,
			Columns:         3,
			Rows:            3,
			Name:            "Simple challenge",
			Cells: toModalCells(cellCreator(
				1, 0, 0,
				0, 1, 0,
				0, 0, 1,
			)),
		}
		res, err := ts.client.CreateGameChallenge(ts.context, connect.NewRequest(&payload))
		testza.AssertNil(t, err, "Expected no errors from CreateGameChallenge")

		// ------------------------------------------------------------
		ts.LogMark("Attempting to play the challenge with id '%s'", res.Msg.Id)
		// ------------------------------------------------------------
		ts.NewGameChallenge(res.Msg.Id)

		checkHistoryLength := func() {
			game := ts.Game()
			t.Helper()
			t.Logf("Moves: %d History-Length: %d %s ID: %s", game.Moves(), game.History.Length(), game.Print(), game.ID)
			if game.Moves() != game.History.Length() {
				t.Errorf("Expected Game.History.Length()=%d to equal game.Moves()=%d", game.History.Length(), game.Moves())
			}

		}

		checkHistoryLength()
		{
			res := ts.SwipeDown()
			testza.AssertEqual(t, int64(1), res.Msg.Moves, "Moves should be 1")
		}
		checkHistoryLength()
		{
			res := ts.SwipeUp()
			testza.AssertEqual(t, int64(2), res.Msg.Moves, "Moves should be 2")
		}
		{
			res := ts.RestartGame()
			testza.AssertEqual(t, int64(0), res.Msg.Moves, "Moves should be 0")
		}
		checkHistoryLength()
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
