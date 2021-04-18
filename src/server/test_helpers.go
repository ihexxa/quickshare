package server

import (
	"time"

	"github.com/ihexxa/gocfg"
	"github.com/ihexxa/quickshare/src/client"
)

func startTestServer(config string) *Server {
	defaultCfg, err := DefaultConfig()
	if err != nil {
		panic(err)
	}

	cfg, err := gocfg.New(NewConfig()).
		Load(
			gocfg.JSONStr(defaultCfg),
			gocfg.JSONStr(config),
		)
	if err != nil {
		panic(err)
	}

	srv, err := NewServer(cfg)
	if err != nil {
		panic(err)
	}

	go srv.Start()
	return srv
}

func waitForReady(addr string) bool {
	retry := 20
	setCl := client.NewSettingsClient(addr)

	for retry > 0 {
		_, _, errs := setCl.Health()
		if len(errs) > 0 {
			time.Sleep(100 * time.Millisecond)
		} else {
			return true
		}
		retry--
	}

	return false
}
