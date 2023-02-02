package api

import (
	"context"

	"github.com/runar-rkmedia/gotally/types"
)

// PersistantStorage ...
type PersistantStorage interface {
	// Deploy() error
	// VoteForBoard(id, user, userName string, funVote int) (*types.Vote, error)
	// GetAllVotes() (map[string]types.Vote, error)
	// GetVotesForBoardByUserName(userName string) (map[string]types.Vote, error)
	SessionStore
}

type SessionStore interface {
	CreateUserSession(ctx context.Context, payload types.CreateUserSessionPayload) (*types.SessionUser, error)
	GetUserBySessionID(ctx context.Context, payload types.GetUserPayload) (*types.SessionUser, error)
	CombinePath(ctx context.Context, payload types.CombinePathPayload) error
	SwipeBoard(ctx context.Context, payload types.SwipePayload) error
	NewGameForUser(ctx context.Context, payload types.NewGamePayload) (types.Game, error)
	RestartGame(ctx context.Context, payload types.RestartGamePayload) (types.Game, error)
	Stats(ctx context.Context) (sess *types.Statistics, err error)
	// TGD, Subject to change
	Dump(ctx context.Context) (types.Dump, error)
}
