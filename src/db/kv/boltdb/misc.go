package boltdb

import (
	"encoding/json"
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
