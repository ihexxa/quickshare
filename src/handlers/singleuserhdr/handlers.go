package singleuserhdr

import (
	"crypto/rand"
	"crypto/sha1"
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
	NewPwdParam      = "newpwd"
	RoleParam        = "role"
	ExpireParam      = "expire"
	InitTimeParam    = "initTime"
	TokenCookie      = "tk"
	AdminRole        = "admin"
	VisitorRole      = "visitor"
	InitNs           = "usersInit"
	UsersNs          = "users"
	RolesNs          = "roles"
)

type SimpleUserHandlers struct {
	cfg  gocfg.ICfg
	deps *depidx.Deps
}

func NewSimpleUserHandlers(cfg gocfg.ICfg, deps *depidx.Deps) (*SimpleUserHandlers, error) {
	var err error
	if err = deps.KV().AddNamespace(InitNs); err != nil {
		return nil, err
	}
	if err = deps.KV().AddNamespace(UsersNs); err != nil {
		return nil, err
	}
	if err = deps.KV().AddNamespace(RolesNs); err != nil {
		return nil, err
	}

	return &SimpleUserHandlers{
		cfg:  cfg,
		deps: deps,
	}, nil
}

func (h *SimpleUserHandlers) IsInited() bool {
	_, ok := h.deps.KV().GetStringIn(InitNs, InitTimeParam)
	return ok
}

func generatePwd() (string, error) {
	size := 10
	buf := make([]byte, size)
	size, err := rand.Read(buf)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", sha1.Sum(buf[:size]))[:8], nil
}

func (h *SimpleUserHandlers) Init(userName string) (string, error) {
	if userName == "" {
		return "", errors.New("user name can not be empty")
	}

	var err error
	tmpPwd, err := generatePwd()
	if err != nil {
		return "", err
	}

	err = h.deps.KV().SetStringIn(UsersNs, userName, tmpPwd)
	if err != nil {
		return "", err
	}
	err = h.deps.KV().SetStringIn(RolesNs, RoleParam, AdminRole)
	if err != nil {
		return "", err
	}
	err = h.deps.KV().SetStringIn(InitNs, InitTimeParam, fmt.Sprintf("%d", time.Now().Unix()))
	if err != nil {
		return "", err
	}

	return tmpPwd, nil
}

func (h *SimpleUserHandlers) Login(c *gin.Context) {
	user, ok1 := c.GetPostForm(UserParam)
	pwd, ok2 := c.GetPostForm(PwdParam)
	if !ok1 || !ok2 {
		c.JSON(q.ErrResp(c, 401, ErrInvalidUser))
		return
	}

	expectedHash, ok := h.deps.KV().GetStringIn(UsersNs, user)
	if !ok {
		c.JSON(q.ErrResp(c, 500, ErrInvalidConfig))
		return
	}

	err := bcrypt.CompareHashAndPassword([]byte(expectedHash), []byte(pwd))
	if err != nil {
		c.JSON(q.ErrResp(c, 401, ErrInvalidUser))
		return
	}

	role, ok := h.deps.KV().GetStringIn(RolesNs, user)
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

func (h *SimpleUserHandlers) SetPwd(c *gin.Context) {
	user, ok1 := c.GetPostForm(UserParam)
	pwd1, ok2 := c.GetPostForm(PwdParam)
	pwd2, ok3 := c.GetPostForm(NewPwdParam)
	if !ok1 || !ok2 || !ok3 {
		c.JSON(q.ErrResp(c, 401, ErrInvalidUser))
		return
	}

	expectedHash, ok := h.deps.KV().GetStringIn(UsersNs, user)
	if !ok {
		c.JSON(q.ErrResp(c, 500, ErrInvalidConfig))
		return
	}

	err := bcrypt.CompareHashAndPassword([]byte(expectedHash), []byte(pwd1))
	if err != nil {
		c.JSON(q.ErrResp(c, 401, ErrInvalidUser))
		return
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(pwd2), 10)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, errors.New("fail to set password")))
		return
	}
	err = h.deps.KV().SetStringIn(UsersNs, user, string(newHash))
	if err != nil {
		c.JSON(q.ErrResp(c, 500, ErrInvalidConfig))
		return
	}

	c.JSON(q.Resp(200))
}
