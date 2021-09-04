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

		// add sharings
		for _, dirPath := range dirPaths {
			err = store.AddSharing(dirPath)
			if err != nil {
				t.Fatal(err)
			}
		}

		// list sharings
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

		// del sharings
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
			if _, ok := sharingMap[dirPath]; ok {
				t.Fatalf("sharing(%s) should not exist", dirPath)
			}
			shared, exist := store.GetSharing(dirPath)
			if shared {
				t.Fatalf("get sharing(%t, %t) should not shared but exist", shared, exist)
			}
		}
	}

	testInfoMethods := func(t *testing.T, store IFileInfoStore) {
		pathInfos := map[string]*FileInfo{
			"admin/item": &FileInfo{
				Shared: false,
				IsDir:  false,
				Sha1:   "file",
			},
			"admin/dir": &FileInfo{
				Shared: true,
				IsDir:  true,
				Sha1:   "dir",
			},
		}
		var err error

		// add infos
		for itemPath, info := range pathInfos {
			err = store.SetInfo(itemPath, info)
			if err != nil {
				t.Fatal(err)
			}
		}

		// get infos
		for itemPath, expected := range pathInfos {
			info, err := store.GetInfo(itemPath)
			if err != nil {
				t.Fatal(err)
			}
			if info.Shared != expected.Shared || info.IsDir != expected.IsDir || info.Sha1 != expected.Sha1 {
				t.Fatalf("info not equaled (%v) (%v)", expected, info)
			}
		}

		// del sharings
		for itemPath := range pathInfos {
			err = store.DelInfo(itemPath)
			if err != nil {
				t.Fatal(err)
			}
		}

		// get infos
		for itemPath := range pathInfos {
			_, err := store.GetInfo(itemPath)
			if !IsNotFound(err) {
				t.Fatal(err)
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
		testInfoMethods(t, store)
	})
}
