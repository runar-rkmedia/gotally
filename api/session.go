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
	// TODO:  Get from context
	session := UserStateFromContext(ctx)
	response := &model.GetSessionResponse{
		Session: &model.Session{
			Game: &model.Game{
				Board:       toModalBoard(&session.Game),
				Score:       session.Game.Score(),
				Moves:       int64(session.Game.Moves()),
				Description: session.Game.Name,
			},
		},
	}
	res := connect.NewResponse(response)
	res.Header().Set("PetV", "v1")
	return res, nil
}

func UserStateFromContext(ctx context.Context) *UserState {
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
