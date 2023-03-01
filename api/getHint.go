package api

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/bufbuild/connect-go"
	model "github.com/runar-rkmedia/gotally/gen/proto/tally/v1"
	"github.com/runar-rkmedia/gotally/tallylogic"
)

func intsTouInt32s(ints []int) []uint32 {
	out := make([]uint32, len(ints))
	for i := 0; i < len(ints); i++ {
		out[i] = uint32(ints[i])
	}
	return out
}

const MaxUint = ^uint(0)
const MaxInt = int(MaxUint >> 1)

func (s *TallyServer) GetHint(
	ctx context.Context,
	req *connect.Request[model.GetHintRequest],
) (*connect.Response[model.GetHintResponse], error) {
	// There are usually a lot of hints available,
	// but some hints leads to more fun games than others
	// In general, multiplication is more fun than addition
	// and swiping is a bit dull.
	// Long hints seem like magic, so we prefer shorter hints.
	// However, too short hints are also boring
	// TODO: introduce a weighted hint and solution-sorter
	session := ContextGetUserState(ctx)

	response := &model.GetHintResponse{
		// Instruction: []*model.Instruction{},
	}
	// Get a single hint. Does not look ahead to do swipes etc.
	if session.Game.Rules.GameMode == tallylogic.GameModeRandom {
		hints := session.GetHint()
		if len(hints) > 0 {

			response.Instructions = make([]*model.Instruction, 1)
			s.l.Debug().
				Bool("deep", false).
				Int("hintsFound", len(hints)).
				Msg("Returning hints")
			// TODO: sort these better
			for _, h := range hints {
				if h.Method != tallylogic.EvalMethodProduct {
					continue
				}
				response.Instructions[0] = &model.Instruction{
					InstructionOneof: &model.Instruction_Combine{
						Combine: &model.Indexes{
							Index: intsTouInt32s(h.Path),
						},
					},
				}
				return connect.NewResponse(response), nil
			}
			for _, h := range hints {
				response.Instructions[0] = &model.Instruction{
					InstructionOneof: &model.Instruction_Combine{
						Combine: &model.Indexes{
							Index: intsTouInt32s(h.Path),
						},
					},
				}
				return connect.NewResponse(response), nil
			}
		}
	}
	response.Instructions = make([]*model.Instruction, 1)
	// Deeper hint, looking ahead to find better hints, attempting to solve the game if possible.
	// h := tallylogic.NewHintCalculator(session.Game, session.Game, session.Game)
	games, err := tallylogic.SolveGame(tallylogic.SolveOptions{
		MaxDepth:     10,
		MaxVisits:    6000,
		MinMoves:     0,
		MaxMoves:     10,
		MaxSolutions: 1,
		MaxTime:      time.Second * 10,
	}, session.Game, nil)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Failed to generate hint"))
	}
	s.l.Debug().
		Bool("deep", true).
		Int("solutions", len(games)).
		Msg("Solver returned solutiosn")
	if len(games) == 0 {
		return connect.NewResponse(response), nil
	}
	if req.Msg.MaxLength == 0 {
		req.Msg.MaxLength = 1
	}
	if req.Msg.HintPreference == model.HintPreference_HINT_PREFERENCE_UNSPECIFIED {
		if session.Game.Rules.GameMode == tallylogic.GameModeRandom {
			req.Msg.HintPreference = model.HintPreference_HINT_PREFERENCE_FIRST_COMBINE
		} else {
			req.Msg.HintPreference = model.HintPreference_HINT_PREFERENCE_SHORT
		}
	}
	switch req.Msg.HintPreference {
	case model.HintPreference_HINT_PREFERENCE_HIGHEST_SCORE:
		sort.Slice(games, func(i, j int) bool {
			return games[i].Score() < games[j].Score()
		})
	case model.HintPreference_HINT_PREFERENCE_SHORT:
		sort.Slice(games, func(i, j int) bool {
			return len(games[i].History) < len(games[j].History)
		})
	case model.HintPreference_HINT_PREFERENCE_MINIMUM_SWIPES:
		sort.Slice(games, func(i, j int) bool {
			var swipeI int
			var swipeJ int
			for _, instr := range games[i].History {
				if tallylogic.GetInstructionType(instr) == tallylogic.InstructionTypeSwipe {
					swipeI++
				}
			}
			for _, instr := range games[j].History {
				if tallylogic.GetInstructionType(instr) == tallylogic.InstructionTypeSwipe {
					swipeJ++
				}
			}
			return swipeI < swipeJ
		})
	case model.HintPreference_HINT_PREFERENCE_FIRST_COMBINE:
		sort.Slice(games, func(i, j int) bool {
			var combineIndexI int
			var combineIndexJ int
			// TODO
			for k, instr := range games[i].History {
				if tallylogic.GetInstructionType(instr) == tallylogic.InstructionTypeCombinePath {
					combineIndexI = k
					break
				}
			}
			for k, instr := range games[j].History {
				if tallylogic.GetInstructionType(instr) == tallylogic.InstructionTypeCombinePath {
					combineIndexJ = k
					break
				}
			}
			if combineIndexI == combineIndexJ {
				return len(games[i].History) < len(games[j].History)
			}
			return combineIndexI < combineIndexJ
		})
	case model.HintPreference_HINT_PREFERENCE_MINIMUM_SWIPES_TO_COMBINE_RATIO:
		sort.Slice(games, func(i, j int) bool {
			var swipeI float32
			var swipeJ float32
			var combineI float32
			var combineJ float32
			for _, instr := range games[i].History {
				t := tallylogic.GetInstructionType(instr)
				if t == tallylogic.InstructionTypeSwipe {
					swipeI++
				} else if t == tallylogic.InstructionTypeCombinePath {
					combineI++
				}
			}
			for _, instr := range games[j].History {
				t := tallylogic.GetInstructionType(instr)
				if t == tallylogic.InstructionTypeSwipe {
					swipeJ++
				} else if t == tallylogic.InstructionTypeCombinePath {
					combineJ++
				}
			}
			ratioI := swipeI / combineI
			ratioJ := swipeJ / combineI
			return ratioI < ratioJ
		})
	}
	bestInstructions := games[0].History
	var length int = len(bestInstructions)
	if req.Msg.MaxLength > 0 && req.Msg.MaxLength < uint32(length) {
		length = int(req.Msg.MaxLength)
	}
	response.Instructions = make([]*model.Instruction, length)
	for i := 0; i < length; i++ {
		h := bestInstructions[i]
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
						Index: intsTouInt32s(t),
					},
				},
			}
		}
	}
	res := connect.NewResponse(response)
	return res, nil
}
