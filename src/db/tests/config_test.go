package tests

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/ihexxa/quickshare/src/db"
	"github.com/ihexxa/quickshare/src/db/rdb/sqlite"
)

var testSiteConfig = &db.SiteConfig{
	ClientCfg: &db.ClientConfig{
		SiteName:   "",
		SiteDesc:   "",
		AllowSetBg: true,
		AutoTheme:  false,
		Bg: &db.BgConfig{
			Url:      "/imgs/bg.jpg",
			Repeat:   "repeat",
			Position: "top",
			Align:    "scroll",
			BgColor:  "#000",
		},
	},
}

func TestSiteStore(t *testing.T) {
	testConfigMethods := func(t *testing.T, store db.IConfigDB) {
		siteCfg := &db.SiteConfig{
			ClientCfg: &db.ClientConfig{
				SiteName:   "quickshare",
				SiteDesc:   "simpel file sharing",
				AllowSetBg: true,
				AutoTheme:  true,
				Bg: &db.BgConfig{
					Url:      "/imgs/bg.jpg",
					Repeat:   "no-repeat",
					Position: "center",
					Align:    "fixed",
					BgColor:  "#ccc",
				},
			},
		}

		ctx := context.TODO()
		err := store.SetClientCfg(ctx, siteCfg.ClientCfg)
		if err != nil {
			t.Fatal(err)
		}
		newSiteCfg, err := store.GetCfg(ctx)
		if err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(newSiteCfg, siteCfg) {
			t.Fatalf("not equal new(%v) original(%v)", newSiteCfg, siteCfg)
		}
	}

	t.Run("config methods basic tests - sqlite", func(t *testing.T) {
		rootPath, err := ioutil.TempDir("./", "qs_sqlite_config_")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(rootPath)

		dbPath := filepath.Join(rootPath, "quickshare.sqlite")
		sqliteDB, err := sqlite.NewSQLite(dbPath)
		if err != nil {
			t.Fatal(err)
		}
		defer sqliteDB.Close()

		store, err := sqlite.NewSQLiteStore(sqliteDB)
		if err != nil {
			t.Fatal("fail to new sqlite store", err)
		}
		err = store.Init(context.TODO(), "admin", "adminPwd", testSiteConfig)
		if err != nil {
			panic(err)
		}

		testConfigMethods(t, store)
	})

	t.Run("idemptent initialization - sqlite", func(t *testing.T) {
		rootPath, err := ioutil.TempDir("./", "qs_sqlite_config_")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(rootPath)

		dbPath := filepath.Join(rootPath, "quickshare.sqlite")
		sqliteDB, err := sqlite.NewSQLite(dbPath)
		if err != nil {
			t.Fatal(err)
		}
		defer sqliteDB.Close()

		store, err := sqlite.NewSQLiteStore(sqliteDB)
		if err != nil {
			t.Fatal("fail to new sqlite store", err)
		}

		for i := 0; i < 2; i++ {
			err = store.Init(context.TODO(), "admin", "adminPwd", testSiteConfig)
			if err != nil {
				panic(err)
			}
		}
	})
}
