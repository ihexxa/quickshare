package server

import (
	"context"
	"crypto/rand"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/ihexxa/gocfg"
	"github.com/ihexxa/quickshare/src/db/rdb/sqlite"
	"github.com/ihexxa/quickshare/src/worker"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/crypto/bcrypt"

	"github.com/ihexxa/quickshare/src/cryptoutil"
	"github.com/ihexxa/quickshare/src/cryptoutil/jwt"
	"github.com/ihexxa/quickshare/src/db"
	"github.com/ihexxa/quickshare/src/depidx"
	"github.com/ihexxa/quickshare/src/fs"
	"github.com/ihexxa/quickshare/src/fs/local"
	"github.com/ihexxa/quickshare/src/idgen"
	"github.com/ihexxa/quickshare/src/idgen/simpleidgen"
	"github.com/ihexxa/quickshare/src/iolimiter"
	"github.com/ihexxa/quickshare/src/search/fileindex"
	"github.com/ihexxa/quickshare/src/worker/localworker"
)

type Initer struct {
	cfg          gocfg.ICfg
	input        io.Reader
	output       io.Writer
	onStartHooks []func(cfg gocfg.ICfg) error
}

func NewIniter(cfg gocfg.ICfg) *Initer {
	return &Initer{
		cfg:    cfg,
		input:  os.Stdin,
		output: os.Stdout,
	}
}

func (it *Initer) InitDeps() *depidx.Deps {
	ider := simpleidgen.New()
	logger := it.initLogger()
	jwtEncDec := it.initJWT(logger)
	workers := it.initWorkerPool(logger)
	filesystem, err := it.initFs(ider, logger)
	if err != nil {
		logger.Fatalf("failed to init DB: %s", err)
	}
	quickshareDb, err := it.initDb(filesystem)
	if err != nil {
		logger.Fatalf("failed to init DB: %s", err)
	}
	rateLimiter := it.initRateLimiter(quickshareDb)
	fileIndex := it.initSearchIndex(filesystem, logger)

	deps := depidx.NewDeps(it.cfg)
	deps.SetDB(quickshareDb)
	deps.SetFS(filesystem)
	deps.SetToken(jwtEncDec)
	deps.SetID(ider)
	deps.SetLog(logger)
	deps.SetLimiter(rateLimiter)
	deps.SetWorkers(workers)
	deps.SetFileIndex(fileIndex)

	return deps
}

