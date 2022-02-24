package settings

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/ihexxa/gocfg"

	"github.com/ihexxa/quickshare/src/db/sitestore"
	"github.com/ihexxa/quickshare/src/depidx"
	q "github.com/ihexxa/quickshare/src/handlers"
)

type SettingsSvc struct {
	cfg  gocfg.ICfg
	deps *depidx.Deps
}

func NewSettingsSvc(cfg gocfg.ICfg, deps *depidx.Deps) (*SettingsSvc, error) {
	return &SettingsSvc{
		cfg:  cfg,
		deps: deps,
	}, nil
}

func (h *SettingsSvc) Health(c *gin.Context) {
	// TODO: currently it checks nothing
	c.JSON(q.Resp(200))
}

type ClientCfgMsg struct {
	SiteName       string              `json:"siteName"`
	SiteDesc       string              `json:"siteDesc"`
	Bg             *sitestore.BgConfig `json:"bg"`
	CaptchaEnabled bool                `json:"captchaEnabled"`
}

func (h *SettingsSvc) GetClientCfg(c *gin.Context) {
	// TODO: add cache
	siteCfg, err := h.deps.SiteStore().GetCfg()
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}

	c.JSON(200, &ClientCfgMsg{
		SiteName:       siteCfg.ClientCfg.SiteName,
		SiteDesc:       siteCfg.ClientCfg.SiteDesc,
		Bg:             siteCfg.ClientCfg.Bg,
		CaptchaEnabled: h.cfg.BoolOr("Users.CaptchaEnabled", true),
	})
}

func (h *SettingsSvc) SetClientCfg(c *gin.Context) {
	var err error
	req := &ClientCfgMsg{}
	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(q.ErrResp(c, 400, err))
		return
	}

	clientCfg := &sitestore.ClientConfig{
		SiteName: req.SiteName,
		SiteDesc: req.SiteDesc,
		Bg:       req.Bg,
		// TODO: captchaEnabled is not persisted in db
	}
	if err = validateClientCfg(clientCfg); err != nil {
		c.JSON(q.ErrResp(c, 400, err))
		return
	}

	err = h.deps.SiteStore().SetClientCfg(clientCfg)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}
	c.JSON(q.Resp(200))
}

func validateClientCfg(cfg *sitestore.ClientConfig) error {
	if cfg.SiteName == "" {
		return errors.New("site name is empty")
	}
	return nil
}

type ClientErrorReport struct {
	Report  string `json:"report"`
	Version string `json:"version"`
}

type ClientErrorReports struct {
	Reports []*ClientErrorReport `json:"reports"`
}

func (h *SettingsSvc) ReportErrors(c *gin.Context) {
	var err error
	req := &ClientErrorReports{}
	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(q.ErrResp(c, 400, err))
		return
	}

	for _, report := range req.Reports {
		h.deps.Log().Errorf("version:%s,error:%s", report.Version, report.Report)
	}
	c.JSON(q.Resp(200))
}
