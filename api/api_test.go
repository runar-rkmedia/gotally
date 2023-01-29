package api

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bufbuild/connect-go"
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
		// t.Log(res)
		if err != nil {
			t.Fatalf("Restart Game failed %s", err)
		}
		t.Error(res)

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
			t.Error(res)
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
}

func newTestApi(t *testing.T) testApi {

	logger.InitLogger(logger.LogConfig{
		Level:      "error",
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
	return a
}
