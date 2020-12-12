package server

import (
	"context"
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/ihexxa/gocfg"

	"github.com/ihexxa/quickshare/src/cryptoutil/jwt"
	"github.com/ihexxa/quickshare/src/depidx"
	"github.com/ihexxa/quickshare/src/fs"
	"github.com/ihexxa/quickshare/src/fs/local"
	"github.com/ihexxa/quickshare/src/handlers/fileshdr"
	"github.com/ihexxa/quickshare/src/handlers/settings"
	"github.com/ihexxa/quickshare/src/handlers/singleuserhdr"
	"github.com/ihexxa/quickshare/src/idgen/simpleidgen"
	"github.com/ihexxa/quickshare/src/kvstore"
	"github.com/ihexxa/quickshare/src/kvstore/boltdbpvd"
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
	kv := boltdbpvd.New(rootPath, 1024)

	deps := depidx.NewDeps(cfg)
	deps.SetFS(filesystem)
	deps.SetToken(jwtEncDec)
	deps.SetKV(kv)
	deps.SetID(ider)

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
			// only write to stdout
			fmt.Print("Please input admin name: ")
			fmt.Scanf("%s", &adminName)
		}

		adminPwd, _ := cfg.String("ENV.DEFAULTADMINPWD")
		if adminPwd == "" {
			adminPwd, err = generatePwd()
			if err != nil {
				return nil, err
			}
			// only write to stdout
			fmt.Printf("password is generated: %s, please update it after login\n", adminPwd)
		}
		adminPwd, err := userHdrs.Init(adminName, adminPwd)
		if err != nil {
			return nil, err
		}

		fmt.Printf("user (%s) is created\n", adminName)
	}

	fileHdrs, err := fileshdr.NewFileHandlers(cfg, deps)
	if err != nil {
		return nil, err
	}

	settingsSvc, err := settings.NewSettingsSvc(cfg, deps)
	if err != nil {
		return nil, err
	}

	// middleware
	router.Use(userHdrs.Auth())
	// tmp static server
	router.Use(static.Serve("/", static.LocalFile("../public", false)))

	// handler
	v1 := router.Group("/v1")

	usersAPI := v1.Group("/users")
	usersAPI.POST("/login", userHdrs.Login)
	usersAPI.POST("/logout", userHdrs.Logout)
	usersAPI.PATCH("/pwd", userHdrs.SetPwd)

	filesAPI := v1.Group("/fs")
	filesAPI.POST("/files", fileHdrs.Create)
	filesAPI.DELETE("/files", fileHdrs.Delete)
	filesAPI.GET("/files", fileHdrs.Download)
	filesAPI.PATCH("/files/chunks", fileHdrs.UploadChunk)
	filesAPI.GET("/files/chunks", fileHdrs.UploadStatus)
	filesAPI.PATCH("/files/copy", fileHdrs.Copy)
	filesAPI.PATCH("/files/move", fileHdrs.Move)

	filesAPI.GET("/dirs", fileHdrs.List)
	filesAPI.POST("/dirs", fileHdrs.Mkdir)
	// files.POST("/dirs/copy", fileHdrs.CopyDir)

	filesAPI.GET("/metadata", fileHdrs.Metadata)

	settingsAPI := v1.Group("/settings")
	settingsAPI.OPTIONS("/health", settingsSvc.Health)

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

func generatePwd() (string, error) {
	size := 10
	buf := make([]byte, size)
	size, err := rand.Read(buf)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", sha1.Sum(buf[:size]))[:6], nil
}
