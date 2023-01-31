package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/bufbuild/connect-go"
	"github.com/go-test/deep"
	"github.com/runar-rkmedia/go-common/logger"
	model "github.com/runar-rkmedia/gotally/gen/proto/tally/v1"
	"github.com/runar-rkmedia/gotally/gen/proto/tally/v1/tallyv1connect"
	"gopkg.in/yaml.v3"
)

func TestApi_Restart(t *testing.T) {

	// resulted in the error-message
	// runtime error: invalid memory address or nil pointer dereference
	t.Run("Should not crash on restart (https://github.com/runar-rkmedia/gotally/issues/11)", func(t *testing.T) {

		ts := newTestApi(t)
		ctx := context.TODO()
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

func jsonCopy(in any, out any) {
	b, err := json.Marshal(in)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(b, out)
	if err != nil {
		panic(err)
	}
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
		res, err := ts.client.RestartGame(ctx, connect.NewRequest(&model.RestartGameRequest{}))
		t.Logf("Response \n%v Err %v\n\n", res, err)
		if err != nil {
			t.Fatalf("Restart Game failed %s", err)
		}
		if res.Msg.Board.Id == "" {
			t.Fatalf("expected board.id to not be empty: %#v", res)
		}
		if diff := deep.Equal(res.Msg.Board, ts.initialSession.Msg.Session.Game.Board); diff != nil {
			yDiff, _ := yaml.Marshal(diff)
			t.Errorf("Resetting should return the board-state to the initial state diff: \n%v\ngot = %v\nwant %v", string(yDiff), res.Msg.Board, ts.initialSession.Msg.Session.Game.Board)
		}
		if res.Msg.Moves != 0 {
			t.Fatalf("Expected Game.Moves to be exactly 0 after reset, but was %d", res.Msg.Moves)
		}
		if res.Msg.Score != 0 {
			t.Fatalf("Expected Game.Moves to be exactly 0 after reset, but was %d", res.Msg.Moves)
		}
	})
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
			res, err := ts.client.RestartGame(ctx, connect.NewRequest(&model.RestartGameRequest{}))
			if err != nil {
				t.Fatalf("Restart Game failed %s", err)
			}
			if res.Msg.Board.Id == "" {
				t.Fatalf("expected board.id to not be empty: %#v", res)
			}
		}

	})
}

func pretty(j any) string {
	b, _ := yaml.Marshal(j)
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
		Level:      "info",
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
	t.Cleanup(func() {
		dumpPath, _ := filepath.Abs("dump.yaml")
		t.Logf("dumping sql-dump to %s", dumpPath)
		dump, err := tally.storage.Dump(context.TODO())
		if err != nil {
			t.Errorf("Failed to dump db: %v", err)
		}
		b, err := yaml.Marshal(dump)
		if err != nil {
			t.Errorf("Failed to marshal dump of db: %v", err)
		}

		os.WriteFile(dumpPath, b, 0755)

	})
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
