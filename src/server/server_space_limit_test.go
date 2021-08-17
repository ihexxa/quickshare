package server

import (
	"fmt"
	"os"
	"testing"

	"github.com/ihexxa/quickshare/src/client"
	q "github.com/ihexxa/quickshare/src/handlers"
	"github.com/ihexxa/quickshare/src/userstore"
)

func TestSpaceLimit(t *testing.T) {
	addr := "http://127.0.0.1:8686"
	root := "testData"
	config := `{
		"users": {
			"enableAuth": true,
			"minUserNameLen": 2,
			"minPwdLen": 4,
			"captchaEnabled": false,
			"uploadSpeedLimit": 409600,
			"downloadSpeedLimit": 409600,
			"spaceLimit": 100,
			"limiterCapacity": 1000,
			"limiterCyc": 1000
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
	// fs := srv.depsFS()
	if !waitForReady(addr) {
		t.Fatal("fail to start server")
	}

	usersCl := client.NewSingleUserClient(addr)
	resp, _, errs := usersCl.Login(adminName, adminPwd)
	if len(errs) > 0 {
		t.Fatal(errs)
	} else if resp.StatusCode != 200 {
		t.Fatal(resp.StatusCode)
	}
	token := client.GetCookie(resp.Cookies(), q.TokenCookie)

	userCount := 1
	userPwd := "1234"
	users := map[string]string{}
	getUserName := func(id int) string {
		return fmt.Sprintf("space_limit_user_%d", id)
	}

	for i := 0; i < userCount; i++ {
		userName := getUserName(i)

		resp, adResp, errs := usersCl.AddUser(userName, userPwd, userstore.UserRole, token)
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if resp.StatusCode != 200 {
			t.Fatal("failed to add user")
		}

		users[userName] = adResp.ID
	}

	resp, _, errs = usersCl.Logout(token)
	if len(errs) > 0 {
		t.Fatal(errs)
	} else if resp.StatusCode != 200 {
		t.Fatal(resp.StatusCode)
	}

	t.Run("test space limitiong: Upload", func(t *testing.T) {
		usersCl := client.NewSingleUserClient(addr)
		resp, _, errs := usersCl.Login(getUserName(0), userPwd)
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if resp.StatusCode != 200 {
			t.Fatal(resp.StatusCode)
		}
		token := client.GetCookie(resp.Cookies(), q.TokenCookie)

		fileContent := ""
		for i := 0; i < 10; i++ {
			fileContent += "0"
		}

		for i := 0; i < 10; i++ {
			ok := assertUploadOK(t, fmt.Sprintf("%s/spacelimit/f_%d", getUserName(0), 0), fileContent, addr, token)
			if !ok {
				t.Fatalf("space limit failed at %d", 0)
			}

			resp, selfResp, errs := usersCl.Self(token)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != 200 {
				t.Fatal("failed to add user")
			} else if selfResp.UsedSpace != int64((i+1)*10) {
				t.Fatal("incorrect used space")
			}
		}

		cl := client.NewFilesClient(addr, token)
		filePath := fmt.Sprintf("%s/spacelimit/f_%d", getUserName(0), 11)
		res, _, errs := cl.Create(filePath, 1)
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if res.StatusCode != 429 {
			t.Fatal("(space limit): this request should be rejected")
		}
	})

	resp, _, errs = usersCl.Logout(token)
	if len(errs) > 0 {
		t.Fatal(errs)
	} else if resp.StatusCode != 200 {
		t.Fatal(resp.StatusCode)
	}
}
