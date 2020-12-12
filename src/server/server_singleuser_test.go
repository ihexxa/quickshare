package server

import (
	"os"
	"testing"

	"github.com/ihexxa/quickshare/src/client"
	su "github.com/ihexxa/quickshare/src/handlers/singleuserhdr"
)

func xTestSingleUserHandlers(t *testing.T) {
	addr := "http://127.0.0.1:8888"
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

	suCl := client.NewSingleUserClient(addr)

	if !waitForReady(addr) {
		t.Fatal("fail to start server")
	}

	t.Run("test single user APIs: Login-SetPwd-Logout-Login", func(t *testing.T) {
		resp, _, errs := suCl.Login(adminName, adminPwd)
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if resp.StatusCode != 200 {
			t.Fatal(resp.StatusCode)
		}

		token := client.GetCookie(resp.Cookies(), su.TokenCookie)

		resp, _, errs = suCl.SetPwd(adminPwd, adminNewPwd, token)
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if resp.StatusCode != 200 {
			t.Fatal(resp.StatusCode)
		}

		resp, _, errs = suCl.Logout(adminName, token)
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if resp.StatusCode != 200 {
			t.Fatal(resp.StatusCode)
		}

		resp, _, errs = suCl.Login(adminName, adminNewPwd)
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if resp.StatusCode != 200 {
			t.Fatal(resp.StatusCode)
		}
	})
}
