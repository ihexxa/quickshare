package sqlite

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ihexxa/quickshare/src/db"
)

func (st *SQLiteStore) addUploadInfoOnly(ctx context.Context, userId uint64, filePath, tmpPath string, fileSize int64) error {
	_, err := st.db.ExecContext(
		ctx,
		`insert into t_file_uploading
		(real_path, tmp_path, user, size, uploaded) values (?, ?, ?, ?, ?)`,
		filePath, tmpPath, userId, fileSize, 0,
	)
	return err
}

func (st *SQLiteStore) AddUploadInfos(ctx context.Context, userId uint64, tmpPath, filePath string, info *db.FileInfo) error {
	st.Lock()
	defer st.Unlock()

	userInfo, err := st.getUser(ctx, userId)
	if err != nil {
		return err
	} else if userInfo.UsedSpace+info.Size > int64(userInfo.Quota.SpaceLimit) {
		return db.ErrQuota
	}

	_, _, _, err = st.getUploadInfo(ctx, userId, filePath)
	if err == nil {
		return db.ErrKeyExisting
	} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	userInfo.UsedSpace += info.Size
	err = st.setUser(ctx, userInfo)
	if err != nil {
		return err
	}

	return st.addUploadInfoOnly(ctx, userId, filePath, tmpPath, info.Size)
}

func (st *SQLiteStore) DelUploadingInfos(ctx context.Context, userId uint64, realPath string) error {
	st.Lock()
	defer st.Unlock()

	return st.delUploadingInfos(ctx, userId, realPath)
}

func (st *SQLiteStore) delUploadingInfos(ctx context.Context, userId uint64, realPath string) error {
	_, size, _, err := st.getUploadInfo(ctx, userId, realPath)
	if err != nil {
		// info may not exist
		return err
	}

	err = st.delUploadInfoOnly(ctx, userId, realPath)
	if err != nil {
		return err
	}

	userInfo, err := st.getUser(ctx, userId)
	if err != nil {
		return err
	}
	userInfo.UsedSpace -= size
	return st.setUser(ctx, userInfo)
}

func (st *SQLiteStore) delUploadInfoOnly(ctx context.Context, userId uint64, filePath string) error {
	_, err := st.db.ExecContext(
		ctx,
		`delete from t_file_uploading
		where real_path=? and user=?`,
		filePath, userId,
	)
	return err
}

// func (st *SQLiteStore) MoveUploadingInfos(ctx context.Context, userId uint64, uploadPath, itemPath string) error {
// 	st.Lock()
// 	defer st.Unlock()

// 	_, size, _, err := st.getUploadInfo(ctx, userId, itemPath)
// 	if err != nil {
// 		return err
// 	}
// 	err = st.delUploadInfoOnly(ctx, userId, itemPath)
// 	if err != nil {
// 		return err
// 	}
// 	return st.addFileInfo(ctx, userId, itemPath, &db.FileInfo{
// 		Size: size,
// 	})
// }

func (st *SQLiteStore) SetUploadInfo(ctx context.Context, userId uint64, filePath string, newUploaded int64) error {
	st.Lock()
	defer st.Unlock()

	var size, uploaded int64
	err := st.db.QueryRowContext(
		ctx,
		`select size, uploaded
		from t_file_uploading
		where real_path=? and user=?`,
		filePath, userId,
	).Scan(&size, &uploaded)
	if err != nil {
		return err
	} else if newUploaded > size {
		return db.ErrGreaterThanSize
	}

	_, err = st.db.ExecContext(
		ctx,
		`update t_file_uploading
		set uploaded=?
		where real_path=? and user=?`,
		newUploaded, filePath, userId,
	)
	return err
}

func (st *SQLiteStore) getUploadInfo(ctx context.Context, userId uint64, filePath string) (string, int64, int64, error) {
	var size, uploaded int64
	err := st.db.QueryRowContext(
		ctx,
		`select size, uploaded
		from t_file_uploading
		where user=? and real_path=?`,
		userId, filePath,
	).Scan(&size, &uploaded)
	if err != nil {
		return "", 0, 0, err
	}

	return filePath, size, uploaded, nil
}

func (st *SQLiteStore) GetUploadInfo(ctx context.Context, userId uint64, filePath string) (string, int64, int64, error) {
	st.RLock()
	defer st.RUnlock()
	return st.getUploadInfo(ctx, userId, filePath)
}

func (st *SQLiteStore) ListUploadInfos(ctx context.Context, userId uint64) ([]*db.UploadInfo, error) {
	st.RLock()
	defer st.RUnlock()

	rows, err := st.db.QueryContext(
		ctx,
		`select real_path, size, uploaded
		from t_file_uploading
		where user=?`,
		userId,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pathname string
	var size, uploaded int64
	infos := []*db.UploadInfo{}
	for rows.Next() {
		err = rows.Scan(
			&pathname,
			&size,
			&uploaded,
		)
		if err != nil {
			return nil, err
		}

		infos = append(infos, &db.UploadInfo{
			RealFilePath: pathname,
			Size:         size,
			Uploaded:     uploaded,
		})
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return infos, nil
}