func (it *Initer) initLogger() *zap.SugaredLogger {
	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   path.Join(it.cfg.GrabString("Fs.Root"), "quickshare.log"),
		MaxSize:    it.cfg.IntOr("Log.MaxSize", 50), // megabytes
		MaxBackups: it.cfg.IntOr("Log.MaxBackups", 2),
		MaxAge:     it.cfg.IntOr("Log.MaxAge", 31), // days
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

func (it *Initer) initJWT(logger *zap.SugaredLogger) cryptoutil.ITokenEncDec {
	secret, ok := it.cfg.String("ENV.TOKENSECRET")
	if !ok {
		b := make([]byte, 32)
		_, err := rand.Read(b)
		if err != nil {
			logger.Fatalf("make rand token error: %s", err)
		}
		secret = string(b)
		logger.Info("warning: TOKENSECRET is not set, a random token is generated")
	}

	return jwt.NewJWTEncDec(secret)
}

func (it *Initer) initRateLimiter(quickshareDb db.IDBQuickshare) iolimiter.ILimiter {
	limiterCap := it.cfg.IntOr("Users.LimiterCapacity", 10000)
	limiterCyc := it.cfg.IntOr("Users.LimiterCyc", 1000)
	return iolimiter.NewIOLimiter(limiterCap, limiterCyc, quickshareDb)
}

func (it *Initer) initWorkerPool(logger *zap.SugaredLogger) worker.IWorkerPool {
	queueSize := it.cfg.GrabInt("Workers.QueueSize")
	sleepCyc := it.cfg.GrabInt("Workers.SleepCyc")
	workerCount := it.cfg.GrabInt("Workers.WorkerCount")

	workers := localworker.NewWorkerPool(queueSize, sleepCyc, workerCount, logger)
	workers.Start()
	return workers
}

func (it *Initer) initSearchIndex(filesystem fs.ISimpleFS, logger *zap.SugaredLogger) fileindex.IFileIndex {
	searchResultLimit := it.cfg.GrabInt("Server.SearchResultLimit")
	fileIndex := fileindex.NewFileTreeIndex(filesystem, "/", searchResultLimit)

	indexInited := false
	indexInfo, err := filesystem.Stat(fileIndexPath)
	if err != nil {
		if !os.IsNotExist(err) {
			logger.Warnf("detect index file error: %s", err)
		} else {
			logger.Warnf("index file not found")
		}
	} else if indexInfo.IsDir() {
		logger.Warnf("file index is a folder, not a file: %s", fileIndexPath)
	} else {
		err = fileIndex.ReadFrom(fileIndexPath)
		if err != nil {
			logger.Warnf("failed to load file index: %s", err)
		} else {
			indexInited = true
		}
	}

	logger.Infof("file index inited(%t)", indexInited)
	return fileIndex
}

func (it *Initer) initFs(idGenerator idgen.IIDGen, logger *zap.SugaredLogger) (fs.ISimpleFS, error) {
	rootPath := it.cfg.GrabString("Fs.Root")
	opensLimit := it.cfg.GrabInt("Fs.OpensLimit")
	openTTL := it.cfg.GrabInt("Fs.OpenTTL")
	readerTTL := it.cfg.GrabInt("Server.WriteTimeout") / 1000 // millisecond -> second

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

	return local.NewLocalFS(rootPath, 0660, opensLimit, openTTL, readerTTL, idGenerator), nil
}

func (it *Initer) initDb(filesystem fs.ISimpleFS) (db.IDBQuickshare, error) {
	dbPath := it.cfg.GrabString("Db.DbPath")
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

		adminName, ok = it.cfg.String("ENV.DEFAULTADMIN")
		if !ok || adminName == "" {
			fmt.Fprintln(it.output, "Please input admin name: ")
			fmt.Fscanf(it.input, "%s", &adminName)
		}

		adminPwd, _ := it.cfg.String("ENV.DEFAULTADMINPWD")
		if adminPwd == "" {
			adminPwd, err = generatePwd()
			if err != nil {
				return nil, fmt.Errorf("generate password error: %w", err)
			}
			fmt.Fprintf(
				it.output,
				"password is generated: %s, please update it immediately after login\n",
				adminPwd,
			)
		}

		pwdHash, err = bcrypt.GenerateFromPassword([]byte(adminPwd), 10)
		if err != nil {
			return nil, fmt.Errorf("hashing password error: %w", err)
		}

		it.cfg.SetString("ENV.DEFAULTADMIN", adminName)
		it.cfg.SetString("ENV.DEFAULTADMINPWD", adminPwd)
		it.cfg.SetString("ENV.DEFAULTADMINPWDHASH", string(pwdHash))

		siteCfg := &db.SiteConfig{
			ClientCfg: &db.ClientConfig{
				SiteName: it.cfg.StringOr("Site.ClientCfg.SiteName", "Quickshare"),
				SiteDesc: it.cfg.StringOr("Site.ClientCfg.SiteDesc", "Quick and simple file sharing"),
				Bg: &db.BgConfig{
					Url:      it.cfg.StringOr("Site.ClientCfg.Bg.Url", ""),
					Repeat:   it.cfg.StringOr("Site.ClientCfg.Bg.Repeat", "repeat"),
					Position: it.cfg.StringOr("Site.ClientCfg.Bg.Position", "center"),
					Align:    it.cfg.StringOr("Site.ClientCfg.Bg.Align", "fixed"),
					BgColor:  it.cfg.StringOr("Site.ClientCfg.Bg.BgColor", ""),
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
