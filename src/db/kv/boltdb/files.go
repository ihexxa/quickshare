package boltdb

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/boltdb/bolt"
	"github.com/ihexxa/quickshare/src/db"
)

const (
	InitNs       = "Init"
	InitTimeKey  = "initTime"
	SchemaVerKey = "SchemaVersion"
	SchemaV1     = "v1"
)

var (
	// ErrEmpty = errors.New("can not hash empty string")
	// ErrNotFound        = errors.New("file info not found")
	// ErrSharingNotFound = errors.New("sharing id not found")
	// ErrConflicted      = errors.New("conflict found in hashing")
	// ErrVerNotFound     = errors.New("file info schema version not found")
	maxHashingTime = 10
)

// type IFileInfoStore interface {
// 	AddSharing(dirPath string) error
// 	DelSharing(dirPath string) error
// 	GetSharing(dirPath string) (bool, bool)
// 	ListSharings(prefix string) (map[string]string, error)
// 	GetInfo(itemPath string) (*db.FileInfo, error)
// 	SetInfo(itemPath string, info *db.FileInfo) error
// 	DelInfo(itemPath string) error
// 	SetSha1(itemPath, sign string) error
// 	GetInfos(itemPaths []string) (map[string]*db.FileInfo, error)
// 	GetSharingDir(hashID string) (string, error)
// 	// upload info
// 	AddUploadInfo(user, filePath, tmpPath string, fileSize int64) error
// 	SetUploadInfo(user, filePath string, newUploaded int64) error
// 	GetUploadInfo(user, filePath string) (string, int64, int64, error)
// 	DelUploadInfo(user, filePath string) error
// 	ListUploadInfo(user string) ([]*db.UploadInfo, error)
// }

// type FileInfoStore struct {
// 	mtx   *sync.RWMutex
// 	store kvstore.IKVStore
// }

// func NewFileInfoStore(store kvstore.IKVStore) (*FileInfoStore, error) {
// 	var err error
// 	for _, nsName := range []string{
// 		db.FileSchemaNs,
// 		db.FileInfoNs,
// 		db.ShareIDNs,
// 	} {
// 		if !store.HasNamespace(nsName) {
// 			if err = store.AddNamespace(nsName); err != nil {
// 				return nil, err
// 			}
// 		}
// 	}

// 	fi := &FileInfoStore{
// 		store: store,
// 		mtx:   &sync.RWMutex{},
// 	}
// 	return fi, nil
// }

func (kv *KVStore) GetFileInfo(ctx context.Context, itemPath string) (*db.FileInfo, error) {
	return kv.getFileInfo(itemPath)
}

