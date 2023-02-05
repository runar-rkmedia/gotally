package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/bufbuild/connect-go"
	"github.com/go-test/deep"
	model "github.com/runar-rkmedia/gotally/gen/proto/tally/v1"
	"github.com/runar-rkmedia/gotally/generated"
	"github.com/runar-rkmedia/gotally/sqlite"
	"gopkg.in/yaml.v3"
)

func TestApi_Restart(t *testing.T) {

	// resulted in the error-message
	// runtime error: invalid memory address or nil pointer dereference
	t.Run("Should not crash on restart (https://github.com/runar-rkmedia/gotally/issues/11)", func(t *testing.T) {

		generated.ReadGeneratedBoardsFromDisk(generated.Options{MaxItems: 3})
		ts := newTestApi(t)
		ctx := context.TODO()
		ts.SwipeUp()
		res, err := ts.client.RestartGame(ctx, connect.NewRequest(&model.RestartGameRequest{}))
		if err != nil {
			t.Fatalf("Restart Game failed %s", strErr(err))
		}
		if res.Msg.Board.Id == "" {
			t.Fatalf("expected board.id to not be empty: %#v", res)
		}
	})
}

func jsonCopy[T any](in T) T {
	b, err := json.Marshal(in)
	if err != nil {
		panic(err)
	}
	var out T
	err = json.Unmarshal(b, &out)
	if err != nil {
		panic(err)
	}
	return out
}

func getConnectErr(err error) *connect.Error {
	if connectErr := new(connect.Error); errors.As(err, &connectErr) {
		return connectErr
	}
	return nil
}
func strErr(err error) string {
	if err == nil {
		return ""
	}
	connectErr := getConnectErr(err)
	if connectErr == nil {
		return err.Error()
	}
	return fmt.Sprintf("[%s] %s %#v %s", connectErr.Code(), connectErr.Message(), connectErr.Meta(), connectErr.Error())
}

func TestApi_GameModes(t *testing.T) {
	t.Run("Modes should not change for challenge", func(t *testing.T) {
		ts := newTestApi(t)
		ctx := context.TODO()
		if ts.initialSession.Msg.Session.Game.Mode == 0 {
			t.Fatalf("the initial game-mode should not be 0: %s", prettyJson(ts.initialSession))
		}
		want := model.GameMode_GAME_MODE_RANDOM_CHALLENGE
		newGameResponse, err := ts.client.NewGame(ctx, connect.NewRequest(&model.NewGameRequest{
			Mode:    want,
			Variant: nil,
		}))
		if err != nil {
			t.Fatalf("New game failed failed %v", strErr(err))
		}
		got := newGameResponse.Msg.Mode
		if got != want {
			dump := ts.GetDBDump()
			var rule *sqlite.Rule
			var game *sqlite.Game
			for _, g := range dump.Games {
				if g.ID == newGameResponse.Msg.Board.Id {
					game = &g
					break
				}
			}
			if game == nil {
				t.Fatalf("failed to find the game in the database during error-checking")
			}
			for _, r := range dump.Rules {
				if r.ID == game.RuleID {
					rule = &r
					break
				}
			}
			t.Logf("The storage has this record of the rule.Mode: %v", rule.Mode)
			t.Fatalf("The game-mode should match. got %v, want %v", got, want)
		}
	})
}

func find[T any](arr []T, f func(t T) bool) *T {
	for _, v := range arr {
		if f(v) {
			return &v
		}
	}
	return nil
}

