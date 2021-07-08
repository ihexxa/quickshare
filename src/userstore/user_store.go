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
	InitTimeKey = "initTime"
)

type User struct {
	ID   uint64
	Name string
	Pwd  string
	Role string
}

type IUserStore interface {
	Init(rootName, rootPwd string) error
	IsInited() bool
	AddUser(user *User) error
	GetUser(id uint64) (*User, error)
	GetUserByName(name string) (*User, error)
	SetName(id uint64, name string) error
	SetPwd(id uint64, pwd string) error
	SetRole(id uint64, role string) error
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
	err := us.AddUser(&User{
		ID:   0,
		Name: rootName,
		Pwd:  rootPwd,
		Role: AdminRole,
	})
	if err != nil {
		return err
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
