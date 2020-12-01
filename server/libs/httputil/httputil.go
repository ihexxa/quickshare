package httputil

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ihexxa/quickshare/server/libs/errutil"
)

type MsgRes struct {
	Code int
	Msg  string
}

var (
	Err400 = MsgRes{Code: http.StatusBadRequest, Msg: "Bad Request"}
	Err401 = MsgRes{Code: http.StatusUnauthorized, Msg: "Unauthorized"}
	Err404 = MsgRes{Code: http.StatusNotFound, Msg: "Not Found"}
	Err412 = MsgRes{Code: http.StatusPreconditionFailed, Msg: "Precondition Failed"}
	Err429 = MsgRes{Code: http.StatusTooManyRequests, Msg: "Too Many Requests"}
	Err500 = MsgRes{Code: http.StatusInternalServerError, Msg: "Internal Server Error"}
	Err503 = MsgRes{Code: http.StatusServiceUnavailable, Msg: "Service Unavailable"}
	Err504 = MsgRes{Code: http.StatusGatewayTimeout, Msg: "Gateway Timeout"}
	Ok200  = MsgRes{Code: http.StatusOK, Msg: "OK"}
)

type HttpUtil interface {
	GetCookie(cookies []*http.Cookie, key string) string
	SetCookie(res http.ResponseWriter, key string, val string)
	Fill(msg interface{}, res http.ResponseWriter) int
}

type QHttpUtil struct {
	CookieDomain   string
	CookieHttpOnly bool
	CookieMaxAge   int
	CookiePath     string
	CookieSecure   bool
	Err            errutil.ErrUtil
}

func (q *QHttpUtil) GetCookie(cookies []*http.Cookie, key string) string {
	for _, cookie := range cookies {
		if cookie.Name == key {
			return cookie.Value
		}
	}
	return ""
}

func (q *QHttpUtil) SetCookie(res http.ResponseWriter, key string, val string) {
	cookie := http.Cookie{
		Name:     key,
		Value:    val,
		Domain:   q.CookieDomain,
		Expires:  time.Now().Add(time.Duration(q.CookieMaxAge) * time.Second),
		HttpOnly: q.CookieHttpOnly,
		MaxAge:   q.CookieMaxAge,
		Secure:   q.CookieSecure,
		Path:     q.CookiePath,
	}

	res.Header().Set("Set-Cookie", cookie.String())
}

func (q *QHttpUtil) Fill(msg interface{}, res http.ResponseWriter) int {
	if msg == nil {
		return 0
	}

	msgBytes, marsErr := json.Marshal(msg)
	if q.Err.IsErr(marsErr) {
		return 0
	}

	wrote, writeErr := res.Write(msgBytes)
	if q.Err.IsErr(writeErr) {
		return 0
	}
	return wrote
}
