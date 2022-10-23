package boltdb

import (
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
