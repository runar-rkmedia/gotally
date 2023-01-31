package api

import (
	"context"
	"errors"
	"fmt"
	"math/rand"

	"github.com/bufbuild/connect-go"
	model "github.com/runar-rkmedia/gotally/gen/proto/tally/v1"
	"github.com/runar-rkmedia/gotally/generated"
	"github.com/runar-rkmedia/gotally/tallylogic"
	logic "github.com/runar-rkmedia/gotally/tallylogic"
	"github.com/runar-rkmedia/gotally/types"
	"gopkg.in/yaml.v2"
)

var (
	ErrSliceUnderflow = errors.New("Slice index underflow")
	ErrSliceOverFlow  = errors.New("Slice index overflow")
)

func devpretty(j any) string {
	b, _ := yaml.Marshal(j)
	return string(b)
}

func (s *TallyServer) NewGame(
	ctx context.Context,
	req *connect.Request[model.NewGameRequest],
) (*connect.Response[model.NewGameResponse], error) {
	session := ContextGetUserState(ctx)
	var mode logic.GameMode
	var template *logic.GameTemplate

	switch req.Msg.Mode {
	case model.GameMode_GAME_MODE_RANDOM_CHALLENGE:
		mode = logic.GameModeRandomChallenge
		index := rand.Intn(len(generated.GeneratedTemplates))
		template = &generated.GeneratedTemplates[index]
	case model.GameMode_GAME_MODE_TUTORIAL:
		mode = logic.GameModeTemplate
		if _i, ok := req.Msg.Variant.(*model.NewGameRequest_LevelIndex); ok {
			i := int(_i.LevelIndex)
			if i < 0 {
				return nil, fmt.Errorf("invalid levelindex, must be positive got %d", i)
			}
			if i >= len(logic.ChallengeGames) {
				return nil, fmt.Errorf("invalid levelindex, got %d, must be lower than %d", i, len(logic.ChallengeGames))
			}
			template = &logic.ChallengeGames[i]
		} else {
			template = &logic.ChallengeGames[0]
		}

	}

	game, err := logic.NewGame(mode, template)
	if err != nil {
		return nil, fmt.Errorf("failed to created game: %w", err)
	}
	payload := types.NewGamePayload{
		Game: toTypeGame(game, session.UserID),
	}
	tg, err := s.storage.NewGameForUser(ctx, payload)
	if err != nil {
		s.l.Error().Err(err).Msg("failed to create new game")
		return nil, fmt.Errorf("internal error while creating new game")
	}
	game.ID = tg.ID
	game, err = tallylogic.RestoreGame(&tg)
	if err != nil {
		return nil, fmt.Errorf("failed to restore game: %w", err)
	}

	session.Game = game
	Store.SetUserState(session)
	response := &model.NewGameResponse{
		Description: session.Game.Description,
		Board:       toModalBoard(&session.Game),
		Score:       session.Game.Score(),
		Moves:       int64(session.Game.Moves()),
	}
	if response.Description == "" {
		response.Description = session.Game.Name
	}
	res := connect.NewResponse(response)
	return res, nil
}
