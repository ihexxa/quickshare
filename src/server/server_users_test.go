package server

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"testing"

	"github.com/ihexxa/quickshare/src/client"
	"github.com/ihexxa/quickshare/src/db"
	"github.com/ihexxa/quickshare/src/db/userstore"
	q "github.com/ihexxa/quickshare/src/handlers"
	su "github.com/ihexxa/quickshare/src/handlers/singleuserhdr"
)

func TestUsersHandlers(t *testing.T) {
	addr := "http://127.0.0.1:8686"
	rootPath := "tmpTestData"
	config := `{
		"users": {
			"enableAuth": true,
			"minUserNameLen": 2,
			"minPwdLen": 4,
			"captchaEnabled": false,
			"uploadSpeedLimit": 409600,
			"downloadSpeedLimit": 409600,
			"spaceLimit": 1024,
			"limiterCapacity": 1000,
			"limiterCyc": 1000,
			"predefinedUsers": [
				{
					"name": "demo",
					"pwd": "Quicksh@re",
					"role": "user"
				}
			]
		},
		"server": {
			"debug": true,
			"host": "127.0.0.1"
		},
		"fs": {
			"root": "tmpTestData"
		},
		"db": {
			"dbPath": "tmpTestData/quickshare"
		}
	}`
	adminName := "qs"
	adminPwd := "quicksh@re"
	adminNewPwd := "quicksh@re2"
	setUpEnv(t, rootPath, adminName, adminPwd)
	defer os.RemoveAll(rootPath)

	srv := startTestServer(config)
	defer srv.Shutdown()
	fs := srv.depsFS()

	usersCl := client.NewSingleUserClient(addr)

	if !isServerReady(addr) {
		t.Fatal("fail to start server")
	}

	var err error

	t.Run("test inited users", func(t *testing.T) {
		usersCl := client.NewSingleUserClient(addr)
		resp, _, errs := usersCl.Login(adminName, adminPwd)
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if resp.StatusCode != 200 {
			t.Fatal(resp.StatusCode)
		}
		token := client.GetCookie(resp.Cookies(), su.TokenCookie)

		resp, lsResp, errs := usersCl.ListUsers(token)
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if resp.StatusCode != 200 {
			t.Fatal(resp.StatusCode)
		}

		if len(lsResp.Users) != 3 {
			t.Fatal(fmt.Errorf("incorrect users size (%d)", len(lsResp.Users)))
		}

		for _, user := range lsResp.Users {
			if user.Name == adminName {
				if user.ID != 0 ||
					user.Role != userstore.AdminRole ||
					user.UsedSpace != 0 ||
					user.Quota.SpaceLimit != 1024*1024*1024 || // TODO: export these
					user.Quota.UploadSpeedLimit != 50*1024*1024 ||
					user.Quota.DownloadSpeedLimit != 50*1024*1024 ||
					!reflect.DeepEqual(user.Preferences, &userstore.DefaultPreferences) {
					t.Fatal(fmt.Errorf("incorrect user info (%v)", user))
				}
			}
			if user.Name == "visitor" {
				if user.Role != userstore.VisitorRole ||
					user.Quota.SpaceLimit != 0 || // TODO: export these
					user.Quota.UploadSpeedLimit != 10*1024*1024 ||
					user.Quota.DownloadSpeedLimit != 10*1024*1024 ||
					!reflect.DeepEqual(user.Preferences, &userstore.DefaultPreferences) {
					t.Fatal(fmt.Errorf("incorrect user info (%v)", user))
				}
			}
			if user.Name == "demo" {
				if user.Role != userstore.UserRole ||
					user.Quota.SpaceLimit != 1024 ||
					user.Quota.UploadSpeedLimit != 409600 ||
					user.Quota.DownloadSpeedLimit != 409600 ||
					!reflect.DeepEqual(user.Preferences, &userstore.DefaultPreferences) {
					t.Fatal(fmt.Errorf("incorrect user info (%v)", user))
				}
			}
		}
	})

	t.Run("test users APIs: Login-Self-SetPwd-Logout-Login", func(t *testing.T) {
		users := []*db.User{
			{
				ID:        0,
				Name:      adminName,
				Pwd:       adminPwd,
				Role:      userstore.AdminRole,
				UsedSpace: 0,
				Quota: &db.Quota{
					UploadSpeedLimit:   50 * 1024 * 1024,
					DownloadSpeedLimit: 50 * 1024 * 1024,
					SpaceLimit:         1024 * 1024 * 1024,
				},
			},
			{
				ID:        0,
				Name:      "demo",
				Pwd:       "Quicksh@re",
				Role:      userstore.UserRole,
				UsedSpace: 0,
				Quota: &db.Quota{
					UploadSpeedLimit:   409600,
					DownloadSpeedLimit: 409600,
					SpaceLimit:         1024,
				},
			},
		}

		for _, user := range users {
			usersCl := client.NewSingleUserClient(addr)
			resp, _, errs := usersCl.Login(user.Name, user.Pwd)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != 200 {
				t.Fatal(resp.StatusCode)
			}
			token := client.GetCookie(resp.Cookies(), su.TokenCookie)

			resp, selfResp, errs := usersCl.Self(token)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != 200 {
				t.Fatal(resp.StatusCode)
			} else if selfResp.Name != user.Name ||
				selfResp.Role != user.Role ||
				selfResp.UsedSpace != 0 ||
				selfResp.Quota.UploadSpeedLimit != user.Quota.UploadSpeedLimit ||
				selfResp.Quota.DownloadSpeedLimit != user.Quota.DownloadSpeedLimit ||
				selfResp.Quota.SpaceLimit != user.Quota.SpaceLimit {
				// TODO: expose default values from userstore
				t.Fatalf("user infos don't match %v", selfResp)
			}
			if selfResp.Role == userstore.AdminRole {
				if selfResp.ID != "0" {
					t.Fatalf("user id don't match %v", selfResp)
				}
			}

			resp, _, errs = usersCl.SetPwd(user.Pwd, adminNewPwd, token)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != 200 {
				t.Fatal(resp.StatusCode)
			}

			resp, _, errs = usersCl.Logout(token)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != 200 {
				t.Fatal(resp.StatusCode)
			}

			resp, _, errs = usersCl.Login(user.Name, adminNewPwd)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != 200 {
				t.Fatal(resp.StatusCode)
			}
		}
	})

	t.Run("test users APIs: Login-AddUser-Logout-Login-Logout", func(t *testing.T) {
		resp, _, errs := usersCl.Login(adminName, adminNewPwd)
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if resp.StatusCode != 200 {
			t.Fatal(resp.StatusCode)
		}
		token := client.GetCookie(resp.Cookies(), su.TokenCookie)

		userName, userPwd := "user_login", "1234"
		resp, auResp, errs := usersCl.AddUser(userName, userPwd, userstore.UserRole, token)
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if resp.StatusCode != 200 {
			t.Fatal(resp.StatusCode)
		}
		// TODO: check id
		fmt.Printf("new user id: %v\n", auResp)

		// check uploading file
		userFsRootFolder := q.FsRootPath(userName, "/")
		_, err = fs.Stat(userFsRootFolder)
		if err != nil {
			t.Fatal(err)
		}
		userUploadFolder := q.UploadFolder(userName)
		_, err = fs.Stat(userUploadFolder)
		if err != nil {
			t.Fatal(err)
		}

		resp, _, errs = usersCl.Logout(token)
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if resp.StatusCode != 200 {
			t.Fatal(resp.StatusCode)
		}

		resp, _, errs = usersCl.Login(userName, userPwd)
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if resp.StatusCode != 200 {
			t.Fatal(resp.StatusCode)
		}

		resp, _, errs = usersCl.DelUser(auResp.ID, token)
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if resp.StatusCode != 200 {
			t.Fatal(resp.StatusCode)
		}

		resp, _, errs = usersCl.Logout(token)
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if resp.StatusCode != 200 {
			t.Fatal(resp.StatusCode)
		}
	})

	t.Run("test users APIs: Login-AddUser-ListUsers-SetUser-ListUsers-DelUser-ListUsers", func(t *testing.T) {
		resp, _, errs := usersCl.Login(adminName, adminNewPwd)
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if resp.StatusCode != 200 {
			t.Fatal(resp.StatusCode)
		}

		token := client.GetCookie(resp.Cookies(), su.TokenCookie)

		userName, userPwd, userRole := "new_user", "1234", userstore.UserRole
		resp, auResp, errs := usersCl.AddUser(userName, userPwd, userRole, token)
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if resp.StatusCode != 200 {
			t.Fatal(resp.StatusCode)
		}
		// TODO: check id
		fmt.Printf("new user id: %v\n", auResp)
		newUserID, err := strconv.ParseUint(auResp.ID, 10, 64)
		if err != nil {
			t.Fatal(err)
		}

		resp, lsResp, errs := usersCl.ListUsers(token)
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if resp.StatusCode != 200 {
			t.Fatal(resp.StatusCode)
		}

		if len(lsResp.Users) != 4 {
			t.Fatal(fmt.Errorf("incorrect users size (%d)", len(lsResp.Users)))
		}
		for _, user := range lsResp.Users {
			if user.ID == 0 {
				if user.Name != adminName ||
					user.Role != userstore.AdminRole {
					t.Fatal(fmt.Errorf("incorrect root info (%v)", user))
				}
			}
			if user.ID == newUserID {
				if user.Name != userName ||
					user.Role != userRole {
					t.Fatal(fmt.Errorf("incorrect user info (%v)", user))
				}
			}
		}

		newRole, newQuota := userstore.AdminRole, &db.Quota{
			SpaceLimit:         3,
			UploadSpeedLimit:   3,
			DownloadSpeedLimit: 3,
		}
		resp, _, errs = usersCl.SetUser(newUserID, newRole, newQuota, token)
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if resp.StatusCode != 200 {
			t.Fatal(resp.StatusCode)
		}

		resp, lsResp, errs = usersCl.ListUsers(token)
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if resp.StatusCode != 200 {
			t.Fatal(resp.StatusCode)
		}
		for _, user := range lsResp.Users {
			if user.ID == newUserID {
				if user.Role != newRole {
					t.Fatal(fmt.Errorf("incorrect role (%v)", user.Role))
				}
				if user.Quota.SpaceLimit != newQuota.SpaceLimit {
					t.Fatal(fmt.Errorf("incorrect quota (%v)", newQuota.SpaceLimit))
				}
				if user.Quota.UploadSpeedLimit != newQuota.UploadSpeedLimit {
					t.Fatal(fmt.Errorf("incorrect quota (%v)", newQuota.UploadSpeedLimit))
				}
				if user.Quota.DownloadSpeedLimit != newQuota.DownloadSpeedLimit {
					t.Fatal(fmt.Errorf("incorrect quota (%v)", newQuota.DownloadSpeedLimit))
				}
			}
		}

		resp, _, errs = usersCl.DelUser(auResp.ID, token)
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if resp.StatusCode != 200 {
			t.Fatal(resp.StatusCode)
		}

		resp, lsResp, errs = usersCl.ListUsers(token)
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if resp.StatusCode != 200 {
			t.Fatal(resp.StatusCode)
		}
		if len(lsResp.Users) != 3 {
			t.Fatal(fmt.Errorf("incorrect users size (%d)", len(lsResp.Users)))
		}

		resp, _, errs = usersCl.Logout(token)
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if resp.StatusCode != 200 {
			t.Fatal(resp.StatusCode)
		}
	})

	t.Run("test roles APIs: Login-AddRole-ListRoles-DelRole-ListRoles-Logout", func(t *testing.T) {
		resp, _, errs := usersCl.Login(adminName, adminNewPwd)
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if resp.StatusCode != 200 {
			t.Fatal(resp.StatusCode)
		}

		token := client.GetCookie(resp.Cookies(), su.TokenCookie)
		roles := []string{"role1", "role2"}

		for _, role := range roles {
			resp, _, errs := usersCl.AddRole(role, token)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != 200 {
				t.Fatal(resp.StatusCode)
			}
		}

		resp, lsResp, errs := usersCl.ListRoles(token)
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if resp.StatusCode != 200 {
			t.Fatal(resp.StatusCode)
		}
		for _, role := range append(roles, []string{
			userstore.AdminRole,
			userstore.UserRole,
			userstore.VisitorRole,
		}...) {
			if !lsResp.Roles[role] {
				t.Fatalf("role(%s) not found", role)
			}
		}

		for _, role := range roles {
			resp, _, errs := usersCl.DelRole(role, token)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != 200 {
				t.Fatal(resp.StatusCode)
			}
		}

		resp, lsResp, errs = usersCl.ListRoles(token)
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if resp.StatusCode != 200 {
			t.Fatal(resp.StatusCode)
		}
		for _, role := range roles {
			if lsResp.Roles[role] {
				t.Fatalf("role(%s) should not exist", role)
			}
		}

		resp, _, errs = usersCl.Logout(token)
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if resp.StatusCode != 200 {
			t.Fatal(resp.StatusCode)
		}
	})

	t.Run("Login,SetPreferences, Self, Logout", func(t *testing.T) {
		usersCl := client.NewSingleUserClient(addr)
		resp, _, errs := usersCl.Login(adminName, adminNewPwd)
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if resp.StatusCode != 200 {
			t.Fatal(resp.StatusCode)
		}

		token := client.GetCookie(resp.Cookies(), su.TokenCookie)

		prefers := []*db.Preferences{
			&db.Preferences{
				Bg: &db.BgConfig{
					Url:      "/bgurl",
					Repeat:   "no-repeat",
					Position: "center",
					Align:    "fixed",
					BgColor:  "#ccc",
				},
				CSSURL:     "/cssurl",
				LanPackURL: "/lanpack",
				Avatar:     "a1",
				Email:      "email1",
			},
			&db.Preferences{
				Bg: &db.BgConfig{
					Url:      "/bgurl2",
					Repeat:   "no-repeat2",
					Position: "center2",
					Align:    "fixed2",
					BgColor:  "#333",
				},
				CSSURL:     "/cssurl2",
				LanPackURL: "/lanpack2",
				Avatar:     "a2",
				Email:      "email2",
			},
		}
		for _, prefer := range prefers {
			resp, _, errs := usersCl.SetPreferences(prefer, token)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != 200 {
				t.Fatal(resp.StatusCode)
			}

			resp, selfResp, errs := usersCl.Self(token)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != 200 {
				t.Fatal(resp.StatusCode)
			} else if !reflect.DeepEqual(selfResp.Preferences, prefer) {
				t.Fatal("preference not equal")
			}
		}
	})
}
