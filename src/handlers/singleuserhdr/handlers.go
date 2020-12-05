package singleuserhdr

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/ihexxa/gocfg"

	"github.com/ihexxa/quickshare/src/depidx"
	q "github.com/ihexxa/quickshare/src/handlers"
)

var ErrInvalidUser = errors.New("invalid user name or password")

type SimpleUserHandlers struct {
	cfg  gocfg.ICfg
	deps *depidx.Deps
}

func NewSimpleUserHandlers(cfg gocfg.ICfg, deps *depidx.Deps) *SimpleUserHandlers {
	return &SimpleUserHandlers{
		cfg:  cfg,
		deps: deps,
	}
}

func (hdr *SimpleUserHandlers) Login(c *gin.Context) {
	userName := c.Query("username")
	pwd := c.Query("pwd")
	if userName == "" || pwd == "" {
		c.JSON(q.ErrResp(c, 400, ErrInvalidUser))
		return
	}

	expectedName, ok1 := hdr.deps.KV().GetString("username")
	expectedPwd, ok2 := hdr.deps.KV().GetString("pwd")
	if !ok1 || !ok2 {
		c.JSON(q.ErrResp(c, 400, ErrInvalidUser))
		return
	}

	if userName != expectedName || pwd != expectedPwd {
		c.JSON(q.ErrResp(c, 400, ErrInvalidUser))
		return
	}
	token, err := hdr.deps.Token().ToToken(map[string]string{
		"username": expectedName,
	})
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}

	// TODO: use config
	c.SetCookie("token", token, 3600, "/", "localhost", false, true)
	c.JSON(q.Resp(200))
}

func (hdr *SimpleUserHandlers) Logout(c *gin.Context) {
	token, err := c.Cookie("token")
	if err != nil {
		c.JSON(q.ErrResp(c, 400, err))
		return
	}

	// TODO: // check if token expired
	_, err = hdr.deps.Token().FromToken(token, map[string]string{"token": ""})
	if err != nil {
		c.JSON(q.ErrResp(c, 400, err))
		return
	}

	c.SetCookie("token", "", 0, "/", "localhost", false, true)
	c.JSON(q.Resp(200))
}
