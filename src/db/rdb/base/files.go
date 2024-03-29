package base

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

func (st *BaseStore) getFileInfo(ctx context.Context, tx *sql.Tx, itemPath string) (*db.FileInfo, error) {
	var infoStr string
	fInfo := &db.FileInfo{}
	var id uint64
	var isDir bool
	var size int64
	var shareId string
	err := tx.QueryRowContext(
		ctx,
		`select id, is_dir, size, share_id, info
		from t_file_info
		where path=?`,
		itemPath,
	).Scan(
		&id,
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
	fInfo.Id = id
	fInfo.IsDir = isDir
	fInfo.Size = size
	fInfo.ShareID = shareId
	fInfo.Shared = shareId != ""
	return fInfo, nil
}

func (st *BaseStore) GetFileInfo(ctx context.Context, itemPath string) (*db.FileInfo, error) {
	tx, err := st.db.BeginTx(ctx, txOpts)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	info, err := st.getFileInfo(ctx, tx, itemPath)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return info, err
}

func (st *BaseStore) ListFileInfos(ctx context.Context, itemPaths []string) (map[string]*db.FileInfo, error) {
	tx, err := st.db.BeginTx(ctx, txOpts)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// TODO: add pagination
	placeholders := []string{}
	values := []any{}
	for i := 0; i < len(itemPaths); i++ {
		placeholders = append(placeholders, "?")
		values = append(values, itemPaths[i])
	}
	rows, err := tx.QueryContext(
		ctx,
		fmt.Sprintf(
			`select id, path, is_dir, size, share_id, info
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
	var id uint64
	fInfos := map[string]*db.FileInfo{}
	for rows.Next() {
		fInfo := &db.FileInfo{}

		err = rows.Scan(&id, &itemPath, &isDir, &size, &shareId, &fInfoStr)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal([]byte(fInfoStr), fInfo)
		if err != nil {
			return nil, err
		}
		fInfo.Id = id
		fInfo.IsDir = isDir
		fInfo.Size = size
		fInfo.ShareID = shareId
		fInfo.Shared = shareId != ""
		fInfos[itemPath] = fInfo
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return fInfos, nil
}

func (st *BaseStore) addFileInfo(ctx context.Context, tx *sql.Tx, infoId, userId uint64, itemPath string, info *db.FileInfo) error {
	infoStr, err := json.Marshal(info)
	if err != nil {
		return err
	}

	location, err := getLocation(itemPath)
	if err != nil {
		return err
	}

	dirPath, itemName := path.Split(itemPath)
	_, err = tx.ExecContext(
		ctx,
		`insert into t_file_info (
			id, path, user, location, parent, name,
			is_dir, size, share_id, info
		)
		values (
			?, ?, ?, ?, ?, ?,
			?, ?, ?, ?
		)`,
		infoId, itemPath, userId, location, dirPath, itemName,
		info.IsDir, info.Size, info.ShareID, infoStr,
	)
	return err
}

func (st *BaseStore) AddFileInfo(ctx context.Context, infoId, userId uint64, itemPath string, info *db.FileInfo) error {
	tx, err := st.db.BeginTx(ctx, txOpts)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = st.addFileInfo(ctx, tx, infoId, userId, itemPath, info)
	if err != nil {
		return err
	}

	// increase used space
	err = st.setUsed(ctx, tx, userId, true, info.Size)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (st *BaseStore) delFileInfo(ctx context.Context, tx *sql.Tx, itemPath string) error {
	_, err := tx.ExecContext(
		ctx,
		`delete from t_file_info
		where path=?
		`,
		itemPath,
	)
	return err
}

func (st *BaseStore) SetSha1(ctx context.Context, itemPath, sign string) error {
	tx, err := st.db.BeginTx(ctx, txOpts)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	info, err := st.getFileInfo(ctx, tx, itemPath)
	if err != nil {
		return err
	}
	info.Sha1 = sign

	infoStr, err := json.Marshal(info)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(
		ctx,
		`update t_file_info
		set info=?
		where path=?`,
		infoStr,
		itemPath,
	)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (st *BaseStore) DelFileInfo(ctx context.Context, userID uint64, itemPath string) error {
	tx, err := st.db.BeginTx(ctx, txOpts)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// get all children and size
	rows, err := tx.QueryContext(
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
	err = st.setUsed(ctx, tx, userID, false, decrSize)
	if err != nil {
		return err
	}

	// delete file info entries
	_, err = tx.ExecContext(
		ctx,
		fmt.Sprintf(
			`delete from t_file_info
			where path in (%s)`,
			strings.Join(placeholders, ","),
		),
		values...,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (st *BaseStore) MoveFileInfo(ctx context.Context, userId uint64, oldPath, newPath string, isDir bool) error {
	tx, err := st.db.BeginTx(ctx, txOpts)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	info, err := st.getFileInfo(ctx, tx, oldPath)
	if err != nil {
		if errors.Is(err, db.ErrFileInfoNotFound) {
			// info for file does not exist so no need to move it
			// e.g. folder info is not created before
			// TODO: but sometimes it could be a bug
			return nil
		}
		return err
	}
	err = st.delFileInfo(ctx, tx, oldPath)
	if err != nil {
		return err
	}
	err = st.addFileInfo(ctx, tx, info.Id, userId, newPath, info)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func getLocation(itemPath string) (string, error) {
	// location is taken from item path
	itemPathParts := strings.Split(itemPath, "/")
	if len(itemPathParts) == 0 {
		return "", fmt.Errorf("invalid item path '%s'", itemPath)
	}
	return itemPathParts[0], nil
}
