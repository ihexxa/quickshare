package userstore

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ihexxa/quickshare/src/kvstore"
)

const (
	AdminRole   = "admin"
	UserRole    = "user"
	VisitorRole = "visitor"
	InitNs      = "usersInit"
	IDsNs       = "ids"
	UsersNs     = "users"
	RoleListNs  = "roleList"
	InitTimeKey = "initTime"

	defaultSpaceLimit         = 1024 * 1024 * 1024 // 1GB
	defaultUploadSpeedLimit   = 50 * 1024 * 1024   // 50MB
	defaultDownloadSpeedLimit = 50 * 1024 * 1024   // 50MB
)

var (
	ErrReachedLimit = errors.New("reached space limit")
)

func IsReachedLimitErr(err error) bool {
	return err == ErrReachedLimit
}

type Quota struct {
	SpaceLimit         int64 `json:"spaceLimit,string"`
	UploadSpeedLimit   int   `json:"uploadSpeedLimit"`
	DownloadSpeedLimit int   `json:"downloadSpeedLimit"`
}

type User struct {
	ID        uint64 `json:"id,string"`
	Name      string `json:"name"`
	Pwd       string `json:"pwd"`
	Role      string `json:"role"`
	UsedSpace int64  `json:"usedSpace,string"`
	Quota     *Quota `json:"quota"`
}

type IUserStore interface {
	Init(rootName, rootPwd string) error
	IsInited() bool
	AddUser(user *User) error
	DelUser(id uint64) error
	GetUser(id uint64) (*User, error)
	GetUserByName(name string) (*User, error)
	SetInfo(id uint64, user *User) error
	SetUsed(id uint64, incr bool, capacity int64) error
	SetPwd(id uint64, pwd string) error
	ListUsers() ([]*User, error)
	AddRole(role string) error
	DelRole(role string) error
	ListRoles() (map[string]bool, error)
}

type KVUserStore struct {
	store kvstore.IKVStore
	mtx   *sync.RWMutex
}

func NewKVUserStore(store kvstore.IKVStore) (*KVUserStore, error) {
	_, ok := store.GetStringIn(InitNs, InitTimeKey)
	if !ok {
		var err error
		for _, nsName := range []string{
			IDsNs,
			UsersNs,
			InitNs,
			RoleListNs,
		} {
			if err = store.AddNamespace(nsName); err != nil {
				return nil, err
			}
		}
	}

	return &KVUserStore{
		store: store,
		mtx:   &sync.RWMutex{},
	}, nil
}

func (us *KVUserStore) Init(rootName, rootPwd string) error {
	var err error
	err = us.AddUser(&User{
		ID:   0,
		Name: rootName,
		Pwd:  rootPwd,
		Role: AdminRole,
		Quota: &Quota{
			SpaceLimit:         defaultSpaceLimit,
			UploadSpeedLimit:   defaultUploadSpeedLimit,
			DownloadSpeedLimit: defaultDownloadSpeedLimit,
		},
	})
	if err != nil {
		return err
	}

	for _, role := range []string{AdminRole, UserRole, VisitorRole} {
		err = us.AddRole(role)
		if err != nil {
			return err
		}
	}

	return us.store.SetStringIn(InitNs, InitTimeKey, fmt.Sprintf("%d", time.Now().Unix()))
}

func (us *KVUserStore) IsInited() bool {
	_, ok := us.store.GetStringIn(InitNs, InitTimeKey)
	return ok
}

func (us *KVUserStore) AddUser(user *User) error {
	us.mtx.Lock()
	defer us.mtx.Unlock()

	userID := fmt.Sprint(user.ID)
	_, ok := us.store.GetStringIn(UsersNs, userID)
	if ok {
		return fmt.Errorf("userID (%d) exists", user.ID)
	}
	if user.Name == "" || user.Pwd == "" {
		return errors.New("user name or password can not be empty")
	}
	_, ok = us.store.GetStringIn(IDsNs, user.Name)
	if ok {
		return fmt.Errorf("user name (%s) exists", user.Name)
	}

	err := us.store.SetStringIn(IDsNs, user.Name, userID)
	if err != nil {
		return err
	}
	userBytes, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return us.store.SetStringIn(UsersNs, userID, string(userBytes))
}

func (us *KVUserStore) DelUser(id uint64) error {
	us.mtx.Lock()
	defer us.mtx.Unlock()

	userID := fmt.Sprint(id)
	name, ok := us.store.GetStringIn(UsersNs, userID)
	if !ok {
		return fmt.Errorf("userID (%s) does not exist", userID)
	}

	// TODO: add complement operations if part of the actions fails
	err1 := us.store.DelStringIn(IDsNs, name)
	err2 := us.store.DelStringIn(UsersNs, userID)
	if err1 != nil || err2 != nil {
		return fmt.Errorf("get id(%s) user(%s)", err1, err2)
	}
	return nil
}

func (us *KVUserStore) GetUser(id uint64) (*User, error) {
	us.mtx.RLock()
	defer us.mtx.RUnlock()

	userID := fmt.Sprint(id)

	infoStr, ok := us.store.GetStringIn(UsersNs, userID)
	if !ok {
		return nil, fmt.Errorf("user (%s) not found", userID)
	}
	user := &User{}
	err := json.Unmarshal([]byte(infoStr), user)
	if err != nil {
		return nil, err
	}

	gotID, ok := us.store.GetStringIn(IDsNs, user.Name)
	if !ok {
		return nil, fmt.Errorf("user id (%s) not found", user.Name)
	} else if gotID != userID {
		return nil, fmt.Errorf("user id (%s) not match: got(%s) expected(%s)", user.Name, gotID, userID)
	}

	// TODO: use sync.Pool instead
	return user, nil

}

