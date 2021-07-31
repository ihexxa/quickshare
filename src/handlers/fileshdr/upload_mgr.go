package fileshdr

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ihexxa/quickshare/src/kvstore"
)

var (
	ErrCreateExisting  = errors.New("create upload info which already exists")
	ErrGreaterThanSize = errors.New("uploaded is greater than file size")
	ErrNotFound        = errors.New("upload info not found")

	uploadsPrefix = "uploads"
)

type UploadInfo struct {
	RealFilePath string `json:"realFilePath"`
	Size         int64  `json:"size"`
	Uploaded     int64  `json:"uploaded"`
}

type UploadMgr struct {
	kv kvstore.IKVStore
}

func UploadNS(user string) string {
	return fmt.Sprintf("%s/%s", uploadsPrefix, user)
}

func NewUploadMgr(kv kvstore.IKVStore) *UploadMgr {
	return &UploadMgr{
		kv: kv,
	}
}

func (um *UploadMgr) AddInfo(user, filePath, tmpPath string, fileSize int64) error {
	ns := UploadNS(user)
	err := um.kv.AddNamespace(ns)
	if err != nil {
		return err
	}

	_, ok := um.kv.GetStringIn(ns, tmpPath)
	if ok {
		return ErrCreateExisting
	}

	info := &UploadInfo{
		RealFilePath: filePath,
		Size:         fileSize,
		Uploaded:     0,
	}
	infoBytes, err := json.Marshal(info)
	if err != nil {
		return err
	}

	return um.kv.SetStringIn(ns, tmpPath, string(infoBytes))
}

func (um *UploadMgr) SetInfo(user, filePath string, newUploaded int64) error {
	realFilePath, fileSize, _, err := um.GetInfo(user, filePath)
	if err != nil {
		return err
	} else if newUploaded > fileSize {
		return ErrGreaterThanSize
	}

	newInfo := &UploadInfo{
		RealFilePath: realFilePath,
		Size:         fileSize,
		Uploaded:     newUploaded,
	}
	newInfoBytes, err := json.Marshal(newInfo)
	if err != nil {
		return err
	}
	return um.kv.SetStringIn(UploadNS(user), filePath, string(newInfoBytes))
}

func (um *UploadMgr) GetInfo(user, filePath string) (string, int64, int64, error) {
	ns := UploadNS(user)
	infoBytes, ok := um.kv.GetStringIn(ns, filePath)
	if !ok {
		return "", 0, 0, ErrNotFound
	}

	info := &UploadInfo{}
	err := json.Unmarshal([]byte(infoBytes), info)
	if err != nil {
		return "", 0, 0, err
	}

	return info.RealFilePath, info.Size, info.Uploaded, nil
}

func (um *UploadMgr) DelInfo(user, filePath string) error {
	return um.kv.DelInt64In(UploadNS(user), filePath)
}

func (um *UploadMgr) ListInfo(user string) ([]*UploadInfo, error) {
	ns := UploadNS(user)
	if !um.kv.HasNamespace(ns) {
		return nil, nil
	}

	infoMap, err := um.kv.ListStringsIn(ns)
	if err != nil {
		return nil, err
	}

	infos := []*UploadInfo{}
	for _, infoStr := range infoMap {
		info := &UploadInfo{}
		err = json.Unmarshal([]byte(infoStr), info)
		if err != nil {
			return nil, err
		}
		infos = append(infos, info)
	}

	return infos, nil
}
