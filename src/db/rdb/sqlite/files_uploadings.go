package sqlite

import (
	"context"

	"github.com/ihexxa/quickshare/src/db"
)

func (st *SQLiteStore) AddUploadInfos(ctx context.Context, uploadId, userId uint64, tmpPath, filePath string, info *db.FileInfo) error {
	st.Lock()
	defer st.Unlock()

	return st.store.AddUploadInfos(ctx, uploadId, userId, tmpPath, filePath, info)
}

func (st *SQLiteStore) DelUploadingInfos(ctx context.Context, userId uint64, realPath string) error {
	st.Lock()
	defer st.Unlock()

	return st.store.DelUploadingInfos(ctx, userId, realPath)
}

func (st *SQLiteStore) MoveUploadingInfos(ctx context.Context, infoId, userId uint64, uploadPath, itemPath string) error {
	st.Lock()
	defer st.Unlock()

	return st.store.MoveUploadingInfos(ctx, infoId, userId, uploadPath, itemPath)
}

func (st *SQLiteStore) SetUploadInfo(ctx context.Context, userId uint64, filePath string, newUploaded int64) error {
	st.Lock()
	defer st.Unlock()

	return st.store.SetUploadInfo(ctx, userId, filePath, newUploaded)
}

func (st *SQLiteStore) GetUploadInfo(ctx context.Context, userId uint64, filePath string) (string, int64, int64, error) {
	st.RLock()
	defer st.RUnlock()

	return st.store.GetUploadInfo(ctx, userId, filePath)
}

func (st *SQLiteStore) ListUploadInfos(ctx context.Context, userId uint64) ([]*db.UploadInfo, error) {
	st.RLock()
	defer st.RUnlock()

	return st.store.ListUploadInfos(ctx, userId)
}
