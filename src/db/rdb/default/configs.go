package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/ihexxa/quickshare/src/db"
)

func (st *SQLiteStore) getCfg(ctx context.Context, tx *sql.Tx) (*db.SiteConfig, error) {
	var configStr string
	err := tx.QueryRowContext(
		ctx,
		`select config
		from t_config
		where id=0`,
	).Scan(&configStr)
	if err != nil {
		return nil, err
	}

	config := &db.SiteConfig{}
	err = json.Unmarshal([]byte(configStr), config)
	if err != nil {
		return nil, err
	}

	if err = db.CheckSiteCfg(config, true); err != nil {
		return nil, err
	}
	return config, nil
}

func (st *SQLiteStore) setCfg(ctx context.Context, tx *sql.Tx, cfg *db.SiteConfig) error {
	if err := db.CheckSiteCfg(cfg, false); err != nil {
		return err
	}

	cfgBytes, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(
		ctx,
		`update t_config
		set config=?
		where id=0`,
		string(cfgBytes),
	)
	return err
}

func (st *SQLiteStore) SetClientCfg(ctx context.Context, cfg *db.ClientConfig) error {
	tx, err := st.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	siteCfg, err := st.getCfg(ctx, tx)
	if err != nil {
		return err
	}
	siteCfg.ClientCfg = cfg

	err = st.setCfg(ctx, tx, siteCfg)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (st *SQLiteStore) GetCfg(ctx context.Context) (*db.SiteConfig, error) {
	tx, err := st.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	return st.getCfg(ctx, tx)
}
