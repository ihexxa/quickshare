package multiusers

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ihexxa/gocfg"
	"golang.org/x/crypto/bcrypt"

	"github.com/ihexxa/quickshare/src/depidx"
	q "github.com/ihexxa/quickshare/src/handlers"
	"github.com/ihexxa/quickshare/src/userstore"
)

var (
	ErrInvalidUser   = errors.New("invalid user name or password")
	ErrInvalidConfig = errors.New("invalid user config")
)

type MultiUsersSvc struct {
	cfg        gocfg.ICfg
	deps       *depidx.Deps
	apiACRules map[string]bool
}

func NewMultiUsersSvc(cfg gocfg.ICfg, deps *depidx.Deps) (*MultiUsersSvc, error) {
	publicPath := filepath.Join("/", cfg.GrabString("Server.PublicPath"))

	apiACRules := map[string]bool{
		// TODO: make these configurable
		// admin rules
		apiRuleCname(userstore.AdminRole, "GET", "/"):                       true,
		apiRuleCname(userstore.AdminRole, "GET", publicPath):                true,
		apiRuleCname(userstore.AdminRole, "POST", "/v1/users/login"):        true,
		apiRuleCname(userstore.AdminRole, "POST", "/v1/users/logout"):       true,
		apiRuleCname(userstore.AdminRole, "GET", "/v1/users/isauthed"):      true,
		apiRuleCname(userstore.AdminRole, "PATCH", "/v1/users/pwd"):         true,
		apiRuleCname(userstore.AdminRole, "POST", "/v1/users/"):             true,
		apiRuleCname(userstore.AdminRole, "POST", "/v1/roles/"):             true,
		apiRuleCname(userstore.AdminRole, "DELETE", "/v1/roles/"):           true,
		apiRuleCname(userstore.AdminRole, "GET", "/v1/roles/"):              true,
		apiRuleCname(userstore.AdminRole, "POST", "/v1/fs/files"):           true,
		apiRuleCname(userstore.AdminRole, "DELETE", "/v1/fs/files"):         true,
		apiRuleCname(userstore.AdminRole, "GET", "/v1/fs/files"):            true,
		apiRuleCname(userstore.AdminRole, "PATCH", "/v1/fs/files/chunks"):   true,
		apiRuleCname(userstore.AdminRole, "GET", "/v1/fs/files/chunks"):     true,
		apiRuleCname(userstore.AdminRole, "PATCH", "/v1/fs/files/copy"):     true,
		apiRuleCname(userstore.AdminRole, "PATCH", "/v1/fs/files/move"):     true,
		apiRuleCname(userstore.AdminRole, "GET", "/v1/fs/dirs"):             true,
		apiRuleCname(userstore.AdminRole, "GET", "/v1/fs/dirs/home"):        true,
		apiRuleCname(userstore.AdminRole, "POST", "/v1/fs/dirs"):            true,
		apiRuleCname(userstore.AdminRole, "GET", "/v1/fs/uploadings"):       true,
		apiRuleCname(userstore.AdminRole, "DELETE", "/v1/fs/uploadings"):    true,
		apiRuleCname(userstore.AdminRole, "GET", "/v1/fs/metadata"):         true,
		apiRuleCname(userstore.AdminRole, "OPTIONS", "/v1/settings/health"): true,
		// user rules
		apiRuleCname(userstore.UserRole, "GET", "/"):                       true,
		apiRuleCname(userstore.UserRole, "GET", publicPath):                true,
		apiRuleCname(userstore.UserRole, "POST", "/v1/users/logout"):       true,
		apiRuleCname(userstore.UserRole, "GET", "/v1/users/isauthed"):      true,
		apiRuleCname(userstore.UserRole, "PATCH", "/v1/users/pwd"):         true,
		apiRuleCname(userstore.UserRole, "POST", "/v1/fs/files"):           true,
		apiRuleCname(userstore.UserRole, "DELETE", "/v1/fs/files"):         true,
		apiRuleCname(userstore.UserRole, "GET", "/v1/fs/files"):            true,
		apiRuleCname(userstore.UserRole, "PATCH", "/v1/fs/files/chunks"):   true,
		apiRuleCname(userstore.UserRole, "GET", "/v1/fs/files/chunks"):     true,
		apiRuleCname(userstore.UserRole, "PATCH", "/v1/fs/files/copy"):     true,
		apiRuleCname(userstore.UserRole, "PATCH", "/v1/fs/files/move"):     true,
		apiRuleCname(userstore.UserRole, "GET", "/v1/fs/dirs"):             true,
		apiRuleCname(userstore.UserRole, "GET", "/v1/fs/dirs/home"):        true,
		apiRuleCname(userstore.UserRole, "POST", "/v1/fs/dirs"):            true,
		apiRuleCname(userstore.UserRole, "GET", "/v1/fs/uploadings"):       true,
		apiRuleCname(userstore.UserRole, "DELETE", "/v1/fs/uploadings"):    true,
		apiRuleCname(userstore.UserRole, "GET", "/v1/fs/metadata"):         true,
		apiRuleCname(userstore.UserRole, "OPTIONS", "/v1/settings/health"): true,
		// visitor rules
		apiRuleCname(userstore.VisitorRole, "GET", "/"):                       true,
		apiRuleCname(userstore.VisitorRole, "GET", publicPath):                true,
		apiRuleCname(userstore.VisitorRole, "POST", "/v1/users/login"):        true,
		apiRuleCname(userstore.VisitorRole, "GET", "/v1/users/isauthed"):      true,
		apiRuleCname(userstore.VisitorRole, "GET", "/v1/fs/files"):            true,
		apiRuleCname(userstore.VisitorRole, "OPTIONS", "/v1/settings/health"): true,
	}

	return &MultiUsersSvc{
		cfg:        cfg,
		deps:       deps,
		apiACRules: apiACRules,
	}, nil
}

