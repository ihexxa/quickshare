package qtube

import (
	"net/http"
)

import (
	"github.com/ihexxa/quickshare/server/libs/fileidx"
)

type Downloader interface {
	ServeFile(res http.ResponseWriter, req *http.Request, fileInfo *fileidx.FileInfo) error
}
