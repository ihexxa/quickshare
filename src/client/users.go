package client

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ihexxa/quickshare/src/db/userstore"
	"github.com/ihexxa/quickshare/src/handlers"
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

func (cl *SingleUserClient) SetUser(ID uint64, role string, quota *userstore.Quota, token *http.Cookie) (*http.Response, string, []error) {
	return cl.r.Patch(cl.url("/v1/users/")).
		Send(multiusers.SetUserReq{
			ID:    ID,
			Role:  role,
			Quota: quota,
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

	if len(errs) > 0 {
		return nil, nil, errs
	}

	auResp := &multiusers.AddUserResp{}
	err := json.Unmarshal([]byte(body), auResp)
	if err != nil {
		errs = append(errs, err)
		return nil, nil, errs
	}
	return resp, auResp, errs
}

func (cl *SingleUserClient) DelUser(id string, token *http.Cookie) (*http.Response, string, []error) {
	return cl.r.Delete(cl.url("/v1/users/")).
		AddCookie(token).
		Param(handlers.UserIDParam, id).
		End()
}

func (cl *SingleUserClient) ListUsers(token *http.Cookie) (*http.Response, *multiusers.ListUsersResp, []error) {
	resp, body, errs := cl.r.Get(cl.url("/v1/users/list")).
		AddCookie(token).
		End()
	if len(errs) > 0 {
		return nil, nil, errs
	}

	lsResp := &multiusers.ListUsersResp{}
	err := json.Unmarshal([]byte(body), lsResp)
	if err != nil {
		errs = append(errs, err)
		return nil, nil, errs
	}
	return resp, lsResp, errs
}

func (cl *SingleUserClient) AddRole(role string, token *http.Cookie) (*http.Response, string, []error) {
	return cl.r.Post(cl.url("/v1/roles/")).
		AddCookie(token).
		Send(multiusers.AddRoleReq{
			Role: role,
		}).
		End()
}

func (cl *SingleUserClient) DelRole(role string, token *http.Cookie) (*http.Response, string, []error) {
	return cl.r.Delete(cl.url("/v1/roles/")).
		AddCookie(token).
		Send(multiusers.DelRoleReq{
			Role: role,
		}).
		End()
}

func (cl *SingleUserClient) ListRoles(token *http.Cookie) (*http.Response, *multiusers.ListRolesResp, []error) {
	resp, body, errs := cl.r.Get(cl.url("/v1/roles/list")).
		AddCookie(token).
		End()
	if len(errs) > 0 {
		return nil, nil, errs
	}

	lsResp := &multiusers.ListRolesResp{}
	err := json.Unmarshal([]byte(body), lsResp)
	if err != nil {
		errs = append(errs, err)
		return nil, nil, errs
	}
	return resp, lsResp, errs
}

func (cl *SingleUserClient) Self(token *http.Cookie) (*http.Response, *multiusers.SelfResp, []error) {
	resp, body, errs := cl.r.Get(cl.url("/v1/users/self")).
		AddCookie(token).
		End()
	if len(errs) > 0 {
		return nil, nil, errs
	}

	selfResp := &multiusers.SelfResp{}
	err := json.Unmarshal([]byte(body), selfResp)
	if err != nil {
		errs = append(errs, err)
		return nil, nil, errs
	}
	return resp, selfResp, errs
}

func (cl *SingleUserClient) SetPreferences(prefers *userstore.Preferences, token *http.Cookie) (*http.Response, string, []error) {
	return cl.r.Patch(cl.url("/v1/users/preferences")).
		Send(multiusers.SetPreferencesReq{
			Preferences: prefers,
		}).
		AddCookie(token).
		End()
}

func (cl *SingleUserClient) IsAuthed(token *http.Cookie) (*http.Response, string, []error) {
	return cl.r.Get(cl.url("/v1/users/isauthed")).
		AddCookie(token).
		End()
}
