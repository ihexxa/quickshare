package settings

import (
	"github.com/gin-gonic/gin"
	"github.com/ihexxa/gocfg"

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
