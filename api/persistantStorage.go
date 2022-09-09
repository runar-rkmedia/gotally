// Code generated by ifacemaker; DO NOT EDIT.

package api

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/runar-rkmedia/gotally/database"
)

// PersistantStorage ...
type PersistantStorage interface {
	Deploy() error
	VoteForBoard(id, user, userName string, funVote int) (*database.Vote, error)
	GetAllVotes() (map[string]database.Vote, error)
	GetVotesForBoardByUserName(userName string) (map[string]database.Vote, error)
}
