package userstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ihexxa/quickshare/src/db"
	"github.com/ihexxa/quickshare/src/kvstore"
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

type KVUserStore struct {
	store kvstore.IKVStore
	mtx   *sync.RWMutex
}

func NewKVUserStore(store kvstore.IKVStore) (*KVUserStore, error) {
	return &KVUserStore{
		store: store,
		mtx:   &sync.RWMutex{},
	}, nil
}

func (us *KVUserStore) Init(ctx context.Context, rootName, rootPwd string) error {
	var err error

	for _, namespace := range []string{
		db.UserSchemaNs,
		db.UserIDsNs,
		db.UsersNs,
		db.RolesNs,
	} {
		_, ok := us.store.GetStringIn(namespace, db.KeyInitTime)
		if !ok {
			if err = us.store.AddNamespace(namespace); err != nil {
				return err
			}
		}
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
		err = us.AddUser(context.TODO(), user)
		if err != nil {
			return err
		}
	}

	for _, role := range []string{db.AdminRole, db.UserRole, db.VisitorRole} {
		err = us.AddRole(role)
		if err != nil {
			return err
		}
	}

	return us.store.SetStringIn(db.UserSchemaNs, db.KeyInitTime, fmt.Sprintf("%d", time.Now().Unix()))
}

func (us *KVUserStore) IsInited() bool {
	_, ok := us.store.GetStringIn(db.UserSchemaNs, db.KeyInitTime)
	return ok
}

func (us *KVUserStore) setUser(user *db.User) error {
	var err error

	if err = db.CheckUser(user, false); err != nil {
		return err
	}

	userID := fmt.Sprint(user.ID)
	err = us.store.SetStringIn(db.UserIDsNs, user.Name, userID)
	if err != nil {
		return err
	}
	userBytes, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return us.store.SetStringIn(db.UsersNs, userID, string(userBytes))
}

func (us *KVUserStore) getUser(id uint64) (*db.User, error) {
	userID := fmt.Sprint(id)
	userBytes, ok := us.store.GetStringIn(db.UsersNs, userID)
	if !ok {
		return nil, ErrUserNotFound
	}

	user := &db.User{}
	err := json.Unmarshal([]byte(userBytes), user)
	if err != nil {
		return nil, err
	}

	if err = db.CheckUser(user, true); err != nil {
		return nil, err
	}
	return user, nil
}

func (us *KVUserStore) getUserByName(name string) (*db.User, error) {
	userID, ok := us.store.GetStringIn(db.UserIDsNs, name)
	if !ok {
		return nil, ErrUserNotFound
	}

	userBytes, ok := us.store.GetStringIn(db.UsersNs, userID)
	if !ok {
		return nil, ErrUserNotFound
	}

	user := &db.User{}
	err := json.Unmarshal([]byte(userBytes), user)
	if err != nil {
		return nil, err
	}

	if err = db.CheckUser(user, true); err != nil {
		return nil, err
	}
	return user, nil
}

func (us *KVUserStore) AddUser(ctx context.Context, user *db.User) error {
	us.mtx.Lock()
	defer us.mtx.Unlock()

	return us.setUser(user)
}

func (us *KVUserStore) DelUser(ctx context.Context, id uint64) error {
	us.mtx.Lock()
	defer us.mtx.Unlock()

	user, err := us.getUser(id)
	if err != nil {
		return err
	}

	// TODO: add complement operations if part of the actions fails
	err1 := us.store.DelStringIn(db.UserIDsNs, user.Name)
	err2 := us.store.DelStringIn(db.UsersNs, fmt.Sprint(user.ID))
	if err1 != nil || err2 != nil {
		return fmt.Errorf("DelUser: err1(%s) err2(%s)", err1, err2)
	}
	return nil
}

func (us *KVUserStore) GetUser(ctx context.Context, id uint64) (*db.User, error) {
	us.mtx.RLock()
	defer us.mtx.RUnlock()

	return us.getUser(id)
}

func (us *KVUserStore) GetUserByName(ctx context.Context, name string) (*db.User, error) {
	us.mtx.RLock()
	defer us.mtx.RUnlock()

	return us.getUserByName(name)
}

