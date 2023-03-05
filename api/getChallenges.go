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
	expectedCells := int(req.Msg.Columns * req.Msg.Rows)
	if expectedCells != len(req.Msg.Cells) {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("The number of cells must match Rows*Columns"))
	}

	payload := types.CreateGameTemplatePayload{
		ID:          s.UidGenerator(),
		CreatedAt:   time.Now(),
		CreatedByID: session.UserID,
		Description: req.Msg.Description,
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

	fmt.Printf("fofofofofo %#v", template.ChallengeNumber)
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

	c, err := s.storage.GetGameChallenges(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get game challenges: %w", err))
	}

	response := &model.GetGameChallengesResponse{}
	response.Challenges = make([]*model.GameChallenge, len(c))

	for i := 0; i < len(c); i++ {
		fmt.Printf("chal %d %#v %#v\n", i, c[i].Cells, c[i].Rules)
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

	}

	return connect.NewResponse(response), nil

}
func intPointerUint32(i *int) uint32 {
	if i == nil {
		return 0
	}
	return uint32(*i)
}
