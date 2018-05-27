package apis

import (
	"fmt"
	"math/rand"
	"net/http"
	"path/filepath"
	"strconv"
	"time"
)

import (
	"github.com/ihexxa/quickshare/server/libs/fileidx"
	"github.com/ihexxa/quickshare/server/libs/httputil"
	"github.com/ihexxa/quickshare/server/libs/httpworker"
)

func (srv *SrvShare) FileInfoHandler(res http.ResponseWriter, req *http.Request) {
	tokenStr := srv.Http.GetCookie(req.Cookies(), srv.Conf.KeyToken)
	if !srv.Walls.PassIpLimit(GetRemoteIp(req.RemoteAddr)) ||
		!srv.Walls.PassLoginCheck(tokenStr, req) {
		srv.Http.Fill(httputil.Err429, res)
		return
	}

	todo := func(res http.ResponseWriter, req *http.Request) interface{} { return httputil.Err404 }
	switch req.Method {
	case http.MethodGet:
		todo = srv.List
	case http.MethodDelete:
		todo = srv.Del
	case http.MethodPatch:
		act := req.FormValue(srv.Conf.KeyAct)
		switch act {
		case srv.Conf.ActShadowId:
			todo = srv.ShadowId
		case srv.Conf.ActPublishId:
			todo = srv.PublishId
		case srv.Conf.ActSetDownLimit:
			todo = srv.SetDownLimit
		case srv.Conf.ActAddLocalFiles:
			todo = srv.AddLocalFiles
		default:
			srv.Http.Fill(httputil.Err404, res)
			return
		}
	default:
		srv.Http.Fill(httputil.Err404, res)
		return
	}

	ack := make(chan error, 1)
	ok := srv.WorkerPool.Put(&httpworker.Task{
		Ack: ack,
		Do:  srv.Wrap(todo),
		Res: res,
		Req: req,
	})
	if !ok {
		srv.Http.Fill(httputil.Err503, res)
	}

	execErr := srv.WorkerPool.IsInTime(ack, time.Duration(srv.Conf.Timeout)*time.Millisecond)
	if srv.Err.IsErr(execErr) {
		srv.Http.Fill(httputil.Err500, res)
	}
}

type ResInfos struct {
	List []*fileidx.FileInfo
}

func (srv *SrvShare) List(res http.ResponseWriter, req *http.Request) interface{} {
	if !srv.Walls.PassOpLimit(srv.Conf.AllUsers, srv.Conf.OpIdGetFInfo) {
		return httputil.Err429
	}

	return srv.list()
}

func (srv *SrvShare) list() interface{} {
	infos := make([]*fileidx.FileInfo, 0)
	for _, info := range srv.Index.List() {
		infos = append(infos, info)
	}

	return &ResInfos{List: infos}
}

func (srv *SrvShare) Del(res http.ResponseWriter, req *http.Request) interface{} {
	shareId := req.FormValue(srv.Conf.KeyShareId)
	if !srv.Walls.PassOpLimit(shareId, srv.Conf.OpIdDelFInfo) {
		return httputil.Err504
	}

	return srv.del(shareId)
}

func (srv *SrvShare) del(shareId string) interface{} {
	if !srv.IsValidShareId(shareId) {
		return httputil.Err400
	}

	fileInfo, found := srv.Index.Get(shareId)
	if !found {
		return httputil.Err404
	}

	srv.Index.Del(shareId)
	fullPath := filepath.Join(srv.Conf.PathLocal, fileInfo.PathLocal)
	if !srv.Fs.DelFile(fullPath) {
		// TODO: may log file name because file not exist or delete is not authenticated
		return httputil.Err500
	}

	return httputil.Ok200
}

func (srv *SrvShare) ShadowId(res http.ResponseWriter, req *http.Request) interface{} {
	if !srv.Walls.PassOpLimit(srv.Conf.AllUsers, srv.Conf.OpIdOpFInfo) {
		return httputil.Err429
	}

	shareId := req.FormValue(srv.Conf.KeyShareId)
	return srv.shadowId(shareId)
}

func (srv *SrvShare) shadowId(shareId string) interface{} {
	if !srv.IsValidShareId(shareId) {
		return httputil.Err400
	}

	info, found := srv.Index.Get(shareId)
	if !found {
		return httputil.Err404
	}

	secretId := srv.Encryptor.Encrypt(
		[]byte(fmt.Sprintf("%s%s", info.PathLocal, genPwd())),
	)
	if !srv.Index.SetId(info.Id, secretId) {
		return httputil.Err412
	}

	return &ShareInfo{ShareId: secretId}
}

func (srv *SrvShare) PublishId(res http.ResponseWriter, req *http.Request) interface{} {
	if !srv.Walls.PassOpLimit(srv.Conf.AllUsers, srv.Conf.OpIdOpFInfo) {
		return httputil.Err429
	}

	shareId := req.FormValue(srv.Conf.KeyShareId)
	return srv.publishId(shareId)
}

func (srv *SrvShare) publishId(shareId string) interface{} {
	if !srv.IsValidShareId(shareId) {
		return httputil.Err400
	}

	info, found := srv.Index.Get(shareId)
	if !found {
		return httputil.Err404
	}

	publicId := srv.Encryptor.Encrypt([]byte(info.PathLocal))
	if !srv.Index.SetId(info.Id, publicId) {
		return httputil.Err412
	}

	return &ShareInfo{ShareId: publicId}
}

func (srv *SrvShare) SetDownLimit(res http.ResponseWriter, req *http.Request) interface{} {
	if !srv.Walls.PassOpLimit(srv.Conf.AllUsers, srv.Conf.OpIdOpFInfo) {
		return httputil.Err429
	}

	shareId := req.FormValue(srv.Conf.KeyShareId)
	downLimit64, downLimitParseErr := strconv.ParseInt(req.FormValue(srv.Conf.KeyDownLimit), 10, 32)
	downLimit := int(downLimit64)
	if srv.Err.IsErr(downLimitParseErr) {
		return httputil.Err400
	}

	return srv.setDownLimit(shareId, downLimit)
}

func (srv *SrvShare) setDownLimit(shareId string, downLimit int) interface{} {
	if !srv.IsValidShareId(shareId) || !srv.IsValidDownLimit(downLimit) {
		return httputil.Err400
	}

	if !srv.Index.SetDownLimit(shareId, downLimit) {
		return httputil.Err404
	}
	return httputil.Ok200
}

func (srv *SrvShare) AddLocalFiles(res http.ResponseWriter, req *http.Request) interface{} {
	return srv.AddLocalFilesImp()
}

func (srv *SrvShare) AddLocalFilesImp() interface{} {
	infos, err := srv.Fs.Readdir(srv.Conf.PathLocal, srv.Conf.LocalFileLimit)
	if srv.Err.IsErr(err) {
		panic(fmt.Sprintf("fail to readdir: %v", err))
	}

	for _, info := range infos {
		info.DownLimit = srv.Conf.DownLimit
		info.State = fileidx.StateDone
		info.PathLocal = info.PathLocal
		info.Id = srv.Encryptor.Encrypt([]byte(info.PathLocal))

		addRet := srv.Index.Add(info)
		switch {
		case addRet == 0 || addRet == -1:
			// TODO: return files not added
			continue
		case addRet == 1:
			break
		default:
			return httputil.Err500
		}
	}

	return httputil.Ok200
}

func genPwd() string {
	return fmt.Sprintf("%d%d%d%d", rand.Intn(10), rand.Intn(10), rand.Intn(10), rand.Intn(10))
}
