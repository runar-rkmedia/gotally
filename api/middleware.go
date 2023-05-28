package api

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/cip8/autoname"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/runar-rkmedia/go-common/logger"
	"github.com/runar-rkmedia/gotally/tallylogic"
	"github.com/runar-rkmedia/gotally/types"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
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
			w.Header().Set("Access-Control-Allow-Headers", "DEV_GAME_OPTIONS, DEV_USERNAME,  content-type, "+tokenHeader)
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
			l := logger.With(l.With().
				Str("reqId", reqID).
				Str("method", r.Method).
				Str("content-type", r.Header.Get("Content-Type")).
				Str("path", r.URL.Path).
				Logger())
			lw := &LogResponseWriter{ResponseWriter: w, collectsBody: l.HasDebug()}
			w = lw
			r = r.WithContext(context.WithValue(r.Context(), ContextKeyLogger, l))
			if debug {
				l.Debug().
					Interface("headers", r.Header).
					Msg("Incoming request")
			}
			if lw.statusCode == 0 {
				lw.statusCode = 200
			}

			next.ServeHTTP(w, r)
			metricHttpCalls.With(prometheus.Labels{
				"code":   strconv.FormatInt(int64(lw.statusCode), 10),
				"method": r.Method,
				"path":   r.URL.Path,
			}).Inc()
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

type s struct {
	Code     connect.Code `json:"code"`
	Message  string       `json:"message"`
	Details  []any        `json:"details,omitempty"`
	Stack    string       `json:"stack"`
	File     string       `json:"file,omitempty"`
	Line     int          `json:"line,omitempty"`
	Function string       `json:"function,omitempty"`
}

func (cess *s) createStack() {

	stack := make([]byte, 4096)
	j := runtime.Stack(stack, false)

	rewinds, s := cleanStackTrace(string(stack[:j]))
	cess.Stack = s
	cess.Message += "file"
	pc, file, line, ok := runtime.Caller(rewinds + 2)
	if ok {
		cess.File = file
		cess.Line = line
		f := runtime.FuncForPC(pc)
		if f != nil {
			cess.Function = shortFuncName(f)
		}
	}
	if logger.IsInteractiveTTY() {
		// Just to simplify reading the stacktrace while developing
		println(fmt.Sprintf("Error in %s %s:%d\n%s", cess.Function, cess.File, cess.Line, cess.Stack))
	}
}

var (
	stackRegex = regexp.MustCompile(`[^\s]*(\/gotally\/)`)
)

func cleanStackTrace(stack string) (rewinds int, clean string) {
	stack = strings.ReplaceAll(stack, "github.com/runar-rkmedia/", "./")
	stack = stackRegex.ReplaceAllString(stack, ".$1")
	split := strings.Split(stack, "\n")
	for i := 0; i < len(split); i++ {
		if strings.Contains(split[i], "runtime/panic.go") {
			i++
			return rewinds / 2, strings.Join(split[i:], "\n")
		}
		rewinds++
	}
	return 0, clean
}

