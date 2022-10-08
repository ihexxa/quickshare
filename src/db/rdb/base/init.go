package base

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/ihexxa/quickshare/src/db"
	_ "github.com/mattn/go-sqlite3"
)

var (
	txOpts = &sql.TxOptions{}
)

type BaseStore struct {
	db db.IDB
}

func NewBaseStore(db db.IDB) *BaseStore {
	return &BaseStore{
		db: db,
	}
}

func (st *BaseStore) Db() db.IDB {
	return st.db
}

func (st *BaseStore) Close() error {
	return st.db.Close()
}

func (st *BaseStore) IsInited() bool {
	// always try to init the db
	return false
}

func (st *BaseStore) Init(ctx context.Context, rootName, rootPwd string, cfg *db.SiteConfig) error {
	tx, err := st.db.BeginTx(ctx, txOpts)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = st.InitUserTable(ctx, tx, rootName, rootPwd)
	if err != nil {
		return err
	}

	if err = st.InitFileTables(ctx, tx); err != nil {
		return err
	}

	if err = st.InitConfigTable(ctx, tx, cfg); err != nil {
		return err
	}

	return tx.Commit()
}

func (st *BaseStore) InitUserTable(ctx context.Context, tx *sql.Tx, rootName, rootPwd string) error {
	_, err := tx.ExecContext(
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

	_, err = tx.ExecContext(
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
		_, err := st.getUser(ctx, tx, user.ID)
		if err != nil {
			if errors.Is(err, db.ErrUserNotFound) {
				err = st.addUser(ctx, tx, user)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}

	return nil
}

func (st *BaseStore) InitFileTables(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(
		ctx,
		`create table if not exists t_file_info (
			id bigint not null,
			path varchar not null,
			user bigint not null,
			location varchar not null,
			parent varchar not null,
			name varchar not null,
			is_dir boolean not null,
			size bigint not null,
			share_id varchar not null,
			info varchar not null,
			primary key(id)
		)`,
	)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(
		ctx,
		`create index if not exists t_file_path on t_file_info (path, location)`,
	)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(
		ctx,
		`create index if not exists t_file_share on t_file_info (share_id, location)`,
	)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(
		ctx,
		`create table if not exists t_file_uploading (
			id bigint not null,
			real_path varchar not null,
			tmp_path varchar not null unique,
			user bigint not null,
			size bigint not null,
			uploaded bigint not null,
			primary key(id)
		)`,
	)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(
		ctx,
		`create index if not exists t_file_uploading_path on t_file_uploading (real_path, user)`,
	)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(
		ctx,
		`create index if not exists t_file_uploading_user on t_file_uploading (user)`,
	)
	return err
}

func (st *BaseStore) InitConfigTable(ctx context.Context, tx *sql.Tx, cfg *db.SiteConfig) error {
	_, err := tx.ExecContext(
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

	_, err = st.getCfg(ctx, tx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			_, err = tx.ExecContext(
				ctx,
				`insert into t_config
				(id, config, modified) values (?, ?, ?)`,
				0, cfgStr, time.Now(),
			)
			return err
		}
		return err
	}

	return nil
}
