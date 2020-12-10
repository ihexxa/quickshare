package client

import (
	"fmt"
	"net/http"

	"github.com/parnurzeal/gorequest"
)

type SettingsClient struct {
	addr string
	r    *gorequest.SuperAgent
}

func NewSettingsClient(addr string) *SettingsClient {
	gr := gorequest.New()
	return &SettingsClient{
		addr: addr,
		r:    gr,
	}
}

func (cl *SettingsClient) url(urlpath string) string {
	return fmt.Sprintf("%s%s", cl.addr, urlpath)
}

func (cl *SettingsClient) Health() (*http.Response, string, []error) {
	return cl.r.Options(cl.url("/v1/settings/health")).
		End()
}
