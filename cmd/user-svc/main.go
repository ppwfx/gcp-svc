package main

import (
	"context"
	"flag"
	"github.com/ppwfx/user-svc/pkg/business"
	"go.uber.org/zap"
	"log"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/ppwfx/user-svc/pkg/communication"
	"github.com/ppwfx/user-svc/pkg/persistence"
	"github.com/ppwfx/user-svc/pkg/types"
)

var args = types.ServeArgs{}

func main() {
	flag.StringVar(&args.Addr, "addr", "", "")
	flag.StringVar(&args.DbConnection, "db-connection", "", "")
	flag.StringVar(&args.HmacSecret, "hmac-secret", "", "")
	flag.StringVar(&args.AllowedSubjectSuffix, "allowed-subject-suffix", "", "")
	flag.Parse()

	v := validator.New()

	err := func() (err error) {
		db, err := persistence.OpenPostgresDB(25, 25, 5*time.Minute, args.DbConnection)
		if err != nil {
			return
		}

		err = persistence.ConnectToPostgresDb(context.Background(), db, 5*time.Second)
		if err != nil {
			return
		}

		c := zap.NewProductionConfig()
		c.DisableStacktrace = true
		l, _ := c.Build()

		err = communication.Serve(v, l.Sugar(), db, args.Addr, args.HmacSecret, args.AllowedSubjectSuffix, business.DefaultArgon2IdOpts)
		if err != nil {
			return
		}

		return
	}()
	if err != nil {
		log.Fatal(err)
	}
}
