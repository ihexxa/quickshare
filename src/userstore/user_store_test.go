package userstore

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/ihexxa/quickshare/src/kvstore/boltdbpvd"
)

func TestUserStores(t *testing.T) {
	rootName, rootPwd := "root", "rootPwd"

	testUserMethods := func(t *testing.T, store IUserStore) {
		root, err := store.GetUser(0)
		if err != nil {
			t.Fatal(err)
		}
		if root.Name != rootName {
			t.Fatal("root user not found")
		}
		if root.Pwd != rootPwd {
			t.Fatalf("passwords not match %s", err)
		}
		if root.Role != AdminRole {
			t.Fatalf("incorrect root fole")
		}

		id, name1, name2 := uint64(1), "test_user1", "test_user2"
		pwd1, pwd2 := "666", "888"
		role1, role2 := UserRole, AdminRole

		err = store.AddUser(&User{
			ID:   id,
			Name: name1,
			Pwd:  pwd1,
			Role: role1,
		})

		user, err := store.GetUser(id)
		if err != nil {
			t.Fatal(err)
		}
		if user.Name != name1 {
			t.Fatalf("names not matched %s %s", name1, user.Name)
		}
		if user.Pwd != pwd1 {
			t.Fatalf("passwords not match %s", err)
		}
		if user.Role != role1 {
			t.Fatalf("roles not matched %s %s", role1, user.Role)
		}

		err = store.SetName(id, name2)
		if err != nil {
			t.Fatal(err)
		}
		err = store.SetPwd(id, pwd2)
		if err != nil {
			t.Fatal(err)
		}
		err = store.SetRole(id, role2)
		if err != nil {
			t.Fatal(err)
		}

		user, err = store.GetUser(id)
		if err != nil {
			t.Fatal(err)
		}
		if user.Name != name2 {
			t.Fatalf("names not matched %s %s", name2, user.Name)
		}
		if user.Pwd != pwd2 {
			t.Fatalf("passwords not match %s", err)
		}
		if user.Role != role2 {
			t.Fatalf("roles not matched %s %s", role2, user.Role)
		}

		user, err = store.GetUserByName(name2)
		if err != nil {
			t.Fatal(err)
		}
		if user.ID != id {
			t.Fatalf("ids not matched %d %d", id, user.ID)
		}
		if user.Pwd != pwd2 {
			t.Fatalf("passwords not match %s", err)
		}
		if user.Role != role2 {
			t.Fatalf("roles not matched %s %s", role2, user.Role)
		}
	}

	testRoleMethods := func(t *testing.T, store IUserStore) {
		roles := []string{"role1", "role2"}
		var err error
		for _, role := range roles {
			err = store.AddRole(role)
			if err != nil {
				t.Fatal(err)
			}
		}

		roleMap, err := store.ListRoles()
		if err != nil {
			t.Fatal(err)
		}

		for _, role := range append(roles, []string{
			AdminRole, UserRole, VisitorRole,
		}...) {
			if !roleMap[role] {
				t.Fatalf("role(%s) not found", role)
			}
		}

		for _, role := range roles {
			err = store.DelRole(role)
			if err != nil {
				t.Fatal(err)
			}
		}

		roleMap, err = store.ListRoles()
		if err != nil {
			t.Fatal(err)
		}
		for _, role := range roles {
			if roleMap[role] {
				t.Fatalf("role(%s) should not exist", role)
			}
		}
	}

	t.Run("testing KVUserStore", func(t *testing.T) {
		rootPath, err := ioutil.TempDir("./", "quickshare_userstore_test_")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(rootPath)

		kvstore := boltdbpvd.New(rootPath, 1024)
		defer kvstore.Close()

		store, err := NewKVUserStore(kvstore)
		if err != nil {
			t.Fatal("fail to new kvstore", err)
		}
		if err = store.Init(rootName, rootPwd); err != nil {
			t.Fatal("fail to init kvstore", err)
		}

		testUserMethods(t, store)
		testRoleMethods(t, store)
	})
}
