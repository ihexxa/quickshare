package main

import (
	"github.com/ihexxa/gocfg"
	goflags "github.com/jessevdk/go-flags"

	"github.com/ihexxa/quickshare/src/server"
)

var opts struct {
	host    string   `short:"h" long:"host" description:"server hostname"`
	port    string   `short:"f" long:"file" description:"A file"`
	configs []string `short:"c" description:"config path"`
}

func main() {
	_, err := goflags.Parse(&opts)
	if err != nil {
		panic(err)
	}

	cfg := gocfg.New(server.NewDefaultConfig())
	if len(opts.configs) > 0 {
		for _, configPath := range opts.configs {
			cfg, err = cfg.Load(gocfg.JSON(configPath))
			if err != nil {
				panic(err)
			}
		}
	}

	srv, err := server.NewServer(cfg)
	if err != nil {
		panic(err)
	}

	err = srv.Start()
	if err != nil {
		panic(err)
	}
}
