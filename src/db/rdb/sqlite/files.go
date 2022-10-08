package sqlite

import (
	"context"

	"github.com/ihexxa/quickshare/src/db"
)

func (st *SQLiteStore) GetFileInfo(ctx context.Context, itemPath string) (*db.FileInfo, error) {
	st.RLock()
	defer st.RUnlock()

	return st.store.GetFileInfo(ctx, itemPath)
}

func (st *SQLiteStore) ListFileInfos(ctx context.Context, itemPaths []string) (map[string]*db.FileInfo, error) {
	st.RLock()
	defer st.RUnlock()

	return st.store.ListFileInfos(ctx, itemPaths)
}

func (st *SQLiteStore) AddFileInfo(ctx context.Context, infoId, userId uint64, itemPath string, info *db.FileInfo) error {
	st.Lock()
	defer st.Unlock()

	return st.store.AddFileInfo(ctx, infoId, userId, itemPath, info)
}

func (st *SQLiteStore) SetSha1(ctx context.Context, itemPath, sign string) error {
	st.Lock()
	defer st.Unlock()

	return st.store.SetSha1(ctx, itemPath, sign)
}

func (st *SQLiteStore) DelFileInfo(ctx context.Context, userID uint64, itemPath string) error {
	st.Lock()
	defer st.Unlock()

	return st.store.DelFileInfo(ctx, userID, itemPath)
}

func (st *SQLiteStore) MoveFileInfo(ctx context.Context, userId uint64, oldPath, newPath string, isDir bool) error {
	st.Lock()
	defer st.Unlock()

	return st.store.MoveFileInfo(ctx, userId, oldPath, newPath, isDir)
}
