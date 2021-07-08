package server

import (
	"fmt"
	"os"
	"testing"

	"github.com/ihexxa/quickshare/src/client"
	su "github.com/ihexxa/quickshare/src/handlers/singleuserhdr"
	"github.com/ihexxa/quickshare/src/userstore"
)

func TestSingleUserHandlers(t *testing.T) {
	addr := "http://127.0.0.1:8686"
	root := "testData"
	config := `{
		"users": {
			"enableAuth": true
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

	usersCl := client.NewSingleUserClient(addr)

	if !waitForReady(addr) {
		t.Fatal("fail to start server")
	}

	t.Run("test users APIs: Login-SetPwd-Logout-Login", func(t *testing.T) {
		resp, _, errs := usersCl.Login(adminName, adminPwd)
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if resp.StatusCode != 200 {
			t.Fatal(resp.StatusCode)
		}

		token := client.GetCookie(resp.Cookies(), su.TokenCookie)

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

	t.Run("test users APIs: Login-AddUser-Logout-Login", func(t *testing.T) {
		resp, _, errs := usersCl.Login(adminName, adminNewPwd)
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if resp.StatusCode != 200 {
			t.Fatal(resp.StatusCode)
		}

		token := client.GetCookie(resp.Cookies(), su.TokenCookie)

		userName, userPwd := "user", "1234"
		resp, auResp, errs := usersCl.AddUser(userName, userPwd, userstore.UserRole, token)
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if resp.StatusCode != 200 {
			t.Fatal(resp.StatusCode)
		}
		// TODO: check id
		fmt.Printf("new user id: %v\n", auResp)

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
	})
}
