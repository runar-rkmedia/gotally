package api

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	model "github.com/runar-rkmedia/gotally/gen/proto/tally/v1"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
)

func (s *TallyServer) CombineCells(
	ctx context.Context,
	req *connect.Request[model.CombineCellsRequest],
) (*connect.Response[model.CombineCellsResponse], error) {
	var length int
	var path []int
	switch t := req.Msg.Selection.(type) {
	case *model.CombineCellsRequest_Indexes:
		fmt.Println("i am index", t.Indexes)
		length = len(t.Indexes.Index)
		path = make([]int, length)
		for i := 0; i < length; i++ {
			path[i] = int(t.Indexes.Index[i])
		}
	case *model.CombineCellsRequest_Coordinate:
		fmt.Println("i am coordinate", t.Coordinate)
		return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("Not implemented: coordinate"))
	}

	fmt.Println("lenth", length)
	if length < 2 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Selection must be have atleast two items"))
	}
	session := UserStateFromContext(ctx)
	if err, invalidIndex := session.Game.ValidatePath(path); err != nil {
		fmt.Println("\npath", path)

		cerr := createError(connect.CodeInvalidArgument, fmt.Errorf("Invalid path at (%d): %w", invalidIndex, err))
		details := []*errdetails.BadRequest_FieldViolation{
			{
				Field: fmt.Sprintf("path-index: %d", invalidIndex),
			},
		}
		if invalidIndex >= 1 {
			if neighbours, ok := session.NeighboursForCellIndex(path[invalidIndex-1]); ok {
				details[0].Description = fmt.Sprintf("The item before the invalid path (%d), has valid neighbours %v", path[invalidIndex-1], neighbours)
			}
		}
		cerr.AddBadRequestDetail(details)

		return nil, cerr.ToConnectError()
	}
	ok := session.EvaluateForPath(path)
	if !ok {
		cerr := connect.NewError(connect.CodeNotFound, fmt.Errorf("path does evaluate to the final item the selection"))
		details := &errdetails.BadRequest_FieldViolation{
			Field: "NoEval",
		}
		if details, detailErr := connect.NewErrorDetail(details); detailErr == nil {
			cerr.AddDetail(details)
		}
		return nil, cerr
	}
	response := model.CombineCellsResponse{
		Board:  toModalBoard(&session.Game),
		Score:  session.Game.Score(),
		Moves:  int64(session.Game.Moves()),
		DidWin: session.Game.IsGameWon(),
	}
	fmt.Println("didwin", response.DidWin)
	res := connect.NewResponse(&response)
	res.Header().Set("PetV", "v1")
	return res, nil
}
