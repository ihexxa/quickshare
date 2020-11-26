package main

import (
	"github.com/ihexxa/quickshare/src/server"

	"github.com/ihexxa/gocfg"
)

func main() {
	cfg := gocfg.New()
	srv, err := server.NewServer(cfg)
	if err != nil {
		panic(err)
	}
	srv.Start()
}
