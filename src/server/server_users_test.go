package server

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/ihexxa/quickshare/src/client"
	q "github.com/ihexxa/quickshare/src/handlers"
	su "github.com/ihexxa/quickshare/src/handlers/singleuserhdr"
	"github.com/ihexxa/quickshare/src/userstore"
)

func TestUsersHandlers(t *testing.T) {
	addr := "http://127.0.0.1:8686"
	root := "testData"
	config := `{
		"users": {
			"enableAuth": true,
			"minUserNameLen": 2,
			"minPwdLen": 4,
			"captchaEnabled": false,
			"captchaFont": "../../public/static/font/Roboto-Light.ttf"
		},
		"server": {
			"debug": true
		},
		"fs": {
			"root": "testData"
		}
	}`
	adminName := "qs"
	adminPwd := "quicksh@re"
	adminNewPwd := "quicksh@re2"
	os.Setenv("DEFAULTADMIN", adminName)
	os.Setenv("DEFAULTADMINPWD", adminPwd)

	os.RemoveAll(root)
	err := os.MkdirAll(root, 0700)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)

	srv := startTestServer(config)
	defer srv.Shutdown()
	fs := srv.depsFS()

	usersCl := client.NewSingleUserClient(addr)

	if !waitForReady(addr) {
		t.Fatal("fail to start server")
	}

	t.Run("test users APIs: Login-Self-SetPwd-Logout-Login", func(t *testing.T) {
		resp, _, errs := usersCl.Login(adminName, adminPwd)
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
		} else if selfResp.ID != "0" ||
			selfResp.Name != adminName ||
			selfResp.Role != userstore.AdminRole {
			t.Fatalf("user infos don't match %v", selfResp)
		}

		resp, _, errs = usersCl.SetPwd(adminPwd, adminNewPwd, token)
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

		resp, _, errs = usersCl.Login(adminName, adminNewPwd)
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if resp.StatusCode != 200 {
			t.Fatal(resp.StatusCode)
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
		userFsRootFolder := q.FsRootPath(auResp.ID, "/")
		_, err = fs.Stat(userFsRootFolder)
		if err != nil {
			t.Fatal(err)
		}
		userUploadFolder := q.UploadFolder(auResp.ID)
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

	t.Run("test users APIs: Login-AddUser-ListUsers-DelUser-ListUsers", func(t *testing.T) {
		resp, _, errs := usersCl.Login(adminName, adminNewPwd)
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if resp.StatusCode != 200 {
			t.Fatal(resp.StatusCode)
		}

		token := client.GetCookie(resp.Cookies(), su.TokenCookie)

		userName, userPwd, userRole := "user_admin", "1234", userstore.UserRole
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

		if len(lsResp.Users) != 2 {
			t.Fatal(fmt.Errorf("incorrect users size (%d)", len(lsResp.Users)))
		} else if lsResp.Users[0].ID != 0 ||
			lsResp.Users[0].Name != adminName ||
			lsResp.Users[0].Role != userstore.AdminRole {
			t.Fatal(fmt.Errorf("incorrect root info (%v)", lsResp.Users[0]))
		} else if lsResp.Users[1].ID != newUserID ||
			lsResp.Users[1].Name != userName ||
			lsResp.Users[1].Role != userRole {
			t.Fatal(fmt.Errorf("incorrect user info (%v)", lsResp.Users[1]))
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
		if len(lsResp.Users) != 1 {
			t.Fatal(fmt.Errorf("incorrect users size (%d)", len(lsResp.Users)))
		} else if lsResp.Users[0].ID != 0 ||
			lsResp.Users[0].Name != adminName ||
			lsResp.Users[0].Role != userstore.AdminRole {
			t.Fatal(fmt.Errorf("incorrect root info (%v)", lsResp.Users[0]))
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
}
