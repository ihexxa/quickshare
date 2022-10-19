package boltdb

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/ihexxa/quickshare/src/db"
	"github.com/ihexxa/quickshare/src/kvstore"
)

type KVStore struct {
	mtx *sync.RWMutex
	db  kvstore.IKVStore
}

func NewKVStore(db kvstore.IKVStore) (*KVStore, error) {
	return &KVStore{
		mtx: &sync.RWMutex{},
		db:  db,
	}, nil
}

func (kv *KVStore) Close() error {
	return kv.db.Close()
}

func (kv *KVStore) Lock() {
	panic("not supported")
}

func (kv *KVStore) Unlock() {
	panic("not supported")
}

func (kv *KVStore) RLock() {
	panic("not supported")
}

func (kv *KVStore) RUnlock() {
	panic("not supported")
}

func (kv *KVStore) IsInited() bool {
	// always try to init the db
	return false
}

func (kv *KVStore) Init(ctx context.Context, rootName, rootPwd string, cfg *db.SiteConfig) error {
	kv.mtx.Lock()
	defer kv.mtx.Unlock()

	err := kv.InitUserTable(ctx, nil, rootName, rootPwd)
	if err != nil {
		return err
	}
	err = kv.InitFileTables(ctx, nil)
	if err != nil {
		return err
	}
	return kv.InitConfigTable(ctx, nil, cfg)
}

func (kv *KVStore) InitUserTable(ctx context.Context, tx *sql.Tx, rootName, rootPwd string) error {
	var err error

	for _, namespace := range []string{
		db.UserSchemaNs,
		db.UserIDsNs,
		db.UsersNs,
		db.RolesNs,
	} {
		_, ok := kv.db.GetStringIn(namespace, db.KeyInitTime)
		if !ok {
			if err = kv.db.AddNamespace(namespace); err != nil {
				return err
			}
		}
	}

	admin := &db.User{
		ID:   0,
		Name: rootName,
		Pwd:  rootPwd,
		Role: db.AdminRole,
		Quota: &db.Quota{
			SpaceLimit:         db.DefaultSpaceLimit,
			UploadSpeedLimit:   db.DefaultUploadSpeedLimit,
			DownloadSpeedLimit: db.DefaultDownloadSpeedLimit,
		},
		Preferences: &db.DefaultPreferences,
	}

	visitor := &db.User{
		ID:   VisitorID,
		Name: VisitorName,
		Pwd:  rootPwd,
		Role: db.VisitorRole,
		Quota: &db.Quota{
			SpaceLimit:         0,
			UploadSpeedLimit:   db.VisitorUploadSpeedLimit,
			DownloadSpeedLimit: db.VisitorDownloadSpeedLimit,
		},
		Preferences: &db.DefaultPreferences,
	}

	for _, user := range []*db.User{admin, visitor} {
		err = kv.AddUser(ctx, user)
		if err != nil {
			return err
		}
	}

	for _, role := range []string{db.AdminRole, db.UserRole, db.VisitorRole} {
		err = kv.AddRole(role)
		if err != nil {
			return err
		}
	}

	return kv.db.SetStringIn(db.UserSchemaNs, db.KeyInitTime, fmt.Sprintf("%d", time.Now().Unix()))
}

func (kv *KVStore) InitFileTables(ctx context.Context, tx *sql.Tx) error {
	for _, nsName := range []string{
		db.FileInfoNs,
		db.ShareIDNs,
	} {
		if !kv.db.HasNamespace(nsName) {
			if err := kv.db.AddNamespace(nsName); err != nil {
				return err
			}
		}
	}

	return nil
}

func (kv *KVStore) InitConfigTable(ctx context.Context, tx *sql.Tx, cfg *db.SiteConfig) error {
	_, ok := kv.db.GetStringIn(NsSite, KeySiteCfg)
	if !ok {
		var err error
		if err = kv.db.AddNamespace(NsSite); err != nil {
			return err
		}

		return kv.db.setCfg(cfg)
	}
	return nil
}
