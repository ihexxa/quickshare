package apis

import (
	"log"
	"net/http"
	"os"
	"strings"
)

import (
	"github.com/ihexxa/quickshare/server/libs/cfg"
	"github.com/ihexxa/quickshare/server/libs/encrypt"
	"github.com/ihexxa/quickshare/server/libs/errutil"
	"github.com/ihexxa/quickshare/server/libs/fileidx"
	"github.com/ihexxa/quickshare/server/libs/fsutil"
	"github.com/ihexxa/quickshare/server/libs/httputil"
	"github.com/ihexxa/quickshare/server/libs/httpworker"
	"github.com/ihexxa/quickshare/server/libs/limiter"
	"github.com/ihexxa/quickshare/server/libs/logutil"
	"github.com/ihexxa/quickshare/server/libs/qtube"
	"github.com/ihexxa/quickshare/server/libs/walls"
)

type AddDep func(*SrvShare)

func NewSrvShare(config *cfg.Config) *SrvShare {
	logger := logutil.NewSlog(os.Stdout, config.AppName)
	setLog := func(srv *SrvShare) {
		srv.Log = logger
	}

	errChecker := errutil.NewErrChecker(!config.Production, logger)
	setErr := func(srv *SrvShare) {
		srv.Err = errChecker
	}

	setWorkerPool := func(srv *SrvShare) {
		workerPoolSize := config.WorkerPoolSize
		taskQueueSize := config.TaskQueueSize
		srv.WorkerPool = httpworker.NewWorkerPool(workerPoolSize, taskQueueSize, logger)
	}

	setWalls := func(srv *SrvShare) {
		encrypterMaker := encrypt.JwtEncrypterMaker
		ipLimiter := limiter.NewRateLimiter(
			config.LimiterCap,
			config.LimiterTtl,
			config.LimiterCyc,
			config.BucketCap,
			config.SpecialCaps,
		)
		opLimiter := limiter.NewRateLimiter(
			config.LimiterCap,
			config.LimiterTtl,
			config.LimiterCyc,
			config.BucketCap,
			config.SpecialCaps,
		)
		srv.Walls = walls.NewAccessWalls(config, ipLimiter, opLimiter, encrypterMaker)
	}

	setIndex := func(srv *SrvShare) {
		srv.Index = fileidx.NewMemFileIndex(config.MaxShares)
	}

	fs := fsutil.NewSimpleFs(errChecker)
	setFs := func(srv *SrvShare) {
		srv.Fs = fs
	}

	setDownloader := func(srv *SrvShare) {
		srv.Downloader = qtube.NewQTube(
			config.PathLocal,
			config.MaxDownBytesPerSec,
			config.MaxRangeLength,
			fs,
		)
	}

	setEncryptor := func(srv *SrvShare) {
		srv.Encryptor = &encrypt.HmacEncryptor{Key: config.SecretKeyByte}
	}

	setHttp := func(srv *SrvShare) {
		srv.Http = &httputil.QHttpUtil{
			CookieDomain:   config.CookieDomain,
			CookieHttpOnly: config.CookieHttpOnly,
			CookieMaxAge:   config.CookieMaxAge,
			CookiePath:     config.CookiePath,
			CookieSecure:   config.CookieSecure,
			Err:            errChecker,
		}
	}

	return InitSrvShare(config, setIndex, setWalls, setWorkerPool, setFs, setDownloader, setEncryptor, setLog, setErr, setHttp)
}

func InitSrvShare(config *cfg.Config, addDeps ...AddDep) *SrvShare {
	srv := &SrvShare{}
	srv.Conf = config
	for _, addDep := range addDeps {
		addDep(srv)
	}

	if !srv.Fs.MkdirAll(srv.Conf.PathLocal, os.FileMode(0775)) {
		panic("fail to make ./files/ folder")
	}

	if res := srv.AddLocalFilesImp(); res != httputil.Ok200 {
		panic("fail to add local files")
	}

	return srv
}

type SrvShare struct {
	Conf       *cfg.Config
	Encryptor  encrypt.Encryptor
	Err        errutil.ErrUtil
	Downloader qtube.Downloader
	Http       httputil.HttpUtil
	Index      fileidx.FileIndex
	Fs         fsutil.FsUtil
	Log        logutil.LogUtil
	Walls      walls.Walls
	WorkerPool httpworker.Workers
}

func (srv *SrvShare) Wrap(serviceFunc httpworker.ServiceFunc) httpworker.DoFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		body := serviceFunc(res, req)

		if body != nil && body != 0 && srv.Http.Fill(body, res) <= 0 {
			log.Println("Wrap: fail to fill body", body, res)
		}
	}
}

func GetRemoteIp(addr string) string {
	addrParts := strings.Split(addr, ":")
	if len(addrParts) > 0 {
		return addrParts[0]
	}
	return "unknown ip"
}