func (us *KVUserStore) GetUserByName(name string) (*User, error) {
	us.mtx.RLock()
	defer us.mtx.RUnlock()

	userID, ok := us.store.GetStringIn(IDsNs, name)
	if !ok {
		return nil, fmt.Errorf("user id (%s) not found", name)
	}
	infoStr, ok := us.store.GetStringIn(UsersNs, userID)
	if !ok {
		return nil, fmt.Errorf("user name (%s) not found", userID)
	}

	user := &User{}
	err := json.Unmarshal([]byte(infoStr), user)
	if err != nil {
		return nil, err
	}
	if user.Name != name {
		return nil, fmt.Errorf("user id (%s) not match: got(%s) expected(%s)", userID, user.Name, name)
	}

	// TODO: use sync.Pool instead
	return user, nil

}

func (us *KVUserStore) SetPwd(id uint64, pwd string) error {
	us.mtx.Lock()
	defer us.mtx.Unlock()

	userID := fmt.Sprint(id)
	infoStr, ok := us.store.GetStringIn(UsersNs, userID)
	if !ok {
		return fmt.Errorf("user (%d) does not exist", id)
	}
	gotUser := &User{}
	err := json.Unmarshal([]byte(infoStr), gotUser)
	if err != nil {
		return err
	} else if gotUser.ID != id {
		return fmt.Errorf("user id key(%d) info(%d) does match", id, gotUser.ID)
	}

	gotUser.Pwd = pwd
	infoBytes, err := json.Marshal(gotUser)
	if err != nil {
		return err
	}
	return us.store.SetStringIn(UsersNs, userID, string(infoBytes))
}

func (us *KVUserStore) SetUsed(id uint64, incr bool, capacity int64) error {
	us.mtx.Lock()
	defer us.mtx.Unlock()

	userID := fmt.Sprint(id)
	infoStr, ok := us.store.GetStringIn(UsersNs, userID)
	if !ok {
		return fmt.Errorf("user (%d) does not exist", id)
	}
	gotUser := &User{}
	err := json.Unmarshal([]byte(infoStr), gotUser)
	if err != nil {
		return err
	} else if gotUser.ID != id {
		return fmt.Errorf("user id key(%d) info(%d) does match", id, gotUser.ID)
	}

	if incr && gotUser.UsedSpace+capacity > int64(gotUser.Quota.SpaceLimit) {
		return ErrReachedLimit
	}

	if incr {
		gotUser.UsedSpace = gotUser.UsedSpace + capacity
	} else {
		gotUser.UsedSpace = gotUser.UsedSpace - capacity
	}
	infoBytes, err := json.Marshal(gotUser)
	if err != nil {
		return err
	}
	return us.store.SetStringIn(UsersNs, userID, string(infoBytes))
}

func (us *KVUserStore) SetInfo(id uint64, user *User) error {
	us.mtx.Lock()
	defer us.mtx.Unlock()

	userID := fmt.Sprint(id)
	infoStr, ok := us.store.GetStringIn(UsersNs, userID)
	if !ok {
		return fmt.Errorf("user (%d) does not exist", id)
	}
	gotUser := &User{}
	err := json.Unmarshal([]byte(infoStr), gotUser)
	if err != nil {
		return err
	} else if gotUser.ID != id {
		return fmt.Errorf("user id key(%d) info(%d) does match", id, gotUser.ID)
	}

	// name and password can not be updated here
	if user.Role != "" {
		gotUser.Role = user.Role
	}
	if user.Quota != nil {
		gotUser.Quota = user.Quota
	}
	if user.UsedSpace > 0 {
		gotUser.UsedSpace = user.UsedSpace
	}

	infoBytes, err := json.Marshal(gotUser)
	if err != nil {
		return err
	}

	return us.store.SetStringIn(UsersNs, userID, string(infoBytes))
}

func (us *KVUserStore) ListUsers() ([]*User, error) {
	us.mtx.RLock()
	defer us.mtx.RUnlock()

	idToInfo, err := us.store.ListStringsIn(UsersNs)
	if err != nil {
		return nil, err
	}

	users := []*User{}
	for _, infoStr := range idToInfo {
		user := &User{}
		err = json.Unmarshal([]byte(infoStr), user)
		if err != nil {
			return nil, err
		}
		user.Pwd = ""

		users = append(users, user)
	}

	return users, nil
}

func (us *KVUserStore) AddRole(role string) error {
	us.mtx.Lock()
	defer us.mtx.Unlock()

	_, ok := us.store.GetBoolIn(RoleListNs, role)
	if ok {
		return fmt.Errorf("role (%s) exists", role)
	}

	return us.store.SetBoolIn(RoleListNs, role, true)
}

func (us *KVUserStore) DelRole(role string) error {
	us.mtx.Lock()
	defer us.mtx.Unlock()

	if role == AdminRole || role == UserRole || role == VisitorRole {
		return errors.New("predefined roles can not be deleted")
	}

	return us.store.DelBoolIn(RoleListNs, role)
}

func (us *KVUserStore) ListRoles() (map[string]bool, error) {
	us.mtx.Lock()
	defer us.mtx.Unlock()

	return us.store.ListBoolsIn(RoleListNs)
}
