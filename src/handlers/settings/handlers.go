package settings

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/ihexxa/gocfg"

	"github.com/ihexxa/quickshare/src/db"
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
	ClientCfg      *db.ClientConfig `json:"clientCfg"`
	CaptchaEnabled bool             `json:"captchaEnabled"`
}

func (h *SettingsSvc) GetClientCfg(c *gin.Context) {
	// TODO: add cache
	siteCfg, err := h.deps.SiteStore().GetCfg(c)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}

	c.JSON(200, &ClientCfgMsg{
		ClientCfg:      siteCfg.ClientCfg,
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

	// TODO: captchaEnabled is not persisted in db
	clientCfg := req.ClientCfg
	if err = validateClientCfg(clientCfg); err != nil {
		c.JSON(q.ErrResp(c, 400, err))
		return
	}

	// update config
	// TODO: refine the model
	h.cfg.SetString("Site.ClientCfg.SiteName", req.ClientCfg.SiteName)
	h.cfg.SetString("Site.ClientCfg.SiteDesc", req.ClientCfg.SiteDesc)
	h.cfg.SetString("Site.ClientCfg.Bg.Url", req.ClientCfg.Bg.Url)
	h.cfg.SetString("Site.ClientCfg.Bg.Repeat", req.ClientCfg.Bg.Repeat)
	h.cfg.SetString("Site.ClientCfg.Bg.Position", req.ClientCfg.Bg.Position)
	h.cfg.SetString("Site.ClientCfg.Bg.Align", req.ClientCfg.Bg.Align)
	h.cfg.SetString("Site.ClientCfg.Bg.BgColor", req.ClientCfg.Bg.BgColor)
	h.cfg.SetBool("Site.ClientCfg.AllowSetBg", req.ClientCfg.AllowSetBg)
	h.cfg.SetBool("Site.ClientCfg.AutoTheme", req.ClientCfg.AutoTheme)

	err = h.deps.SiteStore().SetClientCfg(c, clientCfg)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}
	c.JSON(q.Resp(200))
}

func validateClientCfg(cfg *db.ClientConfig) error {
	if len(cfg.SiteName) == 0 || len(cfg.SiteName) >= 12 {
		return errors.New("site name is too short or too long")
	} else if len(cfg.SiteDesc) >= 64 {
		return errors.New("site description is too short or too long")
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

type WorkerQueueLenResp struct {
	QueueLen int `json:"queueLen"`
}

func (h *SettingsSvc) WorkerQueueLen(c *gin.Context) {
	c.JSON(200, &WorkerQueueLenResp{
		QueueLen: h.deps.Workers().QueueLen(),
	})
}
