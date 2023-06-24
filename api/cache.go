package api

import (
	"fmt"
	"sync"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/runar-rkmedia/gotally/tallylogic"
	"github.com/runar-rkmedia/gotally/types"
)

type stupidcache struct {
	games map[string]*UserState
	sync.RWMutex
}

func (c *stupidcache) GetUserState(s string) *UserState {
	c.RLock()
	defer c.RUnlock()
	return c.games[s]
}
func (c *stupidcache) SetUserState(game *UserState) {
	c.Lock()
	defer c.Unlock()
	if game.SessionID == "" {
		panic("no SessionID in SetUserState")
	}
	c.games[game.SessionID] = game
}

var (
	// deprecated
	Store stupidcache = stupidcache{
		games: map[string]*UserState{},
	}
)

type UserState struct {
	SessionID string
	UserName  string
	UserID    string
	// Current game being played
	tallylogic.Game
	PlayState types.PlayState
}

func NewUserState(mode tallylogic.GameMode, template *tallylogic.GameTemplate, sessionID string, options ...tallylogic.NewGameOptions) (UserState, error) {
	m := UserState{
		SessionID: sessionID,
		UserName:  GenerateNameForUser(),
		UserID:    gonanoid.Must(),
	}
	if m.SessionID == "" {
		return m, fmt.Errorf("SessionID not set")
	}
	game, err := tallylogic.NewGame(mode, template, options...)
	if err != nil {
		return m, err
	}
	game.Rules.ID = gonanoid.Must()
	m.Game = game
	return m, nil
}
