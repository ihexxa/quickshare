package sqlite

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ihexxa/quickshare/src/db"
)

func (st *SQLiteStore) addUploadInfoOnly(ctx context.Context, tx *sql.Tx, userId uint64, tmpPath, filePath string, fileSize int64) error {
	_, err := tx.ExecContext(
		ctx,
		`insert into t_file_uploading (
			real_path, tmp_path, user, size, uploaded
		)
		values (
			?, ?, ?, ?, ?
		)`,
		filePath, tmpPath, userId, fileSize, 0,
	)
	return err
}

func (st *SQLiteStore) AddUploadInfos(ctx context.Context, userId uint64, tmpPath, filePath string, info *db.FileInfo) error {
	tx, err := st.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	userInfo, err := st.getUser(ctx, tx, userId)
	if err != nil {
		return err
	} else if userInfo.UsedSpace+info.Size > int64(userInfo.Quota.SpaceLimit) {
		return db.ErrQuota
	}

	_, _, _, err = st.getUploadInfo(ctx, tx, userId, filePath)
	if err == nil {
		return db.ErrKeyExisting
	} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	userInfo.UsedSpace += info.Size
	err = st.setUser(ctx, tx, userInfo)
	if err != nil {
		return err
	}

	err = st.addUploadInfoOnly(ctx, tx, userId, tmpPath, filePath, info.Size)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (st *SQLiteStore) DelUploadingInfos(ctx context.Context, userId uint64, realPath string) error {
	tx, err := st.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = st.delUploadingInfos(ctx, tx, userId, realPath)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (st *SQLiteStore) delUploadingInfos(ctx context.Context, tx *sql.Tx, userId uint64, realPath string) error {
	_, size, _, err := st.getUploadInfo(ctx, tx, userId, realPath)
	if err != nil {
		// info may not exist
		return err
	}

	err = st.delUploadInfoOnly(ctx, tx, userId, realPath)
	if err != nil {
		return err
	}

	userInfo, err := st.getUser(ctx, tx, userId)
	if err != nil {
		return err
	}
	userInfo.UsedSpace -= size
	return st.setUser(ctx, tx, userInfo)
}

func (st *SQLiteStore) delUploadInfoOnly(ctx context.Context, tx *sql.Tx, userId uint64, filePath string) error {
	_, err := tx.ExecContext(
		ctx,
		`delete from t_file_uploading
		where real_path=? and user=?`,
		filePath, userId,
	)
	return err
}

func (st *SQLiteStore) MoveUploadingInfos(ctx context.Context, userId uint64, uploadPath, itemPath string) error {
	tx, err := st.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, size, _, err := st.getUploadInfo(ctx, tx, userId, itemPath)
	if err != nil {
		return err
	}
	err = st.delUploadInfoOnly(ctx, tx, userId, itemPath)
	if err != nil {
		return err
	}
	err = st.addFileInfo(ctx, tx, userId, itemPath, &db.FileInfo{
		Size: size,
	})
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (st *SQLiteStore) SetUploadInfo(ctx context.Context, userId uint64, filePath string, newUploaded int64) error {
	tx, err := st.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var size, uploaded int64
	err = tx.QueryRowContext(
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

	_, err = tx.ExecContext(
		ctx,
		`update t_file_uploading
		set uploaded=?
		where real_path=? and user=?`,
		newUploaded, filePath, userId,
	)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (st *SQLiteStore) getUploadInfo(
	ctx context.Context, tx *sql.Tx, userId uint64, filePath string,
) (string, int64, int64, error) {
	var size, uploaded int64
	err := tx.QueryRowContext(
		ctx,
		`select size, uploaded
		from t_file_uploading
		where real_path=? and user=?`,
		filePath, userId,
	).Scan(&size, &uploaded)
	if err != nil {
		return "", 0, 0, err
	}

	return filePath, size, uploaded, nil
}

func (st *SQLiteStore) GetUploadInfo(ctx context.Context, userId uint64, filePath string) (string, int64, int64, error) {
	tx, err := st.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return "", 0, 0, err
	}
	defer tx.Rollback()

	return st.getUploadInfo(ctx, tx, userId, filePath)
}

func (st *SQLiteStore) ListUploadInfos(ctx context.Context, userId uint64) ([]*db.UploadInfo, error) {
	tx, err := st.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(
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
