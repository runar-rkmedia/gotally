package api

import (
	"testing"

	"github.com/MarvinJWendt/testza"
)

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
