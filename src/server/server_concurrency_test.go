package server

import (
	"errors"
	"fmt"
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

	getFilePath := func(name string, i int) string {
		return fmt.Sprintf("%s/files/home_file_%d", name, i)
	}

	filesCount := 10
	clientErrs := []error{}
	mockClient := func(name, pwd string, wg *sync.WaitGroup) {
		defer wg.Done()

		userUsersCli := client.NewUsersClient(addr)
		resp, _, errs := userUsersCli.Login(name, pwd)
		if len(errs) > 0 {
			clientErrs = append(clientErrs, errs...)
			return
		} else if resp.StatusCode != 200 {
			clientErrs = append(clientErrs, fmt.Errorf("failed to login"))
			return
		}

		files := map[string]string{}
		content := "12345678"
		for i := range make([]int, filesCount, filesCount) {
			files[getFilePath(name, i)] = content
		}

		userToken := userUsersCli.Token()
		for filePath, content := range files {
			assertUploadOK(t, filePath, content, addr, userToken)
			assertDownloadOK(t, filePath, content, addr, userToken)
		}

		filesCl := client.NewFilesClient(addr, userToken)
		resp, lsResp, errs := filesCl.ListHome()
		if len(errs) > 0 {
			clientErrs = append(clientErrs, errs...)
			return
		} else if resp.StatusCode != 200 {
			clientErrs = append(clientErrs, errors.New("failed to list home"))
			return
		}

		if lsResp.Cwd != fmt.Sprintf("%s/files", name) {
			clientErrs = append(clientErrs, fmt.Errorf("incorrct cwd (%s)", lsResp.Cwd))
			return

		} else if len(lsResp.Metadatas) != len(files) {
			clientErrs = append(clientErrs, fmt.Errorf("incorrct metadata size (%d)", len(lsResp.Metadatas)))
			return
		}

		resp, selfResp, errs := userUsersCli.Self()
		if len(errs) > 0 {
			clientErrs = append(clientErrs, errs...)
			return
		} else if resp.StatusCode != 200 {
			clientErrs = append(clientErrs, errors.New("failed to self"))
			return
		}
		if selfResp.UsedSpace != int64(filesCount*len(content)) {
			clientErrs = append(
				clientErrs,
				fmt.Errorf("usedSpace(%d) doesn't match (%d)", selfResp.UsedSpace, filesCount*len(content)),
			)
			return
		}

		resp, _, errs = filesCl.Delete(getFilePath(name, 0))
		if len(errs) > 0 {
			clientErrs = append(clientErrs, errs...)
			return
		} else if resp.StatusCode != 200 {
			clientErrs = append(clientErrs, errors.New("failed to add user"))
			return
		}

		resp, selfResp, errs = userUsersCli.Self()
		if len(errs) > 0 {
			clientErrs = append(clientErrs, errs...)
			return
		} else if resp.StatusCode != 200 {
			clientErrs = append(clientErrs, errors.New("failed to self"))
			return
		}
		if selfResp.UsedSpace != int64((filesCount-1)*len(content)) {
			clientErrs = append(
				clientErrs,
				fmt.Errorf(
					"usedSpace(%d) doesn't match (%d)",
					selfResp.UsedSpace,
					int64((filesCount-1)*len(content)),
				),
			)
			return
		}
	}

	var wg sync.WaitGroup
	t.Run("Upload and download concurrently", func(t *testing.T) {
		for userName := range users {
			wg.Add(1)
			go mockClient(userName, userPwd, &wg)
		}

		wg.Wait()
		if len(clientErrs) > 0 {
			t.Fatal(joinErrs(clientErrs))
		}
	})
}
