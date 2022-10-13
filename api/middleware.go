package api

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/cip8/autoname"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/runar-rkmedia/go-common/logger"
	"github.com/runar-rkmedia/gotally/tallylogic"
	"github.com/runar-rkmedia/gotally/types"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func CORSHandler() MiddleWare {
	return func(next http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
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
			next.ServeHTTP(w, r)
		}
	}
}

type MiddleWare func(next http.Handler) http.HandlerFunc

func RequestIDHandler(generator func() string) MiddleWare {
	return func(next http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Set request-id
			reqID := r.Header.Get("X-Request-ID")
			if reqID == "" {
				reqID = generator()
				w.Header().Set("X-Request-ID", reqID)
				r.Header.Set("X-Request-ID", reqID)
			}

			next.ServeHTTP(w, r)
		}
	}
}

type LogResponseWriter struct {
	http.ResponseWriter
	statusCode   int
	written      bool
	collectsBody bool
	responseBody []byte
}

func (lw *LogResponseWriter) Write(b []byte) (int, error) {
	lw.written = true
	if lw.collectsBody {
		lw.responseBody = b

	}
	return lw.ResponseWriter.Write(b)
}

// Gets the body if available.
// Handles encodings:
// gzip
// no encoding
//
// attempts to jsonify the response
func (lw *LogResponseWriter) GetBody(l logger.AppLogger) (decodedBytes []byte, unmarshalled map[string]any, err error) {
	if len(lw.responseBody) == 0 {
		return
	}
	contentEncoding := lw.Header().Get("Content-Encoding")
	contentType := lw.Header().Get("Content-Type")
	if contentType == "application/proto" {
		// unmarshalling the proto-message is not really fun
		// There is some information about it here: https://stackoverflow.com/questions/41348512/protobuf-unmarshal-unknown-message
		// At this time, I don't want to implement it.
		return
	}
	switch contentEncoding {
	case "gzip":
		if r, err := gzip.NewReader(bytes.NewReader(lw.responseBody)); err == nil {
			b, err := io.ReadAll(r)
			if err != nil {
				l.Warn().Err(err).Msg("failed to read the decoded response")
				return nil, nil, err
			}
			decodedBytes = b
		} else {
			l.Warn().Err(err).Msg("failed to decode the response")
			return nil, nil, err
		}
	case "":
		decodedBytes = lw.responseBody
	default:
		l.Warn().Msg("unhandled contentEncoding")
		return
	}
	if len(decodedBytes) == 0 {
		return
	}
	if strings.Contains(contentType, "application/json") {
		err := json.Unmarshal(decodedBytes, &unmarshalled)
		if err != nil {
			l.Warn().Err(err).Msg("failed to unmarshal the decoded response")
		}
	}

	return
}
func (lw *LogResponseWriter) WriteHeader(code int) {
	lw.statusCode = code
	lw.ResponseWriter.WriteHeader(code)
}
func Logger(l logger.AppLogger) MiddleWare {
	debug := l.HasDebug()
	return func(next http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {

			reqID := r.Header.Get("X-Request-ID")

			if reqID == "" {
				l.Warn().Msg("No request-id. Is the ordering of middleware correct?")
			}
			l := logger.With(baseLogger.With().
				Str("reqId", reqID).
				Str("method", r.Method).
				Str("content-type", r.Header.Get("Content-Type")).
				Str("path", r.URL.Path).
				Logger())
			lw := &LogResponseWriter{ResponseWriter: w, collectsBody: l.HasDebug()}
			w = lw
			r = r.WithContext(context.WithValue(r.Context(), ContextKeyLogger, l))
			if debug {
				l.Debug().Msg("Incoming request")
			}

			next.ServeHTTP(w, r)
			if lw.statusCode >= 500 {
				span := trace.SpanFromContext(r.Context())
				if lw.collectsBody && len(lw.responseBody) > 0 {
					span.SetAttributes(
						attribute.String("error.details", string(lw.responseBody)),
					)
					l.Error().
						Int("statusCode", lw.statusCode).
						Str("errorDetails", string(lw.responseBody)).
						Msg("Outgoing response")
				} else {
					l.Error().
						Int("statusCode", lw.statusCode).
						Str("responseBody", string(lw.responseBody)).
						Msg("Outgoing response")
				}
			} else if lw.statusCode >= 400 {
				// var result []byte
				// var resultJson map[string]any
				contentEncoding := lw.Header().Get("Content-Encoding")
				result, resultJson, err := lw.GetBody(l)
				l := l.Error().
					Int("status-code", lw.statusCode).
					Str("content-encoding", contentEncoding)
				if err != nil {
					l = l.Err(err)
				}
				if resultJson != nil {
					l = l.Interface("bodyDecodedJson", resultJson)
				} else if result != nil {
					l = l.Bytes("bodyDecoded", result)
				}
				l.Msg("Outgoing response")

			}
		}
	}
}
func ContextGetLogger(ctx context.Context) logger.AppLogger {
	v := ctx.Value(ContextKeyLogger)
	if v == nil {
		l := logger.GetLogger("unspecified http-logger")
		l.Warn().Stack().Err(fmt.Errorf("no logger available in context")).Msg("no logger available in context. using fallback-logger")
		return l
	}

	return v.(logger.AppLogger)
}
func Recovery(withStackTrace bool, l logger.AppLogger) MiddleWare {
	return func(next http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {

			defer func() {

				if panickedErr := recover(); panickedErr != nil {
					span := trace.SpanFromContext(r.Context())
					w.WriteHeader(http.StatusBadGateway)
					type s struct {
						Code    connect.Code `json:"code"`
						Message string       `json:"message"`
						Details []any        `json:"details,omitempty"`
						Stack   string       `json:"stack"`
					}
					// l := ContextGetLogger()
					cess := s{
						Code:    connect.CodeInternal,
						Message: "internal error",
						Details: []any{},
					}
					if withStackTrace {
						stack := make([]byte, 4096)
						j := runtime.Stack(stack, false)
						cess.Stack = string(stack[:j])
						span.SetAttributes(
							attribute.String("panic.stack", cess.Stack),
						)
					}

					l.Error().
						Interface("panickedErr", panickedErr).
						Interface("error", cess).
						Msg("panic")

					if panickedErr != nil {
						cess.Details = append(cess.Details, panickedErr)
					}

					b, err := json.Marshal(cess)
					if err != nil {
						b = ([]byte("internal failure"))
						if err != nil {
							l.Error().
								Err(err).
								Msg("Failed to marshal error to user after panicking.")
						}
					}
					_, err = w.Write(b)
					if err != nil {
						l.Error().
							Err(err).
							Msg("Failed to write response to client after failure. Client disconnennected?")

					}

				}
			}()
			next.ServeHTTP(w, r)

		}
	}
}

