package zerolog

import (
	"context"

	"github.com/agoda-com/opentelemetry-go/otelzerolog"
	"github.com/agoda-com/opentelemetry-logs-go/exporters/otlp/otlplogs"
	"github.com/agoda-com/opentelemetry-logs-go/exporters/otlp/otlplogs/otlplogshttp"
	sdklogs "github.com/agoda-com/opentelemetry-logs-go/sdk/logs"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

func init() {
	InitDefaultLogger()
}

func InitDefaultLogger() {
	ctx := context.Background()
	exporter, _ := otlplogs.NewExporter(ctx, otlplogs.WithClient(otlplogshttp.NewClient()))
	loggerProvider := sdklogs.NewLoggerProvider(
		sdklogs.WithBatcher(
			exporter,
			sdklogs.WithMaxExportBatchSize(512),
		),
	)
	hook := otelzerolog.NewHook(loggerProvider)
	loggerVal := log.With().Caller().Logger()
	loggerVal = loggerVal.Hook(hook)
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.DefaultContextLogger = &loggerVal
}
