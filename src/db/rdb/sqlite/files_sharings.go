package sqlite

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

func (st *SQLiteStore) generateShareID(payload string) (string, error) {
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

func (st *SQLiteStore) IsSharing(ctx context.Context, userId uint64, dirPath string) bool {
	st.RLock()
	defer st.RUnlock()

	// TODO: differentiate error and not exist
	info, err := st.getFileInfo(ctx, userId, dirPath)
	if err != nil {
		return false
	}
	return info.ShareID != ""
}

func (st *SQLiteStore) GetSharingDir(ctx context.Context, hashID string) (string, error) {
	st.RLock()
	defer st.RUnlock()

	var sharedPath string
	err := st.db.QueryRowContext(
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

func (st *SQLiteStore) AddSharing(ctx context.Context, userId uint64, dirPath string) error {
	st.Lock()
	defer st.Unlock()

	shareID, err := st.generateShareID(dirPath)
	if err != nil {
		return err
	}

	_, err = st.getFileInfo(ctx, userId, dirPath)
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

		_, err = st.db.ExecContext(
			ctx,
			`insert into t_file_info
			(path, user, parent, name, is_dir, size, share_id, info) values (?, ?, ?, ?, ?, ?, ?, ?)`,
			dirPath, userId, parentPath, name, true, 0, shareID, infoStr,
		)
		return err
	}

	_, err = st.db.ExecContext(
		ctx,
		`update t_file_info
		set share_id=?
		where path=? and user=?`,
		shareID, dirPath, userId,
	)
	return err
}

func (st *SQLiteStore) DelSharing(ctx context.Context, userId uint64, dirPath string) error {
	st.Lock()
	defer st.Unlock()

	_, err := st.db.ExecContext(
		ctx,
		`update t_file_info
		set share_id=''
		where path=? and user=?`,
		dirPath,
		userId,
	)
	return err
}

func (st *SQLiteStore) ListUserSharings(ctx context.Context, userId uint64) (map[string]string, error) {
	st.RLock()
	defer st.RUnlock()

	rows, err := st.db.QueryContext(
		ctx,
		`select path, share_id
		from t_file_info
		where user=? and share_id <> ''`,
		userId,
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