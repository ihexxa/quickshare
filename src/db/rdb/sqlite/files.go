package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/ihexxa/quickshare/src/db"
)

const (
	InitNs       = "Init"
	InitTimeKey  = "initTime"
	SchemaVerKey = "SchemaVersion"
	SchemaV1     = "v1"
)

var (
	maxHashingTime = 10
)

func (st *SQLiteStore) getFileInfo(ctx context.Context, userId uint64, itemPath string) (*db.FileInfo, error) {
	var infoStr string
	fInfo := &db.FileInfo{}
	var isDir bool
	var size int64
	var shareId string
	err := st.db.QueryRowContext(
		ctx,
		`select is_dir, size, share_id, info
		from t_file_info
		where path=? and user=?
		`,
		itemPath,
		userId,
	).Scan(
		&isDir,
		&size,
		&shareId,
		&infoStr,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, db.ErrFileInfoNotFound
		}
		return nil, err
	}

	err = json.Unmarshal([]byte(infoStr), &fInfo)
	if err != nil {
		return nil, err
	}
	fInfo.IsDir = isDir
	fInfo.Size = size
	fInfo.ShareID = shareId
	fInfo.Shared = shareId != ""
	return fInfo, nil
}

func (st *SQLiteStore) GetFileInfo(ctx context.Context, userId uint64, itemPath string) (*db.FileInfo, error) {
	st.RLock()
	defer st.RUnlock()

	return st.getFileInfo(ctx, userId, itemPath)
}

func (st *SQLiteStore) ListFileInfos(ctx context.Context, itemPaths []string) (map[string]*db.FileInfo, error) {
	st.RLock()
	defer st.RUnlock()

	// TODO: add pagination
	placeholders := []string{}
	values := []any{}
	for i := 0; i < len(itemPaths); i++ {
		placeholders = append(placeholders, "?")
		values = append(values, itemPaths[i])
	}
	rows, err := st.db.QueryContext(
		ctx,
		fmt.Sprintf(
			`select path, is_dir, size, share_id, info
			from t_file_info
			where path in (%s)
			`,
			strings.Join(placeholders, ","),
		),
		values...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fInfoStr, itemPath, shareId string
	var isDir bool
	var size int64
	fInfos := map[string]*db.FileInfo{}
	for rows.Next() {
		fInfo := &db.FileInfo{}

		err = rows.Scan(&itemPath, &isDir, &size, &shareId, &fInfoStr)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal([]byte(fInfoStr), fInfo)
		if err != nil {
			return nil, err
		}
		fInfo.IsDir = isDir
		fInfo.Size = size
		fInfo.ShareID = shareId
		fInfo.Shared = shareId != ""
		fInfos[itemPath] = fInfo
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return fInfos, nil
}

func (st *SQLiteStore) addFileInfo(ctx context.Context, userId uint64, itemPath string, info *db.FileInfo) error {
	infoStr, err := json.Marshal(info)
	if err != nil {
		return err
	}

	dirPath, itemName := path.Split(itemPath)
	_, err = st.db.ExecContext(
		ctx,
		`insert into t_file_info
		(path, user, parent, name, is_dir, size, share_id, info) values (?, ?, ?, ?, ?, ?, ?, ?)`,
		itemPath,
		userId,
		dirPath,
		itemName,
		info.IsDir,
		info.Size,
		info.ShareID,
		infoStr,
	)
	return err
}

func (st *SQLiteStore) AddFileInfo(ctx context.Context, userId uint64, itemPath string, info *db.FileInfo) error {
	st.Lock()
	defer st.Unlock()

	err := st.addFileInfo(ctx, userId, itemPath, info)
	if err != nil {
		return err
	}

	// increase used space
	return st.setUsed(ctx, userId, true, info.Size)
}

func (st *SQLiteStore) delFileInfo(ctx context.Context, userId uint64, itemPath string) error {
	_, err := st.db.ExecContext(
		ctx,
		`delete from t_file_info
		where path=? and user=?
		`,
		itemPath,
		userId,
	)
	return err
}

// func (st *SQLiteStore) DelFileInfo(ctx context.Context, itemPath string) error {
// 	st.Lock()
// 	defer st.Unlock()
// 	return st.delFileInfo(ctx, itemPath)
// }

// sharings

func (st *SQLiteStore) SetSha1(ctx context.Context, userId uint64, itemPath, sign string) error {
	st.Lock()
	defer st.Unlock()

	info, err := st.getFileInfo(ctx, userId, itemPath)
	if err != nil {
		return err
	}
	info.Sha1 = sign

	infoStr, err := json.Marshal(info)
	if err != nil {
		return err
	}

	_, err = st.db.ExecContext(
		ctx,
		`update t_file_info
		set info=?
		where path=? and user=?`,
		infoStr,
		itemPath,
		userId,
	)
	return err
}

func (st *SQLiteStore) DelFileInfo(ctx context.Context, userID uint64, itemPath string) error {
	st.Lock()
	defer st.Unlock()

	// get all children and size
	rows, err := st.db.QueryContext(
		ctx,
		`select path, size
		from t_file_info
		where path = ? or path like ?
		`,
		itemPath,
		fmt.Sprintf("%s/%%", itemPath),
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	var childrenPath string
	var itemSize int64
	placeholders := []string{}
	values := []any{}
	decrSize := int64(0)
	for rows.Next() {
		err = rows.Scan(&childrenPath, &itemSize)
		if err != nil {
			return err
		}
		placeholders = append(placeholders, "?")
		values = append(values, childrenPath)
		decrSize += itemSize
	}

	// decrease used space
	err = st.setUsed(ctx, userID, false, decrSize)
	if err != nil {
		return err
	}

	// delete file info entries
	_, err = st.db.ExecContext(
		ctx,
		fmt.Sprintf(
			`delete from t_file_info
			where path in (%s)`,
			strings.Join(placeholders, ","),
		),
		values...,
	)
	return err
}

func (st *SQLiteStore) MoveFileInfos(ctx context.Context, userId uint64, oldPath, newPath string, isDir bool) error {
	st.Lock()
	defer st.Unlock()

	info, err := st.getFileInfo(ctx, userId, oldPath)
	if err != nil {
		return err
	}
	err = st.delFileInfo(ctx, userId, oldPath)
	if err != nil {
		return err
	}
	return st.addFileInfo(ctx, userId, newPath, info)
}
