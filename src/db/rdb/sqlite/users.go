package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	// "errors"
	"fmt"
	// "sync"
	// "time"

	"github.com/ihexxa/quickshare/src/db"
	"github.com/ihexxa/quickshare/src/db/rdb"
	// "github.com/ihexxa/quickshare/src/kvstore"
)

// TODO: use sync.Pool instead

const (
	VisitorID   = uint64(1)
	VisitorName = "visitor"
)

var (
	ErrReachedLimit     = errors.New("reached space limit")
	ErrUserNotFound     = errors.New("user not found")
	ErrNegtiveUsedSpace = errors.New("used space can not be negative")
)

type SQLiteUsers struct {
	db rdb.IDB
}

func NewSQLiteUsers(db rdb.IDB) (*SQLiteUsers, error) {
	return &SQLiteUsers{db: db}, nil
}

func (u *SQLiteUsers) Init(ctx context.Context, rootName, rootPwd string) error {
	_, err := u.db.ExecContext(
		ctx,
		`create table if not exists t_user (
			id bigint not null,
			name varchar not null,
			pwd varchar not null,
			role integer not null,
			used_space bigint not null,
			quota varchar not null,
			preference varchar not null,
			primary key(id)
		)`,
	)
	if err != nil {
		return err
	}

	admin := &db.User{
		ID:   0,
		Name: rootName,
		Pwd:  rootPwd,
		Role: db.AdminRole,
		Quota: &db.Quota{
			SpaceLimit:         db.DefaultSpaceLimit,
			UploadSpeedLimit:   db.DefaultUploadSpeedLimit,
			DownloadSpeedLimit: db.DefaultDownloadSpeedLimit,
		},
		Preferences: &db.DefaultPreferences,
	}
	visitor := &db.User{
		ID:   VisitorID,
		Name: VisitorName,
		Pwd:  rootPwd,
		Role: db.VisitorRole,
		Quota: &db.Quota{
			SpaceLimit:         0,
			UploadSpeedLimit:   db.VisitorUploadSpeedLimit,
			DownloadSpeedLimit: db.VisitorDownloadSpeedLimit,
		},
		Preferences: &db.DefaultPreferences,
	}
	for _, user := range []*db.User{admin, visitor} {
		err = u.AddUser(ctx, user)
		if err != nil {
			return err
		}
	}

	return nil
}

func (u *SQLiteUsers) IsInited() bool {
	// always try to init the db
	return false
}

// t_users
// id, name, pwd, role, used_space, config
func (u *SQLiteUsers) setUser(ctx context.Context, tx *sql.Tx, user *db.User) error {
	var err error
	if err = db.CheckUser(user, false); err != nil {
		return err
	}

	quotaStr, err := json.Marshal(user.Quota)
	if err != nil {
		return err
	}
	preferencesStr, err := json.Marshal(user.Preferences)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(
		ctx,
		`update t_user
		set name=?, pwd=?, role=?, used_space=?, quota=?, preference=?
		where id=?`,
		user.Name,
		user.Pwd,
		user.Role,
		user.UsedSpace,
		quotaStr,
		preferencesStr,
	)
	return err
}

func (u *SQLiteUsers) getUser(ctx context.Context, tx *sql.Tx, id uint64) (*db.User, error) {
	var err error

	user := &db.User{}
	var quotaStr, preferenceStr string
	err = tx.QueryRowContext(
		ctx,
		`select id, name, pwd, role, used_space, quota, preference
		from t_user
		where id=?`,
		id,
	).Scan(
		&user.ID,
		&user.Name,
		&user.Pwd,
		&user.Role,
		&user.UsedSpace,
		&quotaStr,
		&preferenceStr,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	err = json.Unmarshal([]byte(quotaStr), &user.Quota)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(preferenceStr), &user.Preferences)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *SQLiteUsers) AddUser(ctx context.Context, user *db.User) error {
	quotaStr, err := json.Marshal(user.Quota)
	if err != nil {
		return err
	}
	preferenceStr, err := json.Marshal(user.Preferences)
	if err != nil {
		return err
	}
	_, err = u.db.ExecContext(
		ctx,
		`insert into t_user (id, name, pwd, role, used_space, quota, preference) values (?, ?, ?, ?, ?, ?, ?)`,
		user.ID,
		user.Name,
		user.Pwd,
		user.Role,
		user.UsedSpace,
		quotaStr,
		preferenceStr,
	)
	return err
}

func (u *SQLiteUsers) DelUser(ctx context.Context, id uint64) error {
	_, err := u.db.ExecContext(
		ctx,
		`delete from t_user where id=?`,
		id,
	)
	return err
}

