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
	httpUtil "github.com/ihexxa/quickshare/server/libs/httputil"
	worker "github.com/ihexxa/quickshare/server/libs/httpworker"
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
		srv.Http.Fill(httpUtil.Err404, res)
		return
	}

	tokenStr := srv.Http.GetCookie(req.Cookies(), srv.Conf.KeyToken)
	ipPass := srv.Walls.PassIpLimit(GetRemoteIp(req.RemoteAddr))
	loginPass := srv.Walls.PassLoginCheck(tokenStr, req)
	opPass := srv.Walls.PassOpLimit(GetRemoteIp(req.RemoteAddr), srv.Conf.OpIdUpload)
	if !ipPass || !loginPass || !opPass {
		srv.Http.Fill(httpUtil.Err429, res)
		return
	}

	ack := make(chan error, 1)
	ok := srv.WorkerPool.Put(&worker.Task{
		Ack: ack,
		Do:  srv.Wrap(srv.StartUpload),
		Res: res,
		Req: req,
	})
	if !ok {
		srv.Http.Fill(httpUtil.Err503, res)
	}

	execErr := srv.WorkerPool.IsInTime(ack, time.Duration(srv.Conf.Timeout)*time.Millisecond)
	if srv.Err.IsErr(execErr) {
		srv.Http.Fill(httpUtil.Err500, res)
	}
}

func (srv *SrvShare) UploadHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		srv.Http.Fill(httpUtil.Err404, res)
		return
	}

	tokenStr := srv.Http.GetCookie(req.Cookies(), srv.Conf.KeyToken)
	ipPass := srv.Walls.PassIpLimit(GetRemoteIp(req.RemoteAddr))
	loginPass := srv.Walls.PassLoginCheck(tokenStr, req)
	opPass := srv.Walls.PassOpLimit(GetRemoteIp(req.RemoteAddr), srv.Conf.OpIdUpload)
	if !ipPass || !loginPass || !opPass {
		srv.Http.Fill(httpUtil.Err429, res)
		return
	}

	multiFormErr := req.ParseMultipartForm(srv.Conf.ParseFormBufSize)
	if srv.Err.IsErr(multiFormErr) {
		srv.Http.Fill(httpUtil.Err400, res)
		return
	}

	srv.Log.Println("form", req.Form)
	srv.Log.Println("pform", req.PostForm)
	srv.Log.Println("mform", req.MultipartForm)
	ack := make(chan error, 1)
	ok := srv.WorkerPool.Put(&worker.Task{
		Ack: ack,
		Do:  srv.Wrap(srv.Upload),
		Res: res,
		Req: req,
	})
	if !ok {
		srv.Http.Fill(httpUtil.Err503, res)
	}

	execErr := srv.WorkerPool.IsInTime(ack, time.Duration(srv.Conf.Timeout)*time.Millisecond)
	if srv.Err.IsErr(execErr) {
		srv.Http.Fill(httpUtil.Err500, res)
	}
}

func (srv *SrvShare) FinishUploadHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		srv.Http.Fill(httpUtil.Err404, res)
		return
	}

	tokenStr := srv.Http.GetCookie(req.Cookies(), srv.Conf.KeyToken)
	ipPass := srv.Walls.PassIpLimit(GetRemoteIp(req.RemoteAddr))
	loginPass := srv.Walls.PassLoginCheck(tokenStr, req)
	opPass := srv.Walls.PassOpLimit(GetRemoteIp(req.RemoteAddr), srv.Conf.OpIdUpload)
	if !ipPass || !loginPass || !opPass {
		srv.Http.Fill(httpUtil.Err429, res)
		return
	}

	ack := make(chan error, 1)
	ok := srv.WorkerPool.Put(&worker.Task{
		Ack: ack,
		Do:  srv.Wrap(srv.FinishUpload),
		Res: res,
		Req: req,
	})
	if !ok {
		srv.Http.Fill(httpUtil.Err503, res)
	}

	execErr := srv.WorkerPool.IsInTime(ack, time.Duration(srv.Conf.Timeout)*time.Millisecond)
	if srv.Err.IsErr(execErr) {
		srv.Http.Fill(httpUtil.Err500, res)
	}
}

func (srv *SrvShare) StartUpload(res http.ResponseWriter, req *http.Request) interface{} {
	return srv.startUpload(req.FormValue(srv.Conf.KeyFileName))
}

func (srv *SrvShare) startUpload(fileName string) interface{} {
	if !IsValidFileName(fileName) {
		return httpUtil.Err400
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
		return httpUtil.Err412
	case 1:
		return httpUtil.Err500 // TODO: use correct status code
	default:
		srv.Index.Del(id)
		return httpUtil.Err500
	}

	fullPath := filepath.Join(srv.Conf.PathLocal, info.PathLocal)
	createFileErr := srv.Fs.CreateFile(fullPath)
	switch {
	case createFileErr == fsutil.ErrExists:
		srv.Index.Del(id)
		return httpUtil.Err412
	case createFileErr == fsutil.ErrUnknown:
		srv.Index.Del(id)
		return httpUtil.Err500
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
		return httpUtil.Err400
	}

	return srv.upload(shareId, start, length, chunk)
}

func (srv *SrvShare) upload(shareId string, start int64, length int64, chunk io.Reader) interface{} {
	if !srv.IsValidShareId(shareId) {
		return httpUtil.Err400
	}

	fileInfo, found := srv.Index.Get(shareId)
	if !found {
		return httpUtil.Err404
	}

	if !srv.IsValidStart(start, fileInfo.Uploaded) || !srv.IsValidLength(length) {
		return httpUtil.Err400
	}

	fullPath := filepath.Join(srv.Conf.PathLocal, fileInfo.PathLocal)
	if !srv.Fs.CopyChunkN(fullPath, chunk, start, length) {
		return httpUtil.Err500
	}

	if srv.Index.IncrUploaded(shareId, length) == 0 {
		return httpUtil.Err404
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
		return httpUtil.Err404
	}

	return &ShareInfo{
		ShareId: shareId,
	}
}

func genInfoId(content string, key []byte) string {
	encrypter := encrypt.HmacEncryptor{Key: key}
	return encrypter.Encrypt([]byte(content))
}
