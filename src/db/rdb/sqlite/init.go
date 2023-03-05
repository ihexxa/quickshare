package sqlite

import (
	"context"
	"database/sql"
	"sync"

	"github.com/ihexxa/quickshare/src/db"
	"github.com/ihexxa/quickshare/src/db/rdb/base"
	_ "modernc.org/sqlite"
)

type SQLite struct {
	db.IDB
	dbPath string
}

func NewSQLite(dbPath string) (*SQLite, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	return &SQLite{
		IDB:    db,
		dbPath: dbPath,
	}, nil
}

type SQLiteStore struct {
	store *base.BaseStore
	mtx   *sync.RWMutex
}

func NewSQLiteStore(db db.IDB) (*SQLiteStore, error) {
	return &SQLiteStore{
		store: base.NewBaseStore(db),
		mtx:   &sync.RWMutex{},
	}, nil
}

func (st *SQLiteStore) Close() error {
	return st.store.Close()
}

func (st *SQLiteStore) Lock() {
	st.mtx.Lock()
}

func (st *SQLiteStore) Unlock() {
	st.mtx.Unlock()
}

func (st *SQLiteStore) RLock() {
	st.mtx.RLock()
}

func (st *SQLiteStore) RUnlock() {
	st.mtx.RUnlock()
}

func (st *SQLiteStore) IsInited() bool {
	// always try to init the db
	return false
}

func (st *SQLiteStore) Init(ctx context.Context, rootName, rootPwd string, cfg *db.SiteConfig) error {
	st.Lock()
	defer st.Unlock()

	return st.store.Init(ctx, rootName, rootPwd, cfg)
}

func (st *SQLiteStore) InitUserTable(ctx context.Context, tx *sql.Tx, rootName, rootPwd string) error {
	return st.store.InitUserTable(ctx, tx, rootName, rootPwd)
}

func (st *SQLiteStore) InitFileTables(ctx context.Context, tx *sql.Tx) error {
	return st.store.InitFileTables(ctx, tx)
}

func (st *SQLiteStore) InitConfigTable(ctx context.Context, tx *sql.Tx, cfg *db.SiteConfig) error {
	return st.store.InitConfigTable(ctx, tx, cfg)
}
