package tallylogiccompaction_test

import (
	"context"
	"testing"
	"time"

	"github.com/MarvinJWendt/testza"
	"github.com/runar-rkmedia/gotally/tallylogic"
)

func init() {
	testza.SetShowStartupMessage(false)
}

// This test is ment to check how long it would take to reacreate a previously
// game, based on its starting-parameters, and the history of the game. As long
// as this is fast enough, it could allow very compact storage of gamestate, as
// a database would not need to store a complete state for every instrction in
// the history.
func TestGame_LongPlay(t *testing.T) {
	// 1. Prepare a long game.
	game, err := tallylogic.NewGame(tallylogic.GameModeRandom, nil, tallylogic.NewGameOptions{
		Seed:  123,
		State: 123,
	})
	if err != nil {
		t.Fatal(err)
	}
	gameCopyAtStart := game.Copy()

	// game := mustCreateNewGameForTest(tallylogic.GameModeRandom, nil, tallylogic.NewGameOptions{Seed: 123})()

	t.Log(game.Print())

outer:
	// 1463 is actually the number of moves until this game is game over
	for i := 0; i < 1000; i++ {
		// if game.Moves() <= 17 {
		// 	t.Log(game.Moves(), game.Print())
		// }
		// Note that GetHint here may be slow. Often it is fast (microseconds)
		// But when we run it in a loop like this, it may use up to 40ms (on my machine)
		// so if it takes too long, it might have ran into positions for the game
		// where it requires deeper hints.
		var hints []tallylogic.Hint
		hints = game.GetHintConsistantly(context.TODO(), 1)
		if len(hints) == 0 {
			for _, dir := range []tallylogic.SwipeDirection{tallylogic.SwipeDirectionUp, tallylogic.SwipeDirectionRight, tallylogic.SwipeDirectionDown, tallylogic.SwipeDirectionLeft} {
				wouldChange := game.SoftSwipe(dir)
				var changed bool
				if wouldChange {
					changed = game.Swipe(dir)
					if changed {
						continue outer
					}
				}
			}
			t.Log(game.Print())
			t.Fatalf("Game is over after %d moves", game.Moves())
		}
		for _, v := range hints {
			// fmt.Printf("%2d hint %d/%d %s\n", i, j, len(keys), h[v])
			if v.Swipe != "" {
				if game.Swipe(v.Swipe) {
					break
				}
			}
			game.EvaluateForPath(v.Path)
			continue
		}

	}
	t.Logf("Moves: %d Score: %d", game.Moves(), game.Score())
	t.Log(game.Print())

	// Quick check to see if we have the expected game
	testza.AssertEqual(t, 1000, game.Moves(), "The game should have the expected number of moves for consistency ")
	testza.AssertEqual(t, int64(1665328), game.Score(), "The game should have the expected score for consistency")

	// 2. Attempt to restart the game, and replay the game from history.
	// This should ideally be fast
	start := time.Now()
	for i, v := range game.History {
		kind := tallylogic.GetInstructionType(v)
		switch kind {
		case tallylogic.InstructionTypeSwipe:
			dir, ok := tallylogic.GetInstructionAsSwipe(v)
			if !ok {
				t.Fatal("instrction was not of expected swipe")
			}
			gameCopyAtStart.Swipe(dir)
		default:
			path, ok := tallylogic.GetInstructionAsPath(v)
			if !ok {
				t.Fatal("instruction was not of expected path")
			}
			ok = gameCopyAtStart.EvaluateForPath(path)
			if !ok {
				t.Log(gameCopyAtStart.Moves(), gameCopyAtStart.Print())
				t.Log(v)
				t.Fatalf("expected %d/%d instruction to evaluate, but it did not", i, len(game.History))
			}
		}
	}
	end := time.Now()
	diff := end.Sub(start)
	t.Logf("recreate completed in %s", diff.String())
	testza.AssertLessOrEqual(t, diff.Milliseconds(), 100, "Recreating the game should be fast")
	testza.AssertEqual(t, game.Score(), gameCopyAtStart.Score(), "Expected the two games scores to be equal")
	testza.AssertEqual(t, game.Moves(), gameCopyAtStart.Moves(), "Expected the two games moves to be equal")
	testza.AssertEqual(t, game.Print(), gameCopyAtStart.Print(), "Expected the two games to be equal")
	seedOriginal, stateOriginal := game.Seed()
	seedCopy, stateCopy := game.Seed()
	testza.AssertEqual(t, seedOriginal, seedCopy, "Expected the two games seed to be equal")
	testza.AssertEqual(t, stateOriginal, stateCopy, "Expected the two games seedState to be equal")

}