func (kv *KVStore) ListFileInfos(ctx context.Context, itemPaths []string) (map[string]*db.FileInfo, error) {
	infos := map[string]*db.FileInfo{}
	for _, itemPath := range itemPaths {
		info, err := kv.getFileInfo(itemPath)
		if err != nil {
			if !errors.Is(err, ErrNotFound) {
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

// func (kv *KVStore) SetInfo(itemPath string, info *db.FileInfo) error {
// 	return kv.setFileInfo(itemPath, info)
// }

// func (kv *KVStore) DelFileInfo(itemPath string) error {
// 	return kv.db.DelStringIn(db.FileInfoNs, itemPath)
// }

func (kv *KVStore) getFileInfoByUser(tx *bolt.Tx, userID uint64, itemPath string) (*db.FileInfo, error) {
	var err error

	fileInfoBucket := tx.Bucket([]byte(db.FileInfoNs))
	if fileInfoBucket == nil {
		return nil, db.ErrBucketNotFound
	}

	fileInfoBytes := fileInfoBucket.Get([]byte(itemPath))
	if fileInfoBytes == nil {
		return nil, db.ErrKeyNotFound
	}

	fileInfo := &db.FileInfo{}
	err = json.Unmarshal(fileInfoBytes, fileInfo)
	if err != nil {
		return nil, err
	}

	if err = db.CheckFileInfo(fileInfo, true); err != nil {
		return nil, err
	}
	return fileInfo, nil
}

func (kv *KVStore) setFileInfoByUser(tx *bolt.Tx, userID uint64, itemPath string, fileInfo *db.FileInfo) error {
	var err error

	if err = db.CheckFileInfo(fileInfo, false); err != nil {
		return err
	}

	fileInfoBucket := tx.Bucket([]byte(db.FileInfoNs))
	if fileInfoBucket == nil {
		return db.ErrBucketNotFound
	}

	fileInfoBytes, err := json.Marshal(fileInfo)
	if err != nil {
		return err
	}

	return fileInfoBucket.Put([]byte(itemPath), fileInfoBytes)
}

func (kv *KVStore) MoveInfos(ctx context.Context, userID uint64, oldPath, newPath string, isDir bool) error {
	return kv.db.Bolt().Update(func(tx *bolt.Tx) error {
		var err error

		fileInfoBucket := tx.Bucket([]byte(db.FileInfoNs))
		if fileInfoBucket == nil {
			return db.ErrBucketNotFound
		}

		fileInfo, err := kv.getFileInfoByUser(tx, userID, oldPath)
		if err != nil {
			if errors.Is(err, db.ErrKeyNotFound) && isDir {
				// the file info for the dir does not exist
				return nil
			}
			return err
		}

		// delete old shareID
		if isDir {
			fileInfo.Shared = false
			fileInfo.ShareID = ""
			err = kv.delShareID(tx, oldPath)
			if err != nil {
				return err
			}
		}

		// delete old info
		err = fileInfoBucket.Delete([]byte(oldPath))
		if err != nil {
			return err
		}

		// add new info
		return kv.setFileInfoByUser(tx, userID, newPath, fileInfo)
	})
}

func (kv *KVStore) SetSha1(ctx context.Context, itemPath, sign string) error {
	info, err := kv.getFileInfo(itemPath)
	if err != nil {
		if !errors.Is(err, ErrNotFound) {
			return err
		}
		info = &db.FileInfo{
			IsDir:  false,
			Shared: false,
		}
	}
	info.Sha1 = sign
	return kv.setFileInfo(itemPath, info)
}

func (kv *KVStore) DelFileInfo(userID uint64, userId uint64, itemPath string) error {
	return kv.db.Bolt().Update(func(tx *bolt.Tx) error {
		var err error

		fileInfoBucket := tx.Bucket([]byte(db.FileInfoNs))
		if fileInfoBucket == nil {
			return db.ErrBucketNotFound
		}

		// delete children
		prefixBytes := []byte(itemPath)
		cur := fileInfoBucket.Cursor()
		usedSpaceDecr := int64(0)
		for k, v := cur.Seek(prefixBytes); k != nil && bytes.HasPrefix(k, prefixBytes); k, v = cur.Next() {
			fileInfo := &db.FileInfo{}
			err = json.Unmarshal(v, fileInfo)
			if err != nil {
				return err
			}

			usedSpaceDecr += fileInfo.Size
			childPath := string(k)
			err = fileInfoBucket.Delete([]byte(childPath))
			if err != nil {
				return err
			}

			err = kv.delShareID(tx, childPath)
			if err != nil {
				return err
			}
		}

		// decr used space
		userInfo, err := kv.getUserInfo(tx, userID)
		if err != nil {
			return err
		}
		userInfo.UsedSpace -= usedSpaceDecr
		err = kv.setUserInfo(tx, userID, userInfo)
		if err != nil {
			return err
		}

		return nil
	})
}

func (kv *KVStore) getFileInfo(itemPath string) (*db.FileInfo, error) {
	infoStr, ok := kv.db.GetStringIn(db.FileInfoNs, itemPath)
	if !ok {
		return nil, ErrNotFound
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

func (kv *KVStore) setFileInfo(itemPath string, info *db.FileInfo) error {
	if err := db.CheckFileInfo(info, false); err != nil {
		return err
	}

	infoStr, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("set file info: %w", err)
	}

	err = kv.db.SetStringIn(db.FileInfoNs, itemPath, string(infoStr))
	if err != nil {
		return fmt.Errorf("set file info: %w", err)
	}
	return nil
}
