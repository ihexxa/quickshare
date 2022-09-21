package server

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/ihexxa/gocfg"

	"github.com/ihexxa/quickshare/src/depidx"
	"github.com/ihexxa/quickshare/src/fs"
	"github.com/ihexxa/quickshare/src/handlers/fileshdr"
	"github.com/ihexxa/quickshare/src/handlers/multiusers"
	"github.com/ihexxa/quickshare/src/handlers/settings"
	"github.com/ihexxa/quickshare/src/kvstore"
	qsstatic "github.com/ihexxa/quickshare/static"
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

	deps := initDeps(cfg)
	router := gin.Default()
	adminName := cfg.GrabString("ENV.DEFAULTADMIN")
	router, err := initHandlers(router, adminName, cfg, deps)
	if err != nil {
		return nil, fmt.Errorf("init handlers error: %w", err)
	}

	port := cfg.GrabInt("Server.Port")
	portStr, ok := cfg.String("ENV.PORT")
	if ok && portStr != "" {
		port, err = strconv.Atoi(portStr)
		if err != nil {
			panic(fmt.Sprintf("invalid port: %s", portStr))
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

func initHandlers(router *gin.Engine, adminName string, cfg gocfg.ICfg, deps *depidx.Deps) (*gin.Engine, error) {
	// handlers
	userHdrs, err := multiusers.NewMultiUsersSvc(cfg, deps)
	if err != nil {
		return nil, fmt.Errorf("new users svc error: %w", err)
	}

	_, err = userHdrs.Init(context.TODO(), adminName)
	if err != nil {
		return nil, fmt.Errorf("failed to init user handlers: %w", err)
	}

	fileHdrs, err := fileshdr.NewFileHandlers(cfg, deps)
	if err != nil {
		return nil, fmt.Errorf("new files service error: %w", err)
	}
	settingsSvc, err := settings.NewSettingsSvc(cfg, deps)
	if err != nil {
		return nil, fmt.Errorf("new setting service error: %w", err)
	}

	// middlewares
	router.Use(userHdrs.AuthN())
	router.Use(userHdrs.APIAccessControl())

	publicPath, ok := cfg.String("Server.PublicPath")
	if !ok || publicPath == "" {
		return nil, errors.New("publicPath not found or empty")
	}
	if cfg.BoolOr("Server.Debug", false) {
		router.Use(static.Serve("/", static.LocalFile(publicPath, false)))
	} else {
		embedFs, err := qsstatic.NewEmbedStaticFS()
		if err != nil {
			return nil, err
		}
		router.Use(static.Serve("/", embedFs))
	}

	// handlers
	v1 := router.Group("/v1")

	usersAPI := v1.Group("/users")
	usersAPI.POST("/login", userHdrs.Login)
	usersAPI.POST("/logout", userHdrs.Logout)
	usersAPI.GET("/isauthed", userHdrs.IsAuthed)
	usersAPI.PATCH("/pwd", userHdrs.SetPwd)
	usersAPI.PATCH("/pwd/force-set", userHdrs.ForceSetPwd)
	usersAPI.POST("/", userHdrs.AddUser)
	usersAPI.DELETE("/", userHdrs.DelUser)
	usersAPI.GET("/list", userHdrs.ListUsers)
	usersAPI.GET("/self", userHdrs.Self)
	usersAPI.PATCH("/", userHdrs.SetUser)
	usersAPI.PATCH("/preferences", userHdrs.SetPreferences)
	usersAPI.PUT("/used-space", userHdrs.ResetUsedSpace)

	rolesAPI := v1.Group("/roles")
	// rolesAPI.POST("/", userHdrs.AddRole)
	// rolesAPI.DELETE("/", userHdrs.DelRole)
	rolesAPI.GET("/list", userHdrs.ListRoles)

	captchaAPI := v1.Group("/captchas")
	captchaAPI.GET("/", userHdrs.GetCaptchaID)
	captchaAPI.GET("/imgs", userHdrs.GetCaptchaImg)

	filesAPI := v1.Group("/fs")
	filesAPI.POST("/files", fileHdrs.Create)
	filesAPI.DELETE("/files", fileHdrs.Delete)
	filesAPI.GET("/files", fileHdrs.Download)
	filesAPI.PATCH("/files/chunks", fileHdrs.UploadChunk)
	filesAPI.GET("/files/chunks", fileHdrs.UploadStatus)
	filesAPI.PATCH("/files/copy", fileHdrs.Copy)
	filesAPI.PATCH("/files/move", fileHdrs.Move)

	filesAPI.GET("/dirs", fileHdrs.List)
	filesAPI.GET("/dirs/home", fileHdrs.ListHome)
	filesAPI.POST("/dirs", fileHdrs.Mkdir)
	// files.POST("/dirs/copy", fileHdrs.CopyDir)

	filesAPI.GET("/uploadings", fileHdrs.ListUploadings)
	filesAPI.DELETE("/uploadings", fileHdrs.DelUploading)

	filesAPI.POST("/sharings", fileHdrs.AddSharing)
	filesAPI.DELETE("/sharings", fileHdrs.DelSharing)
	filesAPI.GET("/sharings", fileHdrs.ListSharings)
	filesAPI.GET("/sharings/ids", fileHdrs.ListSharingIDs)
	filesAPI.GET("/sharings/exist", fileHdrs.IsSharing)
	filesAPI.GET("/sharings/dirs", fileHdrs.GetSharingDir)

	filesAPI.GET("/metadata", fileHdrs.Metadata)
	filesAPI.GET("/search", fileHdrs.SearchItems)
	filesAPI.PUT("/reindex", fileHdrs.Reindex)

	filesAPI.POST("/hashes/sha1", fileHdrs.GenerateHash)

	settingsAPI := v1.Group("/settings")
	settingsAPI.OPTIONS("/health", settingsSvc.Health)
	settingsAPI.GET("/client", settingsSvc.GetClientCfg)
	settingsAPI.PATCH("/client", settingsSvc.SetClientCfg)
	settingsAPI.POST("/errors", settingsSvc.ReportErrors)
	settingsAPI.GET("/workers/queue-len", settingsSvc.WorkerQueueLen)

	return router, nil
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
