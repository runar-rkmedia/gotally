package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/flytam/filenamify"
	"github.com/go-test/deep"
	"github.com/runar-rkmedia/go-common/logger"
	model "github.com/runar-rkmedia/gotally/gen/proto/tally/v1"
	tallyv1 "github.com/runar-rkmedia/gotally/gen/proto/tally/v1"
	"github.com/runar-rkmedia/gotally/gen/proto/tally/v1/tallyv1connect"
	"github.com/runar-rkmedia/gotally/generated"
	"github.com/runar-rkmedia/gotally/sqlite"
	"github.com/runar-rkmedia/gotally/tallylogic"
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

func pretty(j any) string {
	b, _ := yaml.Marshal(j)
	return string(b)
}
func prettyJson(j any) string {
	b, _ := json.MarshalIndent(j, "", "  ")
	return string(b)
}

type testApi struct {
	handler        http.Handler
	path           string
	tally          TallyServer
	t              *testing.T
	server         *httptest.Server
	client         tallyv1connect.BoardServiceClient
	defaultHeaders map[string]string
	initialGame    tallylogic.Game
	initialSession connect.Response[model.GetSessionResponse]
}

const (
	logSuccess = "✔️"
	logError   = "️⚠️"
	logInfo    = "️ℹ️"
)

func newTestApi(t *testing.T) testApi {
	t.Helper()

	logger.InitLogger(logger.LogConfig{
		Level:      "debug",
		Format:     "human",
		WithCaller: true,
	})
	_true := true
	tally, path, handler := createApiHandler(true, TallyOptions{
		DatabaseDSN:         fmt.Sprintf("sqlite:file::%s:?mode=memory&cache=shared", mustCreateUUidgenerator()()),
		SkipStatsCollection: &_true,
	})
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	a := testApi{
		handler: handler,
		tally:   tally,
		path:    path,
		t:       t,
		server:  ts,
		defaultHeaders: map[string]string{
			tokenHeader:    mustCreateUUidgenerator()(),
			"DEV_USERNAME": "GO_TESTER",
		},
	}
	t.Cleanup(a.DumpDB)
	// client := connect.NewClient[tallyv1.BoardServiceClient](http.DefaultClient, path)
	a.client = tallyv1connect.NewBoardServiceClient(http.DefaultClient, ts.URL,
		connect.WithProtoJSON(),
		connect.WithInterceptors(connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
			return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
				if a.defaultHeaders != nil {
					for k, v := range a.defaultHeaders {
						if req.Header().Get(k) != "" {
							continue
						}
						req.Header().Set(k, v)
					}
				}

				return next(ctx, req)
			})

		})))

	res, err := a.client.GetSession(context.TODO(), connect.NewRequest(&model.GetSessionRequest{}))
	if err != nil {
		t.Fatalf("Getsession failed %s", strErr(err))
	}
	if res.Msg.Session.Username != "GO_TESTER" {
		t.Fatalf("Expected username to have been set (dev-header) to '%s' but was '%s'", "GO_TESTER", res.Msg.Session.Username)
	}
	a.initialSession = *res
	a.initialGame = a.Game()
	t.Logf("%s A nes session, user and game was created, with game-mode '%s', Name: '%s', Description:  %s",
		logSuccess,
		res.Msg.Session.Game.Mode,
		res.Msg.Session.Game.Board.Name,
		res.Msg.Session.Game.Description,
	)
	t.Logf("%s tallylogic.Game-record: game-mode '%s', Name: '%s', Description:  %s",
		logSuccess,
		a.initialGame.Rules.GameMode,
		a.initialGame.Name,
		a.initialGame.Description,
	)
	return a
}
func (ta *testApi) DumpDB() {
	ta.DumpDBWithPrefix("")
}

