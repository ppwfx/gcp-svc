package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	monitoring "cloud.google.com/go/monitoring/apiv3"
	"github.com/armon/go-metrics"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/ppwfx/user-svc/pkg/business"
	"github.com/ppwfx/user-svc/pkg/communication"
	"github.com/ppwfx/user-svc/pkg/persistence"
	"github.com/ppwfx/user-svc/pkg/types"
	"github.com/ppwfx/user-svc/pkg/utils/loggingutil"
	"github.com/ppwfx/user-svc/pkg/utils/metricsutil"
)

var args = types.ServeArgs{}

func main() {
	flag.StringVar(&args.Port, "port", "8080", "")
	flag.StringVar(&args.PostgresUrl, "postgres-url", "", "")
	flag.StringVar(&args.HmacSecret, "hmac-secret", "", "")
	flag.StringVar(&args.AllowedSubjectSuffix, "allowed-subject-suffix", "", "")
	flag.StringVar(&args.Metrics, "metrics", "", "")
	flag.StringVar(&args.Logging, "logging", "", "")
	flag.StringVar(&args.Migrate, "migrate", "", "")
	flag.BoolVar(&args.ExposePprof, "expose-pprof", false, "")
	flag.IntVar(&args.HttpReadTimeoutSeconds, "http-read-timeout-seconds", 5, "")
	flag.Parse()

	ctx := context.Background()

	err := func() (err error) {
		var logger *zap.SugaredLogger
		switch args.Logging {
		case types.LoggingStackDriver:
			logger, err = loggingutil.NewStackDriverLogger("user-svc", "dev")
			if err != nil {
				err = errors.Wrap(err, "failed to create stackdriver logger")

				return
			}
		default:
			var unsugared *zap.Logger
			unsugared, err = zap.NewDevelopment()
			if err != nil {
				err = errors.Wrap(err, "failed to create development logger")

				return
			}
			logger = unsugared.Sugar()
		}
		defer func() {
			err := logger.Sync()
			if err != nil {
				err = errors.Wrap(err, "failed to flush logger buffer")

				log.Print(err)

				return
			}
		}()

		var metricSink metrics.MetricSink
		switch args.Metrics {
		case types.MetricsStackDriver:
			var metricClient *monitoring.MetricClient
			metricClient, metricSink, err = metricsutil.NewStackDriverMetricSink(ctx, "user-svc", "user-svc")
			if err != nil {
				err = errors.Wrap(err, "failed to create stackdriver metric sink")

				return
			}
			defer func() {
				err := metricClient.Close()
				if err != nil {
					err = errors.Wrap(err, "failed to close monitoring client")

					log.Print(err)

					return
				}
			}()
		default:
			metricSink, err = metricsutil.NewInMemoryMetricSink()
			if err != nil {
				err = errors.Wrap(err, "failed to create in-memory metric sink")

				return
			}
		}

		db, err := persistence.OpenPostgresDB(25, 25, 5*time.Minute, args.PostgresUrl)
		if err != nil {
			err = errors.Wrap(err, "failed to open postgres")

			return
		}

		err = persistence.ConnectToPostgresDb(ctx, db, 5*time.Second)
		if err != nil {
			err = errors.Wrap(err, "failed to connect to postgres")

			return
		}

		if args.Migrate != "" {
			err = persistence.Migrate(logger, args.Migrate, args.PostgresUrl)
			if err != nil {
				err = errors.Wrapf(err, "failed to migrate postgres")

				return
			}
		}

		validate := validator.New()

		mux := http.NewServeMux()
		mux = communication.AddSvcRoutes(mux, validate, logger, metricSink, db, args.HmacSecret, args.AllowedSubjectSuffix, business.DefaultArgon2IdOpts)

		if args.ExposePprof {
			mux = communication.AddPprofRoutes(mux)
		}

		addr := fmt.Sprintf("0.0.0.0:%v", args.Port)
		l, err := net.Listen("tcp", addr)
		if err != nil {
			err = errors.Wrapf(err, "failed to listen on %v", addr)

			return
		}

		s := &http.Server{
			ReadHeaderTimeout: 5 * time.Second,
			WriteTimeout:      5 * time.Second,
			ReadTimeout:       time.Duration(args.HttpReadTimeoutSeconds) * time.Second,
			IdleTimeout:       30 * time.Second,
			Handler:           mux,
		}

		logger.Infof("service listening on %v", addr)

		err = s.Serve(l)
		if err != nil {
			err = errors.Wrapf(err, "failed to serve")

			return
		}

		return
	}()
	if err != nil {
		log.Fatal("failed to run service: ", err)
	}

	return
}
