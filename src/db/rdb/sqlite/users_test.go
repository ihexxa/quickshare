package sqlite

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ihexxa/quickshare/src/db"
)

func TestUserStores(t *testing.T) {
	rootName, rootPwd := "root", "rootPwd"

	testUserMethods := func(t *testing.T, store db.IUserStore) {
		ctx := context.TODO()
		root, err := store.GetUser(ctx, 0)
		if err != nil {
			t.Fatal(err)
		}
		if root.Name != rootName {
			t.Fatal("root user not found")
		}
		if root.Pwd != rootPwd {
			t.Fatalf("passwords not match (%s) (%s)", root.Pwd, rootPwd)
		}
		if root.Role != db.AdminRole {
			t.Fatalf("incorrect root role")
		}
		if root.Quota.SpaceLimit != db.DefaultSpaceLimit {
			t.Fatalf("incorrect root SpaceLimit")
		}
		if root.Quota.UploadSpeedLimit != db.DefaultUploadSpeedLimit {
			t.Fatalf("incorrect root UploadSpeedLimit")
		}
		if root.Quota.DownloadSpeedLimit != db.DefaultDownloadSpeedLimit {
			t.Fatalf("incorrect root DownloadSpeedLimit")
		}
		if !db.ComparePreferences(root.Preferences, &db.DefaultPreferences) {
			t.Fatalf("incorrect preference %v %v", root.Preferences, db.DefaultPreferences)
		}

		visitor, err := store.GetUser(ctx, 1)
		if err != nil {
			t.Fatal(err)
		}
		if visitor.Name != VisitorName {
			t.Fatal("visitor not found")
		}
		if visitor.Pwd != rootPwd {
			t.Fatalf("passwords not match %s", err)
		}
		if visitor.Role != db.VisitorRole {
			t.Fatalf("incorrect visitor role")
		}
		if visitor.Quota.SpaceLimit != 0 {
			t.Fatalf("incorrect visitor SpaceLimit")
		}
		if visitor.Quota.UploadSpeedLimit != db.VisitorUploadSpeedLimit {
			t.Fatalf("incorrect visitor UploadSpeedLimit")
		}
		if visitor.Quota.DownloadSpeedLimit != db.VisitorDownloadSpeedLimit {
			t.Fatalf("incorrect visitor DownloadSpeedLimit")
		}
		if !db.ComparePreferences(visitor.Preferences, &db.DefaultPreferences) {
			t.Fatalf("incorrect preference")
		}

		id, name1 := uint64(2), "test_user1"
		pwd1, pwd2 := "666", "888"
		role1, role2 := db.UserRole, db.AdminRole
		spaceLimit1, upLimit1, downLimit1 := int64(17), 5, 7
		spaceLimit2, upLimit2, downLimit2 := int64(19), 13, 17

		err = store.AddUser(ctx, &db.User{
			ID:   id,
			Name: name1,
			Pwd:  pwd1,
			Role: role1,
			Quota: &db.Quota{
				SpaceLimit:         spaceLimit1,
				UploadSpeedLimit:   upLimit1,
				DownloadSpeedLimit: downLimit1,
			},
			Preferences: &db.DefaultPreferences,
		})
		if err != nil {
			t.Fatal("there should be no error")
		}

		user, err := store.GetUser(ctx, id)
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

		users, err := store.ListUsers(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if len(users) != 3 {
			t.Fatalf("users size should be 3 (%d)", len(users))
		}
		for _, user := range users {
			if user.ID == 0 {
				if user.Name != rootName || user.Role != db.AdminRole {
					t.Fatalf("incorrect root info %v", user)
				}
			}
			if user.ID == id {
				if user.Name != name1 || user.Role != role1 {
					t.Fatalf("incorrect user info %v", user)
				}
			}
			if user.Pwd != "" {
				t.Fatalf("password must be empty")
			}
		}

		err = store.SetPwd(ctx, id, pwd2)
		if err != nil {
			t.Fatal(err)
		}
		store.SetInfo(ctx, id, &db.User{
			ID:   id,
			Role: role2,
			Quota: &db.Quota{
				SpaceLimit:         spaceLimit2,
				UploadSpeedLimit:   upLimit2,
				DownloadSpeedLimit: downLimit2,
			},
		})

		usedIncr, usedDecr := int64(spaceLimit2), int64(7)
		err = store.SetUsed(ctx, id, true, usedIncr)
		if err != nil {
			t.Fatal(err)
		}
		err = store.SetUsed(ctx, id, false, usedDecr)
		if err != nil {
			t.Fatal(err)
		}
		err = store.SetUsed(ctx, id, true, int64(spaceLimit2)-(usedIncr-usedDecr)+1)
		if err == nil || !strings.Contains(err.Error(), "reached space limit") {
			t.Fatal("should reject big file")
		} else {
			err = nil
		}

		user, err = store.GetUser(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
		if user.Pwd != pwd2 {
			t.Fatalf("passwords not match %s %s", user.Pwd, pwd2)
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

		time.Sleep(5 * time.Second)
		newPrefer := &db.Preferences{
			Bg: &db.BgConfig{
				Url:      "/url",
				Repeat:   "repeat",
				Position: "center",
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
		err = store.SetPreferences(ctx, id, newPrefer)
		if err != nil {
			t.Fatal(err)
		}

		user, err = store.GetUserByName(ctx, name1)
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

		err = store.DelUser(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
		users, err = store.ListUsers(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if len(users) != 2 {
			t.Fatalf("users size should be 2 (%d)", len(users))
		}
		for _, user := range users {
			if user.ID == 0 && user.Name != rootName && user.Role != db.AdminRole {
				t.Fatalf("incorrect root info %v", user)
			}
			if user.ID == VisitorID && user.Name != VisitorName && user.Role != db.VisitorRole {
				t.Fatalf("incorrect visitor info %v", user)
			}
		}

		nameToID, err := store.ListUserIDs(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if len(nameToID) != len(users) {
			t.Fatalf("nameToID size (%d) should be same as (%d)", len(nameToID), len(users))
		}
	}

	t.Run("testing UserStore sqlite", func(t *testing.T) {
		rootPath, err := ioutil.TempDir("./", "quickshare_userstore_test_")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(rootPath)

		dbPath := filepath.Join(rootPath, "quickshare.sqlite")
		sqliteDB, err := NewSQLite(dbPath)
		if err != nil {
			t.Fatal(err)
		}
		defer sqliteDB.Close()

		store, err := NewSQLiteUsers(sqliteDB)
		if err != nil {
			t.Fatal("fail to new user store", err)
		}
		if err = store.Init(context.TODO(), rootName, rootPwd); err != nil {
			t.Fatal("fail to init user store", err)
		}

		testUserMethods(t, store)
	})
}
