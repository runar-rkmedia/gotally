package api

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	model "github.com/runar-rkmedia/gotally/gen/proto/tally/v1"
)

func (s *TallyServer) RestartGame(
	ctx context.Context,
	req *connect.Request[model.RestartGameRequest],
) (*connect.Response[model.RestartGameResponse], error) {
	session := ContextGetUserState(ctx)
	s.l.Warn().Msg("Not implemented: RestartGame")

	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("RestartGame is not implemented"))

	response := &model.RestartGameResponse{
		Board: toModalBoard(&session.Game),
		Score: session.Game.Score(),
		Moves: int64(session.Game.Moves()),
	}
	res := connect.NewResponse(response)
	return res, nil
}
