package api

import (
	"context"
	"fmt"
	"time"

	"github.com/bufbuild/connect-go"
	model "github.com/runar-rkmedia/gotally/gen/proto/tally/v1"
	"github.com/runar-rkmedia/gotally/tallylogic/cell"
	"github.com/runar-rkmedia/gotally/types"
)

func (s *TallyServer) CreateGameChallenge(
	ctx context.Context,
	req *connect.Request[model.CreateGameChallengeRequest],
) (*connect.Response[model.CreateGameChallengeResponse], error) {
	if !s.FeatureGameGeneration {
		err := connect.NewError(connect.CodeResourceExhausted, fmt.Errorf("Creating game challenges has been disabled"))
		return nil, err
	}
	session := ContextGetUserState(ctx)
	if req.Msg.Rows > 16 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("rows must be below 16"))
	}
	if req.Msg.Columns > 16 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("columns must be below 16"))
	}
	if req.Msg.Rows <= 2 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("rows must be above 2"))
	}
	if req.Msg.Columns <= 2 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("columns must be above 2"))
	}
	if req.Msg.TargetCellValue == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("TargetCellValue must be set"))
	}
	if req.Msg.IdealMoves == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("IdealMoves must be set"))
	}
	expectedCells := int(req.Msg.Columns * req.Msg.Rows)
	if expectedCells != len(req.Msg.Cells) {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("The number of cells must match Rows*Columns"))
	}

	payload := types.CreateGameTemplatePayload{
		ID:          s.UidGenerator(),
		CreatedAt:   time.Now(),
		CreatedByID: session.UserID,
		Description: req.Msg.Description,
		IdealMoves:  int(req.Msg.IdealMoves),
		IdealScore:  int(req.Msg.IdealScore),
		Name:        req.Msg.Name,
		Cells:       make([]cell.Cell, req.Msg.Rows*req.Msg.Columns),
		Rules: types.Rules{
			ID:              "",
			CreatedAt:       time.Now(),
			Mode:            types.RuleModeChallenge,
			TargetCellValue: req.Msg.TargetCellValue,
			TargetScore:     0,
			MaxMoves:        0,
			Rows:            uint8(req.Msg.Rows),
			Columns:         uint8(req.Msg.Columns),
			RecreateOnSwipe: false,
			NoReSwipe:       false,
			NoMultiply:      false,
			NoAddition:      false,
		},
	}
	if req.Msg.ChallengeNumber != 0 {
		i := int(req.Msg.ChallengeNumber)
		payload.ChallengeNumber = &i
	}
	for i, v := range req.Msg.Cells {
		payload.Cells[i] = cell.NewCell(v.Base, int(v.Twopow))
	}

	if err := payload.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	template, err := s.storage.CreateGameTemplate(ctx, payload)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	response := &model.CreateGameChallengeResponse{
		Id:              template.ID,
		ChallengeNumber: intPointerUint32(template.ChallengeNumber),
	}
	if template.ChallengeNumber != nil {
		response.ChallengeNumber = uint32(*template.ChallengeNumber)
	}

	return connect.NewResponse(response), nil

}
func (s *TallyServer) GetGameChallenges(
	ctx context.Context,
	req *connect.Request[model.GetGameChallengesRequest],
) (*connect.Response[model.GetGameChallengesResponse], error) {
	session := ContextGetUserState(ctx)

	c, err := s.storage.GetGameChallenges(ctx, types.GetGameChallengePayload{StatsForUserID: session.UserID})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get game challenges: %w", err))
	}

	response := &model.GetGameChallengesResponse{}
	response.Challenges = make([]*model.GameChallenge, len(c))

	for i := 0; i < len(c); i++ {
		response.Challenges[i] = &model.GameChallenge{
			Id:              c[i].ID,
			ChallengeNumber: intPointerUint32(c[i].ChallengeNumber),
			IdealMoves:      intPointerUint32(c[i].IdealMoves),
			TargetCellValue: c[i].TargetCellValue,
			Columns:         uint32(c[i].Columns),
			Rows:            uint32(c[i].Rows),
			Name:            c[i].Name,
			Description:     c[i].Description,
			Cells:           toModalCells(c[i].Cells),
		}
		for _, s := range c[i].Stats {
			if s.Score > 0 {
				if response.Challenges[i].CurrentUsersBestScore < s.Score {
					response.Challenges[i].CurrentUsersBestScore = s.Score
				}
			}
			if s.Moves > 0 {
				if response.Challenges[i].CurrentUsersFewestMoves == 0 {
					response.Challenges[i].CurrentUsersFewestMoves = uint32(s.Moves)
				}
				if response.Challenges[i].CurrentUsersFewestMoves > uint32(s.Moves) {
					response.Challenges[i].CurrentUsersFewestMoves = uint32(s.Moves)
				}
			}
		}
		response.Challenges[i].Rating = calculateRating(
			response.Challenges[i].CurrentUsersBestScore,
			response.Challenges[i].IdealScore,
			response.Challenges[i].CurrentUsersFewestMoves,
			response.Challenges[i].IdealMoves,
		)
	}

	return connect.NewResponse(response), nil

}

func calculateRating(score, idealScore uint64, moves, idealMoves uint32) model.Rating {
	if moves == 0 || score == 0 {
		return model.Rating_RATING_UNPLAYED
	}
	if idealScore != 0 {
		panic("ideal-score is not yet implemeted")
	}
	if idealMoves == 0 {
		return model.Rating_RATING_UNSPECIFIED
	}
	diff := idealMoves - moves
	if diff == 0 {
		return model.Rating_RATING_SUPERB
	}
	if diff < 0 {
		return model.Rating_RATING_SUPERB
	}
	switch diff {
	case 1:
		return model.Rating_RATING_GREAT
	case 2:
		return model.Rating_RATING_GOOD
	case 3:
		return model.Rating_RATING_WELL
	}

	return model.Rating_RATING_OK
}
func intPointerUint32(i *int) uint32 {
	if i == nil {
		return 0
	}
	return uint32(*i)
}
func uint32TointPointer(i uint32) *int {
	if i == 0 {
		return nil
	}
	n := int(i)
	return &n
}
