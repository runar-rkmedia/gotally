package api

import (
	"context"

	"github.com/bufbuild/connect-go"
	model "github.com/runar-rkmedia/gotally/gen/proto/tally/v1"
	"github.com/runar-rkmedia/gotally/tallylogic"
)

func intsToInt32s(ints []int) []int32 {
	out := make([]int32, len(ints))
	for i := 0; i < len(ints); i++ {
		out[i] = int32(ints[i])

	}
	return out
}

const MaxUint = ^uint(0)
const MaxInt = int(MaxUint >> 1)

func (s *TallyServer) GetHint(
	ctx context.Context,
	req *connect.Request[model.GetHintRequest],
) (*connect.Response[model.GetHintResponse], error) {
	session := ContextGetUserState(ctx)

	response := &model.GetHintResponse{
		// Instruction: []*model.Instruction{},
	}
	if false && session.Game.Rules.GameMode == tallylogic.GameModeDefault {
		hints := session.GetHint()
		response.Instructions = make([]*model.Instruction, len(hints))
		i := 0
		for _, h := range hints {
			response.Instructions[i] = &model.Instruction{
				InstructionOneof: &model.Instruction_Combine{
					Combine: &model.Indexes{
						Index: intsToInt32s(h.Path),
					},
				},
			}
			i++
		}
	}
	if true || session.Game.Rules.GameMode == tallylogic.GameModeDefault {
		solver := tallylogic.NewBruteSolver(tallylogic.SolveOptions{
			MaxDepth:     10,
			MaxVisits:    100,
			MinMoves:     0,
			MaxMoves:     10,
			MaxSolutions: 10,
		})
		games, err := solver.SolveGame(session.Game)
		if err != nil {
			return nil, err
		}
		if len(games) == 0 {
			return connect.NewResponse(response), nil
		}
		shortestMoves := MaxInt
		bestIndex := 0
		for i := 0; i < len(games); i++ {
			length := len(games[i].History)
			if length < shortestMoves {
				shortestMoves = length
				bestIndex = i
			}
		}
		bestInstructions := games[bestIndex].History
		response.Instructions = make([]*model.Instruction, len(bestInstructions))
		i := 0
		for _, h := range bestInstructions {
			switch t := h.(type) {
			case tallylogic.SwipeDirection:
				response.Instructions[i] = &model.Instruction{
					InstructionOneof: &model.Instruction_Swipe{
						Swipe: toModalDirection(t),
					},
				}
			case []int:
				response.Instructions[i] = &model.Instruction{
					InstructionOneof: &model.Instruction_Combine{
						Combine: &model.Indexes{
							Index: intsToInt32s(t),
						},
					},
				}
			}
			i++
		}
	}
	res := connect.NewResponse(response)
	return res, nil
}
