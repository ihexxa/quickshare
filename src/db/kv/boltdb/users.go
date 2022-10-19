package boltdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/boltdb/bolt"
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

type IUserStore interface {
	Init(rootName, rootPwd string) error
	IsInited() bool
	AddUser(user *db.User) error
	DelUser(id uint64) error
	GetUser(id uint64) (*db.User, error)
	GetUserByName(name string) (*db.User, error)
	SetInfo(id uint64, user *db.User) error
	SetUsed(id uint64, incr bool, capacity int64) error
	ResetUsed(id uint64, used int64) error
	SetPwd(id uint64, pwd string) error
	SetPreferences(id uint64, settings *db.Preferences) error
	ListUsers() ([]*db.User, error)
	ListUserIDs() (map[string]string, error)
	AddRole(role string) error
	DelRole(role string) error
	ListRoles() (map[string]bool, error)
}

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

func (kv *KVStore) Init(rootName, rootPwd string) error {
	var err error

	for _, namespace := range []string{
		db.UserSchemaNs,
		db.UserIDsNs,
		db.UsersNs,
		db.RolesNs,
	} {
		_, ok := kv.db.GetStringIn(namespace, db.KeyInitTime)
		if !ok {
			if err = kv.db.AddNamespace(namespace); err != nil {
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
		err = kv.AddUser(user)
		if err != nil {
			return err
		}
	}

	for _, role := range []string{db.AdminRole, db.UserRole, db.VisitorRole} {
		err = kv.AddRole(role)
		if err != nil {
			return err
		}
	}

	return kv.db.SetStringIn(db.UserSchemaNs, db.KeyInitTime, fmt.Sprintf("%d", time.Now().Unix()))
}

func (kv *KVStore) IsInited() bool {
	_, ok := kv.db.GetStringIn(db.UserSchemaNs, db.KeyInitTime)
	return ok
}

func (kv *KVStore) setUser(user *db.User) error {
	var err error

	if err = db.CheckUser(user, false); err != nil {
		return err
	}

	userID := fmt.Sprint(user.ID)
	err = kv.db.SetStringIn(db.UserIDsNs, user.Name, userID)
	if err != nil {
		return err
	}
	userBytes, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return kv.db.SetStringIn(db.UsersNs, userID, string(userBytes))
}

func (kv *KVStore) getUser(id uint64) (*db.User, error) {
	userID := fmt.Sprint(id)
	userBytes, ok := kv.db.GetStringIn(db.UsersNs, userID)
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

func (kv *KVStore) getUserByName(name string) (*db.User, error) {
	userID, ok := kv.db.GetStringIn(db.UserIDsNs, name)
	if !ok {
		return nil, ErrUserNotFound
	}

	userBytes, ok := kv.db.GetStringIn(db.UsersNs, userID)
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

func (kv *KVStore) AddUser(user *db.User) error {
	kv.mtx.Lock()
	defer kv.mtx.Unlock()

	return kv.setUser(user)
}

func (kv *KVStore) DelUser(id uint64) error {
	kv.mtx.Lock()
	defer kv.mtx.Unlock()

	user, err := kv.getUser(id)
	if err != nil {
		return err
	}

	// TODO: add complement operations if part of the actions fails
	err1 := kv.db.DelStringIn(db.UserIDsNs, user.Name)
	err2 := kv.db.DelStringIn(db.UsersNs, fmt.Sprint(user.ID))
	if err1 != nil || err2 != nil {
		return fmt.Errorf("DelUser: err1(%s) err2(%s)", err1, err2)
	}
	return nil
}

func (kv *KVStore) GetUser(id uint64) (*db.User, error) {
	kv.mtx.RLock()
	defer kv.mtx.RUnlock()

	return kv.getUser(id)
}

func (kv *KVStore) GetUserByName(name string) (*db.User, error) {
	kv.mtx.RLock()
	defer kv.mtx.RUnlock()

	return kv.getUserByName(name)
}

func (kv *KVStore) SetPwd(id uint64, pwd string) error {
	kv.mtx.Lock()
	defer kv.mtx.Unlock()

	user, err := kv.getUser(id)
	if err != nil {
		return err
	}

	user.Pwd = pwd
	return kv.setUser(user)
}

func (kv *KVStore) SetInfo(id uint64, user *db.User) error {
	kv.mtx.Lock()
	defer kv.mtx.Unlock()

	gotUser, err := kv.getUser(id)
	if err != nil {
		return err
	}

	gotUser.Role = user.Role
	gotUser.Quota = user.Quota
	gotUser.UsedSpace = user.UsedSpace
	return kv.setUser(gotUser)
}

func (kv *KVStore) SetPreferences(id uint64, prefers *db.Preferences) error {
	kv.mtx.Lock()
	defer kv.mtx.Unlock()

	user, err := kv.getUser(id)
	if err != nil {
		return err
	}

	user.Preferences = prefers
	return kv.setUser(user)
}

func (kv *KVStore) SetUsed(id uint64, incr bool, capacity int64) error {
	kv.mtx.Lock()
	defer kv.mtx.Unlock()

	gotUser, err := kv.getUser(id)
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

	return kv.setUser(gotUser)
}

func (kv *KVStore) ResetUsed(id uint64, used int64) error {
	kv.mtx.Lock()
	defer kv.mtx.Unlock()

	gotUser, err := kv.getUser(id)
	if err != nil {
		return err
	}

	gotUser.UsedSpace = used
	return kv.setUser(gotUser)
}

func (kv *KVStore) ListUsers() ([]*db.User, error) {
	kv.mtx.RLock()
	defer kv.mtx.RUnlock()

	idToInfo, err := kv.db.ListStringsIn(db.UsersNs)
	if err != nil {
		return nil, err
	}
	nameToID, err := kv.db.ListStringsIn(db.UserIDsNs)
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
					err = kv.db.DelStringIn(db.UsersNs, fmt.Sprint(user.ID))
					if err != nil {
						return nil, err
					}
				}
			}
		} else {
			for name, id := range nameToID {
				_, ok := idToInfo[id]
				if !ok {
					err = kv.db.DelStringIn(db.UserIDsNs, name)
					if err != nil {
						return nil, err
					}
				}
			}
		}
	}

	return users, nil
}

