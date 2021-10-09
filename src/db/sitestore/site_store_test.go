package sitestore

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/ihexxa/quickshare/src/kvstore/boltdbpvd"
)

func TestSiteStore(t *testing.T) {

	testSiteMethods := func(t *testing.T, store ISiteStore) {
		siteCfg := &SiteConfig{
			ClientCfg: &ClientConfig{
				SiteName: "quickshare",
				SiteDesc: "simpel file sharing",
				Bg: &BgConfig{
					Url:      "/imgs/bg.jpg",
					Repeat:   "no-repeat",
					Position: "fixed",
					Align:    "center",
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

		kvstore := boltdbpvd.New(rootPath, 1024)
		defer kvstore.Close()

		store, err := NewSiteStore(kvstore)
		if err != nil {
			t.Fatal("fail to new kvstore", err)
		}

		testSiteMethods(t, store)
	})
}
