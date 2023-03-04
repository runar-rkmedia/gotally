package api

import (
	"context"
	"fmt"
	"net/http"
	"net/http/pprof"
	"strings"
	"time"

	"github.com/bufbuild/connect-go"
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
	// 400 days see https://httpwg.org/http-extensions/draft-ietf-httpbis-rfc6265bis.html#name-the-max-age-attribute
	sessionMaxTime     int  = 60 * 60 * 24 * 400
	setHttpAuthHeader  bool = true
	setHttpsAuthHeader bool = false
	TokenSourceCookie       = "cookie"
	TokenSourceHeader       = "header"
)

type TokenSource string
type ContextKey string

func getSessionIDFromRequest(req *http.Request) (string, TokenSource) {
	if cookieValue, err := getCookieWithValidation(req, tokenHeader); err == nil {
		return cookieValue, TokenSourceCookie
	}
	if headerValue := req.Header.Get(tokenHeader); headerValue != "" {
		return headerValue, TokenSourceHeader
	}
	return "", ""
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

func NewLogInterceptor(l logger.AppLogger) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (res connect.AnyResponse, err error) {
			spec := req.Spec()
			l.Debug().
				Str("procedure", spec.Procedure).
				Interface("streamType", spec.StreamType).
				Bool("isClient", spec.IsClient).
				Interface("headers", req.Header()).
				Str("peer-addr", req.Peer().Addr).
				Msg("Incoming connect-request")
			return next(ctx, req)
		}

	}
}

func createApiHandler(withDebug bool, options ...TallyOptions) (tally TallyServer, path string, handler http.Handler) {
	tally = NewTallyServer(logger.GetLogger("tally-server"), options...)
	path, connectHandler := tallyv1connect.NewBoardServiceHandler(&tally,
		connect.WithInterceptors(NewLogInterceptor(logger.GetLogger("connect-log-interceptor"))),
		connect.WithRecover(func(ctx context.Context, s connect.Spec, h http.Header, err any) error {
			fmt.Println("\n\n\npanic in conenct-handler", err)
			tally.l.Error().Interface("err", err).Msg("Panic recovered (connect-handler)")

			return connect.NewError(connect.CodeInternal, fmt.Errorf("unhandled error recovered"))
		}))

	pipe := []MiddleWare{
		Recovery(withDebug, logger.GetLogger("recovery")),
		CORSHandler(),
		RequestIDHandler(mustCreateUUidgenerator()),
		Logger(logger.GetLogger("request")),
		Authorization(tally.storage, AuthorizationOptions{
			AllowDevelopmentFlags: tally.AllowDevelopmentFlags}),
	}
	return tally, path, pipeline(connectHandler, pipe...)
}

func StartServer(options TallyOptions) {
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

	_, path, han := createApiHandler(debug, options)
	// tally := NewTallyServer(logger.GetLogger("tally-server"))
	mux := http.NewServeMux()
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
	// Allow game-geneartion to occur
	// TODO: perhasp we should use build-frags to hide these instead.
	FeatureGameGeneration bool
	// If set will allow the client to set some options on each request that normally is not allowed.
	// Mostly used for e2e-testing, where the client wants to be in control over the randomization and such
	AllowDevelopmentFlags bool
}

type TallyOptions struct {
	// Connection-string for the database
	DatabaseDSN           string
	SkipStatsCollection   *bool
	FeatureGameGeneration *bool
	// If set will allow the client to set some options on each request that normally is not allowed.
	// Mostly used for e2e-testing, where the client wants to be in control over the randomization and such
	AllowDevelopmentFlags *bool
}

func NewTallyServer(l logger.AppLogger, options ...TallyOptions) TallyServer {
	opt := TallyOptions{}
	for _, o := range options {
		if o.DatabaseDSN != "" {
			opt.DatabaseDSN = o.DatabaseDSN
		}
		if o.SkipStatsCollection != nil {
			opt.SkipStatsCollection = o.SkipStatsCollection
		}
		if o.FeatureGameGeneration != nil {
			opt.FeatureGameGeneration = o.FeatureGameGeneration
		}
		if o.AllowDevelopmentFlags != nil {
			opt.AllowDevelopmentFlags = o.AllowDevelopmentFlags
		}
	}
	db, err := storage.NewSqliteStorage(logger.GetLogger("database"), opt.DatabaseDSN)
	// db, err := database.NewDatabase(logger.GetLoggerWithLevel("db", "info"), "")
	if err != nil {
		baseLogger.Fatal().Err(err).Msg("failed to initialize database")
	}
	ts := TallyServer{
		l:                     l,
		UidGenerator:          mustCreateUUidgenerator(),
		storage:               db,
		FeatureGameGeneration: isTrue(opt.FeatureGameGeneration),
		AllowDevelopmentFlags: isTrue(opt.AllowDevelopmentFlags),
	}
	if opt.SkipStatsCollection != nil && !*opt.SkipStatsCollection {
		go ts.collectStatsAtInterval(time.Second * 15)
	}
	if ts.AllowDevelopmentFlags || ts.FeatureGameGeneration {

		l.Warn().
			Bool("AllowDevelopmentFlags", ts.AllowDevelopmentFlags).
			Bool("FeatureGameGeneration", ts.FeatureGameGeneration).
			Msg("Starting tallyserver with options")
	}
	return ts
}

func isTrue(b *bool) bool {
	if b == nil {
		return false
	}
	return *b == true
}
