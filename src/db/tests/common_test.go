package tests

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/ihexxa/quickshare/src/db/rdb/sqlite"
)

func TestSqliteInit(t *testing.T) {
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
