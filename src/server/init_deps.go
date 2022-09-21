package server

import (
	"context"
	"crypto/rand"
	"crypto/sha1"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/ihexxa/gocfg"
	"github.com/ihexxa/quickshare/src/db/rdb/sqlite"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/crypto/bcrypt"

	"github.com/ihexxa/quickshare/src/cryptoutil/jwt"
	"github.com/ihexxa/quickshare/src/db"
	"github.com/ihexxa/quickshare/src/depidx"
	"github.com/ihexxa/quickshare/src/fs"
	"github.com/ihexxa/quickshare/src/fs/local"
	"github.com/ihexxa/quickshare/src/idgen/simpleidgen"
	"github.com/ihexxa/quickshare/src/iolimiter"
	"github.com/ihexxa/quickshare/src/search/fileindex"
	"github.com/ihexxa/quickshare/src/worker/localworker"
)

func InitCfg(cfg gocfg.ICfg, logger *zap.SugaredLogger) (gocfg.ICfg, error) {
	_, ok := cfg.String("ENV.TOKENSECRET")
	if !ok {
		cfg.SetString("ENV.TOKENSECRET", makeRandToken())
		logger.Info("warning: TOKENSECRET is not set, generated a random token")
	}

	return cfg, nil
}

func initLogger(cfg gocfg.ICfg) *zap.SugaredLogger {
	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   path.Join(cfg.GrabString("Fs.Root"), "quickshare.log"),
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

func makeRandToken() string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		panic(fmt.Sprintf("make rand token error: %s", err))
	}
	return string(b)
}

func mkRoot(rootPath string, logger *zap.SugaredLogger) {
	info, err := os.Stat(rootPath)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(rootPath, 0760)
			if err != nil {
				logger.Fatalf("create root path error: %s", err)
			}
		} else {
			logger.Fatalf("stat root Path error: %s", err)
		}
	} else if !info.IsDir() {
		logger.Fatalf("can not create %s folder: there is a file with same name", rootPath)
	}
}

func initDeps(cfg gocfg.ICfg) *depidx.Deps {
	var err error
	logger := initLogger(cfg)

	rootPath := cfg.GrabString("Fs.Root")
	mkRoot(rootPath, logger)
	opensLimit := cfg.GrabInt("Fs.OpensLimit")
	openTTL := cfg.GrabInt("Fs.OpenTTL")
	readerTTL := cfg.GrabInt("Server.WriteTimeout") / 1000 // millisecond -> second
	ider := simpleidgen.New()
	filesystem := local.NewLocalFS(rootPath, 0660, opensLimit, openTTL, readerTTL, ider)

	secret, _ := cfg.String("ENV.TOKENSECRET")
	jwtEncDec := jwt.NewJWTEncDec(secret)

	quickshareDb, err := initDB(cfg, filesystem)
	if err != nil {
		logger.Errorf("failed to init DB: %s", err)
		os.Exit(1)
	}

	limiterCap := cfg.IntOr("Users.LimiterCapacity", 10000)
	limiterCyc := cfg.IntOr("Users.LimiterCyc", 1000)
	limiter := iolimiter.NewIOLimiter(limiterCap, limiterCyc, quickshareDb)

	deps := depidx.NewDeps(cfg)
	deps.SetDB(quickshareDb)
	deps.SetFS(filesystem)
	deps.SetToken(jwtEncDec)
	deps.SetID(ider)
	deps.SetLog(logger)
	deps.SetLimiter(limiter)

	queueSize := cfg.GrabInt("Workers.QueueSize")
	sleepCyc := cfg.GrabInt("Workers.SleepCyc")
	workerCount := cfg.GrabInt("Workers.WorkerCount")

	workers := localworker.NewWorkerPool(queueSize, sleepCyc, workerCount, logger)
	workers.Start()
	deps.SetWorkers(workers)

	searchResultLimit := cfg.GrabInt("Server.SearchResultLimit")
	fileIndex := fileindex.NewFileTreeIndex(filesystem, "/", searchResultLimit)
	indexInfo, err := filesystem.Stat(fileIndexPath)
	indexInited := false
	if err != nil {
		if !os.IsNotExist(err) {
			logger.Warnf("failed to detect file index: %s", err)
		} else {
			logger.Warnf("no file index found")
		}
	} else if indexInfo.IsDir() {
		logger.Warnf("file index is folder, not file: %s", fileIndexPath)
	} else {
		err = fileIndex.ReadFrom(fileIndexPath)
		if err != nil {
			logger.Infof("failed to load file index: %s", err)
		} else {
			indexInited = true
		}
	}
	logger.Infof("file index inited(%t)", indexInited)
	deps.SetFileIndex(fileIndex)

	return deps
}

