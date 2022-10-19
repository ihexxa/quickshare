package boltdb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/boltdb/bolt"
	"github.com/ihexxa/quickshare/src/db"
)

var (
	ErrGreaterThanSize = errors.New("uploaded is greater than file size")
	ErrUploadNotFound  = errors.New("upload info not found")
)

func (kv *KVStore) getUploadInfo(user, filePath string) (string, int64, int64, error) {
	ns := db.UploadNS(user)
	infoBytes, ok := kv.db.GetStringIn(ns, filePath)
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

func (kv *KVStore) setUploadInfo(user, filePath string, info *db.UploadInfo) error {
	newInfoBytes, err := json.Marshal(info)
	if err != nil {
		return err
	}
	return kv.db.SetStringIn(db.UploadNS(user), filePath, string(newInfoBytes))
}

// func (kv *KVStore) AddUploadInfo(ctx context.Context, user, filePath, tmpPath string, fileSize int64) error {
// 	kv.mtx.Lock()
// 	defer kv.mtx.Unlock()

// 	ns := db.UploadNS(user)
// 	err := kv.db.AddNamespace(ns)
// 	if err != nil {
// 		return err
// 	}

// 	_, _, _, err = kv.getUploadInfo(user, tmpPath)
// 	if err == nil {
// 		return db.ErrCreateExisting
// 	}

// 	return kv.setUploadInfo(user, filePath, &db.UploadInfo{
// 		RealFilePath: filePath,
// 		Size:         fileSize,
// 		Uploaded:     0,
// 	})
// }

func (kv *KVStore) AddUploadInfos(ctx context.Context, userID uint64, tmpPath, filePath string, info *db.FileInfo) error {
	return kv.db.Bolt().Update(func(tx *bolt.Tx) error {
		var err error

		_, err = kv.getUploadInfoByUser(tx, userID, tmpPath)
		if err == nil {
			return db.ErrKeyExisting
		} else if !errors.Is(err, db.ErrBucketNotFound) && !errors.Is(err, db.ErrKeyNotFound) {
			return err
		}

		// check space quota
		userInfo, err := kv.getUserInfo(tx, userID)
		if err != nil {
			return err
		}

		if userInfo.UsedSpace+info.Size > int64(userInfo.Quota.SpaceLimit) {
			return db.ErrQuota
		}

		// update used space
		userInfo.UsedSpace += info.Size
		err = kv.setUserInfo(tx, userID, userInfo)
		if err != nil {
			return err
		}

		// add upload info
		uploadInfo := &db.UploadInfo{
			RealFilePath: filePath,
			Size:         info.Size,
			Uploaded:     0,
		}
		return kv.setUploadInfoByUser(tx, userID, tmpPath, uploadInfo, false)
	})
}

func (kv *KVStore) DelUploadingInfos(ctx context.Context, userID uint64, uploadPath string) error {
	return kv.db.Bolt().Update(func(tx *bolt.Tx) error {
		var err error

		// delete upload info
		uidStr := fmt.Sprint(userID)
		userUploadNS := db.UploadNS(uidStr)

		uploadInfoBucket := tx.Bucket([]byte(userUploadNS))
		if uploadInfoBucket == nil {
			return db.ErrBucketNotFound
		}

		uploadInfoBytes := uploadInfoBucket.Get([]byte(uploadPath))
		if uploadInfoBytes == nil {
			return db.ErrKeyNotFound
		}

		uploadInfo := &db.UploadInfo{}
		err = json.Unmarshal(uploadInfoBytes, uploadInfo)
		if err != nil {
			return err
		}

		err = uploadInfoBucket.Delete([]byte(uploadPath))
		if err != nil {
			return err
		}

		// decr used space
		userInfo, err := kv.getUserInfo(tx, userID)
		if err != nil {
			return err
		}

		userInfo.UsedSpace -= uploadInfo.Size
		return kv.setUserInfo(tx, userID, userInfo)
	})
}

func (kv *KVStore) MoveUploadingInfos(ctx context.Context, userID uint64, uploadPath, itemPath string) error {
	return kv.db.Bolt().Update(func(tx *bolt.Tx) error {
		var err error

		// delete upload info
		uidStr := fmt.Sprint(userID)
		userUploadNS := db.UploadNS(uidStr)

		uploadInfoBucket := tx.Bucket([]byte(userUploadNS))
		if uploadInfoBucket == nil {
			return db.ErrBucketNotFound
		}

		uploadInfoBytes := uploadInfoBucket.Get([]byte(uploadPath))
		if uploadInfoBytes == nil {
			return db.ErrKeyNotFound
		}

		uploadInfo := &db.UploadInfo{}
		err = json.Unmarshal(uploadInfoBytes, uploadInfo)
		if err != nil {
			return err
		}

		err = uploadInfoBucket.Delete([]byte(uploadPath))
		if err != nil {
			return err
		}

		// create file info
		fileInfo := &db.FileInfo{
			IsDir: false,
			Size:  uploadInfo.Size,
		}
		return kv.setFileInfoByUser(tx, userID, itemPath, fileInfo)
	})
}

func (kv *KVStore) SetUploadInfo(ctx context.Context, user, filePath string, newUploaded int64) error {
	kv.mtx.Lock()
	defer kv.mtx.Unlock()

	realFilePath, fileSize, _, err := kv.getUploadInfo(user, filePath)
	if err != nil {
		return err
	} else if newUploaded > fileSize {
		return ErrGreaterThanSize
	}

	return kv.setUploadInfo(user, filePath, &db.UploadInfo{
		RealFilePath: realFilePath,
		Size:         fileSize,
		Uploaded:     newUploaded,
	})
}

func (kv *KVStore) GetUploadInfo(ctx context.Context, user, filePath string) (string, int64, int64, error) {
	kv.mtx.Lock()
	defer kv.mtx.Unlock()

	return kv.getUploadInfo(user, filePath)
}

func (kv *KVStore) DelUploadInfo(user, filePath string) error {
	return kv.db.DelInt64In(db.UploadNS(user), filePath)
}

func (kv *KVStore) ListUploadInfos(ctx context.Context, user string) ([]*db.UploadInfo, error) {
	ns := db.UploadNS(user)
	if !kv.db.HasNamespace(ns) {
		return nil, nil
	}

	infoMap, err := kv.db.ListStringsIn(ns)
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

func (kv *KVStore) getUploadInfoByUser(tx *bolt.Tx, userID uint64, itemPath string) (*db.UploadInfo, error) {
	var err error

	uidStr := fmt.Sprint(userID)
	userUploadNS := db.UploadNS(uidStr)

	uploadInfoBucket := tx.Bucket([]byte(userUploadNS))
	if uploadInfoBucket == nil {
		return nil, db.ErrBucketNotFound

	}

	uploadInfoBytes := uploadInfoBucket.Get([]byte(itemPath))
	if uploadInfoBytes == nil {
		return nil, db.ErrKeyNotFound
	}

	uploadInfo := &db.UploadInfo{}
	err = json.Unmarshal(uploadInfoBytes, uploadInfo)
	if err != nil {
		return nil, err
	}
	return uploadInfo, nil
}

func (kv *KVStore) setUploadInfoByUser(tx *bolt.Tx, userID uint64, uploadPath string, uploadInfo *db.UploadInfo, overWrite bool) error {
	var err error

	uidStr := fmt.Sprint(userID)
	userUploadNS := db.UploadNS(uidStr)
	uploadInfoBucket := tx.Bucket([]byte(userUploadNS))
	if uploadInfoBucket == nil {
		uploadInfoBucket, err = tx.CreateBucket([]byte(userUploadNS))
		if err != nil {
			return err
		}
	}

	existingInfoBytes := uploadInfoBucket.Get([]byte(uploadPath))
	if existingInfoBytes != nil {
		return nil
	}
	uploadInfoBytes, err := json.Marshal(uploadInfo)
	if err != nil {
		return err
	}

	return uploadInfoBucket.Put([]byte(uploadPath), uploadInfoBytes)
}
