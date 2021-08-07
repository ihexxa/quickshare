package server

import (
	"context"
	"crypto/rand"
	"crypto/sha1"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/ihexxa/gocfg"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/crypto/bcrypt"

	"github.com/ihexxa/quickshare/src/cryptoutil/jwt"
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
	"github.com/ihexxa/quickshare/src/userstore"
)

type Server struct {
	server *http.Server
	cfg    gocfg.ICfg
	deps   *depidx.Deps
}

func NewServer(cfg gocfg.ICfg) (*Server, error) {
	if !cfg.BoolOr("Server.Debug", false) {
		gin.SetMode(gin.ReleaseMode)
	}

	deps := initDeps(cfg)
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
		cfg:    cfg,
	}, nil
}

func mkRoot(rootPath string) {
	info, err := os.Stat(rootPath)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(rootPath, 0760)
			if err != nil {
				panic(err)
			}
		} else {
			panic(err)
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

	ider := simpleidgen.New()
	filesystem := local.NewLocalFS(rootPath, 0660, opensLimit, openTTL)
	jwtEncDec := jwt.NewJWTEncDec(secret)
	kv := boltdbpvd.New(rootPath, 1024)
	users, err := userstore.NewKVUserStore(kv)
	if err != nil {
		panic(fmt.Sprintf("fail to init user store: %s", err))
	}

	limiterCap := cfg.IntOr("Users.LimiterCapacity", 4096)
	limiterCyc := cfg.IntOr("Users.LimiterCyc", 3000)
	limiter := iolimiter.NewIOLimiter(limiterCap, limiterCyc, users)

	deps := depidx.NewDeps(cfg)
	deps.SetFS(filesystem)
	deps.SetToken(jwtEncDec)
	deps.SetKV(kv)
	deps.SetUsers(users)
	deps.SetID(ider)
	deps.SetLog(logger)
	deps.SetLimiter(limiter)

	return deps
}

func initHandlers(router *gin.Engine, cfg gocfg.ICfg, deps *depidx.Deps) (*gin.Engine, error) {
	userHdrs, err := multiusers.NewMultiUsersSvc(cfg, deps)
	if err != nil {
		return nil, err
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
				return nil, err
			}
			// only write to stdout
			fmt.Printf("password is generated: %s, please update it after login\n", adminPwd)
		}

		pwdHash, err := bcrypt.GenerateFromPassword([]byte(adminPwd), 10)
		if err != nil {
			return nil, err
		}
		if _, err := userHdrs.Init(adminName, string(pwdHash)); err != nil {
			return nil, err
		}

		deps.Log().Infof("user (%s) is created\n", adminName)
	}

	fileHdrs, err := fileshdr.NewFileHandlers(cfg, deps)
	if err != nil {
		return nil, err
	}

	settingsSvc, err := settings.NewSettingsSvc(cfg, deps)
	if err != nil {
		return nil, err
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

	filesAPI.GET("/metadata", fileHdrs.Metadata)

	settingsAPI := v1.Group("/settings")
	settingsAPI.OPTIONS("/health", settingsSvc.Health)

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
		return err
	}
	return nil
}

func (s *Server) Shutdown() error {
	// TODO: add timeout
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
