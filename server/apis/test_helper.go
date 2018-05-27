package apis

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

import (
	"quickshare/server/libs/fileidx"
	"quickshare/server/libs/qtube"
)

type stubFsUtil struct {
	MockLocalFileInfos []*fileidx.FileInfo
	MockFile           *qtube.StubFile
}

var expectCreateFileName = ""

func (fs *stubFsUtil) CreateFile(fileName string) error {
	if fileName != expectCreateFileName {
		panic(
			fmt.Sprintf("CreateFile: got: %s expect: %s", fileName, expectCreateFileName),
		)
	}
	return nil
}

func (fs *stubFsUtil) CopyChunkN(fullPath string, chunk io.Reader, start int64, len int64) bool {
	return true
}

func (fs *stubFsUtil) ServeFile(res http.ResponseWriter, req *http.Request, fileName string) {
	return
}

func (fs *stubFsUtil) DelFile(fullPath string) bool {
	return true
}

func (fs *stubFsUtil) MkdirAll(path string, mode os.FileMode) bool {
	return true
}

func (fs *stubFsUtil) Readdir(dirname string, n int) ([]*fileidx.FileInfo, error) {
	return fs.MockLocalFileInfos, nil
}

func (fs *stubFsUtil) Open(filePath string) (qtube.ReadSeekCloser, error) {
	return fs.MockFile, nil
}

type stubWriter struct {
	Headers    http.Header
	Response   []byte
	StatusCode int
}

func (w *stubWriter) Header() http.Header {
	return w.Headers
}

func (w *stubWriter) Write(body []byte) (int, error) {
	w.Response = append(w.Response, body...)
	return len(body), nil
}

func (w *stubWriter) WriteHeader(statusCode int) {
	w.StatusCode = statusCode
}

type stubDownloader struct {
	Content string
}

func (d stubDownloader) ServeFile(w http.ResponseWriter, r *http.Request, fileInfo *fileidx.FileInfo) error {
	_, err := w.Write([]byte(d.Content))
	return err
}

func sameInfoWithoutTime(info1, info2 *fileidx.FileInfo) bool {
	return info1.Id == info2.Id &&
		info1.DownLimit == info2.DownLimit &&
		info1.PathLocal == info2.PathLocal &&
		info1.State == info2.State &&
		info1.Uploaded == info2.Uploaded
}

func sameMap(map1, map2 map[string]*fileidx.FileInfo) bool {
	for key, info1 := range map1 {
		info2, found := map2[key]
		if !found || !sameInfoWithoutTime(info1, info2) {
			fmt.Printf("infos are not same: \n%v \n%v", info1, info2)
			return false
		}
	}

	for key, info2 := range map2 {
		info1, found := map1[key]
		if !found || !sameInfoWithoutTime(info1, info2) {
			fmt.Printf("infos are not same: \n%v \n%v", info1, info2)
			return false
		}
	}

	return true
}

type stubEncryptor struct {
	MockResult string
}

func (enc *stubEncryptor) Encrypt(content []byte) string {
	return enc.MockResult
}
