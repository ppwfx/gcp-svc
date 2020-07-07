package main

import (
	"flag"
	"github.com/ppwfx/user-svc/pkg/communication"
	"github.com/ppwfx/user-svc/pkg/types"
	"log"
)

var serveArgs = types.ServeArgs{}

func main() {
	flag.StringVar(&serveArgs.Addr, "addr", "", "")

	flag.Parse()

	err := communication.Serve(serveArgs.Addr)
	if err != nil {
		log.Fatal(err)
	}
}
