package userstore

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ihexxa/quickshare/src/db/sitestore"
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
	VisitorID   = uint64(1)
	VisitorName = "visitor"

	defaultSpaceLimit         = 1024 * 1024 * 1024 // 1GB
	defaultUploadSpeedLimit   = 50 * 1024 * 1024   // 50MB
	defaultDownloadSpeedLimit = 50 * 1024 * 1024   // 50MB
	visitorUploadSpeedLimit   = 10 * 1024 * 1024   // 10MB
	visitorDownloadSpeedLimit = 10 * 1024 * 1024   // 10MB
)

var (
	ErrReachedLimit    = errors.New("reached space limit")
	DefaultPreferences = Preferences{
		Bg: &sitestore.BgConfig{
			Url:      "",
			Repeat:   "no-repeat",
			Position: "center",
			Align:    "fixed",
		},
		CSSURL:     "",
		LanPackURL: "",
		Lan:        "en_US",
	}
)

func IsReachedLimitErr(err error) bool {
	return err == ErrReachedLimit
}

type Quota struct {
	SpaceLimit         int64 `json:"spaceLimit,string"`
	UploadSpeedLimit   int   `json:"uploadSpeedLimit"`
	DownloadSpeedLimit int   `json:"downloadSpeedLimit"`
}

type Preferences struct {
	Bg         *sitestore.BgConfig `json:"bg"`
	CSSURL     string              `json:"cssURL"`
	LanPackURL string              `json:"lanPackURL"`
	Lan        string              `json:"lan"`
}

type UserCfg struct {
	Name string `json:"name"`
	Role string `json:"role"`
	Pwd  string `json:"pwd"`
}

type User struct {
	ID          uint64       `json:"id,string"`
	Name        string       `json:"name"`
	Pwd         string       `json:"pwd"`
	Role        string       `json:"role"`
	UsedSpace   int64        `json:"usedSpace,string"`
	Quota       *Quota       `json:"quota"`
	Preferences *Preferences `json:"preferences"`
}

type IUserStore interface {
	Init(rootName, rootPwd string) error
	IsInited() bool
	AddUser(user *User) error
	DelUser(id uint64) error
	GetUser(id uint64) (*User, error)
	GetUserByName(name string) (*User, error)
	SetInfo(id uint64, user *User) error
	CanIncrUsed(id uint64, capacity int64) (bool, error)
	SetUsed(id uint64, incr bool, capacity int64) error
	SetPwd(id uint64, pwd string) error
	SetPreferences(id uint64, settings *Preferences) error
	ListUsers() ([]*User, error)
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
	adminPreferences := DefaultPreferences
	admin := &User{
		ID:   0,
		Name: rootName,
		Pwd:  rootPwd,
		Role: AdminRole,
		Quota: &Quota{
			SpaceLimit:         defaultSpaceLimit,
			UploadSpeedLimit:   defaultUploadSpeedLimit,
			DownloadSpeedLimit: defaultDownloadSpeedLimit,
		},
		Preferences: &adminPreferences,
	}

	visitorPreferences := DefaultPreferences
	visitor := &User{
		ID:   VisitorID,
		Name: VisitorName,
		Pwd:  rootPwd,
		Role: VisitorRole,
		Quota: &Quota{
			SpaceLimit:         0,
			UploadSpeedLimit:   visitorUploadSpeedLimit,
			DownloadSpeedLimit: visitorDownloadSpeedLimit,
		},
		Preferences: &visitorPreferences,
	}

	for _, user := range []*User{admin, visitor} {
		err = us.AddUser(user)
		if err != nil {
			return err
		}
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
	infoStr, ok := us.store.GetStringIn(UsersNs, userID)
	if !ok {
		return fmt.Errorf("userID (%s) does not exist", userID)
	}
	user := &User{}
	err := json.Unmarshal([]byte(infoStr), user)
	if err != nil {
		return err
	}

	// TODO: add complement operations if part of the actions fails
	err1 := us.store.DelStringIn(IDsNs, user.Name)
	err2 := us.store.DelStringIn(UsersNs, userID)
	if err1 != nil || err2 != nil {
		return fmt.Errorf("DelUser: err1(%s) err2(%s)", err1, err2)
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

func (us *KVUserStore) SetPreferences(id uint64, prefers *Preferences) error {
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

	gotUser.Preferences = prefers
	infoBytes, err := json.Marshal(gotUser)
	if err != nil {
		return err
	}
	return us.store.SetStringIn(UsersNs, userID, string(infoBytes))
}

func (us *KVUserStore) CanIncrUsed(id uint64, capacity int64) (bool, error) {
	us.mtx.Lock()
	defer us.mtx.Unlock()
	userID := fmt.Sprint(id)
	infoStr, ok := us.store.GetStringIn(UsersNs, userID)
	if !ok {
		return false, fmt.Errorf("user (%d) does not exist", id)
	}

	gotUser := &User{}
	err := json.Unmarshal([]byte(infoStr), gotUser)
	if err != nil {
		return false, err
	} else if gotUser.ID != id {
		return false, fmt.Errorf("user id key(%d) info(%d) does match", id, gotUser.ID)
	}

	return gotUser.UsedSpace+capacity <= int64(gotUser.Quota.SpaceLimit), nil
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
	nameToID, err := us.store.ListStringsIn(IDsNs)
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

	// redundant check
	if len(idToInfo) != len(nameToID) {
		if len(idToInfo) > len(nameToID) {
			for _, user := range users {
				_, ok := nameToID[user.Name]
				if !ok {
					err = us.store.DelStringIn(UsersNs, fmt.Sprint(user.ID))
					if err != nil {
						return nil, err
					}
				}
			}
		} else {
			for name, id := range nameToID {
				_, ok := idToInfo[id]
				if !ok {
					err = us.store.DelStringIn(IDsNs, name)
					if err != nil {
						return nil, err
					}
				}
			}
		}
	}

	return users, nil
}

func (us *KVUserStore) ListUserIDs() (map[string]string, error) {
	us.mtx.RLock()
	defer us.mtx.RUnlock()

	return us.store.ListStringsIn(IDsNs)
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
