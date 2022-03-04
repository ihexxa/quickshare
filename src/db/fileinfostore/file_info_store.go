package fileinfostore

import (
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/ihexxa/quickshare/src/db"
	"github.com/ihexxa/quickshare/src/kvstore"
	"github.com/ihexxa/quickshare/src/kvstore/boltdbpvd"
)

const (
	InitNs       = "Init"
	InitTimeKey  = "initTime"
	SchemaVerKey = "SchemaVersion"
	SchemaV1     = "v1"
)

var (
	ErrEmpty           = errors.New("can not hash empty string")
	ErrNotFound        = errors.New("file info not found")
	ErrSharingNotFound = errors.New("sharing id not found")
	ErrConflicted      = errors.New("conflict found in hashing")
	ErrVerNotFound     = errors.New("file info schema version not found")
	maxHashingTime     = 10
)

func IsNotFound(err error) bool {
	return err == ErrNotFound
}

type IFileInfoStore interface {
	AddSharing(dirPath string) error
	DelSharing(dirPath string) error
	GetSharing(dirPath string) (bool, bool)
	ListSharings(prefix string) (map[string]string, error)
	GetInfo(itemPath string) (*db.FileInfo, error)
	SetInfo(itemPath string, info *db.FileInfo) error
	DelInfo(itemPath string) error
	SetSha1(itemPath, sign string) error
	GetInfos(itemPaths []string) (map[string]*db.FileInfo, error)
	GetSharingDir(hashID string) (string, error)
	// upload info
	AddUploadInfo(user, filePath, tmpPath string, fileSize int64) error
	SetUploadInfo(user, filePath string, newUploaded int64) error
	GetUploadInfo(user, filePath string) (string, int64, int64, error)
	DelUploadInfo(user, filePath string) error
	ListUploadInfo(user string) ([]*db.UploadInfo, error)
}

type FileInfoStore struct {
	mtx    *sync.RWMutex
	store  kvstore.IKVStore
	boltdb boltdbpvd.BoltProvider
}

func migrate(fi *FileInfoStore) error {
	ver := "v0"
	schemaVer, ok := fi.store.GetStringIn(InitNs, SchemaVerKey)
	if ok {
		ver = schemaVer
	}

	switch ver {
	case "v0":
		// add ShareID to FileInfos
		infoStrs, err := fi.store.ListStringsIn(db.InfoNs)
		if err != nil {
			return err
		}

		type FileInfoV0 struct {
			IsDir  bool   `json:"isDir"`
			Shared bool   `json:"shared"`
			Sha1   string `json:"sha1"`
		}

		infoV0 := &FileInfoV0{}
		for itemPath, infoStr := range infoStrs {
			err = json.Unmarshal([]byte(infoStr), infoV0)
			if err != nil {
				return fmt.Errorf("list sharing error: %w", err)
			}

			shareID := ""
			if infoV0.IsDir && infoV0.Shared {
				dirShareID, err := fi.getShareID(itemPath)
				if err != nil {
					return err
				}
				shareID = dirShareID

				err = fi.store.SetStringIn(db.ShareIDNs, shareID, itemPath)
				if err != nil {
					return err
				}
			}

			newInfo := &db.FileInfo{
				IsDir:   infoV0.IsDir,
				Shared:  infoV0.Shared,
				ShareID: shareID,
				Sha1:    infoV0.Sha1,
			}
			if err = fi.SetInfo(itemPath, newInfo); err != nil {
				return err
			}
		}

		err = fi.store.SetStringIn(InitNs, SchemaVerKey, SchemaV1)
		if err != nil {
			return err
		}
	case "v1":
		// add size to file info
		infoStrs, err := fi.store.ListStringsIn(db.InfoNs)
		if err != nil {
			return err
		}

		type FileInfoV1 struct {
			IsDir   bool   `json:"isDir"`
			Shared  bool   `json:"shared"`
			Sha1    string `json:"sha1"`
			ShareID string `json:"shareID"` // for short url
		}

		infoV1 := &FileInfoV1{}
		for itemPath, infoStr := range infoStrs {
			err = json.Unmarshal([]byte(infoStr), infoV1)
			if err != nil {
				return fmt.Errorf("list sharing error: %w", err)
			}

			newInfo := &db.FileInfo{
				IsDir:   infoV1.IsDir,
				Shared:  infoV1.Shared,
				ShareID: infoV1.ShareID,
				Sha1:    infoV1.Sha1,
				Size:    0, // need to run an async task to refresh this
			}
			if err = fi.SetInfo(itemPath, newInfo); err != nil {
				return err
			}
		}

		err = fi.store.SetStringIn(InitNs, SchemaVerKey, db.SchemaV2)
		if err != nil {
			return err
		}
	case db.SchemaV2:
		// no need to migrate
	default:
		return fmt.Errorf("file info: unknown schema version (%s)", ver)
	}

	return nil
}

func NewFileInfoStore(store kvstore.IKVStore) (*FileInfoStore, error) {
	var err error
	for _, nsName := range []string{
		InitNs,
		db.InfoNs,
		db.ShareIDNs,
	} {
		if !store.HasNamespace(nsName) {
			if err = store.AddNamespace(nsName); err != nil {
				return nil, err
			}
		}
	}

	boltdb := store.(boltdbpvd.BoltProvider)

	fi := &FileInfoStore{
		store:  store,
		boltdb: boltdb,
		mtx:    &sync.RWMutex{},
	}
	if err = migrate(fi); err != nil {
		return nil, err
	}
	return fi, nil
}

