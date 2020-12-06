package server

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ihexxa/gocfg"

	"github.com/ihexxa/quickshare/src/cryptoutil/jwt"
	"github.com/ihexxa/quickshare/src/depidx"
	"github.com/ihexxa/quickshare/src/fs"
	"github.com/ihexxa/quickshare/src/fs/local"
	"github.com/ihexxa/quickshare/src/handlers/fileshdr"
	"github.com/ihexxa/quickshare/src/handlers/singleuserhdr"
	"github.com/ihexxa/quickshare/src/idgen/simpleidgen"
	"github.com/ihexxa/quickshare/src/kvstore"
	"github.com/ihexxa/quickshare/src/kvstore/boltdbpvd"
	"github.com/ihexxa/quickshare/src/logging/simplelog"
	"github.com/ihexxa/quickshare/src/uploadmgr"
)

type Server struct {
	server *http.Server
	deps   *depidx.Deps
}

func NewServer(cfg gocfg.ICfg) (*Server, error) {
	deps := initDeps(cfg)

	if !cfg.BoolOr("Server.Debug", false) {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	router, err := initHandlers(router, cfg, deps)
	if err != nil {
		return nil, err
	}

	srv := &http.Server{
		Addr:           fmt.Sprintf("%s:%d", cfg.GrabString("Server.Host"), cfg.GrabInt("Server.Port")),
		Handler:        router,
		ReadTimeout:    time.Duration(cfg.GrabInt("Server.ReadTimeout")) * time.Millisecond,
		WriteTimeout:   time.Duration(cfg.GrabInt("Server.WriteTimeout")) * time.Millisecond,
		MaxHeaderBytes: cfg.GrabInt("Server.MaxHeaderBytes"),
	}

	return &Server{
		server: srv,
		deps:   deps,
	}, nil
}

func initDeps(cfg gocfg.ICfg) *depidx.Deps {
	secret, ok := cfg.String("ENV.TOKENSECRET")
	if !ok {
		secret = makeRandToken()
		fmt.Println("warning: TOKENSECRET is not given, using generated token")
	}

	rootPath := cfg.GrabString("Fs.Root")
	opensLimit := cfg.GrabInt("Fs.OpensLimit")
	openTTL := cfg.GrabInt("Fs.OpenTTL")

	ider := simpleidgen.New()
	filesystem := local.NewLocalFS(rootPath, 0660, opensLimit, openTTL)
	jwtEncDec := jwt.NewJWTEncDec(secret)
	logger := simplelog.NewSimpleLogger()
	kv := boltdbpvd.New(rootPath, 1024)

	deps := depidx.NewDeps(cfg)
	deps.SetFS(filesystem)
	deps.SetToken(jwtEncDec)
	deps.SetLog(logger)
	deps.SetKV(kv)
	deps.SetID(ider)

	uploadMgr, err := uploadmgr.NewUploadMgr(deps)
	if err != nil {
		panic(err)
	}
	deps.SetUploader(uploadMgr)

	return deps
}

func initHandlers(router *gin.Engine, cfg gocfg.ICfg, deps *depidx.Deps) (*gin.Engine, error) {
	userHdrs, err := singleuserhdr.NewSimpleUserHandlers(cfg, deps)
	if err != nil {
		return nil, err
	}
	if cfg.BoolOr("Users.EnableAuth", true) && !userHdrs.IsInited() {
		adminName, ok := cfg.String("ENV.DEFAULTADMIN")
		if !ok || adminName == "" {
			// TODO: print sceen only
			fmt.Println("Please input admin name:")
			fmt.Scanf("%s", &adminName)
		}

		adminPwd, _ := cfg.String("ENV.DEFAULTADMINPWD")
		adminPwd, err := userHdrs.Init(adminName, adminPwd)
		if err != nil {
			return nil, err
		}

		fmt.Printf("%s is created, its password is %s, please update it after login\n", adminName, adminPwd)
	}

	fileHdrs, err := fileshdr.NewFileHandlers(cfg, deps)
	if err != nil {
		return nil, err
	}

	// middleware
	router.Use(userHdrs.Auth())

	// handler
	v1 := router.Group("/v1")

	users := v1.Group("/users")
	users.POST("/login", userHdrs.Login)
	users.POST("/logout", userHdrs.Logout)
	users.PATCH("/pwd", userHdrs.SetPwd)

	filesSvc := v1.Group("/fs")
	filesSvc.POST("/files", fileHdrs.Create)
	filesSvc.DELETE("/files", fileHdrs.Delete)
	filesSvc.GET("/files", fileHdrs.Download)
	filesSvc.PATCH("/files/chunks", fileHdrs.UploadChunk)
	filesSvc.GET("/files/chunks", fileHdrs.UploadStatus)
	filesSvc.PATCH("/files/copy", fileHdrs.Copy)
	filesSvc.PATCH("/files/move", fileHdrs.Move)

	filesSvc.GET("/dirs", fileHdrs.List)
	filesSvc.POST("/dirs", fileHdrs.Mkdir)
	// files.POST("/dirs/copy", fileHdrs.CopyDir)

	filesSvc.GET("/metadata", fileHdrs.Metadata)

	return router, nil
}

func (s *Server) Start() error {
	err := s.server.ListenAndServe()
	if err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) Shutdown() error {
	// TODO: add timeout
	return s.server.Shutdown(context.Background())
}

func (s *Server) depsFS() fs.ISimpleFS {
	return s.deps.FS()
}

func (s *Server) depsKVStore() kvstore.IKVStore {
	return s.deps.KV()
}

func makeRandToken() string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return string(b)
}
