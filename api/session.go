package api

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	gonanoid "github.com/jaevor/go-nanoid"
	model "github.com/runar-rkmedia/gotally/gen/proto/tally/v1"
)

func (s *TallyServer) GetSession(
	ctx context.Context,
	req *connect.Request[model.GetSessionRequest],
) (*connect.Response[model.GetSessionResponse], error) {
	session := ContextGetUserState(ctx)
	response := &model.GetSessionResponse{
		Session: &model.Session{
			SessionId: session.SessionID,
			Username:  session.UserName,
			Game: &model.Game{
				Board:       toModalBoard(&session.Game),
				Score:       session.Game.Score(),
				Moves:       int64(session.Game.Moves()),
				Description: session.Game.Description,
			},
		},
	}
	res := connect.NewResponse(response)
	return res, nil
}

func ContextGetUserState(ctx context.Context) *UserState {
	v := ctx.Value(ContextKeyUserState)

	return v.(*UserState)
}

const tokenHeader = "Authorization"

func mustCreateUUidgenerator() func() string {
	s, err := gonanoid.Standard(21)
	if err != nil {
		panic(fmt.Errorf("Failed in MustGenerateUIDLike: %w", err))
	}
	return s
}
