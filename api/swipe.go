package api

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	model "github.com/runar-rkmedia/gotally/gen/proto/tally/v1"
	"github.com/runar-rkmedia/gotally/types"
)

func (s *TallyServer) SwipeBoard(
	ctx context.Context,
	req *connect.Request[model.SwipeBoardRequest],
) (*connect.Response[model.SwipeBoardResponse], error) {

	if req.Msg.Direction == model.SwipeDirection_SWIPE_DIRECTION_UNSPECIFIED {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Direction must be set"))
	}
	session := ContextGetUserState(ctx)
	dir := toGameSwipeDirection(req.Msg.Direction)
	// We copy the game, to rollback the in-memory cache if anything goes wrong
	gameCopy := session.Game.Copy()
	response := &model.SwipeBoardResponse{
		DidChange: session.Game.Swipe(dir),
		Board:     toModalBoard(&session.Game),
		Moves:     int64(session.Game.Moves()),
	}
	if response.DidChange {
		didWin := session.Game.IsGameWon()
		didLose := session.Game.IsGameOver()
		seed, state := session.Game.Seed()
		payload := types.UpdateGamePayload{
			GameID:    session.Game.ID,
			Moves:     session.Game.Moves(),
			Score:     uint64(gameCopy.Score()),
			State:     state,
			Seed:      seed,
			Cells:     session.Cells(),
			History:   session.Game.History.Bytes(),
			PlayState: types.PlayStateCurrent,
		}
		if didWin {
			payload.PlayState = types.PlayStateWon
			response.DidWin = true
		} else if didLose {
			payload.PlayState = types.PlayStateLost
			response.DidWin = didLose
		}
		err := s.storage.UpdateGame(ctx, payload)
		if err != nil {
			s.l.Error().
				Err(err).
				Interface("payload", payload).
				Msg("failed to save the board to storage during swipe-operation")
			// rollback the game in memory
			session.Game = gameCopy
			return nil, fmt.Errorf("intarnal failure while saving the board: %w", err)
		}
	}
	res := connect.NewResponse(response)
	return res, nil
}
