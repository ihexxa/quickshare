package fileinfostore

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/ihexxa/quickshare/src/kvstore/boltdbpvd"
)

func TestUserStores(t *testing.T) {

	testSharingMethods := func(t *testing.T, store IFileInfoStore) {
		dirPaths := []string{"admin/path1", "admin/path1/path2"}
		var err error
		for _, dirPath := range dirPaths {
			err = store.AddSharing(dirPath)
			if err != nil {
				t.Fatal(err)
			}
		}

		prefix := "admin"
		sharingMap, err := store.ListSharings(prefix)
		if err != nil {
			t.Fatal(err)
		}

		for _, sharingDir := range dirPaths {
			if !sharingMap[sharingDir] {
				t.Fatalf("sharing(%s) not found", sharingDir)
			}
			mustTrue, exist := store.GetSharing(sharingDir)
			if !mustTrue || !exist {
				t.Fatalf("get sharing(%t %t) should exist", mustTrue, exist)
			}
		}

		for _, dirPath := range dirPaths {
			err = store.DelSharing(dirPath)
			if err != nil {
				t.Fatal(err)
			}
		}

		sharingMap, err = store.ListSharings(prefix)
		if err != nil {
			t.Fatal(err)
		}
		for _, dirPath := range dirPaths {
			if sharingMap[dirPath] {
				t.Fatalf("sharing(%s) should not exist", dirPath)
			}
			_, exist := store.GetSharing(dirPath)
			if exist {
				t.Fatalf("get sharing(%t) should not exit", exist)
			}
		}
	}

	t.Run("testing FileInfoStore", func(t *testing.T) {
		rootPath, err := ioutil.TempDir("./", "quickshare_userstore_test_")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(rootPath)

		kvstore := boltdbpvd.New(rootPath, 1024)
		defer kvstore.Close()

		store, err := NewFileInfoStore(kvstore)
		if err != nil {
			t.Fatal("fail to new kvstore", err)
		}

		testSharingMethods(t, store)
	})
}
