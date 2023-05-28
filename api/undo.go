package api

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/runar-rkmedia/gotally/dev"
	model "github.com/runar-rkmedia/gotally/gen/proto/tally/v1"
	"github.com/runar-rkmedia/gotally/tallylogic"
	"github.com/runar-rkmedia/gotally/types"
)

func (s *TallyServer) Undo(
	ctx context.Context,
	req *connect.Request[model.UndoRequest],
) (*connect.Response[model.UndoResponse], error) {

	session := ContextGetUserState(ctx)
	if session.Game.Moves() == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Game is already at the beginning, cannot undo"))
	}
	g, err := s.storage.GetOriginalGame(ctx, types.GetOriginalGamePayload{GameID: session.Game.ID})
	if err != nil {
		s.l.Error().
			Err(err).
			Interface("game", g).
			Msg("Failed to get original game before undo")

		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if err := g.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	g2, err := tallylogic.RestoreGame(&g)
	dev.Println("Original game",
		g.StartingCells,
		g2.Rules.StartingCells,
		session.Game.Rules.StartingCells,
		"OPTSeeed", g2.Rules.Options.PrintSeed(),
		session.Game.PrintWithStats(),
		g2.Print())
	session.Game.Rules.Options.Seed = g2.Rules.Options.Seed
	session.Game.Rules.Options.State = g2.Rules.Options.State
	session.Game.ReplaceBasedOn(g2)
	// We copy the game, to rollback the in-memory cache if anything goes wrong
	dev.Println("UNDODOD", session.Game.Rules.Options.PrintSeed(), session.Game.PrintWithStats())
	gameCopy := session.Game.Copy()
	err = session.Game.Undo()
	if err != nil {
		s.l.Error().
			Err(err).
			Interface("game", g).
			Msg("Failed to undo board")

		return nil, connect.NewError(connect.CodeInternal, err)
	}
	response := &model.UndoResponse{
		Board: toModalBoard(&session.Game),
		Moves: int64(session.Game.Moves()),
		Score: session.Score(),
	}
	seed, state := session.Game.Seed()
	payload := types.UpdateGamePayload{
		GameID:    session.Game.ID,
		Moves:     session.Game.Moves(),
		Score:     uint64(gameCopy.Score()),
		State:     state,
		Seed:      seed,
		Cells:     session.Cells(),
		History:   session.Game.History.Bytes(),
		PlayState: types.PlayStateCurrent,
	}
	err = s.storage.UpdateGame(ctx, payload)
	if err != nil {
		s.l.Error().
			Err(err).
			Interface("payload", payload).
			Msg("failed to save the board to storage during undo-operation")
		// rollback the game in memory
		session.Game = gameCopy
		return nil, fmt.Errorf("intarnal failure while saving the board during undo: %w", err)
	}
	res := connect.NewResponse(response)
	return res, nil
}
