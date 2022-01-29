package test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/ihexxa/quickshare/src/kvstore"
	"github.com/ihexxa/quickshare/src/kvstore/boltdbpvd"
	// "github.com/ihexxa/quickshare/src/kvstore/memstore"
)

func TestKVStoreProviders(t *testing.T) {
	var err error
	var ok bool
	key, boolV, intV, int64V, floatV, stringV := "key", true, 2027, int64(2027), 3.1415, "foobar"
	key2, boolV2 := "key2", false

	kvstoreTest := func(store kvstore.IKVStore, t *testing.T) {
		// test bools
		_, ok = store.GetBool(key)
		if ok {
			t.Error("value should not exist")
		}
		err = store.SetBool(key, boolV)
		if err != nil {
			t.Errorf("there should be no error %v", err)
		}
		err = store.SetBool(key2, boolV2)
		if err != nil {
			t.Errorf("there should be no error %v", err)
		}
		boolList, err := store.ListBools()
		if err != nil {
			t.Errorf("there should be no error %v", err)
		}
		if boolList[key] != boolV {
			t.Error("listBool incorrect val1")
		} else if boolList[key2] != boolV2 {
			t.Error("listBool incorrect val2")
		}
		boolVGot, ok := store.GetBool(key)
		if !ok {
			t.Error("value should exit")
		} else if boolVGot != boolV {
			t.Error(fmt.Sprintln("value not equal", boolVGot, boolV))
		}
		err = store.DelBool(key)
		if err != nil {
			t.Errorf("there should be no error %v", err)
		}
		_, ok = store.GetBool(key)
		if ok {
			t.Error("value should not exist")
		}

		// test ints
		_, ok = store.GetInt(key)
		if ok {
			t.Error("value should not exist")
		}
		err = store.SetInt(key, intV)
		if err != nil {
			t.Errorf("there should be no error %v", err)
		}
		intVGot, ok := store.GetInt(key)
		if !ok {
			t.Error("value should exit")
		} else if intVGot != intV {
			t.Error(fmt.Sprintln("value not equal", intVGot, intV))
		}
		err = store.DelInt(key)
		if err != nil {
			t.Errorf("there should be no error %v", err)
		}
		_, ok = store.GetInt(key)
		if ok {
			t.Error("value should not exist")
		}

		// test int64s
		_, ok = store.GetInt64(key)
		if ok {
			t.Error("value should not exist")
		}
		err = store.SetInt64(key, int64V)
		if err != nil {
			t.Errorf("there should be no error %v", err)
		}
		int64VGot, ok := store.GetInt64(key)
		if !ok {
			t.Error("value should exit")
		} else if int64VGot != int64V {
			t.Error(fmt.Sprintln("value not equal", int64VGot, int64V))
		}
		err = store.DelInt64(key)
		if err != nil {
			t.Errorf("there should be no error %v", err)
		}
		_, ok = store.GetInt64(key)
		if ok {
			t.Error("value should not exist")
		}

		// test floats
		_, ok = store.GetFloat(key)
		if ok {
			t.Error("value should not exist")
		}
		err = store.SetFloat(key, floatV)
		if err != nil {
			t.Errorf("there should be no error %v", err)
		}
		floatVGot, ok := store.GetFloat(key)
		if !ok {
			t.Error("value should exit")
		} else if floatVGot != floatV {
			t.Error(fmt.Sprintln("value not equal", floatVGot, floatV))
		}
		err = store.DelFloat(key)
		if err != nil {
			t.Errorf("there should be no error %v", err)
		}
		_, ok = store.GetFloat(key)
		if ok {
			t.Error("value should not exist")
		}

		// test strings
		_, ok = store.GetString(key)
		if ok {
			t.Error("value should not exist")
		}
		err = store.SetString(key, stringV)
		if err != nil {
			t.Errorf("there should be no error %v", err)
		}
		stringVGot, ok := store.GetString(key)
		if !ok {
			t.Error("value should exit")
		} else if stringVGot != stringV {
			t.Error(fmt.Sprintln("value not equal", stringVGot, stringV))
		}
		err = store.DelString(key)
		if err != nil {
			t.Errorf("there should be no error %v", err)
		}
		_, ok = store.GetString(key)
		if ok {
			t.Error("value should not exist")
		}

		// test strings in ns
		ns := "str_namespace"
		err = store.AddNamespace(ns)
		if err != nil {
			t.Errorf("there should be no error %v", err)
		}
		_, ok = store.GetStringIn(ns, key)
		if ok {
			t.Error("value should not exist")
		}
		err = store.SetStringIn(ns, key, stringV)
		if err != nil {
			t.Errorf("there should be no error %v", err)
		}
		stringVGot, ok = store.GetStringIn(ns, key)
		if !ok {
			t.Error("value should exit")
		} else if stringVGot != stringV {
			t.Error(fmt.Sprintln("value not equal", stringVGot, stringV))
		}
		err = store.SetStringIn(ns, key2, stringV)
		if err != nil {
			t.Errorf("there should be no error %v", err)
		}
		kvs, err := store.ListStringsByPrefixIn("key", ns)
		if err != nil {
			t.Errorf("there should be no error %v", err)
		}
		if kvs[key] != stringV {
			t.Errorf("list str key not found")
		}
		if kvs[key2] != stringV {
			t.Errorf("list str key not found")
		}

		err = store.DelStringIn(ns, key)
		if err != nil {
			t.Errorf("there should be no error %v", err)
		}
		_, ok = store.GetStringIn(ns, key)
		if ok {
			t.Error("value should not exist")
		}

		// test locks
		err = store.TryLock(key)
		if err != nil {
			t.Errorf("there should be no error %v", err)
		}
		err = store.TryLock(key)
		if err == nil || err != kvstore.ErrLocked {
			t.Error("there should be locked")
		}
		err = store.TryLock("key2")
		if err != nil {
			t.Errorf("there should be no error %v", err)
		}
		err = store.Unlock(key)
		if err != nil {
			t.Errorf("there should be no error %v", err)
		}
		err = store.Unlock("key2")
		if err != nil {
			t.Errorf("there should be no error %v", err)
		}
	}

	t.Run("test bolt provider", func(t *testing.T) {
		rootPath, err := ioutil.TempDir("./", "quickshare_kvstore_test_")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(rootPath)

		dbPath := filepath.Join(rootPath, "quickshare.db")
		store := boltdbpvd.New(dbPath, 1024)
		defer store.Close()
		kvstoreTest(store, t)
	})

	// t.Run("test in-memory provider", func(t *testing.T) {
	// 	rootPath, err := ioutil.TempDir("./", "quickshare_kvstore_test_")
	// 	if err != nil {
	// 		t.Fatal(err)
	// 	}
	// 	defer os.RemoveAll(rootPath)

	// 	store := memstore.New()
	// 	kvstoreTest(store, t)
	// })
}
