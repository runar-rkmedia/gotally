package api

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	gonanoid "github.com/jaevor/go-nanoid"
	model "github.com/runar-rkmedia/gotally/gen/proto/tally/v1"
	"github.com/runar-rkmedia/gotally/types"
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
				Mode:        toModelGameMode(session.Rules.GameMode),
			},
		},
	}
	switch session.PlayState {
	case types.PlayStateWon:
		response.Session.Game.PlayState = model.PlayState_PLAYSTATE_WON
	case types.PlayStateLost:
		response.Session.Game.PlayState = model.PlayState_PLAYSTATE_LOST
	case types.PlayStateAbandoned:
		response.Session.Game.PlayState = model.PlayState_PLAYSTATE_ABANDONED
	case types.PlayStateCurrent:
		response.Session.Game.PlayState = model.PlayState_PLAYSTATE_CURRENT
	case "":
		s.l.Error().Str("playstate", session.PlayState).Msg("Empty PlayState")
		if session.Game.IsGameWon() {
			response.Session.Game.PlayState = model.PlayState_PLAYSTATE_WON
		} else if session.Game.IsGameOver() {
			response.Session.Game.PlayState = model.PlayState_PLAYSTATE_LOST
		}
	default:
		s.l.Error().Str("playstate", session.PlayState).Msg("Unhandled PlayState")
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
