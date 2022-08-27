package api

import (
	"context"

	"github.com/bufbuild/connect-go"
	model "github.com/runar-rkmedia/gotally/gen/proto/tally/v1"
	"github.com/runar-rkmedia/gotally/live_client/ex"
	logic "github.com/runar-rkmedia/gotally/tallylogic"
)

func (s *TallyServer) NewGame(
	ctx context.Context,
	req *connect.Request[model.NewGameRequest],
) (*connect.Response[model.NewGameResponse], error) {
	// TODO:  Get from context
	sessionID := req.Header().Get("Authorization")
	// game := ex.Cache.GetGame(sessionID)
	game := ex.NewGameModel(logic.GameModeTemplate, &logic.ChallengeGames[1])
	ex.Cache.SetGame(sessionID, game)
	response := &model.NewGameResponse{
		Board: toModalBoard(game.Game),
		Score: game.Game.Score(),
		Moves: int64(game.Game.Moves()),
	}
	res := connect.NewResponse(response)
	return res, nil
}
