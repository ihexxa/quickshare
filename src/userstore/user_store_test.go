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
			t.Fatalf("incorrect root role")
		}
		if root.Quota.SpaceLimit != defaultSpaceLimit {
			t.Fatalf("incorrect root SpaceLimit")
		}
		if root.Quota.UploadSpeedLimit != defaultUploadSpeedLimit {
			t.Fatalf("incorrect root UploadSpeedLimit")
		}
		if root.Quota.DownloadSpeedLimit != defaultDownloadSpeedLimit {
			t.Fatalf("incorrect root DownloadSpeedLimit")
		}

		id, name1 := uint64(1), "test_user1"
		pwd1, pwd2 := "666", "888"
		role1, role2 := UserRole, AdminRole
		spaceLimit1, upLimit1, downLimit1 := 3, 5, 7
		spaceLimit2, upLimit2, downLimit2 := 11, 13, 17

		err = store.AddUser(&User{
			ID:   id,
			Name: name1,
			Pwd:  pwd1,
			Role: role1,
			Quota: &Quota{
				SpaceLimit:         spaceLimit1,
				UploadSpeedLimit:   upLimit1,
				DownloadSpeedLimit: downLimit1,
			},
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
		if user.Quota.SpaceLimit != spaceLimit1 {
			t.Fatalf("space limit not matched %d %d", spaceLimit1, user.Quota.SpaceLimit)
		}
		if user.Quota.UploadSpeedLimit != upLimit1 {
			t.Fatalf("up limit not matched %d %d", upLimit1, user.Quota.UploadSpeedLimit)
		}
		if user.Quota.DownloadSpeedLimit != downLimit1 {
			t.Fatalf("down limit not matched %d %d", downLimit1, user.Quota.DownloadSpeedLimit)
		}

		users, err := store.ListUsers()
		if err != nil {
			t.Fatal(err)
		}
		if len(users) != 2 {
			t.Fatalf("users size should be 2 (%d)", len(users))
		}
		for _, user := range users {
			if user.ID == 0 {
				if users[0].Name != rootName || users[0].Role != AdminRole {
					t.Fatalf("incorrect root info %v", users[0])
				}
			}
			if user.ID == 1 {
				if users[1].Name != name1 || users[1].Role != role1 {
					t.Fatalf("incorrect user info %v", users[1])
				}
			}
		}

		err = store.SetPwd(id, pwd2)
		if err != nil {
			t.Fatal(err)
		}
		store.SetInfo(id, &User{
			ID:   id,
			Role: role2,
			Quota: &Quota{
				SpaceLimit:         spaceLimit2,
				UploadSpeedLimit:   upLimit2,
				DownloadSpeedLimit: downLimit2,
			},
		})

		user, err = store.GetUser(id)
		if err != nil {
			t.Fatal(err)
		}
		if user.Pwd != pwd2 {
			t.Fatalf("passwords not match %s", err)
		}
		if user.Role != role2 {
			t.Fatalf("roles not matched %s %s", role2, user.Role)
		}
		if user.Quota.SpaceLimit != spaceLimit2 {
			t.Fatalf("space limit not matched %d %d", spaceLimit2, user.Quota.SpaceLimit)
		}
		if user.Quota.UploadSpeedLimit != upLimit2 {
			t.Fatalf("up limit not matched %d %d", upLimit2, user.Quota.UploadSpeedLimit)
		}
		if user.Quota.DownloadSpeedLimit != downLimit2 {
			t.Fatalf("down limit not matched %d %d", downLimit2, user.Quota.DownloadSpeedLimit)
		}

		user, err = store.GetUserByName(name1)
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
		if user.Quota.SpaceLimit != spaceLimit2 {
			t.Fatalf("space limit not matched %d %d", spaceLimit2, user.Quota.SpaceLimit)
		}
		if user.Quota.UploadSpeedLimit != upLimit2 {
			t.Fatalf("up limit not matched %d %d", upLimit2, user.Quota.UploadSpeedLimit)
		}
		if user.Quota.DownloadSpeedLimit != downLimit2 {
			t.Fatalf("down limit not matched %d %d", downLimit2, user.Quota.DownloadSpeedLimit)
		}

		err = store.DelUser(id)
		if err != nil {
			t.Fatal(err)
		}
		users, err = store.ListUsers()
		if err != nil {
			t.Fatal(err)
		}
		if len(users) != 1 {
			t.Fatalf("users size should be 2 (%d)", len(users))
		}
		if users[0].ID != 0 || users[0].Name != rootName || users[0].Role != AdminRole {
			t.Fatalf("incorrect root info %v", users[0])
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
