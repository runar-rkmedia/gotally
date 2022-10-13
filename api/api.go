package api

import (
	"fmt"
	"net/http"
	"net/http/pprof"
	"strings"

	"github.com/carlmjohnson/versioninfo"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/runar-rkmedia/go-common/logger"
	"github.com/runar-rkmedia/gotally/gen/proto/tally/v1/tallyv1connect"
	"github.com/runar-rkmedia/gotally/generated"
	web "github.com/runar-rkmedia/gotally/static"
	"github.com/runar-rkmedia/gotally/storage"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
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
	ContextKeyLogger    ContextKey = "LOGGER"
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

	f := initializeOpenTelemetry(baseLogger)
	defer f()
	debug := baseLogger.HasDebug()
	err := generated.ReadGeneratedBoardsFromDisk()
	if err != nil {
		baseLogger.Fatal().Err(err).Msg("failed to read generated files")
	}

	tally := NewTallyServer(logger.GetLogger("tally-server"))
	mux := http.NewServeMux()
	path, connectHandler := tallyv1connect.NewBoardServiceHandler(&tally)
	// http://192.168.10.101:8080/tally.v1.BoardService/GetSession
	// han := CORSHandler()(
	// 	RequestIDHandler(mustCreateUUidgenerator())(mainHandler),
	// )
	pipe := []MiddleWare{
		Recovery(debug, logger.GetLogger("recovery")),
		CORSHandler(),
		RequestIDHandler(mustCreateUUidgenerator()),
		Logger(logger.GetLogger("request")),
		Authorization(tally.storage, AuthorizationOptions{}),
	}
	// Register metrics
	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle("/metrics/", promhttp.Handler())
	// Register pprof handlers
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)

	mux.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	mux.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	mux.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	mux.Handle("/debug/pprof/block", pprof.Handler("block"))
	han := pipeline(connectHandler, pipe...)
	han = otelhttp.NewHandler(han, "gotally-api")
	mux.Handle(path, han)
	mux.Handle("/", http.StripPrefix("/", web.StaticWebHandler()))
	// mux.Handle("/", web.StaticWebHandler())
	address := "localhost:" + port
	baseLogger.Info().
		Str("address", "http://"+address+path).
		Str("version", versioninfo.Version).
		Bool("dirtyBuild", versioninfo.DirtyBuild).
		Str("revision", versioninfo.Revision).
		Msg("Starting server")
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
	storage      PersistantStorage
	l            logger.AppLogger
}

func NewTallyServer(l logger.AppLogger) TallyServer {
	db, err := storage.NewPersistantStorage(logger.GetLogger("database"), "")
	// db, err := database.NewDatabase(logger.GetLoggerWithLevel("db", "info"), "")
	if err != nil {
		baseLogger.Fatal().Err(err).Msg("failed to initialize database")
	}
	ts := TallyServer{
		l:            l,
		UidGenerator: mustCreateUUidgenerator(),
		storage:      db,
	}

	return ts

}
