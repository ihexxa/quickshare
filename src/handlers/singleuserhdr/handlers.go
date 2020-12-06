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

func (h *SimpleUserHandlers) Init(userName, pwd string) (string, error) {
	if userName == "" {
		return "", errors.New("user name can not be empty")
	}

	var err error
	if pwd == "" {
		tmpPwd, err := generatePwd()
		if err != nil {
			return "", err
		}
		pwd = tmpPwd
	}
	pwdHash, err := bcrypt.GenerateFromPassword([]byte(pwd), 10)
	if err != nil {
		return "", err
	}

	err = h.deps.KV().SetStringIn(UsersNs, userName, string(pwdHash))
	if err != nil {
		return "", err
	}
	err = h.deps.KV().SetStringIn(RolesNs, userName, AdminRole)
	if err != nil {
		return "", err
	}
	err = h.deps.KV().SetStringIn(InitNs, InitTimeParam, fmt.Sprintf("%d", time.Now().Unix()))
	if err != nil {
		return "", err
	}

	return pwd, nil
}

type LoginReq struct {
	User string `json:"user"`
	Pwd  string `json:"pwd"`
}

func (h *SimpleUserHandlers) Login(c *gin.Context) {
	req := &LoginReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}

	expectedHash, ok := h.deps.KV().GetStringIn(UsersNs, req.User)
	if !ok {
		c.JSON(q.ErrResp(c, 500, ErrInvalidConfig))
		return
	}

	err := bcrypt.CompareHashAndPassword([]byte(expectedHash), []byte(req.Pwd))
	if err != nil {
		c.JSON(q.ErrResp(c, 401, err))
		return
	}

	role, ok := h.deps.KV().GetStringIn(RolesNs, req.User)
	if !ok {
		c.JSON(q.ErrResp(c, 501, ErrInvalidConfig))
		return
	}
	ttl := h.cfg.GrabInt("Users.CookieTTL")
	token, err := h.deps.Token().ToToken(map[string]string{
		UserParam:   req.User,
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

type LogoutReq struct {
	User string `json:"user"`
}

func (h *SimpleUserHandlers) Logout(c *gin.Context) {
	req := &LogoutReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}

	// token alreay verified in the authn middleware
	c.SetCookie(TokenCookie, "", 0, "/", "nohost", false, true)
	c.JSON(q.Resp(200))
}

type SetPwdReq struct {
	User   string `json:"user"`
	OldPwd string `json:"oldPwd"`
	NewPwd string `json:"newPwd"`
}

func (h *SimpleUserHandlers) SetPwd(c *gin.Context) {
	req := &SetPwdReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(q.ErrResp(c, 400, err))
		return
	}

	expectedHash, ok := h.deps.KV().GetStringIn(UsersNs, req.User)
	if !ok {
		c.JSON(q.ErrResp(c, 500, ErrInvalidConfig))
		return
	}

	err := bcrypt.CompareHashAndPassword([]byte(expectedHash), []byte(req.OldPwd))
	if err != nil {
		c.JSON(q.ErrResp(c, 401, ErrInvalidUser))
		return
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPwd), 10)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, errors.New("fail to set password")))
		return
	}
	err = h.deps.KV().SetStringIn(UsersNs, req.User, string(newHash))
	if err != nil {
		c.JSON(q.ErrResp(c, 500, ErrInvalidConfig))
		return
	}

	c.JSON(q.Resp(200))
}
