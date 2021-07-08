package multiusers

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/ihexxa/quickshare/src/kvstore/boltdbpvd"
	"golang.org/x/crypto/bcrypt"
)

func TestUserStores(t *testing.T) {
	rootName, rootPwd := "root", "rootPwd"

	testUserStore := func(t *testing.T, store IUserStore) {
		root, err := store.GetUser(0)
		if err != nil {
			t.Fatal(err)
		}
		if root.Name != rootName {
			t.Fatal("root user not found")
		}
		if err = bcrypt.CompareHashAndPassword([]byte(root.Pwd), []byte(rootPwd)); err != nil {
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
		if err = bcrypt.CompareHashAndPassword([]byte(user.Pwd), []byte(pwd1)); err != nil {
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
		if err = bcrypt.CompareHashAndPassword([]byte(user.Pwd), []byte(pwd2)); err != nil {
			t.Fatalf("passwords not match %s", err)
		}
		if user.Role != role2 {
			t.Fatalf("roles not matched %s %s", role2, user.Role)
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

		store, err := NewKVUserStore(kvstore, rootName, rootPwd)
		if err != nil {
			t.Fatal("fail to init kvstore", err)
		}

		testUserStore(t, store)
	})
}
