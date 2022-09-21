package sqlite

import (
	"context"
	"encoding/json"

	"github.com/ihexxa/quickshare/src/db"
)

func (st *SQLiteStore) getCfg(ctx context.Context) (*db.SiteConfig, error) {
	var configStr string
	err := st.db.QueryRowContext(
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

func (st *SQLiteStore) setCfg(ctx context.Context, cfg *db.SiteConfig) error {
	if err := db.CheckSiteCfg(cfg, false); err != nil {
		return err
	}

	cfgBytes, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	_, err = st.db.ExecContext(
		ctx,
		`update t_config
		set config=?
		where id=0`,
		string(cfgBytes),
	)
	return err
}

func (st *SQLiteStore) SetClientCfg(ctx context.Context, cfg *db.ClientConfig) error {
	st.Lock()
	defer st.Unlock()

	siteCfg, err := st.getCfg(ctx)
	if err != nil {
		return err
	}
	siteCfg.ClientCfg = cfg

	return st.setCfg(ctx, siteCfg)
}

func (st *SQLiteStore) GetCfg(ctx context.Context) (*db.SiteConfig, error) {
	st.RLock()
	defer st.RUnlock()

	return st.getCfg(ctx)
}
