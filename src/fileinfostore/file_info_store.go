package fileinfostore

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ihexxa/quickshare/src/kvstore"
)

const (
	InitNs      = "Init"
	InfoNs      = "sharing"
	InitTimeKey = "initTime"
)

var (
	ErrNotFound = errors.New("file info not found")
)

func IsNotFound(err error) bool {
	return err == ErrNotFound
}

type FileInfo struct {
	IsDir  bool   `json:"isDir"`
	Shared bool   `json:"shared"`
	Sha1   string `json:"sha1"`
}

type IFileInfoStore interface {
	AddSharing(dirPath string) error
	DelSharing(dirPath string) error
	GetSharing(dirPath string) (bool, bool)
	ListSharings(prefix string) (map[string]bool, error)
	GetInfo(itemPath string) (*FileInfo, error)
	SetInfo(itemPath string, info *FileInfo) error
	DelInfo(itemPath string) error
	SetSha1(itemPath, sign string) error
}

type FileInfoStore struct {
	mtx   *sync.RWMutex
	store kvstore.IKVStore
}

func NewFileInfoStore(store kvstore.IKVStore) (*FileInfoStore, error) {
	_, ok := store.GetStringIn(InitNs, InitTimeKey)
	if !ok {
		var err error
		for _, nsName := range []string{
			InitNs,
			InfoNs,
		} {
			if err = store.AddNamespace(nsName); err != nil {
				return nil, err
			}
		}
	}

	err := store.SetStringIn(InitNs, InitTimeKey, fmt.Sprintf("%d", time.Now().Unix()))
	if err != nil {
		return nil, err
	}

	return &FileInfoStore{
		store: store,
		mtx:   &sync.RWMutex{},
	}, nil
}

func (fi *FileInfoStore) AddSharing(dirPath string) error {
	fi.mtx.Lock()
	defer fi.mtx.Unlock()

	info, err := fi.GetInfo(dirPath)
	if err != nil {
		if !IsNotFound(err) {
			return err
		}
		info = &FileInfo{
			IsDir: true,
		}
	}
	info.Shared = true
	return fi.SetInfo(dirPath, info)
}

func (fi *FileInfoStore) DelSharing(dirPath string) error {
	fi.mtx.Lock()
	defer fi.mtx.Unlock()

	info, err := fi.GetInfo(dirPath)
	if err != nil {
		return err
	}
	info.Shared = false
	return fi.SetInfo(dirPath, info)
}

func (fi *FileInfoStore) GetSharing(dirPath string) (bool, bool) {
	// TODO: add lock
	info, err := fi.GetInfo(dirPath)
	if err != nil {
		// TODO: error is ignored
		return false, false
	}
	return info.IsDir && info.Shared, true
}

func (fi *FileInfoStore) ListSharings(prefix string) (map[string]bool, error) {
	infoStrs, err := fi.store.ListStringsByPrefixIn(prefix, InfoNs)
	if err != nil {
		return nil, err
	}

	info := &FileInfo{}
	sharings := map[string]bool{}
	for itemPath, infoStr := range infoStrs {
		err = json.Unmarshal([]byte(infoStr), info)
		if err != nil {
			return nil, fmt.Errorf("list sharing error: %w", err)
		}
		if info.IsDir && info.Shared {
			sharings[itemPath] = true
		}
	}

	return sharings, nil
}

func (fi *FileInfoStore) GetInfo(itemPath string) (*FileInfo, error) {
	infoStr, ok := fi.store.GetStringIn(InfoNs, itemPath)
	if !ok {
		return nil, ErrNotFound
	}

	info := &FileInfo{}
	err := json.Unmarshal([]byte(infoStr), info)
	if err != nil {
		return nil, fmt.Errorf("get file info: %w", err)
	}
	return info, nil
}

func (fi *FileInfoStore) SetInfo(itemPath string, info *FileInfo) error {
	infoStr, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("set file info: %w", err)
	}

	err = fi.store.SetStringIn(InfoNs, itemPath, string(infoStr))
	if err != nil {
		return fmt.Errorf("set file info: %w", err)
	}
	return nil
}

func (fi *FileInfoStore) DelInfo(itemPath string) error {
	return fi.store.DelStringIn(InfoNs, itemPath)
}

func (fi *FileInfoStore) SetSha1(itemPath, sign string) error {
	fi.mtx.Lock()
	defer fi.mtx.Unlock()

	info, err := fi.GetInfo(itemPath)
	if err != nil {
		if !IsNotFound(err) {
			return err
		}
		info = &FileInfo{
			IsDir:  false,
			Shared: false,
		}
	}
	info.Sha1 = sign
	return fi.SetInfo(itemPath, info)
}
