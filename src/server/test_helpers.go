package server

import "github.com/ihexxa/gocfg"

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
