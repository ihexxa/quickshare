package server

import (
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/ihexxa/quickshare/src/client"
	q "github.com/ihexxa/quickshare/src/handlers"
)

func xTestConcurrency(t *testing.T) {
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

	usersCl := client.NewSingleUserClient(addr)
	resp, _, errs := usersCl.Login(adminName, adminPwd)
	if len(errs) > 0 {
		t.Fatal(errs)
	} else if resp.StatusCode != 200 {
		t.Fatal(resp.StatusCode)
	}
	token := client.GetCookie(resp.Cookies(), q.TokenCookie)

	userCount := 5
	userPwd := "1234"
	users := addUsers(t, addr, userPwd, userCount, token)

	getFilePath := func(name string, i int) string {
		return fmt.Sprintf("%s/files/home_file_%d", name, i)
	}

	filesCount := 10
	mockClient := func(name, pwd string, wg *sync.WaitGroup) {
		usersCl := client.NewSingleUserClient(addr)
		resp, _, errs := usersCl.Login(name, pwd)
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if resp.StatusCode != 200 {
			t.Fatal("failed to add user")
		}
		token := client.GetCookie(resp.Cookies(), q.TokenCookie)

		files := map[string]string{}
		content := "12345678"
		for i := range make([]int, filesCount, filesCount) {
			files[getFilePath(name, i)] = content
		}

		for filePath, content := range files {
			assertUploadOK(t, filePath, content, addr, token)
			assertDownloadOK(t, filePath, content, addr, token)
		}

		filesCl := client.NewFilesClient(addr, token)
		resp, lsResp, errs := filesCl.ListHome()
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if resp.StatusCode != 200 {
			t.Fatal("failed to add user")
		}

		if lsResp.Cwd != fmt.Sprintf("%s/files", name) {
			t.Fatalf("incorrct cwd (%s)", lsResp.Cwd)
		} else if len(lsResp.Metadatas) != len(files) {
			t.Fatalf("incorrct metadata size (%d)", len(lsResp.Metadatas))
		}

		resp, selfResp, errs := usersCl.Self(token)
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if resp.StatusCode != 200 {
			t.Fatal("failed to self")
		}
		if selfResp.UsedSpace != int64(filesCount*len(content)) {
			t.Fatalf("usedSpace(%d) doesn't match (%d)", selfResp.UsedSpace, filesCount*len(content))
		}

		resp, _, errs = filesCl.Delete(getFilePath(name, 0))
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if resp.StatusCode != 200 {
			t.Fatal("failed to add user")
		}

		resp, selfResp, errs = usersCl.Self(token)
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if resp.StatusCode != 200 {
			t.Fatal("failed to self")
		}
		if selfResp.UsedSpace != int64((filesCount-1)*len(content)) {
			t.Fatalf("usedSpace(%d) doesn't match (%d)", selfResp.UsedSpace, int64((filesCount-1)*len(content)))
		}

		wg.Done()
	}

	var wg sync.WaitGroup
	t.Run("Upload and download concurrently", func(t *testing.T) {
		for userName := range users {
			wg.Add(1)
			go mockClient(userName, userPwd, &wg)
		}

		wg.Wait()
	})

	resp, _, errs = usersCl.Logout(token)
	if len(errs) > 0 {
		t.Fatal(errs)
	} else if resp.StatusCode != 200 {
		t.Fatal(resp.StatusCode)
	}
}
