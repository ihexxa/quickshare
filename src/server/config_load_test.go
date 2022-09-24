package server

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/ihexxa/gocfg"

	"github.com/ihexxa/quickshare/src/db"
)

func TestLoadCfg(t *testing.T) {
	argsList := []*Args{
		// default config
		&Args{
			Host:    "",
			Port:    0,
			DbPath:  "",
			Configs: []string{},
		},
		// default config + db
		&Args{
			Host:    "",
			Port:    0,
			DbPath:  "testdata/quickshare.sqlite",
			Configs: []string{},
		},
		// default config + db + config_1 (dbPath is from config_1)
		&Args{
			Host:    "",
			Port:    0,
			DbPath:  "",
			Configs: []string{"testdata/config_1.yml"},
		},
		// // default config + db + config_1 (dbPath is from args)
		&Args{
			Host:    "",
			Port:    0,
			DbPath:  "testdata/quickshare.sqlite",
			Configs: []string{"testdata/config_1.yml"},
		},
		// default config + db + config_1 + config_4
		&Args{
			Host:    "",
			Port:    0,
			DbPath:  "",
			Configs: []string{"testdata/config_1.yml", "testdata/config_4.yml"},
		},
		// config partial overwrite
		&Args{
			Host:   "",
			Port:   0,
			DbPath: "",
			Configs: []string{
				"testdata/config_1.yml",
				"testdata/config_4.yml",
				"testdata/config_partial_users.yml",
				"testdata/config_partial_db.yml",
			},
		},
		// // default config + db + config_1 + config_4 + config_partial_users.yml + config_partial_db.yml + args(db)
		&Args{
			Host:   "",
			Port:   0,
			DbPath: "testdata/quickshare.sqlite",
			Configs: []string{
				"testdata/config_1.yml",
				"testdata/config_4.yml",
				"testdata/config_partial_users.yml",
				"testdata/config_partial_db.yml",
			},
		},
	}

	cfgDefault := DefaultConfigStruct()

	cfgDBOnly := *cfgDefault
	cfgDBOnly.Site = &db.SiteConfig{
		ClientCfg: &db.ClientConfig{
			SiteName: "Quickshare",
			SiteDesc: "Quick and simple file sharing",
			Bg: &db.BgConfig{
				Url:      "",
				Repeat:   "repeat",
				Position: "center",
				Align:    "fixed",
				BgColor:  "",
			},
			AllowSetBg: false,
			AutoTheme:  true,
		},
	}

	cfg1 := &Config{
		Fs: &FSConfig{
			Root:       "1",
			OpensLimit: 1,
			OpenTTL:    1,
		},
		Users: &UsersCfg{
			EnableAuth:         true,
			DefaultAdmin:       "1",
			DefaultAdminPwd:    "1",
			CookieTTL:          1,
			CookieSecure:       true,
			CookieHttpOnly:     true,
			MinUserNameLen:     1,
			MinPwdLen:          1,
			CaptchaWidth:       1,
			CaptchaHeight:      1,
			CaptchaEnabled:     true,
			UploadSpeedLimit:   1,
			DownloadSpeedLimit: 1,
			SpaceLimit:         1,
			LimiterCapacity:    1,
			LimiterCyc:         1,
			PredefinedUsers: []*db.UserCfg{
				&db.UserCfg{
					Name: "1",
					Pwd:  "1",
					Role: "1",
				},
			},
		},
		Secrets: &Secrets{
			TokenSecret: "1",
		},
		Server: &ServerCfg{
			Debug:             true,
			Host:              "1",
			Port:              1,
			ReadTimeout:       1,
			WriteTimeout:      1,
			MaxHeaderBytes:    1,
			PublicPath:        "1",
			SearchResultLimit: 16,
			InitFileIndex:     true,
		},
		Workers: &WorkerPoolCfg{
			QueueSize:   1,
			SleepCyc:    1,
			WorkerCount: 1,
		},
		Site: &db.SiteConfig{
			ClientCfg: &db.ClientConfig{
				SiteName: "Quickshare",
				SiteDesc: "Quick and simple file sharing",
				Bg: &db.BgConfig{
					Url:      "",
					Repeat:   "repeat",
					Position: "center",
					Align:    "fixed",
					BgColor:  "",
				},
				AllowSetBg: false,
				AutoTheme:  true,
			},
		},
		Db: &DbConfig{
			DbPath: "testdata/quickshare.sqlite",
		},
	}

	cfg4 := &Config{
		Fs: &FSConfig{
			Root:       "4",
			OpensLimit: 4,
			OpenTTL:    4,
		},
		Users: &UsersCfg{
			EnableAuth:         false,
			DefaultAdmin:       "4",
			DefaultAdminPwd:    "4",
			CookieTTL:          4,
			CookieSecure:       false,
			CookieHttpOnly:     false,
			MinUserNameLen:     4,
			MinPwdLen:          4,
			CaptchaWidth:       4,
			CaptchaHeight:      4,
			CaptchaEnabled:     false,
			UploadSpeedLimit:   4,
			DownloadSpeedLimit: 4,
			SpaceLimit:         4,
			LimiterCapacity:    4,
			LimiterCyc:         4,
			PredefinedUsers: []*db.UserCfg{
				&db.UserCfg{
					Name: "4",
					Pwd:  "4",
					Role: "4",
				},
			},
		},
		Secrets: &Secrets{
			TokenSecret: "4",
		},
		Server: &ServerCfg{
			Debug:             false,
			Host:              "4",
			Port:              4,
			ReadTimeout:       4,
			WriteTimeout:      4,
			MaxHeaderBytes:    4,
			PublicPath:        "4",
			SearchResultLimit: 16,
			InitFileIndex:     true,
		},
		Workers: &WorkerPoolCfg{
			QueueSize:   4,
			SleepCyc:    4,
			WorkerCount: 4,
		},
		Site: &db.SiteConfig{
			ClientCfg: &db.ClientConfig{
				SiteName: "4",
				SiteDesc: "4",
				Bg: &db.BgConfig{
					Url:      "4",
					Repeat:   "4",
					Position: "4",
					Align:    "4",
					BgColor:  "4",
				},
				AllowSetBg: true,
				AutoTheme:  true,
			},
		},
		Db: &DbConfig{
			DbPath: "4",
		},
	}

	cfg5 := &Config{
		Fs: &FSConfig{
			Root:       "4",
			OpensLimit: 4,
			OpenTTL:    4,
		},
		Users: &UsersCfg{
			EnableAuth:         true,
			DefaultAdmin:       "5",
			DefaultAdminPwd:    "5",
			CookieTTL:          4,
			CookieSecure:       true,
			CookieHttpOnly:     true,
			MinUserNameLen:     5,
			MinPwdLen:          5,
			CaptchaWidth:       5,
			CaptchaHeight:      5,
			CaptchaEnabled:     true,
			UploadSpeedLimit:   5,
			DownloadSpeedLimit: 5,
			SpaceLimit:         5,
			LimiterCapacity:    5,
			LimiterCyc:         5,
			PredefinedUsers: []*db.UserCfg{
				&db.UserCfg{
					Name: "5",
					Pwd:  "5",
					Role: "5",
				},
			},
		},
		Secrets: &Secrets{
			TokenSecret: "4",
		},
		Server: &ServerCfg{
			Debug:             false,
			Host:              "4",
			Port:              4,
			ReadTimeout:       4,
			WriteTimeout:      4,
			MaxHeaderBytes:    4,
			PublicPath:        "4",
			SearchResultLimit: 16,
			InitFileIndex:     true,
		},
		Workers: &WorkerPoolCfg{
			QueueSize:   4,
			SleepCyc:    4,
			WorkerCount: 4,
		},
		Site: &db.SiteConfig{
			ClientCfg: &db.ClientConfig{
				SiteName: "4",
				SiteDesc: "4",
				Bg: &db.BgConfig{
					Url:      "4",
					Repeat:   "4",
					Position: "4",
					Align:    "4",
					BgColor:  "4",
				},
				AllowSetBg: true,
				AutoTheme:  true,
			},
		},
		Db: &DbConfig{
			DbPath: "5",
		},
	}

	cfgWithPartialCfg := &Config{
		Fs: &FSConfig{
			Root:       "4",
			OpensLimit: 4,
			OpenTTL:    4,
		},
		Users: &UsersCfg{
			EnableAuth:         true,
			DefaultAdmin:       "5",
			DefaultAdminPwd:    "5",
			CookieTTL:          4,
			CookieSecure:       true,
			CookieHttpOnly:     true,
			MinUserNameLen:     5,
			MinPwdLen:          5,
			CaptchaWidth:       5,
			CaptchaHeight:      5,
			CaptchaEnabled:     true,
			UploadSpeedLimit:   5,
			DownloadSpeedLimit: 5,
			SpaceLimit:         5,
			LimiterCapacity:    5,
			LimiterCyc:         5,
			PredefinedUsers: []*db.UserCfg{
				&db.UserCfg{
					Name: "5",
					Pwd:  "5",
					Role: "5",
				},
			},
		},
		Secrets: &Secrets{
			TokenSecret: "4",
		},
		Server: &ServerCfg{
			Debug:             false,
			Host:              "4",
			Port:              4,
			ReadTimeout:       4,
			WriteTimeout:      4,
			MaxHeaderBytes:    4,
			PublicPath:        "4",
			SearchResultLimit: 16,
			InitFileIndex:     true,
		},
		Workers: &WorkerPoolCfg{
			QueueSize:   4,
			SleepCyc:    4,
			WorkerCount: 4,
		},
		Site: &db.SiteConfig{
			ClientCfg: &db.ClientConfig{
				SiteName: "4",
				SiteDesc: "4",
				Bg: &db.BgConfig{
					Url:      "4",
					Repeat:   "4",
					Position: "4",
					Align:    "4",
					BgColor:  "4",
				},
				AllowSetBg: true,
				AutoTheme:  true,
			},
		},
		Db: &DbConfig{
			DbPath: "5",
		},
	}

	expects := []*Config{
		cfgDefault,
		&cfgDBOnly,
		cfg1,
		cfg1,
		cfg4,
		cfg5,
		cfgWithPartialCfg,
	}

	testLoadCfg := func(t *testing.T) {
		for i, args := range argsList {
			gotCfg, err := LoadCfg(context.TODO(), args)
			if err != nil {
				t.Fatal(err)
			}

			expectCfgStruct := expects[i]
			expectCfgBytes, err := json.Marshal(expectCfgStruct)
			if err != nil {
				t.Fatal(err)
			}

			expectCfg, err := gocfg.New(NewConfig()).Load(gocfg.JSONStr(string(expectCfgBytes)))
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(gotCfg.Template(), expectCfg.Template()) {
				gotJSON, _ := gotCfg.JSON()
				expectJSON, _ := expectCfg.JSON()
				fmt.Printf("\ngot\n%s\n", gotJSON)
				fmt.Printf("\nexpected\n%s\n", expectJSON)
				t.Fatalf("%d, cfgs are not identical", i)
			}
		}
	}

	t.Run("test LoadCfg", testLoadCfg)
}

