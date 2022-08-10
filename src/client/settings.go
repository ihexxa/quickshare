package client

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ihexxa/quickshare/src/handlers/settings"
	"github.com/parnurzeal/gorequest"
)

type SettingsClient struct {
	addr  string
	token *http.Cookie
	r     *gorequest.SuperAgent
}

func NewSettingsClient(addr string, token *http.Cookie) *SettingsClient {
	gr := gorequest.New()
	return &SettingsClient{
		addr:  addr,
		token: token,
		r:     gr,
	}
}

func (cl *SettingsClient) url(urlpath string) string {
	return fmt.Sprintf("%s%s", cl.addr, urlpath)
}

func (cl *SettingsClient) Health() (*http.Response, string, []error) {
	return cl.r.Options(cl.url("/v1/settings/health")).
		End()
}

func (cl *SettingsClient) GetClientCfg() (*http.Response, *settings.ClientCfgMsg, []error) {
	resp, body, errs := cl.r.Get(cl.url("/v1/settings/client")).
		AddCookie(cl.token).
		End()

	mResp := &settings.ClientCfgMsg{}
	err := json.Unmarshal([]byte(body), mResp)
	if err != nil {
		errs = append(errs, err)
		return nil, nil, errs
	}
	return resp, mResp, nil
}

func (cl *SettingsClient) SetClientCfg(cfgMsg *settings.ClientCfgMsg) (*http.Response, string, []error) {
	return cl.r.Patch(cl.url("/v1/settings/client")).
		AddCookie(cl.token).
		Send(cfgMsg).
		End()
}

func (cl *SettingsClient) ReportErrors(reports *settings.ClientErrorReports) (*http.Response, string, []error) {
	return cl.r.Post(cl.url("/v1/settings/errors")).
		AddCookie(cl.token).
		Send(reports).
		End()
}

func (cl *SettingsClient) WorkerQueueLen() (*http.Response, *settings.WorkerQueueLenResp, []error) {
	resp, body, errs := cl.r.Get(cl.url("/v1/settings/workers/queue-len")).
		AddCookie(cl.token).
		End()

	mResp := &settings.WorkerQueueLenResp{}
	err := json.Unmarshal([]byte(body), mResp)
	if err != nil {
		errs = append(errs, err)
		return nil, nil, errs
	}
	return resp, mResp, nil
}
