package sqlite

import (
	"context"

	"github.com/ihexxa/quickshare/src/db"
)

func (st *SQLiteStore) SetClientCfg(ctx context.Context, cfg *db.ClientConfig) error {
	st.Lock()
	defer st.Unlock()

	return st.store.SetClientCfg(ctx, cfg)
}

func (st *SQLiteStore) GetCfg(ctx context.Context) (*db.SiteConfig, error) {
	st.RLock()
	defer st.RUnlock()

	return st.store.GetCfg(ctx)
}