func TestApi_Consistent_State(t *testing.T) {

	// resulted in the error-message
	// runtime error: invalid memory address or nil pointer dereference
	t.Run("Db, tallylogic and api-response should have the same information", func(t *testing.T) {
		ts := newTestApi(t)
		// For a new session, we are starting with a short tutorial. These have Name and Description
		if ts.initialSession.Msg.Session.Game.Moves != 0 {
			t.Fatalf("Expected Game.Moves to be exactly 0, but was %d", ts.initialSession.Msg.Session.Game.Moves)
		}
		ddump := ts.GetDBDump()
		dbGame := find(ddump.Games, func(t sqlite.Game) bool { return t.ID == ts.initialGame.ID })
		// Check the data-base entry:
		if dbGame.Name.String == "" {
			t.Fatalf("Expected Session.Game.Board.Name to be non-empty, but was %s", dbGame.Name.String)
		}
		if dbGame.Description.String == "" {
			t.Fatalf("Expected Session.Game.Board.Description to be non-empty, but was %s", dbGame.Description.String)
		}
		t.Logf("%s The db-entry looks correct", logSuccess)

		// Check the returned initial response
		if ts.initialSession.Msg.Session.Game.Board.Name != dbGame.Name.String {
			t.Fatalf("The session.Board.Name '%s' did not match the expected Name '%s'",
				ts.initialSession.Msg.Session.Game.Board.Name, dbGame.Name.String,
			)
		}
		if ts.initialSession.Msg.Session.Game.Description != dbGame.Description.String {
			t.Fatalf("The session.Board.Description '%s' did not match expected Description '%s'",
				ts.initialSession.Msg.Session.Game.Description, dbGame.Description.String,
			)
		}
		t.Logf("%s The initial response looks correct", logSuccess)

		// Check the internal tallylogic-state for the game
		if ts.initialGame.Name != dbGame.Name.String {
			t.Fatalf("The initialGame.Name (tallylogic) '%s' did not match the expected Name '%s'",
				ts.initialGame.Name, dbGame.Name.String,
			)
		}
		if ts.initialGame.Description != dbGame.Description.String {
			t.Fatalf("The initialGame.Description (tallylogic) '%s' did not match expected Description '%s'",
				ts.initialGame.Description, dbGame.Description.String,
			)
		}
		t.Logf("%sThe interal tallylogic looks correct", logSuccess)
	})
}
func TestApi_Restart_After_Some_Moves(t *testing.T) {

	// resulted in the error-message
	// runtime error: invalid memory address or nil pointer dereference
	t.Run("Resetting game should reset all", func(t *testing.T) {

		ts := newTestApi(t)
		ctx := context.TODO()
		dump := ts.GetDBDump()
		if len(dump.GameHistories) == 0 {
			t.Fatalf("although an internal technical implementation, a new game should have GameHistories applied, but there were none")
		}
		{
			res := ts.SwipeUp()
			if res.Msg.Moves != 1 {
				t.Fatalf("Expected Game.Moves to be exactly 1, but was %d", res.Msg.Moves)
			}
		}
		{
			res, err := ts.client.RestartGame(ctx, connect.NewRequest(&model.RestartGameRequest{}))
			if err != nil {
				t.Fatalf("Restart Game failed %s", strErr(err))
			}
			if res.Msg.Board.Id == "" {
				t.Fatalf("expected board.id to not be empty: %#v", res)
			}
			testBoardEqualityIgnoreIds(t, res.Msg.Board, ts.initialSession.Msg.Session.Game.Board, func(diff []byte) bool {
				// dump := ts.GetDBDump()
				// t.Log(prettyJson(dump))
				t.Log(prettyJson(ts.initialSession.Msg))
				t.Logf("Initial Game (%s)\n%s", ts.initialGame.Name, ts.initialGame.Print())
				g := ts.Game()
				t.Logf("Game: (%s)\n%s", g.Name, g.Print())
				return false
			})
			if res.Msg.Moves != 0 {
				t.Fatalf("Expected Game.Moves to be exactly 0 after reset, but was %d", res.Msg.Moves)
			}
			if res.Msg.Score != 0 {
				t.Fatalf("Expected Game.Moves to be exactly 0 after reset, but was %d", res.Msg.Moves)
			}
		}
		// Retry one more time, to check if RestartGame does not break the game further
		{
			res := ts.SwipeUp()

			if res.Msg.Moves != 1 {
				t.Fatalf("Expected Game.Moves to be exactly 1, but was %d", res.Msg.Moves)
			}
		}
		{
			res, err := ts.client.RestartGame(ctx, connect.NewRequest(&model.RestartGameRequest{}))
			if err != nil {
				t.Fatalf("Restart Game failed %s", strErr(err))
			}
			if res.Msg.Board.Id == "" {
				t.Fatalf("expected board.id to not be empty: %#v", res)
			}
			testBoardEqualityIgnoreIds(t, res.Msg.Board, ts.initialSession.Msg.Session.Game.Board)
			if res.Msg.Moves != 0 {
				t.Fatalf("Expected Game.Moves to be exactly 0 after reset, but was %d", res.Msg.Moves)
			}
			if res.Msg.Score != 0 {
				t.Fatalf("Expected Game.Moves to be exactly 0 after reset, but was %d", res.Msg.Moves)
			}
		}
		// Retry one more time, to ensure NewGame does not break the game further
		newGameResponse, err := ts.client.NewGame(ctx, connect.NewRequest(&model.NewGameRequest{
			Mode:    *model.GameMode_GAME_MODE_RANDOM.Enum(),
			Variant: nil,
		}))
		if err != nil {
			t.Fatalf("New game failed failed %v", strErr(err))
		}
		if newGameResponse.Msg.Moves != 0 {
			t.Fatalf("Expected Game.Moves to be exactly 0, but was %d", newGameResponse.Msg.Moves)
		}
		{
			res := ts.SwipeDown()
			if res.Msg.Moves != 1 {
				t.Fatalf("Expected Game.Moves to be exactly 1, but was %d", res.Msg.Moves)
			}
		}
		{
			res, err := ts.client.RestartGame(ctx, connect.NewRequest(&model.RestartGameRequest{}))
			if err != nil {
				t.Fatalf("Restart Game failed %s", strErr(err))
			}
			if res.Msg.Board.Id == "" {
				t.Fatalf("expected board.id to not be empty: %#v", res)
			}
			testBoardEqualityIgnoreIds(t, res.Msg.Board, newGameResponse.Msg.Board)
			if res.Msg.Moves != 0 {
				t.Fatalf("Expected Game.Moves to be exactly 0 after reset, but was %d", res.Msg.Moves)
			}
			if res.Msg.Score != 0 {
				t.Fatalf("Expected Game.Moves to be exactly 0 after reset, but was %d", res.Msg.Moves)
			}
		}

	})
}
func testBoardEqualityIgnoreIds(t *testing.T, got *model.Board, want *model.Board, fs ...func(diff []byte) bool) {
	t.Helper()
	boardCopy := jsonCopy(got)
	initalBoardCopy := jsonCopy(want)
	boardCopy.Id = "test_id_overriden_for_comparison_purposes"
	initalBoardCopy.Id = "test_id_overriden_for_comparison_purposes"
	if diff := deep.Equal(boardCopy, initalBoardCopy); diff != nil {
		skipErr := false
		yDiff, _ := yaml.Marshal(diff)
		for _, f := range fs {

			skip := f(yDiff)
			if skip {
				skipErr = true
			}
		}
		if !skipErr {
			t.Errorf("Resetting should return the board-state to the initial state diff: \n%v\ngot = %v\nwant %v", string(yDiff), boardCopy, initalBoardCopy)
		}
	}
}
func TestApi_NewGame(t *testing.T) {
	// This is a variant of the above bug, that manifested in a different error-message
	// the rule '' from the user-session was not found
	t.Run("Should not crash on new game (https://github.com/runar-rkmedia/gotally/issues/11)", func(t *testing.T) {
		ts := newTestApi(t)
		ctx := context.TODO()
		{
			_, err := ts.client.NewGame(ctx, connect.NewRequest(&model.NewGameRequest{
				Mode: model.GameMode_GAME_MODE_TUTORIAL,
			}))
			if err != nil {
				t.Fatalf("New Game failed %s", strErr(err))
			}
		}
		{
			// This should fail, as we cannot restart a game that has no moves
			res, err := ts.client.RestartGame(ctx, connect.NewRequest(&model.RestartGameRequest{}))
			if err == nil {
				t.Fatalf("Expected err for RestartGame, when the game has no moves, but there was no error.\nResponse:\n%#v", res)
			}
			expectedErrMsg := "invalid_argument: The game is already at the start, and cannot be restarted"
			if !strings.Contains(err.Error(), expectedErrMsg) {
				t.Fatalf("Non-expected error: '%s', expected: '%s'", err, expectedErrMsg)
			}
			if res != nil {
				t.Fatalf("Expected response to be nil, but it was: %#v", res)
			}
		}

	})
}
