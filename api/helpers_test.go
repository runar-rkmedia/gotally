package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/MarvinJWendt/testza"
	"github.com/bufbuild/connect-go"
	"github.com/flytam/filenamify"
	"github.com/go-test/deep"
	"github.com/runar-rkmedia/go-common/logger"
	model "github.com/runar-rkmedia/gotally/gen/proto/tally/v1"
	tallyv1 "github.com/runar-rkmedia/gotally/gen/proto/tally/v1"
	"github.com/runar-rkmedia/gotally/gen/proto/tally/v1/tallyv1connect"
	"github.com/runar-rkmedia/gotally/sqlite"
	"github.com/runar-rkmedia/gotally/tallylogic"
	"github.com/runar-rkmedia/gotally/tallylogic/cell"
	"github.com/runar-rkmedia/gotally/types"
	"gopkg.in/yaml.v3"
)

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
	context        context.Context
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
		Level:      "error",
		Format:     "human",
		WithCaller: true,
	})
	_true := true
	tally, path, handler := createApiHandler(true, TallyOptions{
		DatabaseDSN:           fmt.Sprintf("sqlite:file::%s:?mode=memory&cache=shared", mustCreateUUidgenerator()()),
		SkipStatsCollection:   &_true,
		FeatureGameGeneration: &_true,
		AllowDevelopmentFlags: &_true,
	})
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	a := testApi{
		context: context.TODO(),
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
		t.Fatalf("Expected username to have been set (dev-header) to '%s' but was '%s'. Did you enable development-mode for the api-server?", "GO_TESTER", res.Msg.Session.Username)
	}
	a.initialSession = *res
	a.initialGame = a.Game()
	t.Logf("%s A new session, user and game was created, with game-mode '%s', Name: '%s', Description:  %s",
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
func (ts *testApi) Undo() *connect.Response[tallyv1.UndoResponse] {
	ts.t.Helper()
	ctx := context.TODO()
	res, err := ts.client.Undo(ctx, connect.NewRequest(&model.UndoRequest{}))
	if err != nil {
		ts.t.Log(err.Error())
		ts.t.Fatalf("%s Failed during Undo: %#v", logError, err)
	}
	ts.t.Logf("response %#v", res.Msg)
	ts.t.Logf("%s Board Undo", logSuccess)
	return res
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
		ts.t.Log(err.Error())
		ts.t.Fatalf("%s Failed during SwipeBoard: %#v", logError, err)
	}
	ts.t.Logf("response %#v", res.Msg)
	if !res.Msg.DidChange {
		game := ts.Game()
		ts.t.Fatalf("%s board should have changed during swipe '%s', but did not. Perhaps you meant a different swipe-direction? %v", logError, direction, game.Print())
	}
	ts.t.Logf("%s Board Swiped %s", logSuccess, direction)
	return res
}

func (ta *testApi) Session() types.SessionUser {
	s, err := ta.tally.storage.GetUserBySessionID(context.TODO(), types.GetUserPayload{
		ID: ta.initialSession.Msg.Session.SessionId,
	})
	if err != nil {
		ta.t.Error("failed to get session for debugging-purposes: %w", err)
	}

	return *s
}

// DbGame returns the current game from the database
func (ta *testApi) DbGame() *sqlite.Game {
	return ta.DbGameById(ta.initialGame.ID)
}
func (ta *testApi) DbGameById(id string) *sqlite.Game {
	ddump := ta.GetDBDump()
	return find(ddump.Games, func(t sqlite.Game) bool { return t.ID == id })
}

// Game returns the game from the database, as a tallylogic-game
func (ta *testApi) Game() tallylogic.Game {
	s := ta.Session()
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
	Date      time.Time
	Games     []sqlite.Game         //[]Game
	Rules     []sqlite.Rule         //[]Rules
	Sessions  []sqlite.Session      //[]Session
	Users     []sqlite.User         //[]User
	Templates []sqlite.GameTemplate //[]GameTemplate
}

