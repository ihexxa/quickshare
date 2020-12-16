package client

import (
	"fmt"
	"net/http"

	su "github.com/ihexxa/quickshare/src/handlers/singleuserhdr"
	"github.com/parnurzeal/gorequest"
)

type SingleUserClient struct {
	addr string
	r    *gorequest.SuperAgent
}

func NewSingleUserClient(addr string) *SingleUserClient {
	gr := gorequest.New()
	return &SingleUserClient{
		addr: addr,
		r:    gr,
	}
}

func (cl *SingleUserClient) url(urlpath string) string {
	return fmt.Sprintf("%s%s", cl.addr, urlpath)
}

func (cl *SingleUserClient) Login(user, pwd string) (*http.Response, string, []error) {
	return cl.r.Post(cl.url("/v1/users/login")).
		Send(su.LoginReq{
			User: user,
			Pwd:  pwd,
		}).
		End()
}

func (cl *SingleUserClient) Logout(token *http.Cookie) (*http.Response, string, []error) {
	return cl.r.Post(cl.url("/v1/users/logout")).
		AddCookie(token).
		End()
}

func (cl *SingleUserClient) SetPwd(oldPwd, newPwd string, token *http.Cookie) (*http.Response, string, []error) {
	return cl.r.Patch(cl.url("/v1/users/pwd")).
		Send(su.SetPwdReq{
			OldPwd: oldPwd,
			NewPwd: newPwd,
		}).
		AddCookie(token).
		End()
}
