package fileinfostore

import (
	"encoding/json"
	"errors"

	"github.com/ihexxa/quickshare/src/db"
)

var (
	ErrGreaterThanSize = errors.New("uploaded is greater than file size")
	ErrUploadNotFound  = errors.New("upload info not found")
)

func (fi *FileInfoStore) getUploadInfo(user, filePath string) (string, int64, int64, error) {
	ns := db.UploadNS(user)
	infoBytes, ok := fi.store.GetStringIn(ns, filePath)
	if !ok {
		return "", 0, 0, ErrUploadNotFound
	}

	info := &db.UploadInfo{}
	err := json.Unmarshal([]byte(infoBytes), info)
	if err != nil {
		return "", 0, 0, err
	}

	return info.RealFilePath, info.Size, info.Uploaded, nil
}

func (fi *FileInfoStore) setUploadInfo(user, filePath string, info *db.UploadInfo) error {
	newInfoBytes, err := json.Marshal(info)
	if err != nil {
		return err
	}
	return fi.store.SetStringIn(db.UploadNS(user), filePath, string(newInfoBytes))
}

func (fi *FileInfoStore) AddUploadInfo(user, filePath, tmpPath string, fileSize int64) error {
	fi.mtx.Lock()
	defer fi.mtx.Unlock()

	ns := db.UploadNS(user)
	err := fi.store.AddNamespace(ns)
	if err != nil {
		return err
	}

	_, _, _, err = fi.getUploadInfo(user, tmpPath)
	if err == nil {
		return db.ErrCreateExisting
	}

	return fi.setUploadInfo(user, filePath, &db.UploadInfo{
		RealFilePath: filePath,
		Size:         fileSize,
		Uploaded:     0,
	})
}

func (fi *FileInfoStore) SetUploadInfo(user, filePath string, newUploaded int64) error {
	fi.mtx.Lock()
	defer fi.mtx.Unlock()

	realFilePath, fileSize, _, err := fi.getUploadInfo(user, filePath)
	if err != nil {
		return err
	} else if newUploaded > fileSize {
		return ErrGreaterThanSize
	}

	return fi.setUploadInfo(user, filePath, &db.UploadInfo{
		RealFilePath: realFilePath,
		Size:         fileSize,
		Uploaded:     newUploaded,
	})
}

func (fi *FileInfoStore) GetUploadInfo(user, filePath string) (string, int64, int64, error) {
	fi.mtx.Lock()
	defer fi.mtx.Unlock()

	return fi.getUploadInfo(user, filePath)
}

func (fi *FileInfoStore) DelUploadInfo(user, filePath string) error {
	return fi.store.DelInt64In(db.UploadNS(user), filePath)
}

func (fi *FileInfoStore) ListUploadInfo(user string) ([]*db.UploadInfo, error) {
	ns := db.UploadNS(user)
	if !fi.store.HasNamespace(ns) {
		return nil, nil
	}

	infoMap, err := fi.store.ListStringsIn(ns)
	if err != nil {
		return nil, err
	}

	infos := []*db.UploadInfo{}
	for _, infoStr := range infoMap {
		info := &db.UploadInfo{}
		err = json.Unmarshal([]byte(infoStr), info)
		if err != nil {
			return nil, err
		}
		infos = append(infos, info)
	}

	return infos, nil
}
