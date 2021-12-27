package server

import (
	"context"
	"crypto/rand"
	"crypto/sha1"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/ihexxa/gocfg"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/crypto/bcrypt"

	"github.com/ihexxa/quickshare/src/cryptoutil/jwt"
	"github.com/ihexxa/quickshare/src/db/fileinfostore"
	"github.com/ihexxa/quickshare/src/db/sitestore"
	"github.com/ihexxa/quickshare/src/db/userstore"
	"github.com/ihexxa/quickshare/src/depidx"
	"github.com/ihexxa/quickshare/src/fs"
	"github.com/ihexxa/quickshare/src/fs/local"
	"github.com/ihexxa/quickshare/src/handlers/fileshdr"
	"github.com/ihexxa/quickshare/src/handlers/multiusers"
	"github.com/ihexxa/quickshare/src/handlers/settings"
	"github.com/ihexxa/quickshare/src/idgen/simpleidgen"
	"github.com/ihexxa/quickshare/src/iolimiter"
	"github.com/ihexxa/quickshare/src/kvstore"
	"github.com/ihexxa/quickshare/src/kvstore/boltdbpvd"
	"github.com/ihexxa/quickshare/src/worker/localworker"
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
	router, err := initHandlers(router, cfg, deps)
	if err != nil {
		return nil, fmt.Errorf("init handlers error: %w", err)
	}

	err = checkCompatibility(deps)
	if err != nil {
		return nil, fmt.Errorf("fail to check compatibility: %w", err)
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

func checkCompatibility(deps *depidx.Deps) error {
	users, err := deps.Users().ListUsers()
	if err != nil {
		return err
	}

	for _, user := range users {
		fmt.Println(user, user.Preferences)
		if user.Preferences == nil {
			deps.Users().SetPreferences(user.ID, &userstore.DefaultPreferences)
		}
	}

	return nil
}

func mkRoot(rootPath string) {
	info, err := os.Stat(rootPath)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(rootPath, 0760)
			if err != nil {
				panic(fmt.Sprintf("mk root path error: %s", err))
			}
		} else {
			panic(fmt.Sprintf("stat root Path error: %s", err))
		}
	} else if !info.IsDir() {
		panic(fmt.Sprintf("can not create %s folder: there is a file with same name", rootPath))
	}
}

func initDeps(cfg gocfg.ICfg) *depidx.Deps {
	logger := initLogger(cfg)

	secret, ok := cfg.String("ENV.TOKENSECRET")
	if !ok {
		secret = makeRandToken()
		logger.Info("warning: TOKENSECRET is not given, using generated token")
	}

	rootPath := cfg.GrabString("Fs.Root")
	mkRoot(rootPath)
	opensLimit := cfg.GrabInt("Fs.OpensLimit")
	openTTL := cfg.GrabInt("Fs.OpenTTL")
	readerTTL := cfg.GrabInt("Server.WriteTimeout") / 1000 // millisecond -> second

	ider := simpleidgen.New()
	filesystem := local.NewLocalFS(rootPath, 0660, opensLimit, openTTL, readerTTL, ider)
	jwtEncDec := jwt.NewJWTEncDec(secret)
	kv := boltdbpvd.New(rootPath, 1024)
	users, err := userstore.NewKVUserStore(kv)
	if err != nil {
		panic(fmt.Sprintf("fail to init user store: %s", err))
	}
	fileInfos, err := fileinfostore.NewFileInfoStore(kv)
	if err != nil {
		panic(fmt.Sprintf("fail to init file info store: %s", err))
	}
	siteStore, err := sitestore.NewSiteStore(kv)
	if err != nil {
		panic(fmt.Sprintf("fail to new site config store: %s", err))
	}

	err = siteStore.Init(&sitestore.SiteConfig{
		ClientCfg: &sitestore.ClientConfig{
			SiteName: cfg.StringOr("Site.ClientCfg.SiteName", "Quickshare"),
			SiteDesc: cfg.StringOr("Site.ClientCfg.SiteDesc", "quick and simple file sharing"),
			Bg: &sitestore.BgConfig{
				Url:      cfg.StringOr("Site.ClientCfg.Bg.Url", "/static/img/textured_paper.png"),
				Repeat:   cfg.StringOr("Site.ClientCfg.Bg.Repeat", "repeat"),
				Position: cfg.StringOr("Site.ClientCfg.Bg.Position", "fixed"),
				Align:    cfg.StringOr("Site.ClientCfg.Bg.Align", "center"),
			},
		},
	})
	if err != nil {
		panic(fmt.Sprintf("fail to init site config store: %s", err))
	}

	limiterCap := cfg.IntOr("Users.LimiterCapacity", 10000)
	limiterCyc := cfg.IntOr("Users.LimiterCyc", 1000)
	limiter := iolimiter.NewIOLimiter(limiterCap, limiterCyc, users)

	deps := depidx.NewDeps(cfg)
	deps.SetFS(filesystem)
	deps.SetToken(jwtEncDec)
	deps.SetKV(kv)
	deps.SetUsers(users)
	deps.SetFileInfos(fileInfos)
	deps.SetSiteStore(siteStore)
	deps.SetID(ider)
	deps.SetLog(logger)
	deps.SetLimiter(limiter)

	queueSize := cfg.GrabInt("Workers.QueueSize")
	sleepCyc := cfg.GrabInt("Workers.SleepCyc")
	workerCount := cfg.GrabInt("Workers.WorkerCount")

	workers := localworker.NewWorkerPool(queueSize, sleepCyc, workerCount, logger)
	workers.Start()
	deps.SetWorkers(workers)

	return deps
}

