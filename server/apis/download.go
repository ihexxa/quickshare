package apis

import (
	"net/http"
	"time"
)

import (
	"github.com/ihexxa/quickshare/server/libs/fileidx"
	"github.com/ihexxa/quickshare/server/libs/httputil"
	"github.com/ihexxa/quickshare/server/libs/httpworker"
)

func (srv *SrvShare) DownloadHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		srv.Http.Fill(httputil.Err404, res)
	}

	ack := make(chan error, 1)
	ok := srv.WorkerPool.Put(&httpworker.Task{
		Ack: ack,
		Do:  srv.Wrap(srv.Download),
		Res: res,
		Req: req,
	})
	if !ok {
		srv.Http.Fill(httputil.Err503, res)
	}

	// using WriteTimeout instead of Timeout
	// After timeout, connection will be lost, and worker will fail to write and return
	execErr := srv.WorkerPool.IsInTime(ack, time.Duration(srv.Conf.WriteTimeout)*time.Millisecond)
	if srv.Err.IsErr(execErr) {
		srv.Http.Fill(httputil.Err500, res)
	}
}

func (srv *SrvShare) Download(res http.ResponseWriter, req *http.Request) interface{} {
	shareId := req.FormValue(srv.Conf.KeyShareId)
	if !srv.Walls.PassIpLimit(GetRemoteIp(req.RemoteAddr)) ||
		!srv.Walls.PassOpLimit(shareId, srv.Conf.OpIdDownload) {
		return httputil.Err429
	}

	return srv.download(shareId, res, req)
}

func (srv *SrvShare) download(shareId string, res http.ResponseWriter, req *http.Request) interface{} {
	if !srv.IsValidShareId(shareId) {
		return httputil.Err400
	}

	fileInfo, found := srv.Index.Get(shareId)
	switch {
	case !found || fileInfo.State != fileidx.StateDone:
		return httputil.Err404
	case fileInfo.DownLimit == 0:
		return httputil.Err412
	default:
		updated, _ := srv.Index.DecrDownLimit(shareId)
		if updated != 1 {
			return httputil.Err500
		}
	}

	err := srv.Downloader.ServeFile(res, req, fileInfo)
	srv.Err.IsErr(err)
	return 0
}
