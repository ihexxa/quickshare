package base

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ihexxa/quickshare/src/db"
)

func (st *BaseStore) addUploadInfoOnly(ctx context.Context, tx *sql.Tx, uploadId, userId uint64, tmpPath, filePath string, fileSize int64) error {
	_, err := tx.ExecContext(
		ctx,
		`insert into t_file_uploading (
			id, real_path, tmp_path, user, size, uploaded
		)
		values (
			?, ?, ?, ?, ?, ?
		)`,
		uploadId, filePath, tmpPath, userId, fileSize, 0,
	)
	return err
}

func (st *BaseStore) AddUploadInfos(ctx context.Context, uploadId, userId uint64, tmpPath, filePath string, info *db.FileInfo) error {
	tx, err := st.db.BeginTx(ctx, txOpts)
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

	err = st.addUploadInfoOnly(ctx, tx, uploadId, userId, tmpPath, filePath, info.Size)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (st *BaseStore) DelUploadingInfos(ctx context.Context, userId uint64, realPath string) error {
	tx, err := st.db.BeginTx(ctx, txOpts)
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

func (st *BaseStore) delUploadingInfos(ctx context.Context, tx *sql.Tx, userId uint64, realPath string) error {
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

func (st *BaseStore) delUploadInfoOnly(ctx context.Context, tx *sql.Tx, userId uint64, filePath string) error {
	_, err := tx.ExecContext(
		ctx,
		`delete from t_file_uploading
		where real_path=? and user=?`,
		filePath, userId,
	)
	return err
}

func (st *BaseStore) MoveUploadingInfos(ctx context.Context, infoId, userId uint64, uploadPath, itemPath string) error {
	tx, err := st.db.BeginTx(ctx, txOpts)
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
	err = st.addFileInfo(ctx, tx, infoId, userId, itemPath, &db.FileInfo{
		Size: size,
	})
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (st *BaseStore) SetUploadInfo(ctx context.Context, userId uint64, filePath string, newUploaded int64) error {
	tx, err := st.db.BeginTx(ctx, txOpts)
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

func (st *BaseStore) getUploadInfo(ctx context.Context, tx *sql.Tx, userId uint64, filePath string) (string, int64, int64, error) {
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

func (st *BaseStore) GetUploadInfo(ctx context.Context, userId uint64, filePath string) (string, int64, int64, error) {
	tx, err := st.db.BeginTx(ctx, txOpts)
	if err != nil {
		return "", 0, 0, err
	}
	defer tx.Rollback()

	filePath, size, uploaded, err := st.getUploadInfo(ctx, tx, userId, filePath)
	if err != nil {
		return filePath, size, uploaded, err
	}

	err = tx.Commit()
	return filePath, size, uploaded, err
}

func (st *BaseStore) ListUploadInfos(ctx context.Context, userId uint64) ([]*db.UploadInfo, error) {
	tx, err := st.db.BeginTx(ctx, txOpts)
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

	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return infos, nil
}
