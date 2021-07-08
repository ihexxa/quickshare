package multiusers

import (
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/ihexxa/quickshare/src/kvstore"
)

type User struct {
	ID   uint64
	Name string
	Pwd  string
	Role string
}

type IUserStore interface {
	AddUser(user *User) error
	GetUser(id uint64) (*User, error)
	SetName(id uint64, name string) error
	SetPwd(id uint64, pwd string) error
	SetRole(id uint64, role string) error
}

type KVUserStore struct {
	store kvstore.IKVStore
}

func NewKVUserStore(store kvstore.IKVStore, rootName, rootPwd string) (*KVUserStore, error) {
	_, ok := store.GetStringIn(InitNs, InitTimeParam)
	if !ok {
		var err error
		if err = store.AddNamespace(UsersNs); err != nil {
			return nil, err
		}
		if err = store.AddNamespace(PwdsNs); err != nil {
			return nil, err
		}
		if err = store.AddNamespace(RolesNs); err != nil {
			return nil, err
		}
		if err = store.AddNamespace(InitNs); err != nil {
			return nil, err
		}
		if err = store.SetStringIn(InitNs, InitTimeParam, fmt.Sprintf("%d", time.Now().Unix())); err != nil {
			return nil, err
		}
	}

	userStore := &KVUserStore{store: store}
	err := userStore.AddUser(&User{
		ID:   0,
		Name: rootName,
		Pwd:  rootPwd,
		Role: AdminRole,
	})

	return userStore, err
}

func (us *KVUserStore) AddUser(user *User) error {
	userID := fmt.Sprint(user.ID)
	_, ok := us.store.GetStringIn(UsersNs, userID)
	if ok {
		return fmt.Errorf("userID (%d) exists", user.ID)
	}
	if user.Name == "" || user.Pwd == "" {
		return errors.New("user name or password can not be empty")
	}

	var err error
	err = us.store.SetStringIn(UsersNs, userID, user.Name)
	if err != nil {
		return err
	}
	pwdHash, err := bcrypt.GenerateFromPassword([]byte(user.Pwd), 10)
	if err != nil {
		return err
	}
	err = us.store.SetStringIn(PwdsNs, userID, string(pwdHash))
	if err != nil {
		return err
	}
	err = us.store.SetStringIn(RolesNs, userID, user.Role)
	if err != nil {
		return err
	}

	return us.SetRole(user.ID, user.Role)
}

func (us *KVUserStore) GetUser(id uint64) (*User, error) {
	userID := fmt.Sprint(id)

	name, ok := us.store.GetStringIn(UsersNs, userID)
	if !ok {
		return nil, fmt.Errorf("name (%s) not found", userID)
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

func (us *KVUserStore) SetName(id uint64, name string) error {
	userID := fmt.Sprint(id)
	_, ok := us.store.GetStringIn(UsersNs, userID)
	if !ok {
		return fmt.Errorf("Name (%d) does not exist", id)
	}
	if name == "" {
		return fmt.Errorf("Name can not be empty")
	}
	// TODO: check if the new name already exists

	return us.store.SetStringIn(UsersNs, userID, name)
}

func (us *KVUserStore) SetPwd(id uint64, pwd string) error {
	userID := fmt.Sprint(id)
	_, ok := us.store.GetStringIn(PwdsNs, userID)
	if !ok {
		return fmt.Errorf("Pwd (%d) does not exist", id)
	}

	pwdHash, err := bcrypt.GenerateFromPassword([]byte(pwd), 10)
	if err != nil {
		return err
	}
	return us.store.SetStringIn(PwdsNs, userID, string(pwdHash))
}

func (us *KVUserStore) SetRole(id uint64, role string) error {
	userID := fmt.Sprint(id)
	_, ok := us.store.GetStringIn(RolesNs, userID)
	if !ok {
		return fmt.Errorf("Role (%d) does not exist", id)
	}

	return us.store.SetStringIn(RolesNs, userID, role)
}
