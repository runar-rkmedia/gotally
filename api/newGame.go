package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/bufbuild/connect-go"
	model "github.com/runar-rkmedia/gotally/gen/proto/tally/v1"
	"github.com/runar-rkmedia/gotally/tallylogic"
	logic "github.com/runar-rkmedia/gotally/tallylogic"
	"github.com/runar-rkmedia/gotally/types"
)

var (
	ErrSliceUnderflow = errors.New("Slice index underflow")
	ErrSliceOverFlow  = errors.New("Slice index overflow")
)

func devpretty(j any) string {
	b, _ := json.MarshalIndent(j, "", "  ")
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
	case model.GameMode_GAME_MODE_RANDOM:
		mode = logic.GameModeRandom
	case model.GameMode_GAME_MODE_RANDOM_CHALLENGE:
		mode = logic.GameModeRandomChallenge
		if variant, ok := req.Msg.Variant.(*model.NewGameRequest_Id); ok {
			id := variant.Id
			if id == "" {
				return nil, fmt.Errorf("ID is required for Variatn with id %s", id)
			}
			// TODO: Get by id
			challenges, err := s.storage.GetGameChallenges(ctx)
			if err != nil {
				return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to retrieve game-challenges %w", err))
			}
			fmt.Println("\n\n hello", id, len(challenges))
			var challenge types.GameTemplate
			for _, c := range challenges {
				if c.ID == id {
					challenge = c
				}
				break
			}
			if challenge.ID != id {
				return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("challenge not found: '%s'", id))

			}
			// TODO: this is tedious, fix it
			template, err = challengeToTemplate(challenge)
			if err != nil {
				return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to map challenge: %w", err))
			}
			fmt.Println("yaya", challenge.ID, challenge.Rules.ID, challenge.TargetCellValue, challenge.Cells, template.Board.PrintBoard(nil))
		} else {

			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("could not resolve challenge"))
		}
		// if len(generated.GeneratedTemplates) == 0 {
		// 	s.l.Error().Msg("No challanges are available at this time. Please ensure that GeneratedTemplates has been initialized.")
		// 	return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("No challanges are available at this time"))
		// }
		// index := rand.Intn(len(generated.GeneratedTemplates))
		// template = &generated.GeneratedTemplates[index]
	case model.GameMode_GAME_MODE_TUTORIAL:
		mode = logic.GameModeTutorial
		if _i, ok := req.Msg.Variant.(*model.NewGameRequest_LevelIndex); ok {
			i := int(_i.LevelIndex)
			if i < 0 {
				return nil, fmt.Errorf("invalid levelindex, must be positive got %d", i)
			}
			if i >= len(logic.TutorialGames) {
				return nil, fmt.Errorf("invalid levelindex, got %d, must be lower than %d", i, len(logic.TutorialGames))
			}
			template = &logic.TutorialGames[i]
		} else {
			template = &logic.TutorialGames[0]
		}
	default:
		s.l.Error().Str("req.Msg.Mode", req.Msg.Mode.String()).Msg("Unhandled game-mode")
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to create game, unhandled mode"))

	}
	// Some extra validation
	switch req.Msg.Mode {
	case model.GameMode_GAME_MODE_RANDOM_CHALLENGE:
		if template.Rules.TargetCellValue == 0 {
			fmt.Println("\n\nboom", template.Name, template.ID, template.Rules.TargetCellValue, template.Rules.ID)
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("There was an error mapping the challenge, expected TargetCellValue to be set, but it was not"))
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
	if template != nil {
		tg.Rules.TargetCellValue = template.Rules.TargetCellValue
		tg.Rules.TargetScore = template.Rules.TargetScore
		tg.Rules.TargetScore = template.Rules.TargetScore
		tg.Rules.MaxMoves = template.Rules.MaxMoves
	}
	game.ID = tg.ID
	game, err = tallylogic.RestoreGame(&tg)
	if err != nil {

		s.l.Error().Err(err).Msg("failed to restore game during call to newgame")
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("failed to restore game: %w", err))
	}

	session.Game = game
	Store.SetUserState(session)
	response := &model.NewGameResponse{
		Description: session.Game.Description,
		Board:       toModalBoard(&session.Game),
		Score:       session.Game.Score(),
		Moves:       int64(session.Game.Moves()),
		Mode:        toModelGameMode(session.Game.Rules.GameMode),
	}
	if response.Description == "" {
		response.Description = session.Game.Name
	}
	res := connect.NewResponse(response)
	return res, nil
}

func challengeToTemplate(challenge types.GameTemplate) (*logic.GameTemplate, error) {

	template := logic.NewGameTemplate(logic.GameModeRandomChallenge, challenge.ID, challenge.Name, challenge.Description, int(challenge.Rows), int(challenge.Columns))
	r := challenge.Rules
	template.Rules = logic.GameRules{
		ID:              r.ID,
		GameMode:        logic.GameModeRandomChallenge,
		SizeX:           int(r.Columns),
		SizeY:           int(r.Rows),
		TargetCellValue: r.TargetCellValue,
		MaxMoves:        r.MaxMoves,
		TargetScore:     r.TargetScore,
		RecreateOnSwipe: r.RecreateOnSwipe,
		// WithSuperPowers: r.WithSuperPowers,
		// StartingBricks:  r.StartingBricks,
		NoReswipe: r.NoReSwipe,
		Options: logic.NewGameOptions{
			TableBoardOptions: logic.TableBoardOptions{
				EvaluateOptions: logic.EvaluateOptions{
					NoMultiply: false,
					NoAddition: false,
				},
				Cells: challenge.Cells,
			},
			Seed:  0,
			State: 0,
		},
	}
	template = template.SetGoalCheckerLargestValue(r.TargetCellValue)
	template.SetStartingCells(challenge.Cells)
	return template, nil
}
