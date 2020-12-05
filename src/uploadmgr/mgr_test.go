package uploadmgr

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sync"
	"testing"

	"github.com/ihexxa/gocfg"
	"github.com/ihexxa/quickshare/src/depidx"
	"github.com/ihexxa/quickshare/src/fs"
	"github.com/ihexxa/quickshare/src/fs/local"
	"github.com/ihexxa/quickshare/src/kvstore/memstore"
	"github.com/ihexxa/quickshare/src/server"
)

var debug = flag.Bool("d", false, "debug mode")

// TODO: teardown after each test case

func TestUploadMgr(t *testing.T) {
	rootPath, err := ioutil.TempDir("./", "quickshare_test_")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(rootPath)

	newTestUploadMgr := func() (*UploadMgr, fs.ISimpleFS) {
		cfg := gocfg.New()
		err := cfg.Load(gocfg.JSONStr("{}"), server.NewEmptyConfig())
		if err != nil {
			t.Fatal(err)
		}

		filesystem := local.NewLocalFS(rootPath, 0660, 32, 10)
		kvstore := memstore.New()

		deps := depidx.NewDeps(cfg)
		deps.SetFS(filesystem)
		deps.SetKV(kvstore)

		mgr, err := NewUploadMgr(deps)
		if err != nil {
			t.Fatal(err)
		}
		return mgr, filesystem
	}

	t.Run("normal upload", func(t *testing.T) {
		mgr, _ := newTestUploadMgr()
		defer mgr.Close()

		testCases := map[string]string{
			"foo.md":          "",
			"bar.md":          "1",
			"path1/foobar.md": "1110011",
		}

		for filePath, content := range testCases {
			err = mgr.Create(filePath, int64(len([]byte(content))))
			if err != nil {
				t.Fatal(err)
			}

			bytes := []byte(content)
			for i := 0; i < len(bytes); i++ {
				wrote, err := mgr.WriteChunk(filePath, bytes[i:i+1], int64(i))
				if err != nil {
					t.Fatal(err)
				}
				if wrote != 1 {
					t.Fatalf("wrote(%d) != 1", wrote)
				}
			}

			if err = mgr.Sync(); err != nil {
				t.Fatal(err)
			}

			gotBytes, err := ioutil.ReadFile(path.Join(rootPath, filePath))
			if err != nil {
				t.Fatal(err)
			}
			if string(gotBytes) != content {
				t.Errorf("content not same expected(%s) got(%s)", content, string(gotBytes))
			}
		}
	})

	t.Run("concurrently upload", func(t *testing.T) {
		mgr, _ := newTestUploadMgr()
		defer mgr.Close()

		testCases := []map[string]string{
			map[string]string{
				"file20.md":       "111",
				"file21.md":       "2222000",
				"path1/file22.md": "1010011",
				"path2/file22.md": "1010011",
			},
		}

		uploadWorker := func(id int, filePath, content string, wg *sync.WaitGroup) {
			err = mgr.Create(filePath, int64(len([]byte(content))))
			if err != nil {
				t.Fatal(err)
			}

			bytes := []byte(content)
			for i := 0; i < len(bytes); i++ {
				wrote, err := mgr.WriteChunk(filePath, bytes[i:i+1], int64(i))
				if err != nil {
					t.Fatal(err)
				}
				if wrote != 1 {
					t.Fatalf("wrote(%d) != 1", wrote)
				}
				if *debug {
					fmt.Printf("worker-%d wrote %s\n", id, string(bytes[i:i+1]))
				}
			}

			wg.Done()
		}

		for _, files := range testCases {
			wg := &sync.WaitGroup{}
			workerID := 0
			for filePath, content := range files {
				wg.Add(1)
				go uploadWorker(workerID, filePath, content, wg)
				workerID++
			}

			wg.Wait()

			if err = mgr.Sync(); err != nil {
				t.Fatal(err)
			}

			for filePath, content := range files {
				gotBytes, err := ioutil.ReadFile(path.Join(rootPath, filePath))
				if err != nil {
					t.Fatal(err)
				}
				if string(gotBytes) != content {
					t.Errorf("content not same expected(%s) got(%s)", content, string(gotBytes))
				}
			}
		}
	})
}