type AuthorizationOptions struct {
	Debug           bool
	SessionLifeTime time.Duration
}

var caser = cases.Title(language.English)

func GenerateNameForUser() string {
	name := autoname.Generate(" ")
	return caser.String(name)

}

func Authorization(store SessionStore, options AuthorizationOptions) MiddleWare {
	if options.SessionLifeTime == 0 {
		// Its a game, I don't see a reason to use short lifetimes
		// There is no sensitive information.
		options.SessionLifeTime = time.Hour * 24 * 30 * 24

	}
	return func(next http.Handler) http.HandlerFunc {
		// idgenerator := mustCreateUUidgenerator()
		return func(w http.ResponseWriter, r *http.Request) {
			l := ContextGetLogger(r.Context())

			sessionID := getSessionIDFromRequest(r)
			now := time.Now()
			if sessionID == "" || len(sessionID) != tokenLength {
				sessionID = gonanoid.Must()

			}

			// Get the session-state for the user
			var userState *UserState

			// if len(sessionID) == tokenLength {
			// 	userState = Store.GetUserState(sessionID)
			// }
			if userState == nil {
				us, err := store.GetUserBySessionID(r.Context(), types.GetUserPayload{ID: sessionID})
				if err != nil {
					l.Error().Str("sessionID", sessionID).Err(err).Msg("failed to lookup user by session-id")
					_, err := w.Write([]byte("failed to lookup user by session-id"))
					if err != nil {
						l.Warn().Str("sessionID", sessionID).Err(err).Msg("failed to write to ResponseWriter. Did the user hang up?")
					}
					w.WriteHeader(500)
					return
				}
				if us != nil {

					userState = &UserState{
						SessionID: us.Session.ID,
						UserName:  us.UserName,
						UserID:    us.UserID,
					}
					if us.ActiveGame != nil {
						g, err := tallylogic.RestoreGame(us.ActiveGame)
						if err != nil {
							w.WriteHeader(500)
							return
						}
						userState.Game = g
					} else {
						l.Error().Interface("userSession", us).Msg("user does not have an active game")

						w.WriteHeader(500)
						return
					}
				}
			}

			if userState == nil {
				// sessionID = idgenerator()
				if us, err := NewUserState(tallylogic.GameModeDefault, &tallylogic.ChallengeGames[0], sessionID); err != nil {
					l.Fatal().Err(err).Msg("Failed in NewUserState")
				} else {
					userState = &us

					payload := types.CreateUserSessionPayload{
						SessionID:    sessionID,
						InvalidAfter: now.Add(options.SessionLifeTime),
						Username:     userState.UserName,
						Game:         toTypeGame(userState.Game, ""),
					}
					createdUserSession, err := store.CreateUserSession(r.Context(), payload)
					if err != nil {
						l.Error().Err(err).Msg("failed in CreateUserSession")
						w.WriteHeader(500)
						return
					}
					{
						// sanity-checks. All of these null-checks should have ben handled in store.CreateUserSession
						// This is just a loud alert to help development
						// If any of these errors do show up, there is probably something really bad going on, and it is better to stop the serice.
						if createdUserSession == nil {
							l.Fatal().Msg("expected createdUserSession to be set")
							return
						}
						if createdUserSession.ActiveGame == nil {
							l.Fatal().Msg("expected createdUserSession.ActiveGame to be set")
							return
						}
						if createdUserSession.ActiveGame.Rules.ID == "" {
							l.Fatal().Msg("expected createdUserSession.ActiveGame.Rules.ID to be set")
							return
						}
						if createdUserSession.User.ID == "" {
							l.Fatal().Msg("expected createdUserSession.User.ID to be set")
							return
						}
					}
					userState.SessionID = createdUserSession.Session.ID
					sessionID = createdUserSession.Session.ID
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
				isSecure, _ := isSecureRequest(r)
				if isSecure {
					cookie.Secure = true
					cookie.SameSite = http.SameSiteNoneMode
				} else {
					cookie.Secure = false
					if setHttpAuthHeader {
						w.Header().Set(tokenHeader, sessionID)
						l.Warn().Msg("using authorization-header")
					}
				}
				http.SetCookie(w, cookie)
			}
			r = r.WithContext(context.WithValue(r.Context(), ContextKeyUserState, userState))
			next.ServeHTTP(w, r)
		}
	}
}

func pipeline(handler http.Handler, middlewares ...MiddleWare) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}
