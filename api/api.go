package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/runar-rkmedia/go-common/logger"
	"github.com/runar-rkmedia/gotally/gen/proto/tally/v1/tallyv1connect"
	"github.com/runar-rkmedia/gotally/live_client/ex"
	"github.com/runar-rkmedia/gotally/tallylogic"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

var (
	port       = "8080"
	baseLogger logger.AppLogger
)

func getCookieWithValidation(req *http.Request, s string) (string, error) {
	cookie, err := req.Cookie(tokenHeader)
	if err != nil {
		return "", err
	}
	if cookie == nil {
		return "", fmt.Errorf("empty cookie???")
	}

	return cookie.Value, err
}

const (
	/// The tokens should all be of this length
	tokenLength int = 21
	// the context-key for UserState
	ContextKeyUserState ContextKey = "USER_STATE"
	cookieMaxTime       int        = 60000
	setHttpAuthHeader   bool       = true
	setHttpsAuthHeader  bool       = false
)

type ContextKey string

func getSessionIDFromRequest(req *http.Request) string {
	if cookieValue, err := getCookieWithValidation(req, tokenHeader); err == nil {
		return cookieValue
	}
	return req.Header.Get(tokenHeader)
}

func c(condition bool, iftrue, iffalse string) string {
	if condition {
		return iftrue
	}
	return iffalse
}

func isSecureRequest(r *http.Request) (bool, string) {
	proto := r.Header.Get("X-Forwarded-Proto")
	if proto == "" {
		origin := r.Header.Get("Origin")
		proto = strings.Split(origin, "://")[0]
	}

	return proto == "https", proto

}

func StartServer() {
	logger.InitLogger(logger.LogConfig{
		Level:      "debug",
		Format:     "human",
		WithCaller: true,
	})
	baseLogger = logger.GetLogger("base")

	debug := baseLogger.HasDebug()
	err := ex.ReadGeneratedBoardsFromDisk()
	if err != nil {
		baseLogger.Fatal().Err(err).Msg("failed to read generated files")
	}

	tally := NewTallyServer()
	mux := http.NewServeMux()
	path, handler := tallyv1connect.NewBoardServiceHandler(&tally)
	// path = "/api" + path
	mainHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isSecure, _ := isSecureRequest(r)
		// CORS
		w.Header().Set("Access-Control-Expose-Headers", "Date, X-Request-ID"+c(!isSecure, ", "+tokenHeader, ""))
		w.Header().Set("Access-Control-Allow-Headers", "content-type, "+tokenHeader)
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET")
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Set("Access-Control-Max-Age", "60")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		// Set request-id
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = tally.UidGenerator()
			w.Header().Set("X-Request-ID", reqID)
		}
		l := logger.With(baseLogger.With().
			Str("reqId", reqID).
			Logger())

		// Get a sessionID
		sessionID := getSessionIDFromRequest(r)
		fmt.Println("sesss", sessionID)

		// Get the session-state for the user
		var userState *UserState
		if len(sessionID) == tokenLength {
			userState = Store.GetUserState(sessionID)
		}
		if userState == nil {
			// The session is either not set or invalid / not found.
			// Create a new session and a new game
			sessionID = tally.UidGenerator()
			if debug {
				l.Debug().Msg("New user encountered")
			}
			if us, err := NewUserState(tallylogic.GameModeTemplate, &tallylogic.ChallengeGames[0], sessionID); err != nil {
				l.Fatal().Err(err).Msg("Failed in NewUserState")
			} else {
				userState = &us
				Store.SetUserState(userState)
			}
			// Set the cookie /user-session

			cookie := &http.Cookie{
				Name: tokenHeader,
				// TODO: when the server is behind a subpath (e.g.
				// exmaple.com/skiver/), the reverse-proxy in front may not return our
				// path, and we probably need to get it from the config
				Path:   "/",
				Value:  sessionID,
				MaxAge: cookieMaxTime,
				Secure: r.TLS != nil,
				// SameSite: http.SameSiteNoneMode,
				HttpOnly: true,
			}
			fmt.Println("sec", isSecure)
			if isSecure {
				cookie.Secure = true
				cookie.SameSite = http.SameSiteNoneMode
			} else {
				cookie.Secure = false
				fmt.Println("Not secure", setHttpAuthHeader)
				if setHttpAuthHeader {
					w.Header().Set(tokenHeader, sessionID)
					l.Warn().Msg("using authorization-header")
				}
			}
			http.SetCookie(w, cookie)
		}
		r = r.WithContext(context.WithValue(r.Context(), ContextKeyUserState, userState))

		handler.ServeHTTP(w, r)
	})
	mux.Handle("/", mainHandler)
	address := "localhost:" + port
	baseLogger.Info().Str("address", "http://"+address+path).Msg("Starting server")
	if err := http.ListenAndServe(
		"0.0.0.0:"+port,
		h2c.NewHandler(
			mux,
			&http2.Server{}),
	); err != nil {
		panic(err)
	}
}

type TallyServer struct {
	UidGenerator func() string
}

func NewTallyServer() TallyServer {
	return TallyServer{
		UidGenerator: mustCreateUUidgenerator(),
	}

}
