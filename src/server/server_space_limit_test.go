package server

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/ihexxa/quickshare/src/client"
	"github.com/ihexxa/quickshare/src/db/userstore"
	q "github.com/ihexxa/quickshare/src/handlers"
)

func TestSpaceLimit(t *testing.T) {
	addr := "http://127.0.0.1:8686"
	rootPath := "tmpTestData"
	spaceLimit := 1000000
	fileSize := 100000
	if spaceLimit%fileSize != 0 {
		t.Fatal("spaceLimit % fileSize must be zero")
	}
	config := fmt.Sprintf(`{
		"users": {
			"enableAuth": true,
			"minUserNameLen": 2,
			"minPwdLen": 4,
			"captchaEnabled": false,
			"uploadSpeedLimit": 4096000,
			"downloadSpeedLimit": 4096000,
			"spaceLimit": %d,
			"limiterCapacity": 1000,
			"limiterCyc": 1000,
			"predefinedUsers": [
				{
					"name": "test",
					"pwd": "test",
					"role": "admin"
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
	}`, spaceLimit)

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

	t.Run("test space limiting: Upload", func(t *testing.T) {
		usersCl := client.NewSingleUserClient(addr)
		resp, _, errs := usersCl.Login(getUserName(0), userPwd)
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if resp.StatusCode != 200 {
			t.Fatal(resp.StatusCode)
		}
		token := client.GetCookie(resp.Cookies(), q.TokenCookie)

		fileContent := ""
		for i := 0; i < fileSize; i++ {
			fileContent += "0"
		}

		for i := 0; i < spaceLimit/fileSize; i++ {
			ok := assertUploadOK(t, fmt.Sprintf("%s/files/spacelimit/f_%d", getUserName(0), i), fileContent, addr, token)
			if !ok {
				t.Fatalf("space limit failed at %d", i)
			}

			resp, selfResp, errs := usersCl.Self(token)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != 200 {
				t.Fatal("failed to get self")
			} else if selfResp.UsedSpace != int64((i+1)*fileSize) {
				t.Fatal("incorrect used space")
			}
		}

		cl := client.NewFilesClient(addr, token)
		filePath := fmt.Sprintf("%s/files/spacelimit/f_%d", getUserName(0), 11)
		res, _, errs := cl.Create(filePath, 1)
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if res.StatusCode != 403 {
			t.Fatalf("(space limit): this request should be rejected: %d", res.StatusCode)
		}

		for i := 0; i < spaceLimit/fileSize; i++ {
			resp, _, errs := cl.Delete(fmt.Sprintf("%s/files/spacelimit/f_%d", getUserName(0), i))
			if len(errs) > 0 {
				t.Fatalf("failed to delete %d", i)
			} else if resp.StatusCode != 200 {
				t.Fatalf("failed to delete status %d", resp.StatusCode)
			}

			resp, selfResp, errs := usersCl.Self(token)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != 200 {
				t.Fatal("failed to get self")
			} else if selfResp.UsedSpace != int64(spaceLimit)-int64((i+1)*fileSize) {
				t.Fatal("incorrect used space")
			}
		}
	})

	t.Run("usedSpace keeps correct in operations: Mkdir-Create-UploadChunk-AddSharing-Move-IsSharing-List", func(t *testing.T) {
		srcDir := "qs/files/folder/move/src"
		dstDir := "qs/files/folder/move/dst"
		filesCl := client.NewFilesClient(addr, token)

		getUsedSpace := func() int64 {
			resp, selfResp, errs := usersCl.Self(token)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != 200 {
				t.Fatal("failed to get self")
			}

			return selfResp.UsedSpace
		}

		initUsedSpace := getUsedSpace()

		for _, dirPath := range []string{srcDir, dstDir} {
			res, _, errs := filesCl.Mkdir(dirPath)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if res.StatusCode != 200 {
				t.Fatal(res.StatusCode)
			}
		}
		if getUsedSpace() != initUsedSpace {
			t.Fatal("incorrect used space")
		}

		expectedUsedSpace := initUsedSpace
		files := map[string]string{
			"f1.md": "111",
			"f2.md": "22222",
		}
		for fileName, content := range files {
			oldPath := filepath.Join(srcDir, fileName)
			assertUploadOK(t, oldPath, content, addr, token)
			expectedUsedSpace += int64(len(content))
		}
		if getUsedSpace() != expectedUsedSpace {
			t.Fatal("used space incorrect")
		}

		for fileName := range files {
			oldPath := filepath.Join(srcDir, fileName)
			newPath := filepath.Join(dstDir, fileName)
			res, _, errs := filesCl.Move(oldPath, newPath)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if res.StatusCode != 200 {
				t.Fatal(res.StatusCode)
			}

			if getUsedSpace() != expectedUsedSpace {
				t.Fatal("used space incorrect")
			}
		}

		res, _, errs := filesCl.Delete(dstDir)
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if res.StatusCode != 200 {
			t.Fatal(res.StatusCode)
		}

		if getUsedSpace() != initUsedSpace {
			t.Fatal("used space incorrect")
		}
	})

	t.Run("ResetUsedSpace", func(t *testing.T) {
		usersCl := client.NewSingleUserClient(addr)
		resp, _, errs := usersCl.Login("test", "test")
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if resp.StatusCode != 200 {
			t.Fatal(resp.StatusCode)
		}
		token := client.GetCookie(resp.Cookies(), q.TokenCookie)

		ok := assertUploadOK(t, "test/files/spacelimit/byte1", "0", addr, token)
		if !ok {
			t.Fatal("upload failed")
		}

		resp, selfResp, errs := usersCl.Self(token)
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if resp.StatusCode != 200 {
			t.Fatal(resp.StatusCode)
		}
		originalUsedSpace := selfResp.UsedSpace

		uidInt, err := strconv.ParseUint(selfResp.ID, 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		resp, _, errs = usersCl.ResetUsedSpace(uidInt, token)
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if resp.StatusCode != 200 {
			t.Fatal(resp.StatusCode)
		}

		settingsCl := client.NewSettingsClient(addr)
		for i := 0; i < 20; i++ {
			resp, wqlResp, errs := settingsCl.WorkerQueueLen(token)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != 200 {
				t.Fatal(resp.StatusCode)
			}

			if wqlResp.QueueLen == 0 {
				break
			}

			time.Sleep(200)
		}

		resp, selfResp, errs = usersCl.Self(token)
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if resp.StatusCode != 200 {
			t.Fatal(resp.StatusCode)
		} else if selfResp.UsedSpace != originalUsedSpace {
			t.Fatalf("used space not equal %d %d", selfResp.UsedSpace, originalUsedSpace)
		}
	})

	resp, _, errs = usersCl.Logout(token)
	if len(errs) > 0 {
		t.Fatal(errs)
	} else if resp.StatusCode != 200 {
		t.Fatal(resp.StatusCode)
	}
}
