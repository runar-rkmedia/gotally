package api

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/bufbuild/connect-go"
	gonanoid "github.com/matoous/go-nanoid/v2"
	model "github.com/runar-rkmedia/gotally/gen/proto/tally/v1"
	"github.com/runar-rkmedia/gotally/gen/proto/tally/v1/tallyv1connect"
	"github.com/runar-rkmedia/gotally/live_client/ex"
	logic "github.com/runar-rkmedia/gotally/tallylogic"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
)

var (
	port = "8080"
)

func StartServer() {

	// TODO: use session / localstorage / params
	game := ex.NewGameModel(logic.GameModeTemplate, &logic.ChallengeGames[1])
	ex.Cache.SetGame("", game)
	if false {

		if err := run(); err != nil {
			log.Fatal(err)
		}
	} else {
		tally := &TallyServer{}
		mux := http.NewServeMux()
		path, handler := tallyv1connect.NewBoardServiceHandler(tally)
		mux.Handle(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := r.Header.Get("X-Request-ID")
			if reqID == "" {
				reqID = gonanoid.Must()
			}
			w.Header().Set("X-Request-ID", reqID)

			// CORS
			w.Header().Set("Access-Control-Expose-Headers", "Date, X-Request-ID")
			w.Header().Set("Access-Control-Allow-Headers", "content-type")
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET")
			w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
			w.Header().Set("Access-Control-Max-Age", "60")
			if r.Method != http.MethodOptions {
				handler.ServeHTTP(w, r)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		}))
		address := "localhost:" + port
		fmt.Println("starting server on http://" + address + path)
		if err := http.ListenAndServe(
			"localhost:8080",
			h2c.NewHandler(
				mux,
				// handlers.CORS(
				// 	handlers.AllowedOrigins([]string{"http://localhost:5173"}),
				// )(mux),
				&http2.Server{}),
		); err != nil {
			panic(err)
		}
	}
}

func run() error {
	listenOn := "127.0.0.1:" + port
	fmt.Println("starting grpc-listener on ", listenOn)
	listener, err := net.Listen("tcp", listenOn)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", listenOn, err)
	}

	server := grpc.NewServer()
	model.RegisterBoardServiceServer(server, &tallyStoreServiceServer{})
	log.Println("Listening on", listenOn)
	if err := server.Serve(listener); err != nil {
		return fmt.Errorf("failed to serve gRPC server: %w", err)
	}

	return nil
}

type tallyStoreServiceServer struct {
	model.UnimplementedBoardServiceServer
}

type TallyServer struct{}

func (s *TallyServer) SwipeBoard(
	ctx context.Context,
	req *connect.Request[model.SwipeBoardRequest],
) (*connect.Response[model.SwipeBoardResponse], error) {

	if req.Msg.Direction == model.SwipeDirection_SWIPE_DIRECTION_UNSPECIFIED {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Direction must be set"))
	}
	sessionID := req.Header().Get("Authorization")
	game := ex.Cache.GetGame(sessionID)
	dir := toGameSwipeDirection(req.Msg.Direction)
	response := &model.SwipeBoardResponse{
		DidChange: game.Swipe(dir),
		Board:     toModalBoard(game.Game),
	}
	res := connect.NewResponse(response)
	res.Header().Set("PetV", "v1")
	return res, nil
}

type CError struct {
	connect.Error
}

func (c *CError) ToConnectError() *connect.Error {
	return &c.Error
}
func (c *CError) AddBadRequestDetail(violations []*errdetails.BadRequest_FieldViolation) *CError {

	details := &errdetails.BadRequest{
		FieldViolations: violations,
	}
	if details, detailErr := connect.NewErrorDetail(details); detailErr == nil {
		c.AddDetail(details)
	}
	return c
}

func createError(c connect.Code, err error) CError {
	cerr := CError{*connect.NewError(c, err)}
	return cerr
}

func toGameSwipeDirection(dir model.SwipeDirection) logic.SwipeDirection {
	switch dir {
	case model.SwipeDirection_SWIPE_DIRECTION_UP:
		return logic.SwipeDirectionUp
	case model.SwipeDirection_SWIPE_DIRECTION_RIGHT:
		return logic.SwipeDirectionRight
	case model.SwipeDirection_SWIPE_DIRECTION_DOWN:
		return logic.SwipeDirectionDown
	case model.SwipeDirection_SWIPE_DIRECTION_LEFT:
		return logic.SwipeDirectionLeft
	}
	return ""
}

func (s *TallyServer) GetBoard(
	ctx context.Context,
	req *connect.Request[model.GetBoardRequest],
) (*connect.Response[model.GetBoardResponse], error) {
	// TODO:  Get from context
	sessionID := req.Header().Get("Authorization")
	game := ex.Cache.GetGame(sessionID)
	response := &model.GetBoardResponse{
		Board: toModalBoard(game.Game),
	}
	res := connect.NewResponse(response)
	res.Header().Set("PetV", "v1")
	return res, nil
}

func toModalBoard(game logic.Game) *model.Board {
	return &model.Board{
		Columns: int32(game.Rules.SizeX),
		Rows:    int32(game.Rules.SizeX),
		Cell:    toModalCells(game.Cells()),
	}
}

func toModalCells(cells []logic.Cell) []*model.Cell {
	c := make([]*model.Cell, len(cells))
	for i := 0; i < len(cells); i++ {
		base, twopow := cells[i].Raw()
		c[i] = &model.Cell{
			Base:   base,
			Twopow: twopow,
		}

	}
	return c
}