func (us *KVUserStore) SetPwd(ctx context.Context, id uint64, pwd string) error {
	us.mtx.Lock()
	defer us.mtx.Unlock()

	user, err := us.getUser(id)
	if err != nil {
		return err
	}

	user.Pwd = pwd
	return us.setUser(user)
}

func (us *KVUserStore) SetInfo(ctx context.Context, id uint64, user *db.User) error {
	us.mtx.Lock()
	defer us.mtx.Unlock()

	gotUser, err := us.getUser(id)
	if err != nil {
		return err
	}

	gotUser.Role = user.Role
	gotUser.Quota = user.Quota
	gotUser.UsedSpace = user.UsedSpace
	return us.setUser(gotUser)
}

func (us *KVUserStore) SetPreferences(ctx context.Context, id uint64, prefers *db.Preferences) error {
	us.mtx.Lock()
	defer us.mtx.Unlock()

	user, err := us.getUser(id)
	if err != nil {
		return err
	}

	user.Preferences = prefers
	return us.setUser(user)
}

func (us *KVUserStore) SetUsed(ctx context.Context, id uint64, incr bool, capacity int64) error {
	us.mtx.Lock()
	defer us.mtx.Unlock()

	gotUser, err := us.getUser(id)
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

	return us.setUser(gotUser)
}

func (us *KVUserStore) ResetUsed(ctx context.Context, id uint64, used int64) error {
	us.mtx.Lock()
	defer us.mtx.Unlock()

	gotUser, err := us.getUser(id)
	if err != nil {
		return err
	}

	gotUser.UsedSpace = used
	return us.setUser(gotUser)
}

func (us *KVUserStore) ListUsers(ctx context.Context) ([]*db.User, error) {
	us.mtx.RLock()
	defer us.mtx.RUnlock()

	idToInfo, err := us.store.ListStringsIn(db.UsersNs)
	if err != nil {
		return nil, err
	}
	nameToID, err := us.store.ListStringsIn(db.UserIDsNs)
	if err != nil {
		return nil, err
	}

	users := []*db.User{}
	for _, infoStr := range idToInfo {
		user := &db.User{}
		err = json.Unmarshal([]byte(infoStr), user)
		if err != nil {
			return nil, err
		}
		user.Pwd = ""

		if err = db.CheckUser(user, true); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	// redundant check
	if len(idToInfo) != len(nameToID) {
		if len(idToInfo) > len(nameToID) {
			for _, user := range users {
				_, ok := nameToID[user.Name]
				if !ok {
					err = us.store.DelStringIn(db.UsersNs, fmt.Sprint(user.ID))
					if err != nil {
						return nil, err
					}
				}
			}
		} else {
			for name, id := range nameToID {
				_, ok := idToInfo[id]
				if !ok {
					err = us.store.DelStringIn(db.UserIDsNs, name)
					if err != nil {
						return nil, err
					}
				}
			}
		}
	}

	return users, nil
}

func (us *KVUserStore) ListUserIDs(ctx context.Context) (map[string]string, error) {
	us.mtx.RLock()
	defer us.mtx.RUnlock()

	return us.store.ListStringsIn(db.UserIDsNs)
}

func (us *KVUserStore) AddRole(role string) error {
	us.mtx.Lock()
	defer us.mtx.Unlock()

	_, ok := us.store.GetBoolIn(db.RolesNs, role)
	if ok {
		return fmt.Errorf("role (%s) exists", role)
	}

	return us.store.SetBoolIn(db.RolesNs, role, true)
}

func (us *KVUserStore) DelRole(role string) error {
	us.mtx.Lock()
	defer us.mtx.Unlock()

	if role == db.AdminRole || role == db.UserRole || role == db.VisitorRole {
		return errors.New("predefined roles can not be deleted")
	}

	return us.store.DelBoolIn(db.RolesNs, role)
}

func (us *KVUserStore) ListRoles() (map[string]bool, error) {
	us.mtx.Lock()
	defer us.mtx.Unlock()

	return us.store.ListBoolsIn(db.RolesNs)
}
