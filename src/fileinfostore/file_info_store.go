package fileinfostore

import (
	"github.com/ihexxa/quickshare/src/kvstore"
)

const (
	InitNs      = "Init"
	SharingNs   = "sharing"
	InitTimeKey = "initTime"
)

type IFileInfoStore interface {
	AddSharing(dirPath string) error
	DelSharing(dirPath string) error
	GetSharing(dirPath string) (bool, bool)
	ListSharings(prefix string) (map[string]bool, error)
}

type FileInfoStore struct {
	store kvstore.IKVStore
}

func NewFileInfoStore(store kvstore.IKVStore) (*FileInfoStore, error) {
	_, ok := store.GetStringIn(InitNs, InitTimeKey)
	if !ok {
		var err error
		for _, nsName := range []string{
			InitNs,
			SharingNs,
		} {
			if err = store.AddNamespace(nsName); err != nil {
				return nil, err
			}
		}
	}

	return &FileInfoStore{
		store: store,
	}, nil
}

func (us *FileInfoStore) AddSharing(dirPath string) error {
	return us.store.SetBoolIn(SharingNs, dirPath, true)
}

func (us *FileInfoStore) DelSharing(dirPath string) error {
	return us.store.DelBoolIn(SharingNs, dirPath)
}

func (us *FileInfoStore) GetSharing(dirPath string) (bool, bool) {
	return us.store.GetBoolIn(SharingNs, dirPath)
}

func (us *FileInfoStore) ListSharings(prefix string) (map[string]bool, error) {
	return us.store.ListBoolsByPrefixIn(prefix, SharingNs)
}
