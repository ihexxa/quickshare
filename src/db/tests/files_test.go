package tests

import (
	"context"
	"database/sql"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ihexxa/quickshare/src/db"
	"github.com/ihexxa/quickshare/src/db/rdb/sqlite"
)

func TestFileStore(t *testing.T) {
	testSharingMethods := func(t *testing.T, store db.IDBQuickshare) {
		dirPaths := []string{"admin/path1", "admin/path1/path2"}
		var err error
		location := "admin"

		ctx := context.TODO()
		adminId := uint64(0)

		// add some of paths...
		err = store.AddFileInfo(ctx, adminId, "admin/path1", &db.FileInfo{
			// Shared:  false, // deprecated
			IsDir:   false,
			Size:    int64(0),
			ShareID: "",
			Sha1:    "",
		})
		if err != nil {
			t.Fatal(err)
		}

		// add sharings
		for _, dirPath := range dirPaths {
			err = store.AddSharing(ctx, adminId, dirPath)
			if err != nil {
				t.Fatal(err)
			}
		}

		// list sharings
		dirToID, err := store.ListSharingsByLocation(ctx, location)
		if err != nil {
			t.Fatal(err)
		} else if len(dirToID) != len(dirPaths) {
			t.Fatal("share size not match")
		}

		for _, sharingDir := range dirPaths {
			if _, ok := dirToID[sharingDir]; !ok {
				t.Fatalf("sharing(%s) not found", sharingDir)
			}
			mustTrue, err := store.IsSharing(ctx, sharingDir)
			if err != nil {
				t.Fatal(err)
			} else if !mustTrue {
				t.Fatalf("get sharing(%t) should exist", mustTrue)
			}

			info, err := store.GetFileInfo(ctx, sharingDir)
			if err != nil {
				t.Fatal(err)
			} else if len(info.ShareID) != 7 {
				t.Fatalf("incorrect ShareID %s", info.ShareID)
			}

			gotSharingDir, err := store.GetSharingDir(ctx, info.ShareID)
			if err != nil {
				t.Fatal(err)
			} else if sharingDir != gotSharingDir {
				t.Fatalf("sharing dir not consist: (%s) (%s)", sharingDir, gotSharingDir)
			}
		}

		// del sharings
		for _, dirPath := range dirPaths {
			err = store.DelSharing(ctx, adminId, dirPath)
			if err != nil {
				t.Fatal(err)
			}
		}

		// list sharings
		dirToIDAfterDel, err := store.ListSharingsByLocation(ctx, location)
		if err != nil {
			t.Fatal(err)
		} else if len(dirToIDAfterDel) != 0 {
			t.Fatalf("share size not match (%+v)", dirToIDAfterDel)
		}

		for _, dirPath := range dirPaths {
			if _, ok := dirToIDAfterDel[dirPath]; ok {
				t.Fatalf("sharing(%s) should not exist", dirPath)
			}
			shared, err := store.IsSharing(ctx, dirPath)
			if err != nil {
				t.Fatal(err)
			} else if shared {
				t.Fatalf("get sharing(%t) should not shared but exist", shared)
			}

			info, err := store.GetFileInfo(ctx, dirPath)
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

			_, err = store.GetSharingDir(ctx, originalShareID)
			if !errors.Is(err, db.ErrSharingNotFound) {
				t.Fatal("should return ErrSharingNotFound")
			}
		}
	}

	testFileInfoMethods := func(t *testing.T, store db.IDBQuickshare) {
		pathInfos := map[string]*db.FileInfo{
			"admin/origin/item1": &db.FileInfo{
				// Shared:  false, // deprecated
				IsDir:   false,
				Size:    int64(7),
				ShareID: "",
				Sha1:    "item1_sha",
			},
			"admin/origin/item2": &db.FileInfo{
				// Shared:  false, // deprecated
				IsDir:   false,
				Size:    int64(3),
				ShareID: "",
				Sha1:    "item2_sha",
			},
			"admin/origin/dir": &db.FileInfo{
				// Shared:  true, // deprecated
				IsDir:   true,
				Size:    int64(0),
				ShareID: "mockedShareID",
				Sha1:    "",
			},
		}
		var err error

		adminId := uint64(0)
		ctx := context.TODO()
		// add infos
		usedSpace := int64(0)
		itemPaths := []string{}
		for itemPath, info := range pathInfos {
			err = store.AddFileInfo(ctx, adminId, itemPath, info)
			if err != nil {
				t.Fatal(err)
			}
			usedSpace += info.Size
			itemPaths = append(itemPaths, itemPath)
		}

		// get infos
		for itemPath, expected := range pathInfos {
			info, err := store.GetFileInfo(ctx, itemPath)
			if err != nil {
				t.Fatal(err)
			}
			if info.ShareID != expected.ShareID ||
				info.IsDir != expected.IsDir ||
				info.Sha1 != expected.Sha1 ||
				info.Size != expected.Size {
				t.Fatalf("info not equaled (%v) (%v)", expected, info)
			}
		}

		// list infos
		pathToInfo, err := store.ListFileInfos(ctx, itemPaths)
		if err != nil {
			t.Fatal(err)
		} else if len(pathToInfo) != len(pathInfos) {
			t.Fatalf("list result size not match (%v) (%d)", pathToInfo, len(pathInfos))
		}
		for pathname, info := range pathInfos {
			gotInfo, ok := pathToInfo[pathname]
			if !ok {
				t.Fatalf("path not found (%s)", pathname)
			}
			if info.ShareID != gotInfo.ShareID ||
				info.IsDir != gotInfo.IsDir ||
				info.Sha1 != gotInfo.Sha1 ||
				info.Size != gotInfo.Size {
				t.Fatalf("info not equaled (%v) (%v)", gotInfo, info)
			}
		}

		// set sha1
		testSha1 := "sha1"
		for itemPath := range pathInfos {
			err := store.SetSha1(ctx, itemPath, testSha1)
			if err != nil {
				t.Fatal(err)
			}
			info, err := store.GetFileInfo(ctx, itemPath)
			if err != nil {
				t.Fatal(err)
			}
			if info.Sha1 != testSha1 {
				t.Fatalf("sha not equaled (%v) (%v)", info.Sha1, testSha1)
			}
		}

		// move paths
		newPaths := []string{}
		for itemPath, info := range pathInfos {
			newItemPath := strings.ReplaceAll(itemPath, "origin", "new")
			err = store.MoveFileInfo(ctx, adminId, itemPath, newItemPath, info.IsDir)
			if err != nil {
				t.Fatal(err)
			}
			newPaths = append(newPaths, newItemPath)
		}

		// list infos
		pathToInfo, err = store.ListFileInfos(ctx, newPaths)
		if err != nil {
			t.Fatal(err)
		} else if len(pathToInfo) != len(pathInfos) {
			t.Fatalf("list result size not match (%v) (%d)", pathToInfo, len(pathInfos))
		}

		// check used space
		adminInfo, err := store.GetUser(ctx, adminId)
		if err != nil {
			t.Fatal(err)
		} else if adminInfo.UsedSpace != usedSpace {
			t.Fatalf("used space not match (%d) (%d)", adminInfo.UsedSpace, usedSpace)
		}

		// del info
		for _, itemPath := range newPaths {
			err = store.DelFileInfo(ctx, adminId, itemPath)
			if err != nil {
				t.Fatal(err)
			}
		}

		// check used space
		adminInfo, err = store.GetUser(ctx, adminId)
		if err != nil {
			t.Fatal(err)
		} else if adminInfo.UsedSpace != int64(0) {
			t.Fatalf("used space not match (%d) (%d)", adminInfo.UsedSpace, int64(0))
		}

		// list infos
		pathToInfo, err = store.ListFileInfos(ctx, itemPaths)
		if err != nil {
			t.Fatal(err)
		} else if len(pathToInfo) != 0 {
			t.Fatalf("list result should be empty (%v)", pathToInfo)
		}

		for itemPath := range pathInfos {
			_, err := store.GetFileInfo(ctx, itemPath)
			if !errors.Is(err, db.ErrFileInfoNotFound) {
				t.Fatal(err)
			}
		}
	}

	testUploadingMethods := func(t *testing.T, store db.IDBQuickshare) {
		pathInfos := map[string]*db.FileInfo{
			"admin/origin/item1": &db.FileInfo{
				// Shared:  false, // deprecated
				IsDir:   false,
				Size:    int64(7),
				ShareID: "",
				Sha1:    "",
			},
			"admin/origin/item2": &db.FileInfo{
				// Shared:  false, // deprecated
				IsDir:   false,
				Size:    int64(3),
				ShareID: "",
				Sha1:    "",
			},
			"admin/origin/to_delete/item3": &db.FileInfo{
				// Shared:  false, // deprecated
				IsDir:   false,
				Size:    int64(11),
				ShareID: "",
				Sha1:    "",
			},
		}
		var err error

		adminId := uint64(0)
		ctx := context.TODO()

		// add infos
		usedSpace := int64(0)
		usedSpaceAfterDeleting := int64(0)
		itemPaths := []string{}
		pathToTmpPath := map[string]string{}
		for itemPath, info := range pathInfos {
			tmpPath := strings.ReplaceAll(itemPath, "origin", "uploads")
			pathToTmpPath[itemPath] = tmpPath
			err = store.AddUploadInfos(ctx, adminId, tmpPath, itemPath, info)
			if err != nil {
				t.Fatal(err)
			}
			usedSpace += info.Size
			if !strings.Contains(itemPath, "delete") {
				usedSpaceAfterDeleting += info.Size
			}
			itemPaths = append(itemPaths, itemPath)
		}

		// get infos
		for itemPath, info := range pathInfos {
			gotPath, size, uploaded, err := store.GetUploadInfo(ctx, adminId, itemPath)
			if err != nil {
				t.Fatal(err)
			}
			if size != info.Size ||
				uploaded != int64(0) ||
				gotPath != itemPath {
				t.Fatalf("info not equaled (%v)", info)
			}
		}

		// list infos
		uploadInfos, err := store.ListUploadInfos(ctx, adminId)
		if err != nil {
			t.Fatal(err)
		} else if len(uploadInfos) != len(pathInfos) {
			t.Fatalf("list result size not match (%v) (%d)", uploadInfos, len(pathInfos))
		}
		for _, uploadInfo := range uploadInfos {
			expected, ok := pathInfos[uploadInfo.RealFilePath]
			if !ok {
				t.Fatalf("path not found (%s)", uploadInfo.RealFilePath)
			}
			if uploadInfo.Uploaded != int64(0) ||
				expected.Size != uploadInfo.Size {
				t.Fatalf("info not equaled (%d) (%d)", uploadInfo.Uploaded, uploadInfo.Size)
			}
		}

		// check used space
		adminInfo, err := store.GetUser(ctx, adminId)
		if err != nil {
			t.Fatal(err)
		} else if adminInfo.UsedSpace != usedSpace {
			t.Fatalf("used space not match (%d) (%d)", adminInfo.UsedSpace, usedSpace)
		}

		// set uploading
		for itemPath, info := range pathInfos {
			err := store.SetUploadInfo(ctx, adminId, itemPath, int64(info.Size/2))
			if err != nil {
				t.Fatal(err)
			}
			gotPath, size, uploaded, err := store.GetUploadInfo(ctx, adminId, itemPath)
			if err != nil {
				t.Fatal(err)
			}
			if gotPath != itemPath || size != info.Size || uploaded != info.Size/2 {
				t.Fatal("uploaded info not match")
			}
		}

		// check used space
		adminInfo, err = store.GetUser(ctx, adminId)
		if err != nil {
			t.Fatal(err)
		} else if adminInfo.UsedSpace != usedSpace {
			t.Fatalf("used space not match (%d) (%d)", adminInfo.UsedSpace, usedSpace)
		}

		// del info
		for itemPath := range pathInfos {
			err = store.DelUploadingInfos(ctx, adminId, itemPath)
			if err != nil {
				t.Fatal(err)
			}
		}

		// check used space
		adminInfo, err = store.GetUser(ctx, adminId)
		if err != nil {
			t.Fatal(err)
		} else if adminInfo.UsedSpace != int64(0) {
			t.Fatalf("used space not match (%d) (%d)", adminInfo.UsedSpace, int64(0))
		}

		// list infos
		uploadInfos, err = store.ListUploadInfos(ctx, adminId)
		if err != nil {
			t.Fatal(err)
		} else if len(uploadInfos) != 0 {
			t.Fatalf("list result size not match (%v) (%d)", uploadInfos, 0)
		}

		for itemPath := range pathInfos {
			_, _, _, err := store.GetUploadInfo(ctx, adminId, itemPath)
			if !errors.Is(err, sql.ErrNoRows) {
				t.Fatal(err)
			}
		}
	}

	t.Run("testing file info store - sqlite", func(t *testing.T) {
		rootPath, err := ioutil.TempDir("./", "qs_sqlite_test_")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(rootPath)

		dbPath := filepath.Join(rootPath, "files.sqlite")
		sqliteDB, err := sqlite.NewSQLite(dbPath)
		if err != nil {
			t.Fatal(err)
		}
		defer sqliteDB.Close()

		store, err := sqlite.NewSQLiteStore(sqliteDB)
		if err != nil {
			t.Fatal("fail to new sqlite store", err)
		}
		err = store.Init(context.TODO(), "admin", "1234", testSiteConfig)
		if err != nil {
			t.Fatal("fail to init", err)
		}

		testSharingMethods(t, store)
		testFileInfoMethods(t, store)
		testUploadingMethods(t, store)
	})
}
