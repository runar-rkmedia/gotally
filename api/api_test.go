package api

import (
	"context"
	"encoding/json"
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
	"github.com/runar-rkmedia/gotally/gen/proto/tally/v1/tallyv1connect"
	"github.com/runar-rkmedia/gotally/sqlite"
	"gopkg.in/yaml.v3"
)

func TestApi_Restart(t *testing.T) {

	// resulted in the error-message
	// runtime error: invalid memory address or nil pointer dereference
	t.Run("Should not crash on restart (https://github.com/runar-rkmedia/gotally/issues/11)", func(t *testing.T) {

		ts := newTestApi(t)
		ctx := context.TODO()
		_, err := ts.client.SwipeBoard(ctx, connect.NewRequest(&model.SwipeBoardRequest{Direction: model.SwipeDirection_SWIPE_DIRECTION_DOWN}))
		if err != nil {
			t.Fatalf("SwipBoard failed %s", err)
		}
		res, err := ts.client.RestartGame(ctx, connect.NewRequest(&model.RestartGameRequest{}))
		if err != nil {
			t.Fatalf("Restart Game failed %s", err)
		}
		if err != nil {
			t.Fatalf("Restart Game failed %s", err)
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

func TestApi_Restart_After_Some_Moves(t *testing.T) {

	// resulted in the error-message
	// runtime error: invalid memory address or nil pointer dereference
	t.Run("Resetting game should reset all", func(t *testing.T) {

		ts := newTestApi(t)
		if ts.initialSession.Msg.Session.Game.Moves != 0 {
			t.Fatalf("Expected Game.Moves to be exactly 0, but was %d", ts.initialSession.Msg.Session.Game.Moves)
		}
		ctx := context.TODO()
		dump, err := ts.GetDBDump()
		if err != nil {
			t.Fatalf("failed to get dump of database")
		}
		if len(dump.GameHistories) == 0 {
			t.Fatalf("although an internal technical implementation, a new game should have GameHistories applied, but there were none")
		}
		{
			res, err := ts.client.SwipeBoard(ctx, connect.NewRequest(&model.SwipeBoardRequest{
				Direction: model.SwipeDirection_SWIPE_DIRECTION_DOWN,
			}))
			if err != nil {
				t.Fatalf("Swip failed %v", err)
			}
			if res.Msg.Moves != 1 {
				t.Fatalf("Expected Game.Moves to be exactly 1, but was %d", res.Msg.Moves)
			}
		}
		{
			res, err := ts.client.RestartGame(ctx, connect.NewRequest(&model.RestartGameRequest{}))
			if err != nil {
				t.Fatalf("Restart Game failed %s", err)
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
		// Retry one more time, to check if RestartGame does not break the game further
		{
			res, err := ts.client.SwipeBoard(ctx, connect.NewRequest(&model.SwipeBoardRequest{
				Direction: model.SwipeDirection_SWIPE_DIRECTION_DOWN,
			}))
			if err != nil {
				t.Fatalf("Swip failed %v", err)
			}
			if res.Msg.Moves != 1 {
				t.Fatalf("Expected Game.Moves to be exactly 1, but was %d", res.Msg.Moves)
			}
		}
		{
			res, err := ts.client.RestartGame(ctx, connect.NewRequest(&model.RestartGameRequest{}))
			if err != nil {
				t.Fatalf("Restart Game failed %s", err)
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
			t.Fatalf("Swip failed %v", err)
		}
		if newGameResponse.Msg.Moves != 0 {
			t.Fatalf("Expected Game.Moves to be exactly 0, but was %d", newGameResponse.Msg.Moves)
		}
		{
			res, err := ts.client.SwipeBoard(ctx, connect.NewRequest(&model.SwipeBoardRequest{
				Direction: model.SwipeDirection_SWIPE_DIRECTION_DOWN,
			}))
			if err != nil {
				t.Fatalf("Swip failed %v", err)
			}
			if res.Msg.Moves != 1 {
				t.Fatalf("Expected Game.Moves to be exactly 1, but was %d", res.Msg.Moves)
			}
		}
		{
			res, err := ts.client.RestartGame(ctx, connect.NewRequest(&model.RestartGameRequest{}))
			if err != nil {
				t.Fatalf("Restart Game failed %s", err)
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
func testBoardEqualityIgnoreIds(t *testing.T, got *model.Board, want *model.Board) {
	boardCopy := jsonCopy(got)
	initalBoardCopy := jsonCopy(want)
	boardCopy.Id = "foo"
	initalBoardCopy.Id = "foo"
	if diff := deep.Equal(boardCopy, initalBoardCopy); diff != nil {
		yDiff, _ := yaml.Marshal(diff)
		t.Errorf("Resetting should return the board-state to the initial state diff: \n%v\ngot = %v\nwant %v", string(yDiff), boardCopy, initalBoardCopy)
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
				t.Fatalf("New Game failed %s", err)
			}
		}
		{
			// This should fail, as we cannot restart a game that has no moves
			res, err := ts.client.RestartGame(ctx, connect.NewRequest(&model.RestartGameRequest{}))
			if err == nil {
				t.Fatalf("Expected err for RestartGame, when the game has no moves, but there was no error.\nResponse:\n%#v", res)
			}
			expectedErrMsg := "Cannot restart a game already at the start"
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
	initialSession connect.Response[model.GetSessionResponse]
}

func newTestApi(t *testing.T) testApi {

	logger.InitLogger(logger.LogConfig{
		Level:      "fatal",
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
		t.Fatalf("Getsession failed %s", err)
	}
	if res.Msg.Session.Username != "GO_TESTER" {
		t.Fatalf("Expected username to have been set (dev-header) to '%s' but was '%s'", "GO_TESTER", res.Msg.Session.Username)
	}
	a.initialSession = *res
	return a
}
func (ta *testApi) DumpDB() {
	ta.DumpDBWithPrefix("")
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

func (ta *testApi) GetDBDump() (sqliteDump, error) {
	d, err := ta.tally.storage.Dump(context.TODO())
	if err != nil {
		return sqliteDump{}, err
	}

	return sqliteDump{
		Date:          time.Now(),
		Games:         d.Games.([]sqlite.Game),
		GameHistories: d.GameHistories.([]sqlite.GameHistory),
		Rules:         d.Rules.([]sqlite.Rule),
		Sessions:      d.Sessions.([]sqlite.Session),
		Users:         d.Users.([]sqlite.User),
	}, nil
}
