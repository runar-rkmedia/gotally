package api

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	model "github.com/runar-rkmedia/gotally/gen/proto/tally/v1"
	"github.com/runar-rkmedia/gotally/tallylogic"
	"github.com/runar-rkmedia/gotally/types"
)

func (s *TallyServer) RestartGame(
	ctx context.Context,
	req *connect.Request[model.RestartGameRequest],
) (*connect.Response[model.RestartGameResponse], error) {
	session := ContextGetUserState(ctx)

	tg, err := s.storage.RestartGame(ctx, types.RestartGamePayload{
		UserID: session.UserID,
		GameID: session.Game.ID,
	})
	if err != nil {
		cerr := createError(connect.CodeInternal, fmt.Errorf("failed to restart the game: %w", err))
		return nil, cerr.ToConnectError()
	}
	g, err := tallylogic.RestoreGame(&tg)
	if err != nil {
		cerr := createError(connect.CodeInternal, fmt.Errorf("failed to restore the game: %w", err))
		return nil, cerr.ToConnectError()
	}
	session.Game = g
	response := &model.RestartGameResponse{
		Board: toModalBoard(&session.Game),
		Score: session.Game.Score(),
		Moves: int64(session.Game.Moves()),
	}
	res := connect.NewResponse(response)
	return res, nil
}