func (ta *testApi) GetDBDump() sqliteDump {
	d, err := ta.tally.storage.Dump(context.TODO())
	if err != nil {
		ta.t.Fatalf("Failed to dump the database: %v", err)
	}

	return sqliteDump{
		Date:      time.Now(),
		Games:     d.Games.([]sqlite.Game),
		Rules:     d.Rules.([]sqlite.Rule),
		Sessions:  d.Sessions.([]sqlite.Session),
		Users:     d.Users.([]sqlite.User),
		Templates: d.Template.([]sqlite.GameTemplate),
	}
}
func (ts *testApi) NewGame(mode tallyv1.GameMode) (response *connect.Response[model.NewGameResponse]) {
	ts.t.Helper()
	newGameResponse, err := ts.client.NewGame(ts.context, connect.NewRequest(&model.NewGameRequest{
		Mode: mode,
	}))
	if err != nil {
		ts.t.Fatalf("new game failed for mode %v: %v", mode, err)
	}
	testza.AssertEqual(ts.t, mode, newGameResponse.Msg.Mode, "Expected modes to be equal")
	return newGameResponse
}
func (ts *testApi) CombineCellsByIndexPath(indexes ...uint32) (response *connect.Response[model.CombineCellsResponse]) {
	ts.t.Helper()
	combineResponse, err := ts.client.CombineCells(ts.context, connect.NewRequest(&model.CombineCellsRequest{
		Selection: &model.CombineCellsRequest_Indexes{
			Indexes: &model.Indexes{
				Index: indexes,
			},
		},
	}))
	ts.FatatErr("CombineCells failed", err, map[string]any{"input": indexes})
	return combineResponse
}
func (ts *testApi) FatatErr(prefix string, err error, details ...any) {
	ts.t.Helper()
	if err != nil {
		if cerr, ok := err.(*connect.Error); ok {
			ts.t.Fatalf("%s: %s %v", prefix, cerr.Message(), details)
		}
		ts.t.Fatalf("%s: %s %v", prefix, err, details)
	}
}
func (ts *testApi) NewGameChallenge(id string) (response *connect.Response[model.NewGameResponse]) {
	ts.t.Helper()
	newGameResponse, err := ts.client.NewGame(ts.context, connect.NewRequest(&model.NewGameRequest{
		Mode: tallyv1.GameMode_GAME_MODE_RANDOM_CHALLENGE,
		Variant: &tallyv1.NewGameRequest_Id{
			Id: id,
		},
	}))
	if err != nil {
		ts.t.Fatalf("new game-challenge failed for id %s: %v", id, err)
	}
	return newGameResponse
}

func (ts *testApi) CreateDefaultChallenge() (response *connect.Response[model.CreateGameChallengeResponse]) {
	ts.t.Helper()

	payload := &tallyv1.CreateGameChallengeRequest{
		ChallengeNumber: 100,
		IdealMoves:      5,
		TargetCellValue: 79,
		Columns:         3,
		Rows:            3,
		Name:            "Simple challenge",
		Cells: toModalCells(cellCreator(
			0, 0, 4,
			0, 0, 10,
			4, 5, 20,
		)),
	}
	return ts.CreateGameChallenge(payload)
}
func (ts *testApi) CreateGameChallenge(payload *tallyv1.CreateGameChallengeRequest) (response *connect.Response[model.CreateGameChallengeResponse]) {
	ts.t.Helper()

	// ------------------------------------------------------------
	ts.LogMark("Creating an initial challenge for testing")
	// ------------------------------------------------------------
	res, err := ts.client.CreateGameChallenge(ts.context, connect.NewRequest(payload))
	testza.AssertNil(ts.t, err, "Expected no errors from CreateGameChallenge")
	return res
}

func (ts *testApi) RestartGame() (response *connect.Response[model.RestartGameResponse]) {
	ts.t.Helper()
	res, err := ts.client.RestartGame(ts.context, connect.NewRequest(&model.RestartGameRequest{}))
	if err != nil {
		ts.t.Fatalf("restart game failed:  %v", err)
	}
	testza.AssertEqual(ts.t, int64(0), res.Msg.Moves, "Expected moves to be reset")
	return res
}

