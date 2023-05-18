package api

import (
	"testing"

	"github.com/MarvinJWendt/testza"
	tallyv1 "github.com/runar-rkmedia/gotally/gen/proto/tally/v1"
)

func init() {
	testza.SetShowStartupMessage(false)
}

func TestApi_Tutorial(t *testing.T) {
	t.Run("Should be able to solve tutorial 1", func(t *testing.T) {
		ts := newTestApi(t)
		// Expect to have the initial tutorial-game
		ts.expectSimpleBoardEquality(
			0, 0, 5,
			0, 0, 4,
			3, 6, 9,
		)
		// Simply the same test as above, but using a snapshot of the api-response.
		// Technically more accurate, but harder to diagnose
		testza.SnapshotCreateOrValidate(t, "initial-cells", ts.initialSession.Msg.Session.Game.Board.Cells)
		t.Logf("%s Tutorial-game is loaded", logSuccess)

		res := ts.SolveGameWithHints(4)
		testza.AssertTrue(t, res.Msg.DidWin, "expected game to be won (solved)")
	})
}
func TestApi_Tutorial_Restart(t *testing.T) {
	t.Run("Should be able to restart", func(t *testing.T) {
		ts := newTestApi(t)
		ts.NewGame(tallyv1.GameMode_GAME_MODE_TUTORIAL)

		{
			res := ts.SwipeUp()
			testza.AssertGreater(t, res.Msg.Moves, int64(0), "Moves should be 1")
		}
		{
			res := ts.RestartGame()
			testza.AssertEqual(t, res.Msg.Moves, int64(0), "Moves should be 0")
		}
	})
}
