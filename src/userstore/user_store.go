package userstore

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	// "golang.org/x/crypto/bcrypt"

	"github.com/ihexxa/quickshare/src/kvstore"
)

const (
	AdminRole   = "admin"
	UserRole    = "user"
	VisitorRole = "visitor"
	InitNs      = "usersInit"
	IDsNs       = "ids"
	NamesNs     = "users"
	PwdsNs      = "pwds"
	RolesNs     = "roles"
	RoleListNs  = "roleList"
	InitTimeKey = "initTime"
)

type User struct {
	ID   uint64 `json:"id,string"`
	Name string `json:"name"`
	Pwd  string `json:"pwd"`
	Role string `json:"role"`
}

type IUserStore interface {
	Init(rootName, rootPwd string) error
	IsInited() bool
	AddUser(user *User) error
	DelUser(id uint64) error
	GetUser(id uint64) (*User, error)
	GetUserByName(name string) (*User, error)
	SetName(id uint64, name string) error
	SetPwd(id uint64, pwd string) error
	ListUsers() ([]*User, error)
	SetRole(id uint64, role string) error
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
			NamesNs,
			PwdsNs,
			RolesNs,
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
	_, ok := us.store.GetStringIn(NamesNs, userID)
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

	var err error
	err = us.store.SetStringIn(IDsNs, user.Name, userID)
	if err != nil {
		return err
	}
	err = us.store.SetStringIn(NamesNs, userID, user.Name)
	if err != nil {
		return err
	}
	err = us.store.SetStringIn(PwdsNs, userID, user.Pwd)
	if err != nil {
		return err
	}
	return us.store.SetStringIn(RolesNs, userID, user.Role)
}

func (us *KVUserStore) DelUser(id uint64) error {
	us.mtx.Lock()
	defer us.mtx.Unlock()

	userID := fmt.Sprint(id)
	name, ok := us.store.GetStringIn(NamesNs, userID)
	if !ok {
		return fmt.Errorf("userID (%s) exists", userID)
	}

	// TODO: add complement operations if part of the actions fails
	err1 := us.store.DelStringIn(NamesNs, userID)
	err2 := us.store.DelStringIn(IDsNs, name)
	err3 := us.store.DelStringIn(PwdsNs, userID)
	err4 := us.store.DelStringIn(RolesNs, userID)
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		return fmt.Errorf("get name(%s) id(%s) pwd(%s) role(%s)", err1, err2, err3, err4)
	}
	return nil
}

func (us *KVUserStore) GetUser(id uint64) (*User, error) {
	us.mtx.RLock()
	defer us.mtx.RUnlock()

	userID := fmt.Sprint(id)

	name, ok := us.store.GetStringIn(NamesNs, userID)
	if !ok {
		return nil, fmt.Errorf("name (%s) not found", userID)
	}
	gotID, ok := us.store.GetStringIn(IDsNs, name)
	if !ok {
		return nil, fmt.Errorf("user id (%s) not found", name)
	} else if gotID != userID {
		return nil, fmt.Errorf("user id (%s) not match: got(%s) expected(%s)", name, gotID, userID)
	}
	pwd, ok := us.store.GetStringIn(PwdsNs, userID)
	if !ok {
		return nil, fmt.Errorf("pwd (%s) not found", userID)
	}
	role, ok := us.store.GetStringIn(RolesNs, userID)
	if !ok {
		return nil, fmt.Errorf("role (%s) not found", userID)
	}

	// TODO: use sync.Pool instead
	return &User{
		ID:   id,
		Name: name,
		Pwd:  pwd,
		Role: role,
	}, nil

}

func (us *KVUserStore) GetUserByName(name string) (*User, error) {
	us.mtx.RLock()
	defer us.mtx.RUnlock()

	userID, ok := us.store.GetStringIn(IDsNs, name)
	if !ok {
		return nil, fmt.Errorf("user (%s) not found", name)
	}
	gotName, ok := us.store.GetStringIn(NamesNs, userID)
	if !ok {
		return nil, fmt.Errorf("user name (%s) not found", userID)
	} else if gotName != name {
		return nil, fmt.Errorf("user id (%s) not match: got(%s) expected(%s)", name, gotName, name)
	}
	pwd, ok := us.store.GetStringIn(PwdsNs, userID)
	if !ok {
		return nil, fmt.Errorf("pwd (%s) not found", userID)
	}
	role, ok := us.store.GetStringIn(RolesNs, userID)
	if !ok {
		return nil, fmt.Errorf("role (%s) not found", userID)
	}

	uid, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		return nil, err
	}
	// TODO: use sync.Pool instead
	return &User{
		ID:   uid,
		Name: name,
		Pwd:  pwd,
		Role: role,
	}, nil

}

func (us *KVUserStore) SetName(id uint64, name string) error {
	us.mtx.Lock()
	defer us.mtx.Unlock()

	_, ok := us.store.GetStringIn(IDsNs, name)
	if ok {
		return fmt.Errorf("user name (%s) exists", name)
	}

	userID := fmt.Sprint(id)
	_, ok = us.store.GetStringIn(NamesNs, userID)
	if !ok {
		return fmt.Errorf("Name (%d) does not exist", id)
	}
	if name == "" {
		return fmt.Errorf("Name can not be empty")
	}

	err := us.store.SetStringIn(IDsNs, name, userID)
	if err != nil {
		return err
	}
	return us.store.SetStringIn(NamesNs, userID, name)
}

func (us *KVUserStore) SetPwd(id uint64, pwd string) error {
	us.mtx.Lock()
	defer us.mtx.Unlock()

	userID := fmt.Sprint(id)
	_, ok := us.store.GetStringIn(PwdsNs, userID)
	if !ok {
		return fmt.Errorf("Pwd (%d) does not exist", id)
	}

	return us.store.SetStringIn(PwdsNs, userID, pwd)
}

func (us *KVUserStore) SetRole(id uint64, role string) error {
	us.mtx.Lock()
	defer us.mtx.Unlock()

	userID := fmt.Sprint(id)
	_, ok := us.store.GetStringIn(RolesNs, userID)
	if !ok {
		return fmt.Errorf("Role (%d) does not exist", id)
	}

	return us.store.SetStringIn(RolesNs, userID, role)
}

func (us *KVUserStore) ListUsers() ([]*User, error) {
	us.mtx.RLock()
	defer us.mtx.RUnlock()

	idToName, err := us.store.ListStringsIn(NamesNs)
	if err != nil {
		return nil, err
	}

	roles, err := us.store.ListStringsIn(RolesNs)
	if err != nil {
		return nil, err
	}

	users := []*User{}
	for id, name := range idToName {
		intID, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			return nil, err
		}

		users = append(users, &User{
			ID:   intID,
			Name: name,
			Role: roles[id],
		})
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
