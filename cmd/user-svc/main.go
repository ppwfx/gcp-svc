package main

import (
	"flag"
	"github.com/go-playground/validator/v10"
	"github.com/ppwfx/user-svc/pkg/communication"
	"github.com/ppwfx/user-svc/pkg/persistence"
	"github.com/ppwfx/user-svc/pkg/types"
	"log"
	"time"
)

var args = types.ServeArgs{}

func main() {
	flag.StringVar(&args.Addr, "addr", "", "")
	flag.StringVar(&args.DbConnection, "db-connection", "", "")
	flag.StringVar(&args.HmacSecret, "hmac-secret", "", "")
	flag.StringVar(&args.Salt, "salt", "", "")
	flag.Parse()

	v := validator.New()

	err := func() (err error) {
		err = persistence.WaitForDb(args.DbConnection)
		if err != nil {
			return
		}

		db, err := persistence.GetDb(25, 25, 5*time.Minute, args.DbConnection)
		if err != nil {
			return
		}

		err = communication.Serve(v, db, args.Addr, args.HmacSecret, args.Salt)
		if err != nil {
			return
		}

		return
	}()
	if err != nil {
		log.Fatal(err)
	}
}
