package apis

import (
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

import (
	"github.com/ihexxa/quickshare/server/libs/httputil"
	"github.com/ihexxa/quickshare/server/libs/httpworker"
)

func (srv *SrvShare) ClientHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		srv.Http.Fill(httputil.Err404, res)
		return
	}

	ack := make(chan error, 1)
	ok := srv.WorkerPool.Put(&httpworker.Task{
		Ack: ack,
		Do:  srv.Wrap(srv.GetClient),
		Res: res,
		Req: req,
	})
	if !ok {
		srv.Http.Fill(httputil.Err503, res)
		return
	}

	execErr := srv.WorkerPool.IsInTime(ack, time.Duration(srv.Conf.Timeout)*time.Millisecond)
	if srv.Err.IsErr(execErr) {
		srv.Http.Fill(httputil.Err500, res)
	}
}

func (srv *SrvShare) GetClient(res http.ResponseWriter, req *http.Request) interface{} {
	if !srv.Walls.PassIpLimit(GetRemoteIp(req.RemoteAddr)) {
		return httputil.Err504
	}

	return srv.getClient(res, req, req.URL.EscapedPath())
}

func (srv *SrvShare) getClient(res http.ResponseWriter, req *http.Request, relPath string) interface{} {
	if strings.HasSuffix(relPath, "/") {
		relPath = relPath + "index.html"
	}
	if !IsValidClientPath(relPath) {
		return httputil.Err400
	}

	fullPath := filepath.Clean(filepath.Join("./public", relPath))
	http.ServeFile(res, req, fullPath)
	return 0
}

func IsValidClientPath(fullPath string) bool {
	if strings.Contains(fullPath, "..") {
		return false
	}

	return true
}
