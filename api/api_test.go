package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/MarvinJWendt/testza"
	"github.com/bufbuild/connect-go"
	"github.com/go-test/deep"
	model "github.com/runar-rkmedia/gotally/gen/proto/tally/v1"
	"github.com/runar-rkmedia/gotally/generated"
	"github.com/runar-rkmedia/gotally/sqlite"
	"github.com/runar-rkmedia/gotally/types"
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
		dbGame := ts.DbGameById(res.Msg.Board.Id)
		testza.AssertNotEqual(t, "", dbGame.TemplateID.String, "TemplateID should have been set after restart")
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
		challenge := ts.CreateDefaultChallenge()
		newGameResponse, err := ts.client.NewGame(ctx, connect.NewRequest(&model.NewGameRequest{
			Mode: want,
			Variant: &model.NewGameRequest_Id{
				Id: challenge.Msg.Id,
			},
		}))
		if err != nil {
			t.Fatalf("New game failed failed %v", strErr(err))
		}
		got := newGameResponse.Msg.Mode
		game := ts.DbGameById(newGameResponse.Msg.Board.Id)
		testza.AssertNotEqual(t, "", game.TemplateID.String, "TemplateID should be set for challenge")
		if got != want {
			dump := ts.GetDBDump()
			var rule *sqlite.Rule
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
		dbGame := ts.DbGame()
		testza.AssertNotEqual(t, "", dbGame.TemplateID.String, "TemplateID should be set for a new session")
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
func TestApi_Undo(t *testing.T) {
	t.Run("Undo should work (I have not yet decided if Moves should be increased/decreased on Undo)", func(t *testing.T) {
		ts := newTestApi(t)
		m1 := ts.SwipeUp()
		testza.AssertEqual(t, m1.Msg.Moves, int64(1))
		m2 := ts.SwipeLeft()
		testza.AssertEqual(t, m2.Msg.Moves, int64(2))
		m3 := ts.SwipeRight()
		testza.AssertEqual(t, m3.Msg.Moves, int64(3))
		m4 := ts.SwipeDown()
		testza.AssertEqual(t, m4.Msg.Moves, int64(4))
		m5 := ts.Undo()
		testza.AssertEqual(t, m5.Msg.Board.Cells, m3.Msg.Board.Cells)
		m6 := ts.Undo()
		testza.AssertEqual(t, m6.Msg.Board.Cells, m2.Msg.Board.Cells)
		m7 := ts.SwipeRight()
		t.Log(ts.Game().Print())
		ts.CombineCellsByIndexPath(2, 5, 8)
		m9 := ts.Undo()

		testza.AssertEqual(t, m9.Msg.Board.Cells, m7.Msg.Board.Cells)
		m10 := ts.Undo()

		testza.AssertEqual(t, m10.Msg.Board.Cells, m2.Msg.Board.Cells)
		m11 := ts.Undo()

		testza.AssertEqual(t, m11.Msg.Board.Cells, m1.Msg.Board.Cells)
		m12 := ts.SwipeDown()
		fmt.Print(m12)
		// testza.AssertEqual(t, m12.Msg.Moves, int64(2))
		ts.CombineCellsByIndexPath(6, 7, 8)
		ts.Undo()
		ts.CombineCellsByIndexPath(2, 5, 8)
		m := ts.CombineCellsByIndexPath(6, 7, 8)
		testza.AssertTrue(t, m.Msg.DidWin)
		// Check that we are able to get a hint.
		// See https://github.com/runar-rkmedia/gotally/issues/27
		// There was a crash here when getting the hint
		// ts.Undo()
		// ts.GetHint(1)
	})
}
func TestApi_Restart_After_Some_Moves(t *testing.T) {

	// resulted in the error-message
	// runtime error: invalid memory address or nil pointer dereference
	t.Run("Resetting game should reset all", func(t *testing.T) {

		ts := newTestApi(t)
		ctx := context.TODO()
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
func TestApi_ShouldWin(t *testing.T) {
	t.Run("Should win game after solving it", func(t *testing.T) {
		ts := newTestApi(t)
		ts.NewGame(model.GameMode_GAME_MODE_TUTORIAL)
		t.Logf("New game: %s", ts.Game().Print())
		{
			swipeResponse := ts.CombineCellsByIndexPath(2, 5, 8)
			testza.AssertFalse(t, swipeResponse.Msg.DidWin, "Game should not have been won yet")
			testza.AssertFalse(t, swipeResponse.Msg.DidLose, "Game should not have been game over yet")
			t.Logf("After combine cells 1: %s", ts.Game().Print())
		}
		{
			swipeResponse := ts.CombineCellsByIndexPath(6, 7, 8)
			testza.AssertTrue(t, swipeResponse.Msg.DidWin, "Game should be won")
			testza.AssertFalse(t, swipeResponse.Msg.DidLose, "Game should not have been lost")
			t.Logf("After combine cells 2: %s", ts.Game().Print())
		}

		// Check that the game-property her also has been won
		game := ts.Game()
		testza.AssertTrue(t, game.IsGameWon(), "Game should be won")
		testza.AssertFalse(t, game.IsGameOver(), "Game should not have been lost")

		// Check that the database agrees
		dump, err := ts.tally.storage.GetUserBySessionID(ts.context, types.GetUserPayload{
			ID: ts.initialSession.Msg.Session.SessionId,
		})
		testza.AssertNoError(t, err)
		testza.AssertEqual(t, dump.ActiveGame.PlayState, types.PlayStateWon)

	})
}
func TestApi_Challenges(t *testing.T) {
	t.Run("Should return the challenges after creation", func(t *testing.T) {
		ts := newTestApi(t)
		ctx := context.TODO()
		// ------------------------------------------------------------
		ts.LogMark("Check that there are no challenges")
		// ------------------------------------------------------------
		{
			r, err := ts.client.GetGameChallenges(ctx, connect.NewRequest(&model.GetGameChallengesRequest{}))
			if err != nil {
				t.Fatalf("Get Game challenges failed %s", strErr(err))
			}
			testza.AssertEqual(t, 0, len(r.Msg.Challenges), "Expecting there to be zero challenges at start")
		}

		// ------------------------------------------------------------
		ts.LogMark("Insert some challenges")
		// ------------------------------------------------------------

		payloads := []model.CreateGameChallengeRequest{
			{
				ChallengeNumber: 100,
				IdealMoves:      5,
				TargetCellValue: 8,
				Columns:         3,
				Rows:            3,
				Name:            "Simply a test",
				Description:     "Please dont fail me",
				Cells: toModalCells(cellCreator(
					3, 1, 3,
					6, 6, 6,
					9, 6, 3,
				)),
			},
			{
				ChallengeNumber: 200,
				IdealMoves:      5,
				TargetCellValue: 8,
				Columns:         3,
				Rows:            3,
				Name:            "Simply another test",
				Description:     "Please dont fail me now",
				Cells: toModalCells(cellCreator(
					2, 5, 5,
					8, 2, 4,
					2, 4, 4,
				)),
			},
		}
		for i := 0; i < len(payloads); i++ {

			{
				// This should fail, as we cannot restart a game that has no moves
				res, err := ts.client.CreateGameChallenge(ctx, connect.NewRequest(&payloads[i]))
				if err != nil {
					t.Errorf("CreateGameChallengeRequest %d failed: %v", i, err)
				}
				if res == nil {
					t.Errorf("CreateGameChallengeRequest %d result was nil for paylaod %v", i, &payloads[i])
				}
				t.Log(res.Msg)
				testza.AssertEqual(t, payloads[i].ChallengeNumber, res.Msg.ChallengeNumber, fmt.Sprintf("ChallengeNumber for CreateGameChallenge-Response %d should match payload", i))
				testza.AssertNotZero(t, res.Msg.Id, fmt.Sprintf("ID for CreateGameChallenge-Response %d should have an ID", i))
			}
		}
		// ------------------------------------------------------------
		ts.LogMark("Check that we can query for the same set of challenges")
		// ------------------------------------------------------------
		{
			r, err := ts.client.GetGameChallenges(ctx, connect.NewRequest(&model.GetGameChallengesRequest{}))
			if err != nil {
				t.Fatalf("Get Game challenges failed %s", strErr(err))
			}
			testza.AssertEqual(t, 2, len(r.Msg.Challenges), "Expecting api to return the newly created challenges")
			for i := 0; i < len(payloads); i++ {
				// t.Logf("Payload %d:\n%s", i, pretty(&payloads[i]))
				testza.AssertEqual(t, payloads[i].ChallengeNumber, r.Msg.Challenges[i].ChallengeNumber, fmt.Sprintf("Expected the %d challenge to match on ChallengeNumber", i))
				testza.AssertEqual(t, payloads[i].Name, r.Msg.Challenges[i].Name, fmt.Sprintf("Expected the %d challenge to match on Name", i))
				testza.AssertEqual(t, payloads[i].Description, r.Msg.Challenges[i].Description, fmt.Sprintf("Expected the %d challenge to match on Description", i))
				testza.AssertEqual(t, payloads[i].Cells, r.Msg.Challenges[i].Cells, fmt.Sprintf("Expected the %d challenge to match on Cells", i))
				testza.AssertEqual(t, payloads[i].Rows, r.Msg.Challenges[i].Rows, fmt.Sprintf("Expected the %d challenge to match on Rows", i))
				testza.AssertEqual(t, payloads[i].Columns, r.Msg.Challenges[i].Columns, fmt.Sprintf("Expected the %d challenge to match on Columns", i))
				testza.AssertEqual(t, payloads[i].TargetCellValue, r.Msg.Challenges[i].TargetCellValue, fmt.Sprintf("Expected the %d challenge to match on TargetCellValue", i))
				testza.AssertEqual(t, payloads[i].IdealMoves, r.Msg.Challenges[i].IdealMoves, fmt.Sprintf("Expected the %d challenge to match on IdealMoves", i))
			}
		}
		ts.LogMark("Check for invalid payload")
		invalidPayloads := []model.CreateGameChallengeRequest{
			{
				ChallengeNumber: 100,
				IdealMoves:      5,
				TargetCellValue: 8,
				Columns:         3,
				Rows:            3,
				Name:            "Should not be able to insert at the same challengeNumber",
				Description:     "Faile me",
				Cells: toModalCells(cellCreator(
					3, 1, 3,
					6, 6, 6,
					9, 6, 3,
				)),
			},
			{
				IdealMoves:      5,
				TargetCellValue: 8,
				Columns:         3,
				Rows:            3,
				Name:            "Should not be able to submit cells with an invalid number",
				Cells: toModalCells(cellCreator(
					2, 5, 5,
				)),
			},
			{
				IdealMoves:      5,
				TargetCellValue: 8,
				Columns:         3,
				Rows:            0,
				Name:            "Rows must be positive",
				Cells: toModalCells(cellCreator(
					2, 5, 5,
				)),
			},
			{
				IdealMoves:      5,
				TargetCellValue: 8,
				Columns:         0,
				Rows:            3,
				Name:            "Columns must be positive",
				Cells: toModalCells(cellCreator(
					2, 5, 5,
				)),
			},
			{
				IdealMoves: 8,
				Columns:    3,
				Rows:       3,
				Name:       "Target Cell Value must be set",
				Cells: toModalCells(cellCreator(
					3, 1, 3,
					6, 6, 6,
					9, 6, 3,
				)),
			},
			{
				TargetCellValue: 8,
				Columns:         3,
				Rows:            3,
				Name:            "Ideal moves must be set",
				Cells: toModalCells(cellCreator(
					3, 1, 3,
					6, 6, 6,
					9, 6, 3,
				)),
			},
		}
		for i := 0; i < len(invalidPayloads); i++ {

			{
				// This should fail, as we cannot restart a game that has no moves
				_, err := ts.client.CreateGameChallenge(ctx, connect.NewRequest(&invalidPayloads[i]))
				testza.AssertNotNil(t, err, "Request %d should have failed but did not for payload with name '%s'", i, invalidPayloads[i].Name)
			}
		}
		// ------------------------------------------------------------
		ts.LogMark("Check that the database was not modified under the bad requests")
		// ------------------------------------------------------------
		{
			dump := ts.GetDBDump()
			if len(dump.Templates) != 2 {
			outer:
				for i, template := range dump.Templates {
					for j := 0; j < len(payloads); j++ {
						if payloads[j].Name == template.Name {
							continue outer
						}

					}
					t.Logf("Extra template %d:  Name='%s'", i, template.Name)
				}
				testza.AssertEqual(t, len(dump.Templates), 2, "Expected the database to not have any of the invalid templates")

			}
		}

	})
}
