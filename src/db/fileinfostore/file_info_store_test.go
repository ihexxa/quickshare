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
		dirToID, err := store.ListSharings(prefix)
		if err != nil {
			t.Fatal(err)
		}
		for _, sharingDir := range dirPaths {
			if _, ok := dirToID[sharingDir]; !ok {
				t.Fatalf("sharing(%s) not found", sharingDir)
			}
			mustTrue, exist := store.GetSharing(sharingDir)
			if !mustTrue || !exist {
				t.Fatalf("get sharing(%t %t) should exist", mustTrue, exist)
			}

			info, err := store.GetInfo(sharingDir)
			if err != nil {
				t.Fatal(err)
			} else if len(info.ShareID) != 7 {
				t.Fatalf("incorrect ShareID %s", info.ShareID)
			}

			gotSharingDir, err := store.GetSharingDir(info.ShareID)
			if err != nil {
				t.Fatal(err)
			} else if sharingDir != gotSharingDir {
				t.Fatalf("sharing dir not consist: (%s) (%s)", sharingDir, gotSharingDir)
			}
		}

		// del sharings
		for _, dirPath := range dirPaths {
			err = store.DelSharing(dirPath)
			if err != nil {
				t.Fatal(err)
			}
		}

		dirToIDAfterDel, err := store.ListSharings(prefix)
		if err != nil {
			t.Fatal(err)
		}
		for _, dirPath := range dirPaths {
			if _, ok := dirToIDAfterDel[dirPath]; ok {
				t.Fatalf("sharing(%s) should not exist", dirPath)
			}
			shared, exist := store.GetSharing(dirPath)
			if shared {
				t.Fatalf("get sharing(%t, %t) should not shared but exist", shared, exist)
			}

			info, err := store.GetInfo(dirPath)
			if err != nil {
				t.Fatal(err)
			} else if len(info.ShareID) != 0 {
				t.Fatalf("ShareID should be empty %s", info.ShareID)
			}

			// shareIDs are removed, use original dirToID to get shareID
			originalShareID, ok := dirToID[dirPath]
			if !ok {
				t.Fatalf("dir (%s) should exist in originalShareID", dirPath)
			}

			_, err = store.GetSharingDir(originalShareID)
			if err != ErrSharingNotFound {
				t.Fatal("should return ErrSharingNotFound")
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
