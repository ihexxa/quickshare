package server

import (
	"crypto/sha1"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/ihexxa/gocfg"

	"github.com/ihexxa/quickshare/src/client"
	"github.com/ihexxa/quickshare/src/handlers/fileshdr"
)

func startTestServer(config string) *Server {
	cfg, err := gocfg.New(NewDefaultConfig()).
		Load(gocfg.JSONStr(config))
	if err != nil {
		panic(err)
	}

	srv, err := NewServer(cfg)
	if err != nil {
		panic(err)
	}

	go srv.Start()
	return srv
}

func TestFileHandlers(t *testing.T) {
	addr := "http://127.0.0.1:8888"
	root := "./testData"
	chunkSize := 2
	config := `{
		"Server": {
			"Debug": true
		},
		"FS": {
			"Root": "./testData"
		}
	}`

	srv := startTestServer(config)
	defer srv.Shutdown()
	// kv := srv.depsKVStore()
	fs := srv.depsFS()
	defer os.RemoveAll(root)
	cl := client.NewQSClient(addr)

	// TODO: remove this
	time.Sleep(500)

	t.Run("test file APIs: Create-UploadChunk-UploadStatus-Metadata-Delete", func(t *testing.T) {
		for filePath, content := range map[string]string{
			"path1/f1.md":       "11111",
			"path1/path2/f2.md": "101010",
		} {
			fileSize := int64(len([]byte(content)))
			// create a file
			res, _, errs := cl.Create(filePath, fileSize)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if res.StatusCode != 200 {
				t.Fatal(res.StatusCode)
			}

			// check uploading file
			uploadFilePath := path.Join(fileshdr.UploadDir, fmt.Sprintf("%x", sha1.Sum([]byte(filePath))))
			info, err := fs.Stat(uploadFilePath)
			if err != nil {
				t.Fatal(err)
			} else if info.Name() != filepath.Base(uploadFilePath) {
				t.Fatal(info.Name(), filepath.Base(uploadFilePath))
			}

			// upload a chunk
			i := 0
			contentBytes := []byte(content)
			for i < len(contentBytes) {
				right := i + chunkSize
				if right > len(contentBytes) {
					right = len(contentBytes)
				}

				res, _, errs = cl.UploadChunk(filePath, string(contentBytes[i:right]), int64(i))
				i = right
				if len(errs) > 0 {
					t.Fatal(errs)
				} else if res.StatusCode != 200 {
					t.Fatal(res.StatusCode)
				}

				if int64(i) != fileSize {
					_, statusResp, errs := cl.UploadStatus(filePath)
					if len(errs) > 0 {
						t.Fatal(errs)
					} else if statusResp.Path != filePath ||
						statusResp.IsDir ||
						statusResp.FileSize != fileSize ||
						statusResp.Uploaded != int64(i) {
						t.Fatal("incorrect uploadinfo info", statusResp)
					}
				}
			}

			// check uploaded file
			fsFilePath := filepath.Join(fileshdr.FsDir, filePath)
			info, err = fs.Stat(fsFilePath)
			if err != nil {
				t.Fatal(err)
			} else if info.Name() != filepath.Base(fsFilePath) {
				t.Fatal(info.Name(), filepath.Base(fsFilePath))
			}

			// metadata
			_, mRes, errs := cl.Metadata(filePath)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if mRes.Name != info.Name() ||
				mRes.IsDir != info.IsDir() ||
				mRes.Size != info.Size() {
				// TODO: modTime is not checked
				t.Fatal("incorrect uploaded info", mRes)
			}

			// delete file
			res, _, errs = cl.Delete(filePath)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if res.StatusCode != 200 {
				t.Fatal(res.StatusCode)
			}
		}
	})

	t.Run("test file APIs: Mkdir-Create-UploadChunk-List", func(t *testing.T) {
		for dirPath, files := range map[string]map[string]string{
			"dir/path1/": map[string]string{
				"f1.md": "11111",
				"f2.md": "22222222222",
			},
			"dir/path1/path2": map[string]string{
				"f3.md": "3333333",
			},
		} {
			res, _, errs := cl.Mkdir(dirPath)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if res.StatusCode != 200 {
				t.Fatal(res.StatusCode)
			}

			for fileName, content := range files {
				filePath := filepath.Join(dirPath, fileName)

				fileSize := int64(len([]byte(content)))
				// create a file
				res, _, errs := cl.Create(filePath, fileSize)
				if len(errs) > 0 {
					t.Fatal(errs)
				} else if res.StatusCode != 200 {
					t.Fatal(res.StatusCode)
				}

				res, _, errs = cl.UploadChunk(filePath, content, 0)
				if len(errs) > 0 {
					t.Fatal(errs)
				} else if res.StatusCode != 200 {
					t.Fatal(res.StatusCode)
				}
			}

			_, lResp, errs := cl.List(dirPath)
			if len(errs) > 0 {
				t.Fatal(errs)
			}
			for _, metadata := range lResp.Metadatas {
				content, ok := files[metadata.Name]
				if !ok {
					t.Fatalf("%s not found", metadata.Name)
				} else if int64(len(content)) != metadata.Size {
					t.Fatalf("size not match %d %d \n", len(content), metadata.Size)
				}
			}
		}
	})

	t.Run("test file APIs: Mkdir-Create-UploadChunk-Move-List", func(t *testing.T) {
		srcDir := "move/src"
		dstDir := "move/dst"

		for _, dirPath := range []string{srcDir, dstDir} {
			res, _, errs := cl.Mkdir(dirPath)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if res.StatusCode != 200 {
				t.Fatal(res.StatusCode)
			}
		}

		files := map[string]string{
			"f1.md": "111",
			"f2.md": "22222",
		}

		for fileName, content := range files {
			oldPath := filepath.Join(srcDir, fileName)
			newPath := filepath.Join(dstDir, fileName)
			fileSize := int64(len([]byte(content)))

			// create a file
			res, _, errs := cl.Create(oldPath, fileSize)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if res.StatusCode != 200 {
				t.Fatal(res.StatusCode)
			}

			res, _, errs = cl.UploadChunk(oldPath, content, 0)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if res.StatusCode != 200 {
				t.Fatal(res.StatusCode)
			}

			res, _, errs = cl.Move(oldPath, newPath)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if res.StatusCode != 200 {
				t.Fatal(res.StatusCode)
			}
		}

		_, lResp, errs := cl.List(dstDir)
		if len(errs) > 0 {
			t.Fatal(errs)
		}
		for _, metadata := range lResp.Metadatas {
			content, ok := files[metadata.Name]
			if !ok {
				t.Fatalf("%s not found", metadata.Name)
			} else if int64(len(content)) != metadata.Size {
				t.Fatalf("size not match %d %d \n", len(content), metadata.Size)
			}
		}
	})
}