/* "FuncName" or "Receiver.MethodName" */
func shortFuncName(f *runtime.Func) string {
	// f.Name() is like one of these:
	// - "github.com/palantir/shield/package.FuncName"
	// - "github.com/palantir/shield/package.Receiver.MethodName"
	// - "github.com/palantir/shield/package.(*PtrReceiver).MethodName"
	longName := f.Name()

	withoutPath := longName[strings.LastIndex(longName, "/")+1:]
	withoutPackage := withoutPath[strings.Index(withoutPath, ".")+1:]

	shortName := withoutPackage
	shortName = strings.Replace(shortName, "(", "", 1)
	shortName = strings.Replace(shortName, "*", "", 1)
	shortName = strings.Replace(shortName, ")", "", 1)

	return shortName
}
func Recovery(withStackTrace bool, l logger.AppLogger) MiddleWare {
	return func(next http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {

			defer func() {

				if panickedErr := recover(); panickedErr != nil {
					var parsedErr error
					if err, ok := panickedErr.(error); ok {
						parsedErr = err
					} else if err, ok := panickedErr.(string); ok {
						parsedErr = errors.New(err)
					} else {
						parsedErr = fmt.Errorf("panic(%T): %#v", panickedErr, panickedErr)
					}
					span := trace.SpanFromContext(r.Context())
					w.WriteHeader(http.StatusBadGateway)
					// l := ContextGetLogger()
					cess := s{
						Code:    connect.CodeInternal,
						Message: "internal error",
						Details: []any{},
					}
					if withStackTrace {
						cess.createStack()
						span.SetStatus(codes.Error, "panic")
						span.RecordError(parsedErr, trace.WithStackTrace(true))
					}

					l.Error().
						Err(parsedErr).
						Interface("panickedErr", panickedErr).
						Interface("errorC", cess).
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
	// If set will allow the client to set some options on each request that normally is not allowed.
	// Mostly used for e2e-testing, where the client wants to be in control over the randomization and such
	AllowDevelopmentFlags bool
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
		options.SessionLifeTime = time.Duration(sessionMaxTime) * time.Second

	}
	return func(next http.Handler) http.HandlerFunc {
		// idgenerator := mustCreateUUidgenerator()
		return func(w http.ResponseWriter, r *http.Request) {
			l := ContextGetLogger(r.Context())
			ctx := r.Context()
			span := trace.SpanFromContext(ctx)

			sessionID, tokenSource := getSessionIDFromRequest(r)
			now := time.Now()

			// Get the session-state for the user
			var userState *UserState

			// if len(sessionID) == tokenLength {
			// 	userState = Store.GetUserState(sessionID)
			// }
			if sessionID != "" {
				us, err := store.GetUserBySessionID(ctx, types.GetUserPayload{ID: sessionID})
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
					if time.Now().After(us.InvalidAfter) {
						l.Warn().
							Time("InvalidAfter", us.InvalidAfter).
							Time("now", time.Now()).
							Msg("session is invalid")
						us = nil
						// TODO: handle this case
						w.WriteHeader(http.StatusUnauthorized)
						return
					}
				}
				if us != nil {

					userState = &UserState{
						SessionID: us.Session.ID,
						UserName:  us.UserName,
						UserID:    us.UserID,
					}
					if us.ActiveGame != nil {
						l.Debug().Msg("Restoring game")
						g, err := tallylogic.RestoreGame(us.ActiveGame)
						if err != nil {
							l.Error().
								Err(err).
								Str("sessionID", us.Session.ID).
								Str("userID", us.Session.UserID).
								Interface("game", us.ActiveGame).
								Msg("Failed to restore game for user in middleware")
							w.WriteHeader(500)
							return
						}
						userState.Game = g
					} else {
						l.Error().Interface("userSession", us).Msg("user does not have an active game")

						l.Error().Msg("User has no active game")
						w.WriteHeader(500)
						return
					}
				}
			}

			if userState == nil {
				if sessionID == "" {
					tokenSource = "generated_empty"
					sessionID = gonanoid.Must()
				}
				if len(sessionID) != tokenLength {
					l.Warn().
						Str("sessionID", sessionID).
						Int("wantedLength", tokenLength).
						Int("gotLength", len(sessionID)).
						Msg("user-provided session was ignored becuase of wrong length")
					sessionID = gonanoid.Must()
					tokenSource += "_generated_invalid_length"
				}
				// sessionID = idgenerator()
				var gameOptions tallylogic.NewGameOptions

				gameMode := tallylogic.GameModeTutorial
				template := &tallylogic.TutorialGames[0]
				if options.AllowDevelopmentFlags {
					// This is only for testing-purposes, and should only be done while running locally
					// For instance can playwright set properties here to ensure consistant options,
					// like seeding the randomizer used.
					// It expect a base64-encoded json in the header
					if o := DevGetGameOptions(l, options, r); o != nil {
						gameOptions = *o
						gameMode = tallylogic.GameModeRandom
						template = nil
					}
				}
				if us, err := NewUserState(gameMode, template, sessionID, gameOptions); err != nil {
					l.Fatal().Err(err).Msg("Failed in NewUserState")
				} else {
					userState = &us
					if l.HasDebug() {
						l.Debug().
							Interface("userstate", userState).
							Msg("userstate created")
					}
					if options.AllowDevelopmentFlags {
						if u := r.Header.Get("DEV_USERNAME"); u != "" {
							// l.Warn().Msg("got username")
							userState.UserName = u
						}
					}

					payload := types.CreateUserSessionPayload{
						UserID:       userState.UserID,
						SessionID:    sessionID,
						InvalidAfter: now.Add(options.SessionLifeTime),
						Username:     userState.UserName,
						Game:         toTypeGame(userState.Game, ""),
					}
					if template != nil {
						payload.TemplateID = template.ID

					}
					err := payload.Validate()
					if err != nil {
						l.Error().Err(err).
							Interface("payload", payload).
							Msg("payload-validation failed for CreateUserSession")
						w.WriteHeader(500)
						return
					}
					createdUserSession, err := store.CreateUserSession(r.Context(), payload)
					if err != nil {
						l.Error().Err(err).
							Interface("payload", payload).
							Msg("failed in CreateUserSession")
						w.WriteHeader(500)
						return
					}
					l.Info().
						Str("tokenSource", string(tokenSource)).
						Str("userID", createdUserSession.UserID).Msg("A new user was created")
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
				// TODO: also recreate a session with a new cookie if it expires soon (like in the next 200 days)

				cookie := &http.Cookie{
					Name: tokenHeader,
					// TODO: when the server is behind a subpath (e.g.
					// exmaple.com/skiver/), the reverse-proxy in front may not return our
					// path, and we probably need to get it from the config
					Path:   "/",
					Value:  sessionID,
					MaxAge: sessionMaxTime,
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
			span.SetAttributes(
				semconv.EnduserIDKey.String(userState.UserID),
			)
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
