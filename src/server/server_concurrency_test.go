package server

import (
	"os"
	"sync"
	"testing"

	"github.com/ihexxa/quickshare/src/client"
	q "github.com/ihexxa/quickshare/src/handlers"
)

func TestConcurrency(t *testing.T) {
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
			"limiterCyc": 1000
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
	if !isServerReady(addr) {
		t.Fatal("fail to start server")
	}

	adminUsersCli := client.NewUsersClient(addr)
	resp, _, errs := adminUsersCli.Login(adminName, adminPwd)
	if len(errs) > 0 {
		t.Fatal(errs)
	} else if resp.StatusCode != 200 {
		t.Fatal(resp.StatusCode)
	}
	adminToken := client.GetCookie(resp.Cookies(), q.TokenCookie)

	userCount := 5
	userPwd := "1234"
	users := addUsers(t, addr, userPwd, userCount, adminToken)
	filesCount := 10

	var wg sync.WaitGroup
	t.Run("Upload and download concurrently", func(t *testing.T) {
		clients := []*MockClient{}
		for userName := range users {
			client := &MockClient{errs: []error{}}
			clients = append(clients, client)
			wg.Add(1)
			go client.uploadAndDownload(t, addr, userName, userPwd, filesCount, &wg)
		}

		wg.Wait()

		errs := []error{}
		for _, client := range clients {
			if len(client.errs) > 0 {
				errs = append(errs, client.errs...)
			}
		}
		if len(errs) > 0 {
			t.Fatal(joinErrs(errs))
		}
	})
}