func (fi *FileInfoStore) AddSharing(dirPath string) error {
	fi.mtx.Lock()
	defer fi.mtx.Unlock()

	info, err := fi.GetInfo(dirPath)
	if err != nil {
		if !IsNotFound(err) {
			return err
		}
		info = &db.FileInfo{
			IsDir: true,
		}
	}

	// TODO: ensure Atomicity
	shareID, err := fi.getShareID(dirPath)
	if err != nil {
		return err
	}
	err = fi.store.SetStringIn(db.ShareIDNs, shareID, dirPath)
	if err != nil {
		return err
	}

	info.Shared = true
	info.ShareID = shareID
	return fi.SetInfo(dirPath, info)
}

func (fi *FileInfoStore) DelSharing(dirPath string) error {
	fi.mtx.Lock()
	defer fi.mtx.Unlock()

	info, err := fi.GetInfo(dirPath)
	if err != nil {
		return err
	}

	// TODO: ensure Atomicity
	// In the bolt, if the key does not exist
	// then nothing is done and a nil error is returned

	// because before this version, shareIDs are not removed correctly
	// so it iterates all shareIDs and cleans remaining entries
	shareIDtoDir, err := fi.store.ListStringsIn(db.ShareIDNs)
	if err != nil {
		return err
	}

	for shareID, shareDir := range shareIDtoDir {
		if shareDir == dirPath {
			err = fi.store.DelStringIn(db.ShareIDNs, shareID)
			if err != nil {
				return err
			}
		}
	}

	info.ShareID = ""
	info.Shared = false
	return fi.SetInfo(dirPath, info)
}

func (fi *FileInfoStore) GetSharing(dirPath string) (bool, bool) {
	fi.mtx.Lock()
	defer fi.mtx.Unlock()

	// TODO: differentiate error and not exist
	info, err := fi.GetInfo(dirPath)
	if err != nil {
		return false, false
	}
	return info.IsDir && info.Shared, true
}

func (fi *FileInfoStore) ListSharings(prefix string) (map[string]string, error) {
	infoStrs, err := fi.store.ListStringsByPrefixIn(prefix, db.InfoNs)
	if err != nil {
		return nil, err
	}

	info := &db.FileInfo{}
	sharings := map[string]string{}
	for itemPath, infoStr := range infoStrs {
		err = json.Unmarshal([]byte(infoStr), info)
		if err != nil {
			return nil, fmt.Errorf("list sharing error: %w", err)
		}

		if info.IsDir && info.Shared {
			sharings[itemPath] = info.ShareID
		}
	}

	return sharings, nil
}

func (fi *FileInfoStore) GetInfo(itemPath string) (*db.FileInfo, error) {
	infoStr, ok := fi.store.GetStringIn(db.InfoNs, itemPath)
	if !ok {
		return nil, ErrNotFound
	}

	info := &db.FileInfo{}
	err := json.Unmarshal([]byte(infoStr), info)
	if err != nil {
		return nil, fmt.Errorf("get file info: %w", err)
	}
	return info, nil
}

func (fi *FileInfoStore) GetInfos(itemPaths []string) (map[string]*db.FileInfo, error) {
	infos := map[string]*db.FileInfo{}
	for _, itemPath := range itemPaths {
		info, err := fi.GetInfo(itemPath)
		if err != nil {
			if !IsNotFound(err) {
				return nil, err
			}
			continue
		}
		infos[itemPath] = info
	}

	return infos, nil
}

func (fi *FileInfoStore) SetInfo(itemPath string, info *db.FileInfo) error {
	infoStr, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("set file info: %w", err)
	}

	err = fi.store.SetStringIn(db.InfoNs, itemPath, string(infoStr))
	if err != nil {
		return fmt.Errorf("set file info: %w", err)
	}
	return nil
}

func (fi *FileInfoStore) DelInfo(itemPath string) error {
	return fi.store.DelStringIn(db.InfoNs, itemPath)
}

func (fi *FileInfoStore) SetSha1(itemPath, sign string) error {
	fi.mtx.Lock()
	defer fi.mtx.Unlock()

	info, err := fi.GetInfo(itemPath)
	if err != nil {
		if !IsNotFound(err) {
			return err
		}
		info = &db.FileInfo{
			IsDir:  false,
			Shared: false,
		}
	}
	info.Sha1 = sign
	return fi.SetInfo(itemPath, info)
}

func (fi *FileInfoStore) getShareID(payload string) (string, error) {
	if len(payload) == 0 {
		return "", ErrEmpty
	}

	for i := 0; i < maxHashingTime; i++ {
		msg := strings.Repeat(payload, i+1)
		h := sha1.New()
		_, err := io.WriteString(h, msg)
		if err != nil {
			return "", err
		}

		shareID := fmt.Sprintf("%x", h.Sum(nil))[:7]
		shareDir, ok := fi.store.GetStringIn(db.ShareIDNs, shareID)
		if !ok {
			return shareID, nil
		} else if ok && shareDir == payload {
			return shareID, nil
		}
	}

	return "", ErrConflicted
}

func (fi *FileInfoStore) GetSharingDir(hashID string) (string, error) {
	dirPath, ok := fi.store.GetStringIn(db.ShareIDNs, hashID)
	if !ok {
		return "", ErrSharingNotFound
	}
	return dirPath, nil
}
