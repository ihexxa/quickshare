package depidx

import (
	"github.com/ihexxa/gocfg"
	"go.uber.org/zap"

	"github.com/ihexxa/quickshare/src/cron"
	"github.com/ihexxa/quickshare/src/cryptoutil"
	"github.com/ihexxa/quickshare/src/db/boltstore"
	"github.com/ihexxa/quickshare/src/db/fileinfostore"
	"github.com/ihexxa/quickshare/src/db/sitestore"
	"github.com/ihexxa/quickshare/src/db/userstore"
	"github.com/ihexxa/quickshare/src/fs"
	"github.com/ihexxa/quickshare/src/idgen"
	"github.com/ihexxa/quickshare/src/iolimiter"
	"github.com/ihexxa/quickshare/src/kvstore"
	"github.com/ihexxa/quickshare/src/search/fileindex"
	"github.com/ihexxa/quickshare/src/worker"
)

type IUploader interface {
	Create(filePath string, size int64) error
	WriteChunk(filePath string, chunk []byte, off int64) (int, error)
	Status(filePath string) (int64, bool, error)
	Close() error
	Sync() error
}

type Deps struct {
	fs        fs.ISimpleFS
	token     cryptoutil.ITokenEncDec
	kv        kvstore.IKVStore
	users     userstore.IUserStore
	fileInfos fileinfostore.IFileInfoStore
	siteStore sitestore.ISiteStore
	id        idgen.IIDGen
	logger    *zap.SugaredLogger
	limiter   iolimiter.ILimiter
	workers   worker.IWorkerPool
	boltStore *boltstore.BoltStore
	cron      cron.ICron
	fileIndex fileindex.IFileIndex
}

func NewDeps(cfg gocfg.ICfg) *Deps {
	return &Deps{}
}

func (deps *Deps) FS() fs.ISimpleFS {
	return deps.fs
}

func (deps *Deps) SetFS(filesystem fs.ISimpleFS) {
	deps.fs = filesystem
}

func (deps *Deps) Token() cryptoutil.ITokenEncDec {
	return deps.token
}

func (deps *Deps) SetToken(tokenMaker cryptoutil.ITokenEncDec) {
	deps.token = tokenMaker
}

func (deps *Deps) KV() kvstore.IKVStore {
	return deps.kv
}

func (deps *Deps) SetKV(kvstore kvstore.IKVStore) {
	deps.kv = kvstore
}

func (deps *Deps) ID() idgen.IIDGen {
	return deps.id
}

func (deps *Deps) SetID(ider idgen.IIDGen) {
	deps.id = ider
}

func (deps *Deps) Log() *zap.SugaredLogger {
	return deps.logger
}

func (deps *Deps) SetLog(logger *zap.SugaredLogger) {
	deps.logger = logger
}

func (deps *Deps) Users() userstore.IUserStore {
	return deps.users
}

func (deps *Deps) SetUsers(users userstore.IUserStore) {
	deps.users = users
}

func (deps *Deps) FileInfos() fileinfostore.IFileInfoStore {
	return deps.fileInfos
}

func (deps *Deps) SetFileInfos(fileInfos fileinfostore.IFileInfoStore) {
	deps.fileInfos = fileInfos
}

func (deps *Deps) SiteStore() sitestore.ISiteStore {
	return deps.siteStore
}

func (deps *Deps) SetSiteStore(siteStore sitestore.ISiteStore) {
	deps.siteStore = siteStore
}

func (deps *Deps) Limiter() iolimiter.ILimiter {
	return deps.limiter
}

func (deps *Deps) SetLimiter(limiter iolimiter.ILimiter) {
	deps.limiter = limiter
}

func (deps *Deps) Workers() worker.IWorkerPool {
	return deps.workers
}

func (deps *Deps) SetWorkers(workers worker.IWorkerPool) {
	deps.workers = workers
}

func (deps *Deps) BoltStore() *boltstore.BoltStore {
	return deps.boltStore
}

func (deps *Deps) SetBoltStore(boltStore *boltstore.BoltStore) {
	deps.boltStore = boltStore
}

func (deps *Deps) Cron() cron.ICron {
	return deps.cron
}

func (deps *Deps) SetCron(cronImp cron.ICron) {
	deps.cron = cronImp
}

func (deps *Deps) FileIndex() fileindex.IFileIndex {
	return deps.fileIndex
}

func (deps *Deps) SetIFileIndex(index fileindex.IFileIndex) {
	deps.fileIndex = index
}
