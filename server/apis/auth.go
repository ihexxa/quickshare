package apis

import (
	"net/http"
	"time"
)

import (
	"quickshare/server/libs/httputil"
	"quickshare/server/libs/httpworker"
)

func (srv *SrvShare) LoginHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		srv.Http.Fill(httputil.Err404, res)
		return
	}

	act := req.FormValue(srv.Conf.KeyAct)
	todo := func(res http.ResponseWriter, req *http.Request) interface{} { return httputil.Err404 }
	switch act {
	case srv.Conf.ActLogin:
		todo = srv.Login
	case srv.Conf.ActLogout:
		todo = srv.Logout
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
		return
	}

	execErr := srv.WorkerPool.IsInTime(ack, time.Duration(srv.Conf.Timeout)*time.Millisecond)
	if srv.Err.IsErr(execErr) {
		srv.Http.Fill(httputil.Err500, res)
	}
}

func (srv *SrvShare) Login(res http.ResponseWriter, req *http.Request) interface{} {
	// all users need to pass same wall to login
	if !srv.Walls.PassIpLimit(GetRemoteIp(req.RemoteAddr)) ||
		!srv.Walls.PassOpLimit(srv.Conf.AllUsers, srv.Conf.OpIdLogin) {
		return httputil.Err504
	}

	return srv.login(
		req.FormValue(srv.Conf.KeyAdminId),
		req.FormValue(srv.Conf.KeyAdminPwd),
		res,
	)
}

func (srv *SrvShare) login(adminId string, adminPwd string, res http.ResponseWriter) interface{} {
	if adminId != srv.Conf.AdminId ||
		adminPwd != srv.Conf.AdminPwd {
		return httputil.Err401
	}

	token := srv.Walls.MakeLoginToken(srv.Conf.AdminId)
	if token == "" {
		return httputil.Err500
	}

	srv.Http.SetCookie(res, srv.Conf.KeyToken, token)
	return httputil.Ok200
}

func (srv *SrvShare) Logout(res http.ResponseWriter, req *http.Request) interface{} {
	srv.Http.SetCookie(res, srv.Conf.KeyToken, "-")
	return httputil.Ok200
}

func (srv *SrvShare) IsValidLength(length int64) bool {
	return length > 0 && length <= srv.Conf.MaxUpBytesPerSec
}

func (srv *SrvShare) IsValidStart(start, expectStart int64) bool {
	return start == expectStart
}

func (srv *SrvShare) IsValidShareId(shareId string) bool {
	// id could be 0 for dev environment
	if srv.Conf.Production {
		return len(shareId) == 64
	}
	return true
}

func (srv *SrvShare) IsValidDownLimit(limit int) bool {
	return limit >= -1
}

func IsValidFileName(fileName string) bool {
	return fileName != "" && len(fileName) < 240
}
