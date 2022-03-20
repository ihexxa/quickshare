package userstore

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ihexxa/quickshare/src/db"
	"github.com/ihexxa/quickshare/src/db/sitestore"
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
		if !db.ComparePreferences(root.Preferences, &DefaultPreferences) {
			t.Fatalf("incorrect preference %v %v", root.Preferences, DefaultPreferences)
		}

		visitor, err := store.GetUser(1)
		if err != nil {
			t.Fatal(err)
		}
		if visitor.Name != VisitorName {
			t.Fatal("visitor not found")
		}
		if visitor.Pwd != rootPwd {
			t.Fatalf("passwords not match %s", err)
		}
		if visitor.Role != VisitorRole {
			t.Fatalf("incorrect visitor role")
		}
		if visitor.Quota.SpaceLimit != 0 {
			t.Fatalf("incorrect visitor SpaceLimit")
		}
		if visitor.Quota.UploadSpeedLimit != visitorUploadSpeedLimit {
			t.Fatalf("incorrect visitor UploadSpeedLimit")
		}
		if visitor.Quota.DownloadSpeedLimit != visitorDownloadSpeedLimit {
			t.Fatalf("incorrect visitor DownloadSpeedLimit")
		}
		if !db.ComparePreferences(visitor.Preferences, &DefaultPreferences) {
			t.Fatalf("incorrect preference")
		}

		id, name1 := uint64(2), "test_user1"
		pwd1, pwd2 := "666", "888"
		role1, role2 := UserRole, AdminRole
		spaceLimit1, upLimit1, downLimit1 := int64(17), 5, 7
		spaceLimit2, upLimit2, downLimit2 := int64(19), 13, 17

		err = store.AddUser(&db.User{
			ID:   id,
			Name: name1,
			Pwd:  pwd1,
			Role: role1,
			Quota: &db.Quota{
				SpaceLimit:         spaceLimit1,
				UploadSpeedLimit:   upLimit1,
				DownloadSpeedLimit: downLimit1,
			},
		})
		if err != nil {
			t.Fatal("there should be no error")
		}

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
		if len(users) != 3 {
			t.Fatalf("users size should be 3 (%d)", len(users))
		}
		for _, user := range users {
			if user.ID == 0 {
				if user.Name != rootName || user.Role != AdminRole {
					t.Fatalf("incorrect root info %v", user)
				}
			}
			if user.ID == id {
				if user.Name != name1 || user.Role != role1 {
					t.Fatalf("incorrect user info %v", user)
				}
			}
		}

		err = store.SetPwd(id, pwd2)
		if err != nil {
			t.Fatal(err)
		}
		store.SetInfo(id, &db.User{
			ID:   id,
			Role: role2,
			Quota: &db.Quota{
				SpaceLimit:         spaceLimit2,
				UploadSpeedLimit:   upLimit2,
				DownloadSpeedLimit: downLimit2,
			},
		})

		usedIncr, usedDecr := int64(spaceLimit2), int64(7)
		err = store.SetUsed(id, true, usedIncr)
		if err != nil {
			t.Fatal(err)
		}
		err = store.SetUsed(id, false, usedDecr)
		if err != nil {
			t.Fatal(err)
		}
		err = store.SetUsed(id, true, int64(spaceLimit2)-(usedIncr-usedDecr)+1)
		if err == nil || !strings.Contains(err.Error(), "reached space limit") {
			t.Fatal("should reject big file")
		} else {
			err = nil
		}

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
		if user.UsedSpace != usedIncr-usedDecr {
			t.Fatalf("used space not matched %d %d", user.UsedSpace, usedIncr-usedDecr)
		}

		newPrefer := &db.Preferences{
			Bg: &sitestore.BgConfig{
				Url:      "/url",
				Repeat:   "repeat",
				Position: "pos",
				Align:    "fixed",
				BgColor:  "#333",
			},
			CSSURL:     "/cssurl",
			LanPackURL: "lanPackURL",
			Lan:        "zhCN",
			Theme:      "dark",
			Avatar:     "/avatar",
			Email:      "foo@gmail.com",
		}
		err = store.SetPreferences(id, newPrefer)
		if err != nil {
			t.Fatal(err)
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
		if !db.ComparePreferences(user.Preferences, newPrefer) {
			t.Fatalf("preferences not matched %v %v", user.Preferences, newPrefer)
		}

		err = store.DelUser(id)
		if err != nil {
			t.Fatal(err)
		}
		users, err = store.ListUsers()
		if err != nil {
			t.Fatal(err)
		}
		if len(users) != 2 {
			t.Fatalf("users size should be 2 (%d)", len(users))
		}
		for _, user := range users {
			if user.ID == 0 && user.Name != rootName && user.Role != AdminRole {
				t.Fatalf("incorrect root info %v", user)
			}
			if user.ID == VisitorID && user.Name != VisitorName && user.Role != VisitorRole {
				t.Fatalf("incorrect visitor info %v", user)
			}
		}

		nameToID, err := store.ListUserIDs()
		if err != nil {
			t.Fatal(err)
		}
		if len(nameToID) != len(users) {
			t.Fatalf("nameToID size (%d) should be same as (%d)", len(nameToID), len(users))
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

		dbPath := filepath.Join(rootPath, "quickshare.db")
		kvstore := boltdbpvd.New(dbPath, 1024)
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
