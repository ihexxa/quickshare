package main

import (
	"fmt"
	"strings"

	"github.com/ihexxa/gocfg"
	goflags "github.com/jessevdk/go-flags"

	"github.com/ihexxa/quickshare/src/server"
)

var opts struct {
	Host    string   `short:"h" long:"host" description:"server hostname"`
	Port    int      `short:"p" long:"port" description:"server port"`
	Debug   bool     `short:"d" long:"debug" description:"debug mode"`
	Configs []string `short:"c" description:"config path"`
}

func main() {
	_, err := goflags.Parse(&opts)
	if err != nil {
		panic(err)
	}
	defaultCfg, err := server.DefaultConfig()
	if err != nil {
		panic(err)
	}

	cfg, err := gocfg.New(server.NewConfig()).Load(gocfg.JSONStr(defaultCfg))
	if err != nil {
		panic(err)
	}
	if len(opts.Configs) > 0 {
		for _, configPath := range opts.Configs {
			if strings.HasSuffix(configPath, ".yml") || strings.HasSuffix(configPath, ".yaml") {
				cfg, err = cfg.Load(gocfg.YAML(configPath))
			} else if strings.HasSuffix(configPath, ".json") {
				cfg, err = cfg.Load(gocfg.JSON(configPath))
			} else {
				panic(fmt.Sprintf("unknown config file type (.yml .yaml .json are supported): %s", configPath))
			}

			if err != nil {
				panic(err)
			}
		}
	}

	if opts.Host != "" {
		cfg.SetString("Server.Host", opts.Host)
	}
	if opts.Port != 0 {
		cfg.SetInt("Server.Port", opts.Port)
	}
	if opts.Debug {
		cfg.SetBool("Server.Debug", opts.Debug)
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
