package server

import (
	"context"

	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ihexxa/gocfg"

	"github.com/ihexxa/quickshare/src/depidx"
	"github.com/ihexxa/quickshare/src/fs"
	"github.com/ihexxa/quickshare/src/kvstore"
)

type Server struct {
	server     *http.Server
	cfg        gocfg.ICfg
	deps       *depidx.Deps
	signalChan chan os.Signal
}

func NewServer(cfg gocfg.ICfg) (*Server, error) {
	if !cfg.BoolOr("Server.Debug", false) {
		gin.SetMode(gin.ReleaseMode)
	}

	initer := NewIniter(cfg)
	deps := initer.InitDeps()
	router, err := initer.InitHandlers(deps)
	if err != nil {
		return nil, fmt.Errorf("init handlers error: %w", err)
	}

	port := cfg.GrabInt("Server.Port")
	portStr, ok := cfg.String("ENV.PORT")
	if ok && portStr != "" {
		port, err = strconv.Atoi(portStr)
		if err != nil {
			deps.Log().Fatalf("invalid port: %s", portStr)
		}
		cfg.SetInt("Server.Port", port)
	}

	srv := &http.Server{
		Addr:           fmt.Sprintf("%s:%d", cfg.GrabString("Server.Host"), port),
		Handler:        router,
		ReadTimeout:    time.Duration(cfg.GrabInt("Server.ReadTimeout")) * time.Millisecond,
		WriteTimeout:   time.Duration(cfg.GrabInt("Server.WriteTimeout")) * time.Millisecond,
		MaxHeaderBytes: cfg.GrabInt("Server.MaxHeaderBytes"),
	}

	return &Server{
		server: srv,
		deps:   deps,
		cfg:    cfg,
	}, nil
}

func (s *Server) Start() error {
	s.signalChan = make(chan os.Signal, 4)
	signal.Notify(s.signalChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-s.signalChan
		if sig != nil {
			s.deps.Log().Infow(
				fmt.Sprintf("received signal %s: shutting down", sig.String()),
			)
		}
		s.Shutdown()
	}()

	s.deps.Log().Infow(
		"quickshare is starting",
		"hostname:port",
		fmt.Sprintf(
			"%s:%d",
			s.cfg.GrabString("Server.Host"),
			s.cfg.GrabInt("Server.Port"),
		),
	)

	err := s.server.ListenAndServe()
	if err != http.ErrServerClosed {
		return fmt.Errorf("listen error: %w", err)
	}
	return nil
}

func (s *Server) Shutdown() error {
	// TODO: add timeout
	err := s.deps.FileIndex().WriteTo(fileIndexPath)
	if err != nil {
		s.deps.Log().Errorf("failed to persist file index: %s", err)
	}
	s.deps.Workers().Stop()
	err = s.deps.FS().Close()
	if err != nil {
		s.deps.Log().Errorf("failed to close file system: %s", err)
	}
	err = s.deps.DB().Close()
	if err != nil {
		s.deps.Log().Errorf("failed to close database: %s", err)
	}
	err = s.server.Shutdown(context.Background())
	if err != nil {
		s.deps.Log().Errorf("failed to shutdown server: %s", err)
	}

	s.deps.Log().Sync()
	return nil
}

func (s *Server) depsFS() fs.ISimpleFS {
	return s.deps.FS()
}

func (s *Server) depsKVStore() kvstore.IKVStore {
	return s.deps.KV()
}
