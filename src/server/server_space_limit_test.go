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
	rootPath := "testData"
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
			"uploadSpeedLimit": 409600,
			"downloadSpeedLimit": 409600,
			"spaceLimit": %d,
			"limiterCapacity": 1000,
			"limiterCyc": 1000
		},
		"server": {
			"debug": true,
			"host": "127.0.0.1"
		},
		"fs": {
			"root": "testData"
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
		} else if res.StatusCode != 429 {
			t.Fatal("(space limit): this request should be rejected")
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

	resp, _, errs = usersCl.Logout(token)
	if len(errs) > 0 {
		t.Fatal(errs)
	} else if resp.StatusCode != 200 {
		t.Fatal(resp.StatusCode)
	}
}
