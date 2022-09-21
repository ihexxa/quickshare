package default

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ihexxa/quickshare/src/db"
	_ "github.com/mattn/go-sqlite3"
)

type DefaultDB struct {
	db.IDB
	dbPath string
}

func NewDefaultDB(driverName, dbPath string) (*DefaultDB, error) {
	db, err := sql.Open(driverName, dbPath)
	if err != nil {
		return nil, err
	}

	return &DefaultDB{
		IDB:    db,
		dbPath: dbPath,
	}, nil
}

type DefaultStore struct {
	db  db.IDB
}

func NewDefaultStore(db db.IDB) (*DefaultStore, error) {
	return &DefaultStore{
		db:  db,
	}, nil
}

func (st *DefaultStore) Close() error {
	return st.db.Close()
}

func (st *DefaultStore) Lock() {
	st.mtx.Lock()
}

func (st *DefaultStore) Unlock() {
	st.mtx.Unlock()
}

func (st *DefaultStore) RLock() {
	st.mtx.RLock()
}

func (st *DefaultStore) RUnlock() {
	st.mtx.RUnlock()
}

func (st *DefaultStore) IsInited() bool {
	// always try to init the db
	return false
}

func (st *DefaultStore) Init(ctx context.Context, rootName, rootPwd string, cfg *db.SiteConfig) error {
	err := st.InitUserTable(ctx, rootName, rootPwd)
	if err != nil {
		return err
	}

	if err = st.InitFileTables(ctx); err != nil {
		return err
	}

	return st.InitConfigTable(ctx, cfg)
}

func (st *DefaultStore) InitUserTable(ctx context.Context, rootName, rootPwd string) error {
	_, err := st.db.ExecContext(
		ctx,
		`create table if not exists t_user (
			id bigint not null,
			name varchar not null unique,
			pwd varchar not null,
			role integer not null,
			used_space bigint not null,
			quota varchar not null,
			preference varchar not null,
			primary key(id)
		)`,
	)
	if err != nil {
		return err
	}

	_, err = st.db.ExecContext(
		ctx,
		`create index if not exists i_user_name on t_user (name)`,
	)
	if err != nil {
		return err
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
		ID:   db.VisitorID,
		Name: db.VisitorName,
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
		err = st.AddUser(ctx, user)
		if err != nil {
			return err
		}
	}

	return nil
}

func (st *DefaultStore) InitFileTables(ctx context.Context) error {
	_, err := st.db.ExecContext(
		ctx,
		`create table if not exists t_file_info (
			path varchar not null,
			user bigint not null,
			location varchar not null,
			parent varchar not null,
			name varchar not null,
			is_dir boolean not null,
			size bigint not null,
			share_id varchar not null,
			info varchar not null,
			primary key(path)
		)`,
	)
	if err != nil {
		return err
	}

	_, err = st.db.ExecContext(
		ctx,
		`create index if not exists t_file_share on t_file_info (share_id, location)`,
	)
	if err != nil {
		return err
	}

	_, err = st.db.ExecContext(
		ctx,
		`create table if not exists t_file_uploading (
			real_path varchar not null,
			tmp_path varchar not null unique,
			user bigint not null,
			size bigint not null,
			uploaded bigint not null,
			primary key(real_path)
		)`,
	)
	if err != nil {
		return err
	}

	_, err = st.db.ExecContext(
		ctx,
		`create index if not exists t_file_uploading_path on t_file_uploading (real_path, user)`,
	)
	if err != nil {
		return err
	}

	_, err = st.db.ExecContext(
		ctx,
		`create index if not exists t_file_uploading_user on t_file_uploading (user)`,
	)
	return err
}

func (st *DefaultStore) InitConfigTable(ctx context.Context, cfg *db.SiteConfig) error {
	st.Lock()
	defer st.Unlock()

	_, err := st.db.ExecContext(
		ctx,
		`create table if not exists t_config (
			id bigint not null,
			config varchar not null,
			modified datetime not null,
			primary key(id)
		)`,
	)
	if err != nil {
		return err
	}

	cfgStr, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	_, err = st.db.ExecContext(
		ctx,
		`insert into t_config
		(id, config, modified) values (?, ?, ?)`,
		0, cfgStr, time.Now(),
	)
	return err
}
