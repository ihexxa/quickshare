package server

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"testing"

	"github.com/ihexxa/gocfg"
	"github.com/ihexxa/quickshare/src/depidx"
)

func TestInit(t *testing.T) {
	dbFileName := "test_init.sqlite"
	adminName := "admin"
	adminPwd := "1234"

	prepareCfg := func(initEnv bool, rootPath, dbFileName string) *gocfg.Cfg {
		config :=
			fmt.Sprintf(
				`{
				"users": {
					"enableAuth": true,
					"minUserNameLen": 2,
					"minPwdLen": 4,
					"captchaEnabled": false,
					"uploadSpeedLimit": 409600,
					"downloadSpeedLimit": 409600,
					"spaceLimit": 1000,
					"limiterCapacity": 1000,
					"limiterCyc": 1000,
					"predefinedUsers": [
						{
							"name": "test_user",
							"pwd": "Quicksh@re",
							"role": "user"
						}
					]
				},
				"server": {
					"debug": true,
					"host": "127.0.0.1",
					"initFileIndex": true
				},
				"fs": {
					"root": "%s"
				},
				"db": {
					"dbPath": "%s"
				}
			}`,
				rootPath,
				dbFileName,
			)

		if initEnv {
			os.Setenv("DEFAULTADMIN", adminName)
			os.Setenv("DEFAULTADMINPWD", adminPwd)
		} else {
			os.Unsetenv("DEFAULTADMIN")
			os.Unsetenv("DEFAULTADMINPWD")
		}
		os.RemoveAll(rootPath)
		err := os.MkdirAll(rootPath, 0700)
		if err != nil {
			t.Fatal(err)
		}

		defaultCfg, err := DefaultConfig()
		if err != nil {
			t.Fatal(err)
		}
		cfg, err := gocfg.New(NewConfig()).
			Load(
				gocfg.JSONStr(defaultCfg),
				gocfg.JSONStr(config),
			)
		if err != nil {
			t.Fatal(err)
		}

		return cfg
	}

	prepareTestDeps := func() (*depidx.Deps, string, *gocfg.Cfg, *Initer) {
		rootPath := fmt.Sprintf("tmpTestData/t_%d", rand.Int())
		cfg := prepareCfg(true, rootPath, dbFileName)
		initer := NewIniter(cfg)
		return initer.InitDeps(), rootPath, cfg, initer
	}

	t.Run("deps/fs: log, db are created", func(t *testing.T) {
		deps, rootPath, _, _ := prepareTestDeps()
		defer os.RemoveAll(rootPath)

		for _, itemPath := range []string{
			dbFileName,
			"quickshare.log",
		} {
			_, err := deps.FS().Stat(itemPath)
			if err != nil {
				t.Fatalf("item(%s) not found: %s", itemPath, err)
			}
		}
	})

	t.Run("deps/db: tables are inited", func(t *testing.T) {
		deps, rootPath, _, _ := prepareTestDeps()
		defer os.RemoveAll(rootPath)

		ctx := context.TODO()
		_, err := deps.DB().ListUsers(ctx)
		if err != nil {
			t.Fatal(err)
		}

		_, err = deps.DB().ListFileInfos(ctx, []string{dbFileName})
		if err != nil {
			t.Fatal(err)
		}

		_, err = deps.DB().ListUploadInfos(ctx, 0)
		if err != nil {
			t.Fatal(err)
		}

		_, err = deps.DB().ListSharingsByLocation(ctx, "/")
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("handlers/fs: home folders are created ", func(t *testing.T) {
		deps, rootPath, _, initer := prepareTestDeps()
		initer.InitHandlers(deps)
		defer os.RemoveAll(rootPath)

		for _, itemPath := range []string{
			dbFileName,
			"quickshare.log",
			"admin",
			"test_user",
		} {
			_, err := deps.FS().Stat(itemPath)
			if err != nil {
				t.Fatalf("item(%s) not found: %s", itemPath, err)
			}
		}
	})

	t.Run("handlers/db: db tables are inited ", func(t *testing.T) {
		deps, rootPath, _, initer := prepareTestDeps()
		initer.InitHandlers(deps)
		defer os.RemoveAll(rootPath)

		ctx := context.TODO()

		// check users
		users, err := deps.DB().ListUsers(ctx)
		if err != nil {
			t.Fatal(err)
		}
		expectedUsers := map[string]bool{
			"admin":     true,
			"test_user": true,
			"visitor":   true,
		}
		if len(expectedUsers) != len(users) {
			t.Fatal("users size not match")
		}
		for _, user := range users {
			if !expectedUsers[user.Name] {
				t.Fatalf("user(%s) not found", user.Name)
			}
		}
	})

	t.Run("deps: idempotancy", func(t *testing.T) {
		_, rootPath, _, initer := prepareTestDeps()
		defer os.RemoveAll(rootPath)
		// init again
		deps := initer.InitDeps()
		initer.InitHandlers(deps)

		for _, itemPath := range []string{
			dbFileName,
			"quickshare.log",
		} {
			_, err := deps.FS().Stat(itemPath)
			if err != nil {
				t.Fatalf("item(%s) not found: %s", itemPath, err)
			}
		}
	})

	t.Run("use input admin name", func(t *testing.T) {
		rootPath := fmt.Sprintf("tmpTestData/t_%d", rand.Int())
		cfg := prepareCfg(false, rootPath, dbFileName)
		initer := NewIniter(cfg)
		defer os.RemoveAll(rootPath)

		// prepare password
		inBuf := bytes.NewBuffer(make([]byte, 1024))
		outBuf := bytes.NewBuffer(make([]byte, 1024))
		inputAdminName := "patrick_star"
		_, err := io.WriteString(inBuf, inputAdminName)
		if err != nil {
			t.Fatal(err)
		}
		initer.input = inBuf
		initer.output = outBuf

		initer.InitDeps()

		outputs, err := io.ReadAll(outBuf)
		if err != nil {
			t.Fatal(err)
		}

		index := strings.Index(string(outputs), "password is generated: ") + len("password is generated: ")
		if cfg.GrabString("ENV.DEFAULTADMINPWD") != string(outputs[index:index+6]) {
			t.Fatalf(
				"pwd not match: (%s) (%s)",
				cfg.GrabString("ENV.DEFAULTADMIN"),
				string(outputs[index:index+6]),
			)
		}
	})
}
