package apis

import (
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"time"
)

import (
	"github.com/ihexxa/quickshare/server/libs/encrypt"
	"github.com/ihexxa/quickshare/server/libs/fileidx"
	"github.com/ihexxa/quickshare/server/libs/fsutil"
	"github.com/ihexxa/quickshare/server/libs/httputil"
	"github.com/ihexxa/quickshare/server/libs/httpworker"
)

const DefaultId = "0"

type ByteRange struct {
	ShareId string
	Start   int64
	Length  int64
}

type ShareInfo struct {
	ShareId string
}

func (srv *SrvShare) StartUploadHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		srv.Http.Fill(httputil.Err404, res)
		return
	}

	tokenStr := srv.Http.GetCookie(req.Cookies(), srv.Conf.KeyToken)
	ipPass := srv.Walls.PassIpLimit(GetRemoteIp(req.RemoteAddr))
	loginPass := srv.Walls.PassLoginCheck(tokenStr, req)
	opPass := srv.Walls.PassOpLimit(GetRemoteIp(req.RemoteAddr), srv.Conf.OpIdUpload)
	if !ipPass || !loginPass || !opPass {
		srv.Http.Fill(httputil.Err429, res)
		return
	}

	ack := make(chan error, 1)
	ok := srv.WorkerPool.Put(&httpworker.Task{
		Ack: ack,
		Do:  srv.Wrap(srv.StartUpload),
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

func (srv *SrvShare) UploadHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		srv.Http.Fill(httputil.Err404, res)
		return
	}

	tokenStr := srv.Http.GetCookie(req.Cookies(), srv.Conf.KeyToken)
	ipPass := srv.Walls.PassIpLimit(GetRemoteIp(req.RemoteAddr))
	loginPass := srv.Walls.PassLoginCheck(tokenStr, req)
	opPass := srv.Walls.PassOpLimit(GetRemoteIp(req.RemoteAddr), srv.Conf.OpIdUpload)
	if !ipPass || !loginPass || !opPass {
		srv.Http.Fill(httputil.Err429, res)
		return
	}

	multiFormErr := req.ParseMultipartForm(srv.Conf.ParseFormBufSize)
	if srv.Err.IsErr(multiFormErr) {
		srv.Http.Fill(httputil.Err400, res)
		return
	}

	ack := make(chan error, 1)
	ok := srv.WorkerPool.Put(&httpworker.Task{
		Ack: ack,
		Do:  srv.Wrap(srv.Upload),
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

func (srv *SrvShare) FinishUploadHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		srv.Http.Fill(httputil.Err404, res)
		return
	}

	tokenStr := srv.Http.GetCookie(req.Cookies(), srv.Conf.KeyToken)
	ipPass := srv.Walls.PassIpLimit(GetRemoteIp(req.RemoteAddr))
	loginPass := srv.Walls.PassLoginCheck(tokenStr, req)
	opPass := srv.Walls.PassOpLimit(GetRemoteIp(req.RemoteAddr), srv.Conf.OpIdUpload)
	if !ipPass || !loginPass || !opPass {
		srv.Http.Fill(httputil.Err429, res)
		return
	}

	ack := make(chan error, 1)
	ok := srv.WorkerPool.Put(&httpworker.Task{
		Ack: ack,
		Do:  srv.Wrap(srv.FinishUpload),
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

func (srv *SrvShare) StartUpload(res http.ResponseWriter, req *http.Request) interface{} {
	return srv.startUpload(req.FormValue(srv.Conf.KeyFileName))
}

func (srv *SrvShare) startUpload(fileName string) interface{} {
	if !IsValidFileName(fileName) {
		return httputil.Err400
	}

	id := DefaultId
	if srv.Conf.Production {
		id = genInfoId(fileName, srv.Conf.SecretKeyByte)
	}

	info := &fileidx.FileInfo{
		Id:        id,
		DownLimit: srv.Conf.DownLimit,
		ModTime:   time.Now().UnixNano(),
		PathLocal: fileName,
		Uploaded:  0,
		State:     fileidx.StateStarted,
	}

	switch srv.Index.Add(info) {
	case 0:
		// go on
	case -1:
		return httputil.Err412
	case 1:
		return httputil.Err500 // TODO: use correct status code
	default:
		srv.Index.Del(id)
		return httputil.Err500
	}

	fullPath := filepath.Join(srv.Conf.PathLocal, info.PathLocal)
	createFileErr := srv.Fs.CreateFile(fullPath)
	switch {
	case createFileErr == fsutil.ErrExists:
		srv.Index.Del(id)
		return httputil.Err412
	case createFileErr == fsutil.ErrUnknown:
		srv.Index.Del(id)
		return httputil.Err500
	default:
		srv.Index.SetState(id, fileidx.StateUploading)
		return &ByteRange{
			ShareId: id,
			Start:   0,
			Length:  srv.Conf.MaxUpBytesPerSec,
		}
	}
}

func (srv *SrvShare) Upload(res http.ResponseWriter, req *http.Request) interface{} {
	shareId := req.FormValue(srv.Conf.KeyShareId)
	start, startErr := strconv.ParseInt(req.FormValue(srv.Conf.KeyStart), 10, 64)
	length, lengthErr := strconv.ParseInt(req.FormValue(srv.Conf.KeyLen), 10, 64)
	chunk, _, chunkErr := req.FormFile(srv.Conf.KeyChunk)

	if srv.Err.IsErr(startErr) ||
		srv.Err.IsErr(lengthErr) ||
		srv.Err.IsErr(chunkErr) {
		return httputil.Err400
	}

	return srv.upload(shareId, start, length, chunk)
}

func (srv *SrvShare) upload(shareId string, start int64, length int64, chunk io.Reader) interface{} {
	if !srv.IsValidShareId(shareId) {
		return httputil.Err400
	}

	fileInfo, found := srv.Index.Get(shareId)
	if !found {
		return httputil.Err404
	}

	if !srv.IsValidStart(start, fileInfo.Uploaded) || !srv.IsValidLength(length) {
		return httputil.Err400
	}

	fullPath := filepath.Join(srv.Conf.PathLocal, fileInfo.PathLocal)
	if !srv.Fs.CopyChunkN(fullPath, chunk, start, length) {
		return httputil.Err500
	}

	if srv.Index.IncrUploaded(shareId, length) == 0 {
		return httputil.Err404
	}

	return &ByteRange{
		ShareId: shareId,
		Start:   start + length,
		Length:  srv.Conf.MaxUpBytesPerSec,
	}
}

func (srv *SrvShare) FinishUpload(res http.ResponseWriter, req *http.Request) interface{} {
	shareId := req.FormValue(srv.Conf.KeyShareId)
	return srv.finishUpload(shareId)
}

func (srv *SrvShare) finishUpload(shareId string) interface{} {
	if !srv.Index.SetState(shareId, fileidx.StateDone) {
		return httputil.Err404
	}

	return &ShareInfo{
		ShareId: shareId,
	}
}

func genInfoId(content string, key []byte) string {
	encrypter := encrypt.HmacEncryptor{Key: key}
	return encrypter.Encrypt([]byte(content))
}