func initDB(cfg gocfg.ICfg, filesystem fs.ISimpleFS) (db.IDBQuickshare, error) {
	dbPath := cfg.GrabString("Db.DbPath")
	dbDir := path.Dir(dbPath)

	sqliteDB, err := sqlite.NewSQLite(path.Join(filesystem.Root(), dbPath))
	if err != nil {
		return nil, fmt.Errorf("failed to create path for db: %w", err)
	}
	dbQuickshare, err := sqlite.NewSQLiteStore(sqliteDB)
	if err != nil {
		return nil, fmt.Errorf("failed to create quickshare db: %w", err)
	}

	inited := true
	_, err = filesystem.Stat(dbPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			inited = false
		} else {
			return nil, fmt.Errorf("failed to stat db: %w", err)
		}
	}

	var ok bool
	var adminName string
	var pwdHash []byte
	if !inited {
		err := filesystem.MkdirAll(dbDir)
		if err != nil {
			return nil, fmt.Errorf("failed to create path for db: %w", err)
		}

		adminName, ok = cfg.String("ENV.DEFAULTADMIN")
		if !ok || adminName == "" {
			fmt.Println("Please input admin name: ")
			fmt.Scanf("%s", &adminName)
		}

		adminPwd, _ := cfg.String("ENV.DEFAULTADMINPWD")
		if adminPwd == "" {
			adminPwd, err = generatePwd()
			if err != nil {
				return nil, fmt.Errorf("generate password error: %w", err)
			}
			fmt.Printf("password is generated: %s, please update it immediately after login\n", adminPwd)
		}

		pwdHash, err = bcrypt.GenerateFromPassword([]byte(adminPwd), 10)
		if err != nil {
			return nil, fmt.Errorf("hashing password error: %w", err)
		}

		cfg.SetString("ENV.DEFAULTADMIN", adminName)
		cfg.SetString("ENV.DEFAULTADMINPWD", string(pwdHash))

		siteCfg := &db.SiteConfig{
			ClientCfg: &db.ClientConfig{
				SiteName: cfg.StringOr("Site.ClientCfg.SiteName", "Quickshare"),
				SiteDesc: cfg.StringOr("Site.ClientCfg.SiteDesc", "Quick and simple file sharing"),
				Bg: &db.BgConfig{
					Url:      cfg.StringOr("Site.ClientCfg.Bg.Url", ""),
					Repeat:   cfg.StringOr("Site.ClientCfg.Bg.Repeat", "repeat"),
					Position: cfg.StringOr("Site.ClientCfg.Bg.Position", "center"),
					Align:    cfg.StringOr("Site.ClientCfg.Bg.Align", "fixed"),
					BgColor:  cfg.StringOr("Site.ClientCfg.Bg.BgColor", ""),
				},
			},
		}
		err = dbQuickshare.Init(context.TODO(), adminName, string(pwdHash), siteCfg)
		if err != nil {
			return nil, fmt.Errorf("failed to init tables: %w %s", err, dbPath)
		}
	}

	return dbQuickshare, nil
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
