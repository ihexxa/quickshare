package server

import (
	"os"
	"reflect"
	"testing"

	"github.com/ihexxa/quickshare/src/client"
	"github.com/ihexxa/quickshare/src/db/sitestore"
	q "github.com/ihexxa/quickshare/src/handlers"
)

func TestSettingsHandlers(t *testing.T) {
	addr := "http://127.0.0.1:8686"
	rootPath := "testData"
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
			"limiterCyc": 1000,
			"predefinedUsers": [
				{
					"name": "demo",
					"pwd": "Quicksh@re",
					"role": "user"
				}
			]
		},
		"server": {
			"debug": true,
			"host": "127.0.0.1"
		},
		"fs": {
			"root": "testData"
		}
	}`
	adminName := "qs"
	adminPwd := "quicksh@re"
	userPwd := "1234"
	// adminNewPwd := "quicksh@re2"
	setUpEnv(t, rootPath, adminName, adminPwd)
	defer os.RemoveAll(rootPath)

	srv := startTestServer(config)
	defer srv.Shutdown()
	// fs := srv.depsFS()

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
	adminToken := client.GetCookie(resp.Cookies(), q.TokenCookie)
	users := addUsers(t, addr, userPwd, 1, adminToken)

	t.Run("get/set client config", func(t *testing.T) {
		settingsCl := client.NewSettingsClient(addr)
		cfgs := []*sitestore.ClientConfig{
			&sitestore.ClientConfig{
				SiteName: "quickshare",
				SiteDesc: "quickshare",
				Bg: &sitestore.BgConfig{
					Url:      "",
					Repeat:   "",
					Position: "",
					Align:    "",
				},
			},
		}

		for _, cfg := range cfgs {
			resp, _, errs := settingsCl.SetClientCfg(cfg, adminToken)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != 200 {
				t.Fatal(resp.StatusCode)
			}

			resp, clientCfgMsg, errs := settingsCl.GetClientCfg(adminToken)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != 200 {
				t.Fatal(resp.StatusCode)
			}

			if !reflect.DeepEqual(cfg, clientCfgMsg.ClientCfg) {
				t.Fatalf("client cfgs are not equal: got(%v) expected(%v)", clientCfgMsg.ClientCfg, cfg)
			}

			for userName := range users {
				resp, _, errs := usersCl.Login(userName, userPwd)
				if len(errs) > 0 {
					t.Fatal(errs)
				} else if resp.StatusCode != 200 {
					t.Fatal(resp.StatusCode)
				}
				userToken := client.GetCookie(resp.Cookies(), q.TokenCookie)

				resp, clientCfgMsg, errs := settingsCl.GetClientCfg(userToken)
				if len(errs) > 0 {
					t.Fatal(errs)
				} else if resp.StatusCode != 200 {
					t.Fatal(resp.StatusCode)
				}

				if !reflect.DeepEqual(cfg, clientCfgMsg.ClientCfg) {
					t.Fatalf("client cfgs are not equal for user: got(%v) expected(%v)", clientCfgMsg.ClientCfg, cfg)
				}
			}
		}
	})
}
