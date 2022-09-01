package api

import (
	"fmt"
	"sync"

	"github.com/jfyne/live"
	"github.com/runar-rkmedia/gotally/tallylogic"
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
	Store stupidcache = stupidcache{
		games: map[string]*UserState{},
	}
	CookieStore = live.NewCookieStore("cookie", []byte("eeeee"))
)

type UserState struct {
	UserName     string
	SelfVotes    map[string]int
	GamesStarted int
	// Current game being played
	tallylogic.Game
	// The current game that is being played, but as a snapshot for the start of the game
	// This is to be able to reset the game
	GameSnapshotAtStart tallylogic.Game
	SessionID           string
}

func NewUserState(mode tallylogic.GameMode, template *tallylogic.GameTemplate, sessionID string) (UserState, error) {
	m := UserState{
		SelfVotes:    map[string]int{},
		SessionID:    sessionID,
		GamesStarted: 1}
	if m.SessionID == "" {
		return m, fmt.Errorf("SessionID not set")
	}
	game, err := tallylogic.NewGame(mode, template)
	if err != nil {
		return m, err
	}
	m.Game = game
	m.GameSnapshotAtStart = game.Copy()
	return m, nil
}
