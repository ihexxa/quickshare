package depidx

import (
	"github.com/ihexxa/gocfg"
	"github.com/ihexxa/quickshare/src/cryptoutil"
	"github.com/ihexxa/quickshare/src/fs"
	"github.com/ihexxa/quickshare/src/idgen"
	"github.com/ihexxa/quickshare/src/kvstore"
)

type IUploader interface {
	Create(filePath string, size int64) error
	WriteChunk(filePath string, chunk []byte, off int64) (int, error)
	Status(filePath string) (int64, bool, error)
	Close() error
	Sync() error
}

type Deps struct {
	fs       fs.ISimpleFS
	token    cryptoutil.ITokenEncDec
	kv       kvstore.IKVStore
	uploader IUploader
	id       idgen.IIDGen
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
