package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"github.com/ppwfx/user-svc/pkg/business"
	"github.com/ppwfx/user-svc/pkg/communication"
	"github.com/ppwfx/user-svc/pkg/persistence"
	"github.com/ppwfx/user-svc/pkg/types"
	"github.com/ppwfx/user-svc/pkg/utils"
)

var args = types.ServeArgs{}

func main() {
	flag.StringVar(&args.Addr, "addr", "", "")
	flag.StringVar(&args.DbConnection, "db-connection", "", "")
	flag.StringVar(&args.HmacSecret, "hmac-secret", "", "")
	flag.StringVar(&args.AllowedSubjectSuffix, "allowed-subject-suffix", "", "")
	flag.Parse()

	ctx, _ := context.WithCancel(context.Background())

	err := func() (err error) {
		l, err := utils.NewProductionLogger("user-svc", "dev")
		if err != nil {
			err = errors.Wrap(err, "failed to create logger")

			return
		}
		defer func() {
			err := l.Sync()
			if err != nil {
				err = errors.Wrap(err, "failed to flush logger buffer")

				log.Print(err)

				return
			}
		}()

		c, m, err := utils.NewProductionMetrics(ctx, "user-svc")
		defer func() {
			err := c.Close()
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

		err = persistence.ConnectToPostgresDb(ctx, m, db, 5*time.Second)
		if err != nil {
			err = errors.Wrap(err, "failed to connect to postgres")

			return
		}

		v := validator.New()

		err = communication.Serve(v, l, m, db, args.Addr, args.HmacSecret, args.AllowedSubjectSuffix, business.DefaultArgon2IdOpts)
		if err != nil {
			err = errors.Wrap(err, "failed to listen")

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