// TODO: use better comparing method
func Equal(cfg1, cfg2 *gocfg.Cfg) bool {
	// check cfg1
	for k1, v1 := range cfg1.Bools() {
		v2, ok := cfg2.Bool(k1)
		if !ok || v2 != v1 {
			fmt.Println(k1, v1, v2)
			return false
		}
	}
	for k1, v1 := range cfg1.Ints() {
		v2, ok := cfg2.Int(k1)
		if !ok || v2 != v1 {
			fmt.Println(k1, v1, v2)
			return false
		}
	}
	for k1, v1 := range cfg1.Floats() {
		v2, ok := cfg2.Float(k1)
		if !ok || v2 != v1 {
			fmt.Println(k1, v1, v2)
			return false
		}
	}
	for k1, v1 := range cfg1.Strings() {
		v2, ok := cfg2.String(k1)
		if !ok || v2 != v1 {
			fmt.Println(k1, v1, v2)
			return false
		}
	}

	// check cfg2
	for k2, v2 := range cfg2.Bools() {
		v1, ok := cfg1.Bool(k2)
		if !ok || v1 != v2 {
			fmt.Println(k2, v1, v2)
			return false
		}
	}
	for k2, v2 := range cfg2.Ints() {
		v1, ok := cfg1.Int(k2)
		if !ok || v1 != v2 {
			fmt.Println(k2, v1, v2)
			return false
		}
	}
	for k2, v2 := range cfg2.Floats() {
		v1, ok := cfg1.Float(k2)
		if !ok || v1 != v2 {
			fmt.Println(k2, v1, v2)
			return false
		}
	}
	for k2, v2 := range cfg2.Strings() {
		v1, ok := cfg1.String(k2)
		if !ok || v1 != v2 {
			fmt.Println(k2, v1, v2)
			return false
		}
	}

	return true
}
