package uploadmgr

import (
	"errors"
	"fmt"
	"path"

	"github.com/ihexxa/quickshare/src/depidx"
)

// TODO:
// uploading resumption test
// rename file after uploaded
// differetiate file and dir

var ErrBadData = errors.New("file size or uploaded not found for a file")
var ErrUploaded = errors.New("file already uploaded")
var ErrWriteUploaded = errors.New("try to write acknowledge part")

type UploadMgr struct {
	deps *depidx.Deps
}

func NewUploadMgr(deps *depidx.Deps) (*UploadMgr, error) {
	if deps.KV() == nil {
		return nil, errors.New("kvstore is not found in deps")
	}
	if deps.FS() == nil {
		return nil, errors.New("fs is not found in deps")
	}

	return &UploadMgr{
		deps: deps,
	}, nil
}

func fileSizeKey(filePath string) string     { return fmt.Sprintf("%s:size", filePath) }
func fileUploadedKey(filePath string) string { return fmt.Sprintf("%s:uploaded", filePath) }

func (mgr *UploadMgr) Create(filePath string, size int64) error {
	// _, found := mgr.deps.KV().GetBool(filePath)
	// if found {
	// 	return os.ErrExist
	// }

	dirPath := path.Dir(filePath)
	if dirPath != "" {
		err := mgr.deps.FS().MkdirAll(dirPath)
		if err != nil {
			return err
		}
	}

	err := mgr.deps.FS().Create(filePath)
	if err != nil {
		return err
	}

	// mgr.deps.KV().SetBool(filePath, true)
	// mgr.deps.KV().SetInt64(fileSizeKey(filePath), size)
	// mgr.deps.KV().SetInt64(fileUploadedKey(filePath), 0)
	return nil
}

func (mgr *UploadMgr) WriteChunk(filePath string, chunk []byte, off int64) (int, error) {
	// _, found := mgr.deps.KV().GetBool(filePath)
	// if !found {
	// 	return 0, os.ErrNotExist
	// }

	// fileSize, ok1 := mgr.deps.KV().GetInt64(fileSizeKey(filePath))
	// uploaded, ok2 := mgr.deps.KV().GetInt64(fileUploadedKey(filePath))
	// if !ok1 || !ok2 {
	// 	return 0, ErrBadData
	// } else if uploaded == fileSize {
	// 	return 0, ErrUploaded
	// } else if off != uploaded {
	// 	return 0, ErrWriteUploaded
	// }

	wrote, err := mgr.deps.FS().WriteAt(filePath, chunk, off)
	if err != nil {
		return wrote, err
	}

	// mgr.deps.KV().SetInt64(fileUploadedKey(filePath), off+int64(wrote))
	return wrote, nil
}

func (mgr *UploadMgr) Status(filePath string) (int64, bool, error) {
	// _, found := mgr.deps.KV().GetBool(filePath)
	// if !found {
	// 	return 0, false, os.ErrNotExist
	// }

	fileSize, ok1 := mgr.deps.KV().GetInt64(fileSizeKey(filePath))
	fileUploaded, ok2 := mgr.deps.KV().GetInt64(fileUploadedKey(filePath))
	if !ok1 || !ok2 {
		return 0, false, ErrBadData
	}
	return fileUploaded, fileSize == fileUploaded, nil
}

func (mgr *UploadMgr) Close() error {
	return mgr.deps.FS().Close()
}

func (mgr *UploadMgr) Sync() error {
	return mgr.deps.FS().Sync()
}
