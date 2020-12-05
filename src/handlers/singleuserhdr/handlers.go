package singleuserhdr

import (
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ihexxa/gocfg"
	"golang.org/x/crypto/bcrypt"

	"github.com/ihexxa/quickshare/src/depidx"
	q "github.com/ihexxa/quickshare/src/handlers"
)

var (
	ErrInvalidUser   = errors.New("invalid user name or password")
	ErrInvalidConfig = errors.New("invalid user config")
	UserParam        = "user"
	PwdParam         = "pwd"
	RoleParam        = "role"
	ExpireParam      = "expire"
	TokenCookie      = "tk"
	AdminRole        = "admin"
	VisitorRole      = "visitor"
	UsersNamespace   = "users"
	RolesNamespace   = "roles"
)

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

func (h *SimpleUserHandlers) Login(c *gin.Context) {
	user, ok1 := c.GetPostForm(UserParam)
	pwd, ok2 := c.GetPostForm(PwdParam)
	if !ok1 || !ok2 {
		c.JSON(q.ErrResp(c, 401, ErrInvalidUser))
		return
	}

	expectedHash, ok := h.deps.KV().GetStringIn(UsersNamespace, user)
	if !ok {
		c.JSON(q.ErrResp(c, 500, ErrInvalidConfig))
		return
	}

	err := bcrypt.CompareHashAndPassword([]byte(expectedHash), []byte(pwd))
	if err != nil {
		c.JSON(q.ErrResp(c, 401, ErrInvalidUser))
		return
	}

	role, ok := h.deps.KV().GetStringIn(RolesNamespace, user)
	if !ok {
		c.JSON(q.ErrResp(c, 500, ErrInvalidConfig))
		return
	}
	ttl := h.cfg.GrabInt("Users.CookieTTL")
	token, err := h.deps.Token().ToToken(map[string]string{
		UserParam:   user,
		RoleParam:   role,
		ExpireParam: fmt.Sprintf("%d", time.Now().Unix()+int64(ttl)),
	})
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}

	hostname := h.cfg.GrabString("Server.Host")
	secure := h.cfg.GrabBool("Users.CookieSecure")
	httpOnly := h.cfg.GrabBool("Users.CookieHttpOnly")
	c.SetCookie(TokenCookie, token, ttl, "/", hostname, secure, httpOnly)

	c.JSON(q.Resp(200))
}

func (h *SimpleUserHandlers) Logout(c *gin.Context) {
	// token alreay verified in the authn middleware
	c.SetCookie(TokenCookie, "", 0, "/", "nohost", false, true)
	c.JSON(q.Resp(200))
}
