package base

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ihexxa/quickshare/src/db"
)

func (st *BaseStore) setUser(ctx context.Context, tx *sql.Tx, user *db.User) error {
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
		user.ID,
	)
	return err
}

func (st *BaseStore) getUser(ctx context.Context, tx *sql.Tx, id uint64) (*db.User, error) {
	user := &db.User{}
	var quotaStr, preferenceStr string
	err := tx.QueryRowContext(
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
			return nil, db.ErrUserNotFound
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

func (st *BaseStore) addUser(ctx context.Context, tx *sql.Tx, user *db.User) error {
	quotaStr, err := json.Marshal(user.Quota)
	if err != nil {
		return err
	}
	preferenceStr, err := json.Marshal(user.Preferences)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(
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

func (st *BaseStore) AddUser(ctx context.Context, user *db.User) error {
	tx, err := st.db.BeginTx(ctx, txOpts)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = st.addUser(ctx, tx, user)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (st *BaseStore) DelUser(ctx context.Context, id uint64) error {
	tx, err := st.db.BeginTx(ctx, txOpts)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(
		ctx,
		`delete from t_user where id=?`,
		id,
	)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (st *BaseStore) GetUser(ctx context.Context, id uint64) (*db.User, error) {
	tx, err := st.db.BeginTx(ctx, txOpts)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	user, err := st.getUser(ctx, tx, id)
	if err != nil {
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return user, err
}

func (st *BaseStore) GetUserByName(ctx context.Context, name string) (*db.User, error) {
	tx, err := st.db.BeginTx(ctx, txOpts)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	user := &db.User{}
	var quotaStr, preferenceStr string
	err = tx.QueryRowContext(
		ctx,
		`select id, name, pwd, role, used_space, quota, preference
		from t_user
		where name=?`,
		name,
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
			return nil, db.ErrUserNotFound
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

	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (st *BaseStore) SetPwd(ctx context.Context, id uint64, pwd string) error {
	tx, err := st.db.BeginTx(ctx, txOpts)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(
		ctx,
		`update t_user
		set pwd=?
		where id=?`,
		pwd,
		id,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// role + quota
func (st *BaseStore) SetInfo(ctx context.Context, id uint64, user *db.User) error {
	tx, err := st.db.BeginTx(ctx, txOpts)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	quotaStr, err := json.Marshal(user.Quota)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(
		ctx,
		`update t_user
		set role=?, quota=?
		where id=?`,
		user.Role, quotaStr,
		id,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (st *BaseStore) SetPreferences(ctx context.Context, id uint64, prefers *db.Preferences) error {
	tx, err := st.db.BeginTx(ctx, txOpts)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	preferenceStr, err := json.Marshal(prefers)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(
		ctx,
		`update t_user
		set preference=?
		where id=?`,
		preferenceStr,
		id,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (st *BaseStore) SetUsed(ctx context.Context, id uint64, incr bool, capacity int64) error {
	tx, err := st.db.BeginTx(ctx, txOpts)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = st.setUsed(ctx, tx, id, incr, capacity)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (st *BaseStore) setUsed(ctx context.Context, tx *sql.Tx, id uint64, incr bool, capacity int64) error {
	gotUser, err := st.getUser(ctx, tx, id)
	if err != nil {
		return err
	}

	if incr && gotUser.UsedSpace+capacity > int64(gotUser.Quota.SpaceLimit) {
		return db.ErrReachedLimit
	}

	if incr {
		gotUser.UsedSpace = gotUser.UsedSpace + capacity
	} else {
		if gotUser.UsedSpace-capacity < 0 {
			return db.ErrNegtiveUsedSpace
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
	return err
}

func (st *BaseStore) ResetUsed(ctx context.Context, id uint64, used int64) error {
	tx, err := st.db.BeginTx(ctx, txOpts)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(
		ctx,
		`update t_user
		set used_space=?
		where id=?`,
		used,
		id,
	)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (st *BaseStore) ListUsers(ctx context.Context) ([]*db.User, error) {
	tx, err := st.db.BeginTx(ctx, txOpts)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// TODO: support pagination
	rows, err := tx.QueryContext(
		ctx,
		`select id, name, role, used_space, quota, preference
		from t_user`,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, db.ErrUserNotFound
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

	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (st *BaseStore) ListUserIDs(ctx context.Context) (map[string]string, error) {
	users, err := st.ListUsers(ctx)
	if err != nil {
		return nil, err
	}

	nameToId := map[string]string{}
	for _, user := range users {
		nameToId[user.Name] = fmt.Sprint(user.ID)
	}
	return nameToId, nil
}

func (st *BaseStore) AddRole(role string) error {
	// TODO: implement this after adding grant/revoke
	panic("not implemented")
}

func (st *BaseStore) DelRole(role string) error {
	// TODO: implement this after adding grant/revoke
	panic("not implemented")
}

func (st *BaseStore) ListRoles() (map[string]bool, error) {
	// TODO: implement this after adding grant/revoke
	panic("not implemented")
}
