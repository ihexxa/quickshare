package server

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/ihexxa/quickshare/src/client"
	q "github.com/ihexxa/quickshare/src/handlers"
	"github.com/ihexxa/quickshare/src/handlers/fileshdr"
)

func TestPermissions(t *testing.T) {
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
			"spaceLimit": 1000,
			"limiterCapacity": 1000,
			"limiterCyc": 1000,
			"predefinedUsers": [				
				{
					"name": "admin",
					"pwd": "1234",
					"role": "admin"
				},
				{
					"name": "user",
					"pwd": "1234",
					"role": "user"
				},
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
	setUpEnv(t, rootPath, adminName, adminPwd)
	defer os.RemoveAll(rootPath)

	srv := startTestServer(config)
	defer srv.Shutdown()
	fs := srv.depsFS()
	if !isServerReady(addr) {
		t.Fatal("fail to start server")
	}

	// adminUsersCl := client.NewSingleUserClient(addr)
	// resp, _, errs := adminUsersCl.Login(adminName, adminPwd)
	// if len(errs) > 0 {
	// 	t.Fatal(errs)
	// } else if resp.StatusCode != 200 {
	// 	t.Fatal(resp.StatusCode)
	// }
	// adminToken := client.GetCookie(resp.Cookies(), q.TokenCookie)
	// cl := client.NewFilesClient(addr, adminToken)

	var err error
	// TODO: remove all files under home folder before testing
	// or the count of files is incorrect

	// tests only check the status code for checking permission
	t.Run("Users API Permissions", func(t *testing.T) {
		testUsersAPIs := func(user string, pwd string, requireAuth bool, expectedCodes map[string]int) {
			cl := client.NewSingleUserClient(addr)
			var token *http.Cookie
			if requireAuth {
				resp, _, errs := cl.Login(user, pwd)
				if len(errs) > 0 {
					t.Fatal(errs)
				} else if resp.StatusCode != 200 {
					t.Fatal(resp.StatusCode)
				}
				token = client.GetCookie(resp.Cookies(), q.TokenCookie)
			}

			expectedCode := expectedCodes["SetPwd"]
			newPwd := "12345"
			resp, _, errs := cl.SetPwd(pwd, newPwd, token)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != expectedCode {
				t.Fatal(resp.StatusCode)
			}
			// set back the pwd
			resp, _, errs = cl.SetPwd(newPwd, pwd, token)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != expectedCode {
				t.Fatal(resp.StatusCode)
			}

			expectedCode = expectedCodes["Self"]
			resp, selfResp, errs := cl.Self(token)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != expectedCode {
				t.Fatal(resp.StatusCode)
			}

			expectedCode = expectedCodes["SetPreferences"]
			prefer := selfResp.Preferences
			resp, _, errs = cl.SetPreferences(prefer, token)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != expectedCode {
				t.Fatal(resp.StatusCode)
			}

			expectedCode = expectedCodes["IsAuthed"]
			resp, _, errs = cl.IsAuthed(token)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != expectedCode {
				t.Fatalf("%s %d", user, resp.StatusCode)
			}

			// test user operations
			expectedCode = expectedCodes["AddUser"]
			tmpUser, tmpPwd, tmpRole := "tmpUser", "1234", "admin"
			resp, addUserResp, errs := cl.AddUser(tmpUser, tmpPwd, tmpRole, token)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != expectedCode {
				t.Fatalf("%s %d", user, resp.StatusCode)
			}

			expectedCode = expectedCodes["ListUsers"]
			resp, _, errs = cl.ListUsers(token)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != expectedCode {
				t.Fatalf("%s %d", user, resp.StatusCode)
			}

			// TODO: the id here should be uint64
			uintID, err := strconv.ParseUint(addUserResp.ID, 64, 10)
			if err != nil {
				t.Fatal(err)
			}
			newRole := "user"
			expectedCode = expectedCodes["SetUser"]
			resp, _, errs = cl.SetUser(uintID, newRole, selfResp.Quota, token)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != expectedCode {
				t.Fatalf("%s %d", user, resp.StatusCode)
			}

			expectedCode = expectedCodes["DelUser"]
			resp, _, errs = cl.DelUser(addUserResp.ID, token)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != expectedCode {
				t.Fatalf("%s %d", user, resp.StatusCode)
			}

			// test role operations
			expectedCode = expectedCodes["AddRole"]
			tmpNewRole := "tmpNewRole"
			resp, _, errs = cl.AddRole(tmpNewRole, token)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != expectedCode {
				t.Fatalf("%s %d", user, resp.StatusCode)
			}

			expectedCode = expectedCodes["ListRoles"]
			resp, _, errs = cl.ListRoles(token)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != expectedCode {
				t.Fatalf("%s %d", user, resp.StatusCode)
			}

			expectedCode = expectedCodes["DelRole"]
			resp, _, errs = cl.DelRole(tmpNewRole, token)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != expectedCode {
				t.Fatalf("%s %d", user, resp.StatusCode)
			}

			if requireAuth {
				resp, _, errs := cl.Logout(token)
				if len(errs) > 0 {
					t.Fatal(errs)
				} else if resp.StatusCode != 200 {
					t.Fatal(resp.StatusCode)
				}
			}
		}

		testUsersAPIs("admin", "1234", true, map[string]int{
			"SetPwd":         200,
			"Self":           200,
			"SetPreferences": 200,
			"IsAuthed":       200,
			"AddUser":        200,
			"ListUsers":      200,
			"SetUser":        200,
			"DelUser":        200,
			"AddRole":        200,
			"ListRoles":      200,
			"DelRole":        200,
		})

		testUsersAPIs("user", "1234", true, map[string]int{
			"SetPwd":         200,
			"Self":           200,
			"SetPreferences": 200,
			"IsAuthed":       200,
			"AddUser":        401,
			"ListUsers":      401,
			"SetUser":        401,
			"DelUser":        401,
			"AddRole":        401,
			"ListRoles":      401,
			"DelRole":        401,
		})

		testUsersAPIs("visitor", "", false, map[string]int{
			"SetPwd":         401,
			"Self":           401,
			"SetPreferences": 401,
			"IsAuthed":       401,
			"AddUser":        401,
			"ListUsers":      401,
			"SetUser":        401,
			"DelUser":        401,
			"AddRole":        401,
			"ListRoles":      401,
			"DelRole":        401,
		})
	})
}
