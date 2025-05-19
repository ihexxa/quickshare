package client

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ihexxa/quickshare/src/db"
	"github.com/ihexxa/quickshare/src/handlers"
	"github.com/ihexxa/quickshare/src/handlers/multiusers"
	"github.com/parnurzeal/gorequest"
)

type UsersClient struct {
	addr  string
	token *http.Cookie
	r     *gorequest.SuperAgent
}

func NewUsersClient(addr string) *UsersClient {
	gr := gorequest.New()
	return &UsersClient{
		addr: addr,
		r:    gr,
	}
}

func (cl *UsersClient) url(urlpath string) string {
	return fmt.Sprintf("%s%s", cl.addr, urlpath)
}

func (cl *UsersClient) SetToken(token *http.Cookie) {
	cl.token = token
}

func (cl *UsersClient) Token() *http.Cookie {
	return cl.token
}

func (cl *UsersClient) Login(user, pwd string) (*http.Response, string, []error) {
	resp, body, errs := cl.r.Post(cl.url("/v2/public/login")).
		Send(multiusers.LoginReq{
			User: user,
			Pwd:  pwd,
		}).
		End()

	if len(errs) == 0 && resp.StatusCode == 200 {
		// it may overwrite the token
		httpResp := (*http.Response)(resp)
		cl.token = GetCookie(httpResp.Cookies(), handlers.TokenCookie)
	} else {
		cl.token = &http.Cookie{}
	}
	return resp, body, errs
}

func (cl *UsersClient) Logout() (*http.Response, string, []error) {
	return cl.r.Post(cl.url("/v2/my/logout")).
		AddCookie(cl.token).
		End()
}

func (cl *UsersClient) SetPwd(oldPwd, newPwd string) (*http.Response, string, []error) {
	return cl.r.Patch(cl.url("/v2/my/pwd")).
		Send(multiusers.SetPwdReq{
			OldPwd: oldPwd,
			NewPwd: newPwd,
		}).
		AddCookie(cl.token).
		End()
}

func (cl *UsersClient) ForceSetPwd(userID, newPwd string) (*http.Response, string, []error) {
	return cl.r.Patch(cl.url("/v2/admin/users/pwd/force-set")).
		Send(multiusers.ForceSetPwdReq{
			ID:     userID,
			NewPwd: newPwd,
		}).
		AddCookie(cl.token).
		End()
}

func (cl *UsersClient) SetUser(ID uint64, role string, quota *db.Quota) (*http.Response, string, []error) {
	return cl.r.Patch(cl.url("/v2/admin/users/")).
		Send(multiusers.SetUserReq{
			ID:    ID,
			Role:  role,
			Quota: quota,
		}).
		AddCookie(cl.token).
		End()
}

func (cl *UsersClient) AddUser(name, pwd, role string) (*http.Response, *multiusers.AddUserResp, []error) {
	resp, body, errs := cl.r.Post(cl.url("/v2/admin/users/")).
		AddCookie(cl.token).
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

func (cl *UsersClient) DelUser(id string) (*http.Response, string, []error) {
	return cl.r.Delete(cl.url("/v2/admin/users/")).
		AddCookie(cl.token).
		Param(handlers.UserIDParam, id).
		End()
}

func (cl *UsersClient) ListUsers() (*http.Response, *multiusers.ListUsersResp, []error) {
	resp, body, errs := cl.r.Get(cl.url("/v2/admin/users/list")).
		AddCookie(cl.token).
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

// func (cl *UsersClient) AddRole(role string) (*http.Response, string, []error) {
// 	return cl.r.Post(cl.url("/v1/roles/")).
// 		AddCookie(cl.token).
// 		Send(multiusers.AddRoleReq{
// 			Role: role,
// 		}).
// 		End()
// }

// func (cl *UsersClient) DelRole(role string) (*http.Response, string, []error) {
// 	return cl.r.Delete(cl.url("/v1/roles/")).
// 		AddCookie(cl.token).
// 		Send(multiusers.DelRoleReq{
// 			Role: role,
// 		}).
// 		End()
// }

func (cl *UsersClient) ListRoles() (*http.Response, *multiusers.ListRolesResp, []error) {
	resp, body, errs := cl.r.Get(cl.url("/v2/admin/roles/list")).
		AddCookie(cl.token).
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

func (cl *UsersClient) Self() (*http.Response, *multiusers.SelfResp, []error) {
	resp, body, errs := cl.r.Get(cl.url("/v2/my/self")).
		AddCookie(cl.token).
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

func (cl *UsersClient) SetPreferences(prefers *db.Preferences) (*http.Response, string, []error) {
	return cl.r.Patch(cl.url("/v2/my/preferences")).
		Send(multiusers.SetPreferencesReq{
			Preferences: prefers,
		}).
		AddCookie(cl.token).
		End()
}

func (cl *UsersClient) IsAuthed() (*http.Response, string, []error) {
	return cl.r.Get(cl.url("/v2/my/isauthed")).
		AddCookie(cl.token).
		End()
}
