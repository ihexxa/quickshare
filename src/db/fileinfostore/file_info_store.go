package fileinfostore

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/ihexxa/quickshare/src/db"
	"github.com/ihexxa/quickshare/src/kvstore"
)

const (
	InitNs       = "Init"
	InitTimeKey  = "initTime"
	SchemaVerKey = "SchemaVersion"
	SchemaV1     = "v1"
)

var (
	// db.ErrEmpty = errors.New("can not hash empty string")
	// db.ErrFileInfoNotFound        = errors.New("file info not found")
	// db.ErrSharingNotFound = errors.New("sharing id not found")
	// db.ErrConflicted  = errors.New("conflict found in hashing")
	// db.ErrVerNotFound = errors.New("file info schema version not found")
	maxHashingTime = 10
)

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
	mtx   *sync.RWMutex
	store kvstore.IKVStore
}

func NewFileInfoStore(store kvstore.IKVStore) (*FileInfoStore, error) {
	var err error
	for _, nsName := range []string{
		db.FileSchemaNs,
		db.FileInfoNs,
		db.ShareIDNs,
	} {
		if !store.HasNamespace(nsName) {
			if err = store.AddNamespace(nsName); err != nil {
				return nil, err
			}
		}
	}

	fi := &FileInfoStore{
		store: store,
		mtx:   &sync.RWMutex{},
	}
	return fi, nil
}

func (fi *FileInfoStore) getInfo(itemPath string) (*db.FileInfo, error) {
	infoStr, ok := fi.store.GetStringIn(db.FileInfoNs, itemPath)
	if !ok {
		return nil, db.ErrFileInfoNotFound
	}

	info := &db.FileInfo{}
	err := json.Unmarshal([]byte(infoStr), info)
	if err != nil {
		return nil, fmt.Errorf("get file info: %w", err)
	}

	if err = db.CheckFileInfo(info, true); err != nil {
		return nil, err
	}
	return info, nil
}

func (fi *FileInfoStore) GetFileInfo(ctx context.Context, itemPath string) (*db.FileInfo, error) {
	return fi.getInfo(itemPath)
}

func (fi *FileInfoStore) ListFileInfos(ctx context.Context, itemPaths []string) (map[string]*db.FileInfo, error) {
	infos := map[string]*db.FileInfo{}
	for _, itemPath := range itemPaths {
		info, err := fi.getInfo(itemPath)
		if err != nil {
			if !errors.Is(err, db.ErrFileInfoNotFound) {
				// TODO: try to make info data consistent with fs
				return nil, err
			}
			continue
		}
		if err = db.CheckFileInfo(info, true); err != nil {
			return nil, err
		}
		infos[itemPath] = info
	}

	return infos, nil
}

func (fi *FileInfoStore) setInfo(itemPath string, info *db.FileInfo) error {
	if err := db.CheckFileInfo(info, false); err != nil {
		return err
	}

	infoStr, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("set file info: %w", err)
	}

	err = fi.store.SetStringIn(db.FileInfoNs, itemPath, string(infoStr))
	if err != nil {
		return fmt.Errorf("set file info: %w", err)
	}
	return nil
}

func (fi *FileInfoStore) SetFileInfo(ctx context.Context, itemPath string, info *db.FileInfo) error {
	return fi.setInfo(itemPath, info)
}

func (fi *FileInfoStore) DelFileInfo(ctx context.Context, itemPath string) error {
	return fi.store.DelStringIn(db.FileInfoNs, itemPath)
}

// sharings

func (fi *FileInfoStore) SetSha1(ctx context.Context, itemPath, sign string) error {
	fi.mtx.Lock()
	defer fi.mtx.Unlock()

	info, err := fi.getInfo(itemPath)
	if err != nil {
		if !errors.Is(err, db.ErrFileInfoNotFound) {
			return err
		}
		info = &db.FileInfo{
			IsDir:  false,
			Shared: false,
		}
	}
	info.Sha1 = sign
	return fi.setInfo(itemPath, info)
}

func (fi *FileInfoStore) getShareID(payload string) (string, error) {
	if len(payload) == 0 {
		return "", db.ErrEmpty
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

	return "", db.ErrConflicted
}

func (fi *FileInfoStore) GetSharingDir(ctx context.Context, hashID string) (string, error) {
	dirPath, ok := fi.store.GetStringIn(db.ShareIDNs, hashID)
	if !ok {
		return "", db.ErrSharingNotFound
	}
	return dirPath, nil
}

func (fi *FileInfoStore) AddSharing(ctx context.Context, dirPath string) error {
	fi.mtx.Lock()
	defer fi.mtx.Unlock()

	info, err := fi.getInfo(dirPath)
	if err != nil {
		if !errors.Is(err, db.ErrFileInfoNotFound) {
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
	return fi.setInfo(dirPath, info)
}

func (fi *FileInfoStore) DelSharing(ctx context.Context, dirPath string) error {
	fi.mtx.Lock()
	defer fi.mtx.Unlock()

	info, err := fi.getInfo(dirPath)
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
	return fi.setInfo(dirPath, info)
}

func (fi *FileInfoStore) GetSharing(ctx context.Context, dirPath string) (bool, bool) {
	fi.mtx.Lock()
	defer fi.mtx.Unlock()

	// TODO: differentiate error and not exist
	info, err := fi.getInfo(dirPath)
	if err != nil {
		return false, false
	}
	return info.IsDir && info.Shared, true
}

func (fi *FileInfoStore) ListSharings(ctx context.Context, prefix string) (map[string]string, error) {
	infoStrs, err := fi.store.ListStringsByPrefixIn(prefix, db.FileInfoNs)
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
