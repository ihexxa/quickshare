package sqlitecgo

import (
	"context"
)

func (st *SQLiteStore) IsSharing(ctx context.Context, dirPath string) (bool, error) {
	st.RLock()
	defer st.RUnlock()

	return st.store.IsSharing(ctx, dirPath)
}

func (st *SQLiteStore) GetSharingDir(ctx context.Context, hashID string) (string, error) {
	st.RLock()
	defer st.RUnlock()

	return st.store.GetSharingDir(ctx, hashID)
}

func (st *SQLiteStore) AddSharing(ctx context.Context, infoId, userId uint64, dirPath string) error {
	st.Lock()
	defer st.Unlock()

	return st.store.AddSharing(ctx, infoId, userId, dirPath)
}

func (st *SQLiteStore) DelSharing(ctx context.Context, userId uint64, dirPath string) error {
	st.Lock()
	defer st.Unlock()

	return st.store.DelSharing(ctx, userId, dirPath)
}

func (st *SQLiteStore) ListSharingsByLocation(ctx context.Context, location string) (map[string]string, error) {
	st.RLock()
	defer st.RUnlock()

	return st.store.ListSharingsByLocation(ctx, location)
}
