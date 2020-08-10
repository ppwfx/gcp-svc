package main

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"github.com/ppwfx/user-svc/pkg/business"
	"github.com/ppwfx/user-svc/pkg/communication"
	"github.com/ppwfx/user-svc/pkg/persistence"
	"github.com/ppwfx/user-svc/pkg/types"
	"github.com/ppwfx/user-svc/pkg/utils/loggingutil"
	"github.com/ppwfx/user-svc/pkg/utils/metricsutil"
)

var args = types.ServeArgs{}

func main() {
	flag.StringVar(&args.Addr, "addr", "", "")
	flag.StringVar(&args.DbConnection, "db-connection", "", "")
	flag.StringVar(&args.HmacSecret, "hmac-secret", "", "")
	flag.StringVar(&args.AllowedSubjectSuffix, "allowed-subject-suffix", "", "")
	flag.Parse()

	ctx := context.Background()

	err := func() (err error) {
		logger, err := loggingutil.NewProductionLogger("user-svc", "dev")
		if err != nil {
			err = errors.Wrap(err, "failed to create logger")

			return
		}
		defer func() {
			err := logger.Sync()
			if err != nil {
				err = errors.Wrap(err, "failed to flush logger buffer")

				log.Print(err)

				return
			}
		}()

		monitoringClient, metrics, err := metricsutil.NewProductionMetricSink(ctx, "user-svc", "user-svc")
		if err != nil {
			err = errors.Wrap(err, "failed to get metrics")

			return
		}
		defer func() {
			err := monitoringClient.Close()
			if err != nil {
				err = errors.Wrap(err, "failed to close monitoring client")

				log.Print(err)

				return
			}
		}()

		db, err := persistence.OpenPostgresDB(25, 25, 5*time.Minute, args.DbConnection)
		if err != nil {
			err = errors.Wrap(err, "failed to open postgres")

			return
		}

		err = persistence.ConnectToPostgresDb(ctx, db, 5*time.Second)
		if err != nil {
			err = errors.Wrap(err, "failed to connect to postgres")

			return
		}

		validate := validator.New()

		mux := http.NewServeMux()
		mux = communication.AddSvcRoutes(mux, validate, logger, metrics, db, args.HmacSecret, args.AllowedSubjectSuffix, business.DefaultArgon2IdOpts)

		l, err := net.Listen("tcp", args.Addr)
		if err != nil {
			err = errors.Wrapf(err, "failed to listen on %v", args.Addr)

			return
		}

		s := &http.Server{
			ReadHeaderTimeout: 5 * time.Second,
			WriteTimeout:      5 * time.Second,
			ReadTimeout:       5 * time.Second,
			IdleTimeout:       30 * time.Second,
			Handler:           mux,
		}

		logger.Infof("service listening on %v", args.Addr)

		err = s.Serve(l)
		if err != nil {
			err = errors.Wrapf(err, "failed to serve")

			return
		}

		return
	}()
	if err != nil {
		err = errors.Wrap(err, "failed to run service")

		return
	}

	return
}
