package server

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/ihexxa/quickshare/src/client"
	"github.com/ihexxa/quickshare/src/handlers/fileshdr"
)

func TestFileHandlers(t *testing.T) {
	addr := "http://127.0.0.1:8686"
	root := "testData"
	config := `{
		"users": {
			"enableAuth": false
		},
		"server": {
			"debug": true
		},
		"fs": {
			"root": "testData"
		}
	}`

	os.RemoveAll(root)
	err := os.MkdirAll(root, 0700)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)

	srv := startTestServer(config)
	defer srv.Shutdown()
	fs := srv.depsFS()
	cl := client.NewFilesClient(addr)

	if !waitForReady(addr) {
		t.Fatal("fail to start server")
	}

	assertUploadOK := func(t *testing.T, filePath, content string) bool {
		cl := client.NewFilesClient(addr)

		fileSize := int64(len([]byte(content)))
		res, _, errs := cl.Create(filePath, fileSize)
		if len(errs) > 0 {
			t.Error(errs)
			return false
		} else if res.StatusCode != 200 {
			t.Error(res.StatusCode)
			return false
		}

		base64Content := base64.StdEncoding.EncodeToString([]byte(content))
		res, _, errs = cl.UploadChunk(filePath, base64Content, 0)
		if len(errs) > 0 {
			t.Error(errs)
			return false
		} else if res.StatusCode != 200 {
			t.Error(res.StatusCode)
			return false
		}

		return true
	}

	assetDownloadOK := func(t *testing.T, filePath, content string) bool {
		var (
			res      *http.Response
			body     string
			errs     []error
			fileSize = int64(len([]byte(content)))
		)

		cl := client.NewFilesClient(addr)

		rd := rand.Intn(3)
		switch rd {
		case 0:
			res, body, errs = cl.Download(filePath, map[string]string{})
		case 1:
			res, body, errs = cl.Download(filePath, map[string]string{
				"Range": fmt.Sprintf("bytes=0-%d", fileSize-1),
			})
		case 2:
			res, body, errs = cl.Download(filePath, map[string]string{
				"Range": fmt.Sprintf("bytes=0-%d, %d-%d", (fileSize-1)/2, (fileSize-1)/2+1, fileSize-1),
			})
		}

		if len(errs) > 0 {
			t.Error(errs)
			return false
		} else if res.StatusCode != 200 && res.StatusCode != 206 {
			t.Error(res.StatusCode)
			return false
		}
		switch rd {
		case 0:
			if body != content {
				t.Errorf("body not equal got(%s) expect(%s)\n", body, content)
				return false
			}
		case 1:

			if body[2:] != content { // body returned by gorequest contains the first CRLF
				t.Errorf("body not equal got(%s) expect(%s)\n", body[2:], content)
				return false
			}
		default:
			body = body[2:] // body returned by gorequest contains the first CRLF
			realBody := ""
			boundaryEnd := strings.Index(body, "\r\n")
			boundary := body[0:boundaryEnd]
			bodyParts := strings.Split(body, boundary)

			for i, bodyPart := range bodyParts {
				if i == 0 || i == len(bodyParts)-1 {
					continue
				}
				start := strings.Index(bodyPart, "\r\n\r\n")

				fmt.Printf("<%s>", bodyPart[start+4:len(bodyPart)-2]) // ignore the last CRLF
				realBody += bodyPart[start+4 : len(bodyPart)-2]
			}
			if realBody != content {
				t.Errorf("multi body not equal got(%s) expect(%s)\n", realBody, content)
				return false
			}
		}

		return true
	}

	t.Run("test files APIs: Create-UploadChunk-UploadStatus-Metadata-Delete", func(t *testing.T) {
		for filePath, content := range map[string]string{
			"path1/f1.md":       "1111 1111 1111 1111",
			"path1/path2/f2.md": "1010 1010 1111 0000 0010",
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
				right := i + rand.Intn(3) + 1
				if right > len(contentBytes) {
					right = len(contentBytes)
				}

				chunk := contentBytes[i:right]
				chunkBase64 := base64.StdEncoding.EncodeToString(chunk)
				res, _, errs = cl.UploadChunk(filePath, chunkBase64, int64(i))
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

			err = fs.Sync()
			if err != nil {
				t.Fatal(err)
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

	t.Run("test dirs APIs: Mkdir-Create-UploadChunk-List", func(t *testing.T) {
		for dirPath, files := range map[string]map[string]string{
			"dir/path1": map[string]string{
				"f1.md": "11111",
				"f2.md": "22222222222",
			},
			"dir/path2/path2": map[string]string{
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
				assertUploadOK(t, filePath, content)
			}

			err = fs.Sync()
			if err != nil {
				t.Fatal(err)
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

	t.Run("test operation APIs: Mkdir-Create-UploadChunk-Move-List", func(t *testing.T) {
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
			// fileSize := int64(len([]byte(content)))
			assertUploadOK(t, oldPath, content)

			res, _, errs := cl.Move(oldPath, newPath)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if res.StatusCode != 200 {
				t.Fatal(res.StatusCode)
			}
		}

		err = fs.Sync()
		if err != nil {
			t.Fatal(err)
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

	t.Run("test download APIs: Download(normal, ranges)", func(t *testing.T) {
		for filePath, content := range map[string]string{
			"download/path1/f1":    "123456",
			"download/path1/path2": "12345678",
		} {
			assertUploadOK(t, filePath, content)

			err = fs.Sync()
			if err != nil {
				t.Fatal(err)
			}

			assetDownloadOK(t, filePath, content)
		}
	})

	t.Run("test concurrently uploading & downloading", func(t *testing.T) {
		type mockFile struct {
			FilePath string
			Content  string
		}
		wg := &sync.WaitGroup{}

		startClient := func(files []*mockFile) {
			for i := 0; i < 5; i++ {
				for _, file := range files {
					if !assertUploadOK(t, fmt.Sprintf("%s_%d", file.FilePath, i), file.Content) {
						break
					}

					err = fs.Sync()
					if err != nil {
						t.Fatal(err)
					}

					if !assetDownloadOK(t, fmt.Sprintf("%s_%d", file.FilePath, i), file.Content) {
						break
					}
				}
			}

			wg.Done()
		}

		for _, clientFiles := range [][]*mockFile{
			[]*mockFile{
				&mockFile{"concurrent/p0/f0", "00"},
				&mockFile{"concurrent/f0.md", "0000 0000 0000 0"},
			},
			[]*mockFile{
				&mockFile{"concurrent/p1/f1", "11"},
				&mockFile{"concurrent/f1.md", "1111 1111 1"},
			},
			[]*mockFile{
				&mockFile{"concurrent/p2/f2", "22"},
				&mockFile{"concurrent/f2.md", "222"},
			},
		} {
			wg.Add(1)
			go startClient(clientFiles)
		}

		wg.Wait()
	})

	t.Run("test uploading APIs: Create, ListUploadings, DelUploading)", func(t *testing.T) {
		files := map[string]string{
			"uploadings/path1/f1":    "123456",
			"uploadings/path1/path2": "12345678",
		}

		for filePath, content := range files {
			fileSize := int64(len([]byte(content)))
			res, _, errs := cl.Create(filePath, fileSize)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if res.StatusCode != 200 {
				t.Fatal(res.StatusCode)
			}
		}

		res, lResp, errs := cl.ListUploadings()
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if res.StatusCode != 200 {
			t.Fatal(res.StatusCode)
		}

		gotInfos := map[string]*fileshdr.UploadInfo{}
		for _, info := range lResp.UploadInfos {
			gotInfos[info.RealFilePath] = info
		}
		for filePath, content := range files {
			info, ok := gotInfos[filePath]
			if !ok {
				t.Fatalf("uploading(%s) not found", filePath)
			} else if info.Uploaded != 0 {
				t.Fatalf("uploading(%s) uploaded is not correct", filePath)
			} else if info.Size != int64(len([]byte(content))) {
				t.Fatalf("uploading(%s) size is not correct", filePath)
			}
		}

		for filePath := range files {
			res, _, errs := cl.DelUploading(filePath)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if res.StatusCode != 200 {
				t.Fatal(res.StatusCode)
			}
		}

		res, lResp, errs = cl.ListUploadings()
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if res.StatusCode != 200 {
			t.Fatal(res.StatusCode)
		} else if len(lResp.UploadInfos) != 0 {
			t.Fatalf("info is not deleted, info len(%d)", len(lResp.UploadInfos))
		}
	})
}
