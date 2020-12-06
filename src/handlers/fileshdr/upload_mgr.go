package fileshdr

import (
	"errors"
	"fmt"
	"os"

	"github.com/ihexxa/quickshare/src/kvstore"
)

var (
	isDirKey    = "isDir"
	fileSizeKey = "fileSize"
	uploadedKey = "uploaded"
	filePathKey = "fileName"
)

type UploadMgr struct {
	kv kvstore.IKVStore
}

func NewUploadMgr(kv kvstore.IKVStore) *UploadMgr {
	return &UploadMgr{
		kv: kv,
	}
}

func (um *UploadMgr) AddInfo(fileName, tmpName string, fileSize int64, isDir bool) error {
	err := um.kv.SetInt64(infoKey(tmpName, fileSizeKey), fileSize)
	if err != nil {
		return err
	}
	err = um.kv.SetInt64(infoKey(tmpName, uploadedKey), 0)
	if err != nil {
		return err
	}
	return um.kv.SetString(infoKey(tmpName, filePathKey), fileName)
}

func (um *UploadMgr) IncreUploaded(fileName string, newUploaded int64) error {
	fileSize, ok := um.kv.GetInt64(infoKey(fileName, fileSizeKey))
	if !ok {
		return fmt.Errorf("file size %s not found", fileName)
	}
	preUploaded, ok := um.kv.GetInt64(infoKey(fileName, uploadedKey))
	if !ok {
		return fmt.Errorf("file uploaded %s not found", fileName)
	}
	if newUploaded+preUploaded <= fileSize {
		um.kv.SetInt64(infoKey(fileName, uploadedKey), newUploaded+preUploaded)
		return nil
	}
	return errors.New("uploaded is greater than file size")
}

func (um *UploadMgr) GetInfo(fileName string) (string, int64, int64, error) {
	realFilePath, ok := um.kv.GetString(infoKey(fileName, filePathKey))
	if !ok {
		return "", 0, 0, os.ErrNotExist
	}
	fileSize, ok := um.kv.GetInt64(infoKey(fileName, fileSizeKey))
	if !ok {
		return "", 0, 0, os.ErrNotExist
	}
	uploaded, ok := um.kv.GetInt64(infoKey(fileName, uploadedKey))
	if !ok {
		return "", 0, 0, os.ErrNotExist
	}

	return realFilePath, fileSize, uploaded, nil
}

func (um *UploadMgr) DelInfo(fileName string) error {
	if err := um.kv.DelInt64(infoKey(fileName, fileSizeKey)); err != nil {
		return err
	}
	if err := um.kv.DelInt64(infoKey(fileName, uploadedKey)); err != nil {
		return err
	}
	return um.kv.DelString(infoKey(fileName, filePathKey))
}

func infoKey(fileName, key string) string {
	return fmt.Sprintf("%s:%s", fileName, key)
}