func (kv *KVStore) ListUserIDs() (map[string]string, error) {
	kv.mtx.RLock()
	defer kv.mtx.RUnlock()

	return kv.db.ListStringsIn(db.UserIDsNs)
}

func (kv *KVStore) getUserInfo(tx *bolt.Tx, userID uint64) (*db.User, error) {
	var err error

	usersBucket := tx.Bucket([]byte(db.UsersNs))
	if usersBucket == nil {
		return nil, db.ErrBucketNotFound
	}

	uidStr := fmt.Sprint(userID)
	infoBytes := usersBucket.Get([]byte(uidStr))
	if infoBytes == nil {
		return nil, db.ErrKeyNotFound
	}

	userInfo := &db.User{}
	err = json.Unmarshal(infoBytes, userInfo)
	if err != nil {
		return nil, err
	} else if userInfo.ID != userID {
		return nil, fmt.Errorf("user id key(%d) info(%d) does match", userID, userInfo.ID)
	}

	if err = db.CheckUser(userInfo, true); err != nil {
		return nil, err
	}

	return userInfo, nil
}

func (kv *KVStore) setUserInfo(tx *bolt.Tx, userID uint64, userInfo *db.User) error {
	var err error

	if err = db.CheckUser(userInfo, false); err != nil {
		return err
	}

	usersBucket := tx.Bucket([]byte(db.UsersNs))
	if usersBucket == nil {
		return db.ErrBucketNotFound
	}

	userInfoBytes, err := json.Marshal(userInfo)
	if err != nil {
		return err
	}

	uidStr := fmt.Sprint(userID)
	return usersBucket.Put([]byte(uidStr), userInfoBytes)
}

func (kv *KVStore) AddRole(role string) error {
	kv.mtx.Lock()
	defer kv.mtx.Unlock()

	_, ok := kv.db.GetBoolIn(db.RolesNs, role)
	if ok {
		return fmt.Errorf("role (%s) exists", role)
	}

	return kv.db.SetBoolIn(db.RolesNs, role, true)
}

func (kv *KVStore) DelRole(role string) error {
	kv.mtx.Lock()
	defer kv.mtx.Unlock()

	if role == db.AdminRole || role == db.UserRole || role == db.VisitorRole {
		return errors.New("predefined roles can not be deleted")
	}

	return kv.db.DelBoolIn(db.RolesNs, role)
}

func (kv *KVStore) ListRoles() (map[string]bool, error) {
	kv.mtx.Lock()
	defer kv.mtx.Unlock()

	return kv.db.ListBoolsIn(db.RolesNs)
}