func (h *MultiUsersSvc) Init(adminName, adminPwd string) (string, error) {
	var err error

	userID := "0"
	fsPath := q.HomePath(userID, "/")
	if err = h.deps.FS().MkdirAll(fsPath); err != nil {
		return "", err
	}
	uploadingsPath := q.GetTmpPath(userID, "/")
	if err = h.deps.FS().MkdirAll(uploadingsPath); err != nil {
		return "", err
	}

	// TODO: return "" for being compatible with singleuser service, should remove this
	err = h.deps.Users().Init(adminName, adminPwd)
	return "", err
}

func (h *MultiUsersSvc) IsInited() bool {
	return h.deps.Users().IsInited()
}

type LoginReq struct {
	User string `json:"user"`
	Pwd  string `json:"pwd"`
}

func (h *MultiUsersSvc) Login(c *gin.Context) {
	req := &LoginReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}

	user, err := h.deps.Users().GetUserByName(req.User)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Pwd), []byte(req.Pwd))
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}

	ttl := h.cfg.GrabInt("Users.CookieTTL")
	token, err := h.deps.Token().ToToken(map[string]string{
		q.UserIDParam: fmt.Sprint(user.ID),
		q.UserParam:   user.Name,
		q.RoleParam:   user.Role,
		q.ExpireParam: fmt.Sprintf("%d", time.Now().Unix()+int64(ttl)),
	})
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}

	secure := h.cfg.GrabBool("Users.CookieSecure")
	httpOnly := h.cfg.GrabBool("Users.CookieHttpOnly")
	c.SetCookie(q.TokenCookie, token, ttl, "/", "", secure, httpOnly)

	c.JSON(q.Resp(200))
}

type LogoutReq struct{}

func (h *MultiUsersSvc) Logout(c *gin.Context) {
	// token alreay verified in the authn middleware
	secure := h.cfg.GrabBool("Users.CookieSecure")
	httpOnly := h.cfg.GrabBool("Users.CookieHttpOnly")
	c.SetCookie(q.TokenCookie, "", 0, "/", "", secure, httpOnly)
	c.JSON(q.Resp(200))
}

func (h *MultiUsersSvc) IsAuthed(c *gin.Context) {
	// token alreay verified in the authn middleware
	role := c.MustGet(q.RoleParam).(string)
	if role == userstore.VisitorRole {
		c.JSON(q.ErrResp(c, 401, q.ErrUnauthorized))
		return
	}
	c.JSON(q.Resp(200))
}

type SetPwdReq struct {
	OldPwd string `json:"oldPwd"`
	NewPwd string `json:"newPwd"`
}

func (h *MultiUsersSvc) SetPwd(c *gin.Context) {
	req := &SetPwdReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(q.ErrResp(c, 400, err))
		return
	} else if req.OldPwd == req.NewPwd {
		c.JSON(q.ErrResp(c, 400, errors.New("password is not updated")))
		return
	}

	claims, err := h.getUserInfo(c)
	if err != nil {
		c.JSON(q.ErrResp(c, 401, err))
		return
	}

	uid, err := strconv.ParseUint(claims[q.UserIDParam], 10, 64)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}
	user, err := h.deps.Users().GetUser(uid)
	if err != nil {
		c.JSON(q.ErrResp(c, 401, err))
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Pwd), []byte(req.OldPwd))
	if err != nil {
		c.JSON(q.ErrResp(c, 401, ErrInvalidUser))
		return
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPwd), 10)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, errors.New("fail to set password")))
		return
	}

	err = h.deps.Users().SetPwd(uid, string(newHash))
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}

	c.JSON(q.Resp(200))
}