func (u *SQLiteUsers) GetUser(ctx context.Context, id uint64) (*db.User, error) {
	tx, err := u.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, err
	}

	user, err := u.getUser(ctx, tx, id)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return user, err
}

func (u *SQLiteUsers) GetUserByName(ctx context.Context, name string) (*db.User, error) {
	user := &db.User{}
	var quotaStr, preferenceStr string
	err := u.db.QueryRowContext(
		ctx,
		`select id, name, role, used_space, quota, preference
		from t_user
		where name=?`,
		name,
	).Scan(
		&user.ID,
		&user.Name,
		&user.Role,
		&user.UsedSpace,
		&quotaStr,
		&preferenceStr,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	err = json.Unmarshal([]byte(quotaStr), &user.Quota)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(preferenceStr), &user.Preferences)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *SQLiteUsers) SetPwd(ctx context.Context, id uint64, pwd string) error {
	_, err := u.db.ExecContext(
		ctx,
		`update t_user
		set pwd=?
		where id=?`,
		pwd,
		id,
	)
	return err
}

// role + quota
func (u *SQLiteUsers) SetInfo(ctx context.Context, id uint64, user *db.User) error {
	quotaStr, err := json.Marshal(user.Quota)
	if err != nil {
		return err
	}

	_, err = u.db.ExecContext(
		ctx,
		`update t_user
		set role=?, quota=?
		where id=?`,
		user.Role, quotaStr,
		id,
	)
	return err
}

func (u *SQLiteUsers) SetPreferences(ctx context.Context, id uint64, prefers *db.Preferences) error {
	preferenceStr, err := json.Marshal(prefers)
	if err != nil {
		return err
	}

	_, err = u.db.ExecContext(
		ctx,
		`update t_user
		set preference=?
		where id=?`,
		preferenceStr,
		id,
	)
	return err
}

func (u *SQLiteUsers) SetUsed(ctx context.Context, id uint64, incr bool, capacity int64) error {
	tx, err := u.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	gotUser, err := u.getUser(ctx, tx, id)
	if err != nil {
		return err
	}

	if incr && gotUser.UsedSpace+capacity > int64(gotUser.Quota.SpaceLimit) {
		return ErrReachedLimit
	}

	if incr {
		gotUser.UsedSpace = gotUser.UsedSpace + capacity
	} else {
		if gotUser.UsedSpace-capacity < 0 {
			return ErrNegtiveUsedSpace
		}
		gotUser.UsedSpace = gotUser.UsedSpace - capacity
	}

	_, err = tx.ExecContext(
		ctx,
		`update t_user
		set used_space=?
		where id=?`,
		gotUser.UsedSpace,
		gotUser.ID,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (u *SQLiteUsers) ResetUsed(ctx context.Context, id uint64, used int64) error {
	_, err := u.db.ExecContext(
		ctx,
		`update t_user
		set used_space=?
		where id=?`,
		used,
		id,
	)
	return err
}

func (u *SQLiteUsers) ListUsers(ctx context.Context) ([]*db.User, error) {
	// TODO: support pagination
	rows, err := u.db.QueryContext(
		ctx,
		`select id, name, role, used_space, quota, preference
		from t_user`,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	defer rows.Close() // TODO: check error

	users := []*db.User{}
	for rows.Next() {
		user := &db.User{}
		var quotaStr, preferenceStr string
		err = rows.Scan(
			&user.ID,
			&user.Name,
			&user.Role,
			&user.UsedSpace,
			&quotaStr,
			&preferenceStr,
		)
		err = json.Unmarshal([]byte(quotaStr), &user.Quota)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal([]byte(preferenceStr), &user.Preferences)
		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return users, nil
}

func (u *SQLiteUsers) ListUserIDs(ctx context.Context) (map[string]string, error) {
	users, err := u.ListUsers(ctx)
	if err != nil {
		return nil, err
	}

	nameToId := map[string]string{}
	for _, user := range users {
		nameToId[user.Name] = fmt.Sprint(user.ID)
	}
	return nameToId, nil
}

func (u *SQLiteUsers) AddRole(role string) error {
	// TODO: implement this after adding grant/revoke
	panic("not implemented")
}

func (u *SQLiteUsers) DelRole(role string) error {
	// TODO: implement this after adding grant/revoke
	panic("not implemented")
}

func (u *SQLiteUsers) ListRoles() (map[string]bool, error) {
	// TODO: implement this after adding grant/revoke
	panic("not implemented")
}
