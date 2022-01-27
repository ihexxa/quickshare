package server

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/ihexxa/quickshare/src/client"
	"github.com/ihexxa/quickshare/src/db/sitestore"
	q "github.com/ihexxa/quickshare/src/handlers"
	"github.com/ihexxa/quickshare/src/handlers/settings"
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
	setUpEnv(t, rootPath, adminName, adminPwd)
	defer os.RemoveAll(rootPath)

	srv := startTestServer(config)
	defer srv.Shutdown()
	fs := srv.depsFS()

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
			clientCfgMsg := &settings.ClientCfgMsg{
				SiteName: cfg.SiteName,
				SiteDesc: cfg.SiteDesc,
				Bg:       cfg.Bg,
			}
			resp, _, errs := settingsCl.SetClientCfg(clientCfgMsg, adminToken)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != 200 {
				t.Fatal(resp.StatusCode)
			}

			resp, clientCfgMsgGot, errs := settingsCl.GetClientCfg(adminToken)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != 200 {
				t.Fatal(resp.StatusCode)
			}

			cfgEqual := func(cfg1, cfg2 *settings.ClientCfgMsg) bool {
				return cfg1.SiteName == cfg2.SiteName &&
					cfg1.SiteDesc == cfg2.SiteDesc &&
					reflect.DeepEqual(cfg1.Bg, cfg2.Bg)
			}

			if !cfgEqual(clientCfgMsg, clientCfgMsgGot) {
				t.Fatalf("client cfgs are not equal: got(%v) expected(%v)", clientCfgMsg, clientCfgMsgGot)
			}

			for userName := range users {
				resp, _, errs := usersCl.Login(userName, userPwd)
				if len(errs) > 0 {
					t.Fatal(errs)
				} else if resp.StatusCode != 200 {
					t.Fatal(resp.StatusCode)
				}
				userToken := client.GetCookie(resp.Cookies(), q.TokenCookie)

				resp, clientCfgMsgGot, errs := settingsCl.GetClientCfg(userToken)
				if len(errs) > 0 {
					t.Fatal(errs)
				} else if resp.StatusCode != 200 {
					t.Fatal(resp.StatusCode)
				}

				if !cfgEqual(clientCfgMsg, clientCfgMsgGot) {
					t.Fatalf("client cfgs are not equal: got(%v) expected(%v)", clientCfgMsg, clientCfgMsgGot)
				}
			}
		}
	})

	t.Run("ReportErrors", func(t *testing.T) {
		settingsCl := client.NewSettingsClient(addr)
		reports := &settings.ClientErrorReports{
			Reports: []*settings.ClientErrorReport{
				&settings.ClientErrorReport{
					Report:  `{state: "{}", error: "empty state1"}`,
					Version: "0.0.1",
				},
				&settings.ClientErrorReport{
					Report:  `{state: "{}", error: "empty state2"}`,
					Version: "0.0.1",
				},
			},
		}

		reportResp, _, errs := settingsCl.ReportErrors(reports, adminToken)
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if reportResp.StatusCode != 200 {
			t.Fatal(reportResp.StatusCode)
		}

		file, id, err := fs.GetFileReader("quickshare.log")
		if err != nil {
			t.Fatal(err)
		}
		defer fs.CloseReader(fmt.Sprint(id))

		// TODO: it is flaky
		time.Sleep(time.Duration(1) * time.Second)

		content, err := ioutil.ReadAll(file)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(content), `"msg":"version:0.0.1,error:{state: \"{}\", error: \"empty state1\"}"`) {
			t.Fatalf("log does not contain error: %s", content)
		}
		if !strings.Contains(string(content), `"msg":"version:0.0.1,error:{state: \"{}\", error: \"empty state2\"}"`) {
			t.Fatalf("log does not contain error: %s", content)
		}
	})
}
