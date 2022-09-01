package api

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	model "github.com/runar-rkmedia/gotally/gen/proto/tally/v1"
)

func (s *TallyServer) SwipeBoard(
	ctx context.Context,
	req *connect.Request[model.SwipeBoardRequest],
) (*connect.Response[model.SwipeBoardResponse], error) {

	if req.Msg.Direction == model.SwipeDirection_SWIPE_DIRECTION_UNSPECIFIED {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Direction must be set"))
	}
	session := UserStateFromContext(ctx)
	dir := toGameSwipeDirection(req.Msg.Direction)
	response := &model.SwipeBoardResponse{
		DidChange: session.Swipe(dir),
		Board:     toModalBoard(&session.Game),
		Moves:     int64(session.Game.Moves()),
	}
	res := connect.NewResponse(response)
	res.Header().Set("PetV", "v1")
	return res, nil
}