func (ts *testApi) SolveGameWithHints(expectMaxHints int) (response *connect.Response[model.CombineCellsResponse]) {
	ts.t.Helper()
	for i := 1; i <= expectMaxHints; i++ {
		res := ts.GetHint(1)

		ts.t.Log(res.Msg.Instructions)
		instr := res.Msg.Instructions

		moves := int64(ts.Game().Moves())
		for _, v := range instr {
			switch x := v.InstructionOneof.(type) {
			case nil:
				ts.t.Fatal("Instruction is nil")
			case *tallyv1.Instruction_Swipe:
				ts.t.Log(ts.Game().Print())
				res := ts.Swipe(x.Swipe)
				moves++
				testza.AssertEqual(ts.t, res.Msg.Moves, moves, "Expected moves to have changed after swipe")
			case *tallyv1.Instruction_Combine:
				req := connect.Request[tallyv1.CombineCellsRequest]{
					Msg: &tallyv1.CombineCellsRequest{
						Selection: &tallyv1.CombineCellsRequest_Indexes{
							Indexes: x.Combine,
						},
					},
				}
				res, err := ts.client.CombineCells(ts.context, &req)
				if err != nil {
					ts.t.Fatalf("CombineCells failed for instruction %s: %v", v, err)
				}
				ts.t.Log(ts.Game().Print())
				response = res
				moves++
				testza.AssertEqual(ts.t, res.Msg.Moves, moves, "Expected moves to have changed after combining cells")
				if res.Msg.DidWin {
					ts.t.Log("Game is won!")
					return res
				}

			default:
				ts.t.Fatal("Unhandled instruction: ", reflect.TypeOf(x), x)
			}

		}
	}
	testza.AssertTrue(ts.t, response.Msg.DidWin, "expected game to be won (solved)")
	return
}
func (ts *testApi) GetHint(expectMinHints int) *connect.Response[model.GetHintResponse] {
	ts.t.Helper()

	res, err := ts.client.GetHint(ts.context, connect.NewRequest(&model.GetHintRequest{}))
	testza.AssertNoError(ts.t, err, "GetHint should not fail")
	if err != nil {
		ts.t.Fail()
	}
	testza.AssertGreaterOrEqual(ts.t, len(res.Msg.Instructions), expectMinHints, "Expected hints to be provided")
	ts.t.Logf("Instruction: %s", res.Msg.Instructions)
	if expectMinHints > 0 {
		for i, v := range res.Msg.Instructions {
			g := ts.Game()
			testza.AssertNotNil(ts.t, v.InstructionOneof, "res.Msg.Instructions[%d] was unexpectedly empty: %s for move %d in game '%s':\n%s", i, prettyJson(v), g.Moves(), g.Name, g.Print())
		}
	}
	return res
}
func (ts *testApi) LogMark(s string, args ...any) {
	ts.t.Helper()
	ts.t.Logf("\n------------------------------------------------------------\n-- %s --\n------------------------------------------------------------\n", fmt.Sprintf(s, args...))
}
func (ts *testApi) expectSimpleBoardEquality(wantValues ...int64) {
	ts.t.Helper()
	g := ts.Game()
	gameCells := g.Cells()
	gotCellValues := make([]int64, len(gameCells))
	for i := 0; i < len(gameCells); i++ {
		gotCellValues[i] = gameCells[i].Value()
	}

	if diff := deep.Equal(gotCellValues, wantValues); diff != nil {
		cells := cellCreator(wantValues...)
		ts.t.Logf("%#v", cells)
		gb := tallylogic.NewTableBoard(3, 3, tallylogic.TableBoardOptions{
			Cells: cells,
		})
		ts.t.Fatalf("Board did not match: Got\n%s\nWant\n%s\n\n (%v != %v)", g.Print(), gb.String(), gotCellValues, wantValues)
	}

}
func cellCreator(vals ...int64) []cell.Cell {
	cells := make([]cell.Cell, len(vals))
	for i, v := range vals {
		cells[i] = cell.NewCell(v, 0)
	}
	return cells
}
