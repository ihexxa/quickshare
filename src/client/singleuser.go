package client

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ihexxa/quickshare/src/handlers/multiusers"
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
		Send(multiusers.LoginReq{
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
		Send(multiusers.SetPwdReq{
			OldPwd: oldPwd,
			NewPwd: newPwd,
		}).
		AddCookie(token).
		End()
}

func (cl *SingleUserClient) AddUser(name, pwd, role string, token *http.Cookie) (*http.Response, *multiusers.AddUserResp, []error) {
	resp, body, errs := cl.r.Post(cl.url("/v1/users/")).
		AddCookie(token).
		Send(multiusers.AddUserReq{
			Name: name,
			Pwd:  pwd,
			Role: role,
		}).
		End()

	auResp := &multiusers.AddUserResp{}
	err := json.Unmarshal([]byte(body), auResp)
	if err != nil {
		errs = append(errs, err)
		return nil, nil, errs
	}
	return resp, auResp, nil
}
