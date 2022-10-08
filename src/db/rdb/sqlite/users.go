package sqlite

import (
	"context"

	"github.com/ihexxa/quickshare/src/db"
)

func (st *SQLiteStore) AddUser(ctx context.Context, user *db.User) error {
	st.Lock()
	defer st.Unlock()

	return st.store.AddUser(ctx, user)
}

func (st *SQLiteStore) DelUser(ctx context.Context, id uint64) error {
	st.Lock()
	defer st.Unlock()

	return st.store.DelUser(ctx, id)
}

func (st *SQLiteStore) GetUser(ctx context.Context, id uint64) (*db.User, error) {
	st.RLock()
	defer st.RUnlock()

	return st.store.GetUser(ctx, id)
}

func (st *SQLiteStore) GetUserByName(ctx context.Context, name string) (*db.User, error) {
	st.RLock()
	defer st.RUnlock()

	return st.store.GetUserByName(ctx, name)
}

func (st *SQLiteStore) SetPwd(ctx context.Context, id uint64, pwd string) error {
	st.Lock()
	defer st.Unlock()

	return st.store.SetPwd(ctx, id, pwd)
}

// role + quota
func (st *SQLiteStore) SetInfo(ctx context.Context, id uint64, user *db.User) error {
	st.Lock()
	defer st.Unlock()

	return st.store.SetInfo(ctx, id, user)
}

func (st *SQLiteStore) SetPreferences(ctx context.Context, id uint64, prefers *db.Preferences) error {
	st.Lock()
	defer st.Unlock()

	return st.store.SetPreferences(ctx, id, prefers)
}

func (st *SQLiteStore) SetUsed(ctx context.Context, id uint64, incr bool, capacity int64) error {
	st.Lock()
	defer st.Unlock()

	return st.store.SetUsed(ctx, id, incr, capacity)
}

func (st *SQLiteStore) ResetUsed(ctx context.Context, id uint64, used int64) error {
	st.Lock()
	defer st.Unlock()

	return st.store.ResetUsed(ctx, id, used)
}

func (st *SQLiteStore) ListUsers(ctx context.Context) ([]*db.User, error) {
	st.RLock()
	defer st.RUnlock()

	return st.store.ListUsers(ctx)
}

func (st *SQLiteStore) ListUserIDs(ctx context.Context) (map[string]string, error) {
	st.RLock()
	defer st.RUnlock()

	return st.store.ListUserIDs(ctx)
}

func (st *SQLiteStore) AddRole(role string) error {
	// TODO: implement this after adding grant/revoke
	panic("not implemented")
}

func (st *SQLiteStore) DelRole(role string) error {
	// TODO: implement this after adding grant/revoke
	panic("not implemented")
}

func (st *SQLiteStore) ListRoles() (map[string]bool, error) {
	// TODO: implement this after adding grant/revoke
	panic("not implemented")
}
