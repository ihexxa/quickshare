package boltstore

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/boltdb/bolt"
	"github.com/ihexxa/quickshare/src/db"
)

type BoltStore struct {
	boltdb *bolt.DB
}

func NewBoltStore(boltdb *bolt.DB) (*BoltStore, error) {
	bs := &BoltStore{
		boltdb: boltdb,
	}
	return bs, nil
}

func (bs *BoltStore) getUserInfo(tx *bolt.Tx, userID uint64) (*db.User, error) {
	var err error

	usersBucket := tx.Bucket([]byte(db.UsersNs))
	if usersBucket == nil {
		return nil, db.ErrBucketNotFound
	}

	uidStr := fmt.Sprint(userID)
	infoBytes := usersBucket.Get([]byte(uidStr))
	if infoBytes == nil {
		return nil, db.ErrKeyNotFound
	}

	userInfo := &db.User{}
	err = json.Unmarshal(infoBytes, userInfo)
	if err != nil {
		return nil, err
	} else if userInfo.ID != userID {
		return nil, fmt.Errorf("user id key(%d) info(%d) does match", userID, userInfo.ID)
	}

	if err = db.CheckUser(userInfo, true); err != nil {
		return nil, err
	}

	return userInfo, nil
}

func (bs *BoltStore) setUserInfo(tx *bolt.Tx, userID uint64, userInfo *db.User) error {
	var err error

	if err = db.CheckUser(userInfo, false); err != nil {
		return err
	}

	usersBucket := tx.Bucket([]byte(db.UsersNs))
	if usersBucket == nil {
		return db.ErrBucketNotFound
	}

	userInfoBytes, err := json.Marshal(userInfo)
	if err != nil {
		return err
	}

	uidStr := fmt.Sprint(userID)
	return usersBucket.Put([]byte(uidStr), userInfoBytes)
}

func (bs *BoltStore) getUploadInfo(tx *bolt.Tx, userID uint64, itemPath string) (*db.UploadInfo, error) {
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

func (bs *BoltStore) setUploadInfo(tx *bolt.Tx, userID uint64, uploadPath string, uploadInfo *db.UploadInfo, overWrite bool) error {
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

func (bs *BoltStore) getFileInfo(tx *bolt.Tx, userID uint64, itemPath string) (*db.FileInfo, error) {
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

func (bs *BoltStore) setFileInfo(tx *bolt.Tx, userID uint64, itemPath string, fileInfo *db.FileInfo) error {
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

func (bs *BoltStore) AddUploadInfos(userID uint64, tmpPath, filePath string, info *db.FileInfo) error {
	return bs.boltdb.Update(func(tx *bolt.Tx) error {
		var err error

		_, err = bs.getUploadInfo(tx, userID, tmpPath)
		if err == nil {
			return db.ErrKeyExisting
		} else if !errors.Is(err, db.ErrBucketNotFound) && !errors.Is(err, db.ErrKeyNotFound) {
			return err
		}

		// check space quota
		userInfo, err := bs.getUserInfo(tx, userID)
		if err != nil {
			return err
		}

		if userInfo.UsedSpace+info.Size > int64(userInfo.Quota.SpaceLimit) {
			return db.ErrQuota
		}

		// update used space
		userInfo.UsedSpace += info.Size
		err = bs.setUserInfo(tx, userID, userInfo)
		if err != nil {
			return err
		}

		// add upload info
		uploadInfo := &db.UploadInfo{
			RealFilePath: filePath,
			Size:         info.Size,
			Uploaded:     0,
		}
		return bs.setUploadInfo(tx, userID, tmpPath, uploadInfo, false)
	})
}

func (bs *BoltStore) DelUploadingInfos(userID uint64, uploadPath string) error {
	return bs.boltdb.Update(func(tx *bolt.Tx) error {
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
		userInfo, err := bs.getUserInfo(tx, userID)
		if err != nil {
			return err
		}

		userInfo.UsedSpace -= uploadInfo.Size
		return bs.setUserInfo(tx, userID, userInfo)
	})
}

func (bs *BoltStore) MoveUploadingInfos(userID uint64, uploadPath, itemPath string) error {
	return bs.boltdb.Update(func(tx *bolt.Tx) error {
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
		return bs.setFileInfo(tx, userID, itemPath, fileInfo)
	})
}

func (bs *BoltStore) delShareID(tx *bolt.Tx, itemPath string) error {
	var err error

	shareIDBucket := tx.Bucket([]byte(db.ShareIDNs))
	if shareIDBucket == nil {
		return db.ErrBucketNotFound
	}

	shareIDtoDir := map[string]string{}
	shareIDBucket.ForEach(func(k, v []byte) error {
		shareIDtoDir[string(k)] = string(v)
		return nil
	})

	// because before this version, shareIDs are not removed correctly
	// so it iterates all shareIDs and cleans remaining entries
	for shareID, shareDir := range shareIDtoDir {
		if shareDir == itemPath {
			err = shareIDBucket.Delete([]byte(shareID))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (bs *BoltStore) DelInfos(userID uint64, itemPath string, isDir bool) error {
	return bs.boltdb.Update(func(tx *bolt.Tx) error {
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
			if fileInfo.IsDir {
				err = bs.delShareID(tx, childPath)
				if err != nil {
					return err
				}
			}
		}

		// decr used space
		userInfo, err := bs.getUserInfo(tx, userID)
		if err != nil {
			return err
		}
		userInfo.UsedSpace -= usedSpaceDecr
		err = bs.setUserInfo(tx, userID, userInfo)
		if err != nil {
			return err
		}

		return nil
	})
}

func (bs *BoltStore) MoveInfos(userID uint64, oldPath, newPath string, isDir bool) error {
	return bs.boltdb.Update(func(tx *bolt.Tx) error {
		var err error

		fileInfoBucket := tx.Bucket([]byte(db.FileInfoNs))
		if fileInfoBucket == nil {
			return db.ErrBucketNotFound
		}

		fileInfo, err := bs.getFileInfo(tx, userID, oldPath)
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
			err = bs.delShareID(tx, oldPath)
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
		return bs.setFileInfo(tx, userID, newPath, fileInfo)
	})
}
