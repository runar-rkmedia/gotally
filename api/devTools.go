package api

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/runar-rkmedia/go-common/logger"
	"github.com/runar-rkmedia/gotally/tallylogic"
)

func DevGetGameOptions(l logger.AppLogger, options AuthorizationOptions, r *http.Request) *tallylogic.NewGameOptions {

	if !options.AllowDevelopmentFlags {
		return nil
	}
	// This is only for testing-purposes, and should only be done while running locally
	// For instance can playwright set properties here to ensure consistant options,
	// like seeding the randomizer used.
	// It expect a base64-encoded json in the header
	if o := r.Header.Get("DEV_GAME_OPTIONS"); o != "" {
		b, err := base64.StdEncoding.DecodeString(o)
		if err != nil {
			l.Warn().Err(err).Str("base64-header", o).Msg("user attempted to set game-options via headers, but the base64-decoding failed")
		} else {
			var gameOptions tallylogic.NewGameOptions
			err := json.Unmarshal(b, &gameOptions)
			if err != nil {
				l.Warn().Err(err).Str("base64-header", o).Msg("user attempted to set game-options via headers, but the unmarshalling failed")
				return nil
			}
			l.Debug().Err(err).Interface("options", gameOptions).Msg("game-options set via headers")
			return &gameOptions

		}
	}
	return nil
}
