package sitestore

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/ihexxa/quickshare/src/db"
	"github.com/ihexxa/quickshare/src/kvstore/boltdbpvd"
)

func TestSiteStore(t *testing.T) {

	testSiteMethods := func(t *testing.T, store ISiteStore) {
		siteCfg := &db.SiteConfig{
			ClientCfg: &db.ClientConfig{
				SiteName: "quickshare",
				SiteDesc: "simpel file sharing",
				Bg: &db.BgConfig{
					Url:      "/imgs/bg.jpg",
					Repeat:   "no-repeat",
					Position: "center",
					Align:    "fixed",
					BgColor:  "#ccc",
				},
			},
		}

		err := store.SetClientCfg(siteCfg.ClientCfg)
		if err != nil {
			t.Fatal(err)
		}
		newSiteCfg, err := store.GetCfg()
		if err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(newSiteCfg, siteCfg) {
			t.Fatalf("not equal new(%v) original(%v)", newSiteCfg, siteCfg)
		}
	}

	t.Run("Get/Set", func(t *testing.T) {
		rootPath, err := ioutil.TempDir("./", "quickshare_sitestore_test_")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(rootPath)

		dbPath := filepath.Join(rootPath, "quickshare.db")
		kvstore := boltdbpvd.New(dbPath, 1024)
		defer kvstore.Close()

		store, err := NewSiteStore(kvstore)
		if err != nil {
			t.Fatal("fail to new kvstore", err)
		}
		err = store.Init(&db.SiteConfig{
			ClientCfg: &db.ClientConfig{
				SiteName: "",
				SiteDesc: "",
				Bg: &db.BgConfig{
					Url:      "/imgs/bg.jpg",
					Repeat:   "repeat",
					Position: "top",
					Align:    "scroll",
					BgColor:  "#000",
				},
			},
		})
		if err != nil {
			panic(err)
		}

		testSiteMethods(t, store)
	})
}
