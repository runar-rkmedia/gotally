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
	// queries

	// Returns various statistics about the system
	Stats(ctx context.Context) (sess *types.Statistics, err error)
	// Returns a User, from their session-id
	GetUserBySessionID(ctx context.Context, payload types.GetUserPayload) (*types.SessionUser, error)
	// TGD, Subject to change
	Dump(ctx context.Context) (types.Dump, error)
	// mutations

	// Creates a new Session for a user
	CreateUserSession(ctx context.Context, payload types.CreateUserSessionPayload) (*types.SessionUser, error)
	// Game-mechanic for combing cells from a path
	CombinePath(ctx context.Context, payload types.CombinePathPayload) error
	// Game-mechanic for swiping the board.
	SwipeBoard(ctx context.Context, payload types.SwipePayload) error
	// Creates a new game for the user
	NewGameForUser(ctx context.Context, payload types.NewGamePayload) (types.Game, error)
	// Restarts the current active game
	RestartGame(ctx context.Context, payload types.RestartGamePayload) (types.Game, error)
}
