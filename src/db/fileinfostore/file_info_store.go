package fileinfostore

import (
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/ihexxa/quickshare/src/kvstore"
)

const (
	InitNs       = "Init"
	InfoNs       = "sharing"
	ShareIDNs    = "sharingKey"
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

type FileInfo struct {
	IsDir   bool   `json:"isDir"`
	Shared  bool   `json:"shared"`
	ShareID string `json:"shareID"` // for short url
	Sha1    string `json:"sha1"`
}

type IFileInfoStore interface {
	AddSharing(dirPath string) error
	DelSharing(dirPath string) error
	GetSharing(dirPath string) (bool, bool)
	ListSharings(prefix string) (map[string]string, error)
	GetInfo(itemPath string) (*FileInfo, error)
	SetInfo(itemPath string, info *FileInfo) error
	DelInfo(itemPath string) error
	SetSha1(itemPath, sign string) error
	GetInfos(itemPaths []string) (map[string]*FileInfo, error)
	GetSharingDir(hashID string) (string, error)
}

type FileInfoStore struct {
	mtx   *sync.RWMutex
	store kvstore.IKVStore
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
		infoStrs, err := fi.store.ListStringsIn(InfoNs)
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

				err = fi.store.SetStringIn(ShareIDNs, shareID, itemPath)
				if err != nil {
					return err
				}
			}

			newInfo := &FileInfo{
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
		// no op
	default:
		return fmt.Errorf("file info: unknown schema version (%s)", ver)
	}

	return nil
}

func NewFileInfoStore(store kvstore.IKVStore) (*FileInfoStore, error) {
	var err error
	for _, nsName := range []string{
		InitNs,
		InfoNs,
		ShareIDNs,
	} {
		if !store.HasNamespace(nsName) {
			if err = store.AddNamespace(nsName); err != nil {
				return nil, err
			}
		}
	}

	// err = store.SetStringIn(InitNs, SchemaVerKey, SchemaV1)
	// if err != nil {
	// 	return nil, err
	// }

	fi := &FileInfoStore{
		store: store,
		mtx:   &sync.RWMutex{},
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
		info = &FileInfo{
			IsDir: true,
		}
	}

	// TODO: ensure Atomicity
	shareID, err := fi.getShareID(dirPath)
	if err != nil {
		return err
	}
	err = fi.store.SetStringIn(ShareIDNs, shareID, dirPath)
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
	info.Shared = false
	info.ShareID = ""

	// TODO: ensure Atomicity
	// In the bolt, if the key does not exist
	// then nothing is done and a nil error is returned
	err = fi.store.DelStringIn(ShareIDNs, info.ShareID)
	if err != nil {
		return err
	}

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
	infoStrs, err := fi.store.ListStringsByPrefixIn(prefix, InfoNs)
	if err != nil {
		return nil, err
	}

	info := &FileInfo{}
	sharings := map[string]string{}
	for itemPath, infoStr := range infoStrs {
		err = json.Unmarshal([]byte(infoStr), info)
		if err != nil {
			return nil, fmt.Errorf("list sharing error: %w", err)
		}

		fmt.Println(infoStr)
		if info.IsDir && info.Shared {
			sharings[itemPath] = info.ShareID
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

func (fi *FileInfoStore) GetInfos(itemPaths []string) (map[string]*FileInfo, error) {
	infos := map[string]*FileInfo{}
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
		if _, ok := fi.store.GetStringIn(ShareIDNs, shareID); !ok {
			return shareID, nil
		}
	}

	return "", ErrConflicted
}

func (fi *FileInfoStore) GetSharingDir(hashID string) (string, error) {
	dirPath, ok := fi.store.GetStringIn(ShareIDNs, hashID)
	if !ok {
		return "", ErrSharingNotFound
	}
	return dirPath, nil
}
