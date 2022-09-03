package main

import (
	"context"
	"fmt"
	"os"

	goflags "github.com/jessevdk/go-flags"

	serverPkg "github.com/ihexxa/quickshare/src/server"
)

var args = &serverPkg.Args{}

func main() {
	_, err := goflags.Parse(args)
	if err != nil {
		panic(err)
	}

	ctx := context.TODO()
	cfg, err := serverPkg.LoadCfg(ctx, args)
	if err != nil {
		fmt.Printf("failed to load config: %s", err)
		os.Exit(1)
	}

	srv, err := serverPkg.NewServer(cfg)
	if err != nil {
		fmt.Printf("failed to new server: %s", err)
		os.Exit(1)
	}

	err = srv.Start()
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		os.Exit(1)
	}
}