type AddUserReq struct {
	Name string `json:"name"`
	Pwd  string `json:"pwd"`
	Role string `json:"role"`
}

type AddUserResp struct {
	ID string `json:"id"`
}

func (h *MultiUsersSvc) AddUser(c *gin.Context) {
	req := &AddUserReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(q.ErrResp(c, 400, err))
		return
	}

	// Role and duplicated name will be validated by the store
	var err error
	if err = h.isValidUserName(req.Name); err != nil {
		c.JSON(q.ErrResp(c, 400, err))
		return
	} else if err = h.isValidPwd(req.Pwd); err != nil {
		c.JSON(q.ErrResp(c, 400, err))
		return
	}

	uid := h.deps.ID().Gen()
	pwdHash, err := bcrypt.GenerateFromPassword([]byte(req.Pwd), 10)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}

	// TODO: following operations must be atomic
	// TODO: check if the folders already exists
	userID := c.MustGet(q.UserIDParam).(string)
	homePath := q.HomePath(userID, "/")
	if err = h.deps.FS().MkdirAll(homePath); err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}
	uploadingsPath := q.GetTmpPath(userID, "/")
	if err = h.deps.FS().MkdirAll(uploadingsPath); err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}

	err = h.deps.Users().AddUser(&userstore.User{
		ID:   uid,
		Name: req.Name,
		Pwd:  string(pwdHash),
		Role: req.Role,
	})
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}

	c.JSON(200, &AddUserResp{ID: fmt.Sprint(uid)})
}

type AddRoleReq struct {
	Role string `json:"role"`
}

func (h *MultiUsersSvc) AddRole(c *gin.Context) {
	var err error
	req := &AddRoleReq{}
	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(q.ErrResp(c, 400, err))
		return
	}

	if err = h.isValidRole(req.Role); err != nil {
		c.JSON(q.ErrResp(c, 400, err))
		return
	}

	err = h.deps.Users().AddRole(req.Role)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}

	c.JSON(q.Resp(200))
}

type DelRoleReq struct {
	Role string `json:"role"`
}

func (h *MultiUsersSvc) DelRole(c *gin.Context) {
	var err error
	req := &DelRoleReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(q.ErrResp(c, 400, err))
		return
	}

	if err = h.isValidRole(req.Role); err != nil {
		c.JSON(q.ErrResp(c, 400, err))
		return
	}

	err = h.deps.Users().DelRole(req.Role)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}

	c.JSON(q.Resp(200))
}

type ListRolesReq struct{}
type ListRolesResp struct {
	Roles map[string]bool `json:"roles"`
}

func (h *MultiUsersSvc) ListRoles(c *gin.Context) {
	roles, err := h.deps.Users().ListRoles()
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}

	c.JSON(200, &ListRolesResp{Roles: roles})
}

func (h *MultiUsersSvc) getUserInfo(c *gin.Context) (map[string]string, error) {
	tokenStr, err := c.Cookie(q.TokenCookie)
	if err != nil {
		return nil, err
	}
	claims, err := h.deps.Token().FromToken(
		tokenStr,
		map[string]string{
			q.UserIDParam: "",
			q.UserParam:   "",
			q.RoleParam:   "",
			q.ExpireParam: "",
		},
	)
	if err != nil {
		return nil, err
	} else if claims[q.UserIDParam] == "" || claims[q.UserParam] == "" {
		return nil, ErrInvalidConfig
	}

	return claims, nil
}

func (h *MultiUsersSvc) isValidUserName(userName string) error {
	minUserNameLen := h.cfg.GrabInt("Users.MinUserNameLen")
	if len(userName) < minUserNameLen {
		return errors.New("name is too short")
	}
	return nil
}

func (h *MultiUsersSvc) isValidPwd(pwd string) error {
	minPwdLen := h.cfg.GrabInt("Users.MinPwdLen")
	if len(pwd) < minPwdLen {
		return errors.New("password is too short")
	}
	return nil
}

func (h *MultiUsersSvc) isValidRole(role string) error {
	if role == userstore.AdminRole || role == userstore.UserRole || role == userstore.VisitorRole {
		return errors.New("predefined roles can not be added/deleted")
	}
	return h.isValidUserName(role)
}