func initHandlers(router *gin.Engine, cfg gocfg.ICfg, deps *depidx.Deps) (*gin.Engine, error) {
	// handlers
	userHdrs, err := multiusers.NewMultiUsersSvc(cfg, deps)
	if err != nil {
		return nil, fmt.Errorf("new users svc error: %w", err)
	}
	if cfg.BoolOr("Users.EnableAuth", true) && !userHdrs.IsInited() {
		adminName, ok := cfg.String("ENV.DEFAULTADMIN")
		if !ok || adminName == "" {
			// only write to stdout
			deps.Log().Info("Please input admin name: ")
			fmt.Scanf("%s", &adminName)
		}

		adminPwd, _ := cfg.String("ENV.DEFAULTADMINPWD")
		if adminPwd == "" {
			adminPwd, err = generatePwd()
			if err != nil {
				return nil, fmt.Errorf("generate pwd error: %w", err)
			}
			// only write to stdout
			fmt.Printf("password is generated: %s, please update it after login\n", adminPwd)
		}

		pwdHash, err := bcrypt.GenerateFromPassword([]byte(adminPwd), 10)
		if err != nil {
			return nil, fmt.Errorf("generate pwd error: %w", err)
		}
		if _, err := userHdrs.Init(adminName, string(pwdHash)); err != nil {
			return nil, fmt.Errorf("init admin error: %w", err)
		}

		deps.Log().Infof("admin(%s) is created", adminName)
	}

	fileHdrs, err := fileshdr.NewFileHandlers(cfg, deps)
	if err != nil {
		return nil, fmt.Errorf("new files service error: %w", err)
	}

	settingsSvc, err := settings.NewSettingsSvc(cfg, deps)
	if err != nil {
		return nil, fmt.Errorf("new setting service error: %w", err)
	}

	publicPath, ok := cfg.String("Server.PublicPath")
	if !ok || publicPath == "" {
		return nil, errors.New("publicPath not found or empty")
	}

	// middleware
	router.Use(userHdrs.AuthN())
	router.Use(userHdrs.APIAccessControl())
	// tmp static server
	router.Use(static.Serve("/", static.LocalFile(publicPath, false)))

	// handler
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

	rolesAPI := v1.Group("/roles")
	rolesAPI.POST("/", userHdrs.AddRole)
	rolesAPI.DELETE("/", userHdrs.DelRole)
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
	filesAPI.GET("/sharings/exist", fileHdrs.IsSharing)

	filesAPI.GET("/metadata", fileHdrs.Metadata)

	filesAPI.POST("/hashes/sha1", fileHdrs.GenerateHash)

	settingsAPI := v1.Group("/settings")
	settingsAPI.OPTIONS("/health", settingsSvc.Health)
	settingsAPI.GET("/client", settingsSvc.GetClientCfg)
	settingsAPI.PATCH("/client", settingsSvc.SetClientCfg)
	settingsAPI.POST("/errors", settingsSvc.ReportError)

	return router, nil
}

func initLogger(cfg gocfg.ICfg) *zap.SugaredLogger {
	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   filepath.Join(cfg.GrabString("Fs.Root"), "quickshare.log"),
		MaxSize:    cfg.IntOr("Log.MaxSize", 50), // megabytes
		MaxBackups: cfg.IntOr("Log.MaxBackups", 2),
		MaxAge:     cfg.IntOr("Log.MaxAge", 31), // days
	})
	stdoutWriter := zapcore.AddSync(os.Stdout)

	multiWriter := zapcore.NewMultiWriteSyncer(fileWriter, stdoutWriter)
	gin.DefaultWriter = multiWriter
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		multiWriter,
		zap.InfoLevel,
	)
	return zap.New(core).Sugar()
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
	s.deps.Workers().Stop()
	s.deps.FS().Close()
	s.deps.Log().Sync()
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
		panic(fmt.Sprintf("make rand token error: %s", err))
	}
	return string(b)
}

func generatePwd() (string, error) {
	size := 10
	buf := make([]byte, size)
	size, err := rand.Read(buf)
	if err != nil {
		return "", fmt.Errorf("generate pwd error: %w", err)
	}

	return fmt.Sprintf("%x", sha1.Sum(buf[:size]))[:6], nil
}
