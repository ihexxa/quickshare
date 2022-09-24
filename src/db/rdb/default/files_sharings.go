package default

import (
	"context"
	"crypto/sha1"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"time"

	"github.com/ihexxa/quickshare/src/db"
)

func (st *DefaultStore) generateShareID(payload string) (string, error) {
	if len(payload) == 0 {
		return "", db.ErrEmpty
	}

	msg := fmt.Sprintf("%s-%d", payload, time.Now().Unix())
	h := sha1.New()
	_, err := io.WriteString(h, msg)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil))[:7], nil
}

func (st *DefaultStore) IsSharing(ctx context.Context, dirPath string) (bool, error) {
	tx, err := st.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	var shareId string
	err = tx.QueryRowContext(
		ctx,
		`select share_id
		from t_file_info
		where path=?`,
		dirPath,
	).Scan(
		&shareId,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, db.ErrFileInfoNotFound
		}
		return false, err
	}

	return shareId != "", nil
}

func (st *DefaultStore) GetSharingDir(ctx context.Context, hashID string) (string, error) {
	tx, err := st.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	var sharedPath string
	err = tx.QueryRowContext(
		ctx,
		`select path
		from t_file_info
		where share_id=?
		`,
		hashID,
	).Scan(
		&sharedPath,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", db.ErrSharingNotFound
		}
		return "", err
	}

	return sharedPath, nil
}

func (st *DefaultStore) AddSharing(ctx context.Context, userId uint64, dirPath string) error {
	tx, err := st.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	shareID, err := st.generateShareID(dirPath)
	if err != nil {
		return err
	}

	location, err := getLocation(dirPath)
	if err != nil {
		return err
	}

	_, err = st.getFileInfo(ctx, tx, dirPath)
	if err != nil && !errors.Is(err, db.ErrFileInfoNotFound) {
		return err
	}

	if errors.Is(err, db.ErrFileInfoNotFound) {
		// insert new
		parentPath, name := path.Split(dirPath)
		info := &db.FileInfo{Shared: true} // TODO: deprecate shared in info
		infoStr, err := json.Marshal(info)
		if err != nil {
			return err
		}

		_, err = tx.ExecContext(
			ctx,
			`insert into t_file_info (
				path, user, location, parent, name,
				is_dir, size, share_id, info
			)
			values (
				?, ?, ?, ?, ?,
				?, ?, ?, ?
			)`,
			dirPath, userId, location, parentPath, name,
			true, 0, shareID, infoStr,
		)
		if err != nil {
			return err
		}
	} else {
		_, err = tx.ExecContext(
			ctx,
			`update t_file_info
			set share_id=?
			where path=?`,
			shareID, dirPath,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (st *DefaultStore) DelSharing(ctx context.Context, userId uint64, dirPath string) error {
	tx, err := st.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(
		ctx,
		`update t_file_info
		set share_id=''
		where path=?`,
		dirPath,
	)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (st *DefaultStore) ListSharingsByLocation(ctx context.Context, location string) (map[string]string, error) {
	tx, err := st.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(
		ctx,
		`select path, share_id
		from t_file_info
		where share_id<>'' and location=?`,
		location,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pathname, shareId string
	pathToShareId := map[string]string{}
	for rows.Next() {
		err = rows.Scan(&pathname, &shareId)
		if err != nil {
			return nil, err
		}
		pathToShareId[pathname] = shareId
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return pathToShareId, nil
}