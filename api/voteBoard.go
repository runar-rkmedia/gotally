package api

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	model "github.com/runar-rkmedia/gotally/gen/proto/tally/v1"
)

func (s *TallyServer) VoteBoard(
	ctx context.Context,
	req *connect.Request[model.VoteBoardRequest],
) (*connect.Response[model.VoteBoardResponse], error) {
	session := ContextGetUserState(ctx)
	boardId := session.Game.BoardID()
	if boardId == "" {
		cerr := createError(connect.CodeInvalidArgument, fmt.Errorf("Cannot vote for this board"))
		return nil, cerr.ToConnectError()
	}
	if req.Msg.FunVote == model.Vote_VOTE_UNSPECIFIED {
		cerr := createError(connect.CodeInvalidArgument, fmt.Errorf("FunVote is required"))
		return nil, cerr.ToConnectError()
	}
	if req.Msg.UserName != "" && req.Msg.UserName != session.UserName {
		if len(req.Msg.UserName) > 200 {
			cerr := createError(connect.CodeInvalidArgument, fmt.Errorf("userName is too long"))
			return nil, cerr.ToConnectError()
		}
		session.UserName = req.Msg.UserName
		Store.SetUserState(session)
	}
	vote, err := s.storage.VoteForBoard(boardId, session.SessionID, req.Msg.UserName, int(req.Msg.FunVote))
	if err != nil {
		cerr := createError(connect.CodeInvalidArgument, fmt.Errorf("failure during vote: %w", err))
		return nil, cerr.ToConnectError()
	}
	session.SelfVotes[vote.ID] = vote.FunVote
	response := &model.VoteBoardResponse{
		Id:        vote.ID,
		SessionId: vote.User,
		UserName:  vote.UserName,
		FunVote:   model.Vote(vote.FunVote),
	}
	res := connect.NewResponse(response)
	return res, nil
}