func (ts *testApi) SwipeUp() *connect.Response[tallyv1.SwipeBoardResponse] {
	ts.t.Helper()
	return ts.Swipe(model.SwipeDirection_SWIPE_DIRECTION_UP)
}
func (ts *testApi) SwipeRight() *connect.Response[tallyv1.SwipeBoardResponse] {
	ts.t.Helper()
	return ts.Swipe(model.SwipeDirection_SWIPE_DIRECTION_RIGHT)
}
func (ts *testApi) SwipeDown() *connect.Response[tallyv1.SwipeBoardResponse] {
	ts.t.Helper()
	return ts.Swipe(model.SwipeDirection_SWIPE_DIRECTION_DOWN)
}
func (ts *testApi) SwipeLeft() *connect.Response[tallyv1.SwipeBoardResponse] {
	ts.t.Helper()
	return ts.Swipe(model.SwipeDirection_SWIPE_DIRECTION_LEFT)
}
func (ts *testApi) Swipe(direction model.SwipeDirection) *connect.Response[tallyv1.SwipeBoardResponse] {
	ts.t.Helper()
	ctx := context.TODO()
	res, err := ts.client.SwipeBoard(ctx, connect.NewRequest(&model.SwipeBoardRequest{
		Direction: direction,
	}))
	if err != nil {
		ts.t.Fatalf("%s Failed during SwipeBoard: %#v", logError, err)
	}
	ts.t.Logf("response %#v", res.Msg)
	if !res.Msg.DidChange {
		game := ts.Game()
		ts.t.Fatalf("%s board should have changed during swipe '%s', but did not. Perhaps you meant a differen swipe-direction? %v", logError, direction, game.Print())
	}
	ts.t.Logf("%s Board Swiped %s", logSuccess, direction)
	return res
}

func (ta *testApi) Game() tallylogic.Game {
	s, err := ta.tally.storage.GetUserBySessionID(context.TODO(), types.GetUserPayload{
		ID: ta.initialSession.Msg.Session.SessionId,
	})
	if err != nil {
		ta.t.Error("failed to get game for debugging-purposes: %w", err)
	}
	game, err := tallylogic.RestoreGame(s.ActiveGame)
	if err != nil {
		ta.t.Error("failed to restore game for debugging-purposes: %w", err)
	}
	return game
}
func (ta *testApi) DumpDBWithPrefix(prefix string) {

	fname, err := filenamify.FilenamifyV2("dump_" + prefix + ta.t.Name())
	if err != nil {
		panic(err)
	}
	dumpPath, err := filepath.Abs(filepath.Join("..", fname+".json"))
	if err != nil {
		panic(err)
	}
	ta.t.Logf("dumping sql-dump to %s", dumpPath)
	dump, err := ta.tally.storage.Dump(context.TODO())
	if err != nil {
		ta.t.Errorf("Failed to dump db: %v", err)
	}
	b, err := json.Marshal(dump)
	if err != nil {
		ta.t.Errorf("Failed to marshal dump of db: %v", err)
	}

	os.WriteFile(dumpPath, b, 0755)
}

// Temp hack since types.Dump has not yet received any good typing
// does not really matter, though
type sqliteDump struct {
	Date          time.Time
	Games         []sqlite.Game        //[]Game
	GameHistories []sqlite.GameHistory //[]any
	Rules         []sqlite.Rule        //[]Rules
	Sessions      []sqlite.Session     //[]Session
	Users         []sqlite.User        //[]User
}

func (ta *testApi) GetDBDump() sqliteDump {
	d, err := ta.tally.storage.Dump(context.TODO())
	if err != nil {
		ta.t.Fatalf("Failed to dump the database: %v", err)
	}

	return sqliteDump{
		Date:          time.Now(),
		Games:         d.Games.([]sqlite.Game),
		GameHistories: d.GameHistories.([]sqlite.GameHistory),
		Rules:         d.Rules.([]sqlite.Rule),
		Sessions:      d.Sessions.([]sqlite.Session),
		Users:         d.Users.([]sqlite.User),
	}
}
