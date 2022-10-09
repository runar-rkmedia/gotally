package api

import (
	"fmt"
	"os"
	"strings"

	"github.com/carlmjohnson/versioninfo"
	_ "github.com/honeycombio/honeycomb-opentelemetry-go"
	"github.com/honeycombio/opentelemetry-go-contrib/launcher"
	"github.com/runar-rkmedia/go-common/logger"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.6.1"
)

func initializeOpenTelemetry(l logger.AppLogger) func() {
	if _, ok := os.LookupEnv("HONEYCOMB_API_KEY"); !ok {

		l.Info().Msg("Tracing via honeycomb is disabled")
		return func() {}
	}
	version := versioninfo.Version
	env := "production"
	if !strings.HasPrefix(version, "v") {
		env = "development"
		version = "v0.0.1-dev"
	}
	attr := []attribute.KeyValue{
		semconv.ServiceVersionKey.String(""),
		attribute.String("environment", env),
	}
	attributes := make(map[string]string, len(attr))
	for _, v := range attr {
		attributes[string(v.Key)] = v.Value.AsString()

	}

	fmt.Println(attributes)
	fmt.Println(attr)
	fmt.Println(version, versioninfo.Version)
	// use honeycomb distro to setup OpenTelemetry SDK
	name := os.Getenv("OTEL_SERVICE_NAME")
	baseLogger.Info().Str("servicename", name).Msg("Tracing via honeycomb is enabled")
	otelShutdown, err := launcher.ConfigureOpenTelemetry(
		launcher.WithResourceAttributes(attributes),
	)
	if err != nil {
		l.Fatal().Err(err).Msg("error setting up OTEL SDK")
	}
	return otelShutdown
}
