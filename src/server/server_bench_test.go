package server

import (
	"os"
	"sync"
	"testing"

	"github.com/ihexxa/quickshare/src/client"
	q "github.com/ihexxa/quickshare/src/handlers"
)

func BenchmarkUploadAndDownload(b *testing.B) {
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
			"root": "tmpTestData",
			"opensLimit": 1024
		},
		"db": {
			"dbPath": "tmpTestData/quickshare"
		}
	}`

	adminName := "qs"
	adminPwd := "quicksh@re"
	setUpEnv(b, rootPath, adminName, adminPwd)
	defer os.RemoveAll(rootPath)

	srv := startTestServer(config)
	defer srv.Shutdown()
	if !isServerReady(addr) {
		b.Fatal("fail to start server")
	}

	adminUsersCli := client.NewUsersClient(addr)
	resp, _, errs := adminUsersCli.Login(adminName, adminPwd)
	if len(errs) > 0 {
		b.Fatal(errs)
	} else if resp.StatusCode != 200 {
		b.Fatal(resp.StatusCode)
	}
	adminToken := client.GetCookie(resp.Cookies(), q.TokenCookie)

	userCount := 5
	userPwd := "1234"
	users := addUsers(b, addr, userPwd, userCount, adminToken)
	filesCount := 30
	rounds := 1

	clients := []*MockClient{}
	for i := 0; i < b.N; i++ {
		for j := 0; j < rounds; j++ {
			var wg sync.WaitGroup
			for userName := range users {
				client := &MockClient{errs: []error{}}
				clients = append(clients, client)
				wg.Add(1)
				go client.uploadAndDownload(b, addr, userName, userPwd, filesCount, &wg)
			}

			wg.Wait()

			errs := []error{}
			for _, client := range clients {
				if len(client.errs) > 0 {
					errs = append(errs, client.errs...)
				}
			}
			if len(errs) > 0 {
				b.Fatal(joinErrs(errs))
			}
		}
	}
}
