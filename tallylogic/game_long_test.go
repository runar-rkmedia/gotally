package tallylogic_test

import (
	"context"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/MarvinJWendt/testza"
	"github.com/runar-rkmedia/gotally/tallylogic"
)

func init() {
	testza.SetShowStartupMessage(false)
}

// CAUTION: THERE IS CURRENTLY A RACE-CONDITION here.
// This should be fixed with running t

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

	// the random source is used for swiping when there are no other valid moves.
	// changing the seed affects how many moves is possible before one is game over.
	rnd := rand.NewSource(133)
	swipeDirections := []tallylogic.SwipeDirection{
		tallylogic.SwipeDirectionUp,
		tallylogic.SwipeDirectionRight,
		tallylogic.SwipeDirectionDown,
		tallylogic.SwipeDirectionLeft}
	// 2531 is actually the number of moves until this game is game over
	// it could probably go on a bit longer with a more tactical use of swipe-direction
	expectedMoves := 2530
outer:
	for i := 0; i < expectedMoves; i++ {
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
			swipes := map[int]struct{}{}
			for len(swipes) < 4 {
				diri := int(rnd.Int63()) % len(swipeDirections)
				if _, existing := swipes[diri]; existing {
					continue
				}
				swipes[diri] = struct{}{}
				dir := swipeDirections[diri]
				wouldChange := game.SoftSwipe(dir)
				var changed bool
				if wouldChange {
					changed = game.Swipe(dir)
					// t.Logf("Swiped %s after %d attempts", dir, len(swipes))
					if changed {
						continue outer
					}
				}
				// t.Logf("attemted to swipe in %v directions", len(swipes))
			}
			t.Log(game.Print())
			t.Fatalf("Game is over after %d moves", game.Moves())
		}
		for _, v := range hints {
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
	testza.AssertEqual(t, expectedMoves, game.Moves(), "The game should have the expected number of moves for consistency ")
	testza.AssertEqual(t, int64(2456090364), game.Score(), "The game should have the expected score for consistency")

	// 2. Attempt to restart the game, and replay the game from history.
	// This should ideally be fast
	start := time.Now()
	history, err := game.History.All()
	if err != nil {
		t.Fatalf("failed to get history: %v", err)
	}
	for i, v := range history {
		switch {
		case v.IsSwipe:
			gameCopyAtStart.Swipe(v.Direction)
		case v.IsPath:
			ok := gameCopyAtStart.EvaluateForPath(v.Path)
			if !ok {
				t.Log(gameCopyAtStart.Moves(), gameCopyAtStart.Print())
				t.Log(v)
				t.Fatalf("expected %d/%d instruction to evaluate, but it did not", i, len(history))
			}
		default:
			panic("NotImplemented: Helper in test")
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
	t.Logf("Wrote %d bytes to %s for game. with length %d", game.History.Size(), outFile, game.History.Length())
	testza.AssertEqual(t, 2102, game.History.Size(), "game history-size should match expected value (it should be low, like below 1 byte per move)")

}

var outFile = "./testdata/longplay-game-data.bin"

func TestGame_LongPlayFromBinary(t *testing.T) {
	// Test a previously stored game-history, in binary-form
	game, err := tallylogic.NewGame(tallylogic.GameModeRandom, nil, tallylogic.NewGameOptions{
		Seed:  123,
		State: 123,
	})
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(outFile)
	testza.AssertNoError(t, err)
	err = game.History.Restore(b)
	testza.AssertNoError(t, err)
	start := time.Now()
	history, err := game.History.All()
	if err != nil {
		t.Fatalf("failed to get history: %v", err)
	}
	for i, v := range history {
		switch {
		case v.IsSwipe:
			game.Swipe(v.Direction)
		case v.IsPath:
			ok := game.EvaluateForPath(v.Path)
			if !ok {
				t.Log(game.Moves(), game.Print())
				t.Log(v)
				t.Fatalf("expected %d/%d instruction to evaluate, but it did not", i, len(history))
			}
		default:
			panic("NotImplemented: Helper in test")
		}
	}
	end := time.Now()
	diff := end.Sub(start)
	t.Logf("recreate completed in %s", diff.String())
}
func BenchmarkLongPlayFromBinaryViaAll(b *testing.B) {
	bytes, err := os.ReadFile(outFile)
	if err != nil {
		b.Fatalf("failed to read file %s: %v", outFile, err)
	}
	for i := 0; i < b.N; i++ {
		// Test a previously stored game-history, in binary-form
		game, err := tallylogic.NewGame(tallylogic.GameModeRandom, nil, tallylogic.NewGameOptions{
			Seed:  123,
			State: 123,
		})
		if err != nil {
			b.Fatal(err)
		}

		err = game.History.Restore(bytes)
		if err != nil {
			b.Fatalf("failed to restore from file %s %v", outFile, err)
		}
		if game.History.Size() != 2102 {
			b.Fatalf("expected size of restore to be 2102")
		}
		history, err := game.History.All()
		if err != nil {
			b.Fatalf("failed to get history: %v", err)
		}
		for i, v := range history {
			switch {
			case v.IsSwipe:
				game.Swipe(v.Direction)
			case v.IsPath:
				ok := game.EvaluateForPath(v.Path)
				if !ok {
					b.Log(game.Moves(), game.Print())
					b.Log(v)
					b.Fatalf("expected %d/%d instruction to evaluate, but it did not", i, len(history))
				}
			default:
				panic("NotImplemented: Helper in test")
			}
		}
	}
}
func BenchmarkLongPlayFromBinaryViaIterate(b *testing.B) {
	bytes, err := os.ReadFile(outFile)
	if err != nil {
		b.Fatalf("failed to read file %s: %v", outFile, err)
	}
	for i := 0; i < b.N; i++ {
		// Test a previously stored game-history, in binary-form
		game, err := tallylogic.NewGame(tallylogic.GameModeRandom, nil, tallylogic.NewGameOptions{
			Seed:  123,
			State: 123,
		})
		if err != nil {
			b.Fatal(err)
		}

		err = game.History.Restore(bytes)
		if err != nil {
			b.Fatalf("failed to restore from file %s %v", outFile, err)
		}
		if game.History.Size() != 2102 {
			b.Fatalf("expected size of restore to be 2102")
		}
		game.History.Iterate(
			func(dir tallylogic.SwipeDirection, i int) error {
				game.Swipe(dir)
				return nil
			},
			func(path []int, i int) error {

				game.EvaluateForPath(path)
				return nil
			},
			func(helper tallylogic.Helper, i int) error {

				panic("NotImplemented: Helper in test")
				return nil
			},
		)
	}
}
