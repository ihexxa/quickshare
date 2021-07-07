package multiusers

import (
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ihexxa/gocfg"
	"golang.org/x/crypto/bcrypt"

	"github.com/ihexxa/quickshare/src/depidx"
	q "github.com/ihexxa/quickshare/src/handlers"
	"github.com/ihexxa/quickshare/src/kvstore"
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
	PwdsNs           = "pwds"
	RolesNs          = "roles"
)

type User struct {
	ID   uint64
	Name string
	Pwd  string
	Role string
}

type UserStore interface {
	AddUser(user *User) error
	GetUser(id uint64) (*User, error)
	SetPwd(id uint64, pwd string) error
	SetRole(id uint64, role string) error
}

type KVUserStore struct {
	store kvstore.IKVStore
}

func NewKVUserStore(store kvstore.IKVStore, rootName, rootPwd string) (*KVUserStore, error) {
	_, ok := store.GetStringIn(InitNs, InitTimeParam)
	if !ok {
		var err error
		if err = store.AddNamespace(UsersNs); err != nil {
			return nil, err
		}
		if err = store.AddNamespace(PwdsNs); err != nil {
			return nil, err
		}
		if err = store.AddNamespace(RolesNs); err != nil {
			return nil, err
		}
		if err = store.AddNamespace(InitNs); err != nil {
			return nil, err
		}
		if err = store.SetStringIn(InitNs, InitTimeParam, fmt.Sprintf("%d", time.Now().Unix())); err != nil {
			return nil, err
		}
	}

	userStore := &KVUserStore{store: store}
	err := userStore.AddUser(&User{
		ID:   0,
		Name: rootName,
		Pwd:  rootPwd,
		Role: AdminRole,
	})

	return userStore, err
}

func (us *KVUserStore) AddUser(user *User) error {
	userID := fmt.Sprint(user.ID)
	_, ok := us.store.GetStringIn(UsersNs, userID)
	if ok {
		return fmt.Errorf("userID (%d) exists", user.ID)
	}
	if user.Name == "" || user.Pwd == "" {
		return errors.New("user name or password can not be empty")
	}

	var err error
	err = us.store.SetStringIn(UsersNs, userID, user.Name)
	if err != nil {
		return err
	}
	err = us.SetPwd(user.ID, user.Pwd)
	if err != nil {
		return err
	}
	return us.SetRole(user.ID, user.Role)
}

func (us *KVUserStore) GetUser(id uint64) (*User, error) {
	userID := fmt.Sprint(id)

	name, ok := us.store.GetStringIn(UsersNs, userID)
	if !ok {
		return nil, fmt.Errorf("name (%s) not found", userID)
	}
	pwd, ok := us.store.GetStringIn(PwdsNs, userID)
	if !ok {
		return nil, fmt.Errorf("pwd (%s) not found", userID)
	}
	role, ok := us.store.GetStringIn(RolesNs, userID)
	if !ok {
		return nil, fmt.Errorf("role (%s) not found", userID)
	}

	// TODO: use sync.Pool instead
	return &User{
		ID:   id,
		Name: name,
		Pwd:  pwd,
		Role: role,
	}, nil

}

func (us *KVUserStore) SetPwd(id uint64, pwd string) error {
	userID := fmt.Sprint(id)
	_, ok := us.store.GetStringIn(PwdsNs, userID)
	if !ok {
		return fmt.Errorf("Pwd (%d) does not exist", id)
	}

	pwdHash, err := bcrypt.GenerateFromPassword([]byte(pwd), 10)
	if err != nil {
		return err
	}
	return us.store.SetStringIn(PwdsNs, userID, string(pwdHash))
}

func (us *KVUserStore) SetRole(id uint64, role string) error {
	userID := fmt.Sprint(id)
	_, ok := us.store.GetStringIn(RolesNs, userID)
	if !ok {
		return fmt.Errorf("Role (%d) does not exist", id)
	}

	return us.store.SetStringIn(RolesNs, userID, role)
}

type MultiUsersSvc struct {
	cfg  gocfg.ICfg
	deps *depidx.Deps
}

func NewMultiUsersSvc(cfg gocfg.ICfg, deps *depidx.Deps) (*MultiUsersSvc, error) {
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

	return &MultiUsersSvc{
		cfg:  cfg,
		deps: deps,
	}, nil
}

func (h *MultiUsersSvc) IsInited() bool {
	_, ok := h.deps.KV().GetStringIn(InitNs, InitTimeParam)
	return ok
}

type LoginReq struct {
	User string `json:"user"`
	Pwd  string `json:"pwd"`
}

func (h *MultiUsersSvc) checkPwd(user, pwd string) error {
	expectedHash, ok := h.deps.KV().GetStringIn(UsersNs, user)
	if !ok {
		return ErrInvalidConfig
	}

	return bcrypt.CompareHashAndPassword([]byte(expectedHash), []byte(pwd))
}

func (h *MultiUsersSvc) Login(c *gin.Context) {
	req := &LoginReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}

	if err := h.checkPwd(req.User, req.Pwd); err != nil {
		c.JSON(q.ErrResp(c, 500, err))
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

	secure := h.cfg.GrabBool("Users.CookieSecure")
	httpOnly := h.cfg.GrabBool("Users.CookieHttpOnly")
	c.SetCookie(TokenCookie, token, ttl, "/", "", secure, httpOnly)

	c.JSON(q.Resp(200))
}

type LogoutReq struct {
}

func (h *MultiUsersSvc) Logout(c *gin.Context) {
	// token alreay verified in the authn middleware
	secure := h.cfg.GrabBool("Users.CookieSecure")
	httpOnly := h.cfg.GrabBool("Users.CookieHttpOnly")
	c.SetCookie(TokenCookie, "", 0, "/", "", secure, httpOnly)
	c.JSON(q.Resp(200))
}

func (h *MultiUsersSvc) IsAuthed(c *gin.Context) {
	// token alreay verified in the authn middleware
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

	expectedHash, ok := h.deps.KV().GetStringIn(UsersNs, claims[UserParam])
	if !ok {
		c.JSON(q.ErrResp(c, 500, ErrInvalidConfig))
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(expectedHash), []byte(req.OldPwd))
	if err != nil {
		c.JSON(q.ErrResp(c, 401, ErrInvalidUser))
		return
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPwd), 10)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, errors.New("fail to set password")))
		return
	}
	err = h.deps.KV().SetStringIn(UsersNs, claims[UserParam], string(newHash))
	if err != nil {
		c.JSON(q.ErrResp(c, 500, ErrInvalidConfig))
		return
	}

	c.JSON(q.Resp(200))
}

func (h *MultiUsersSvc) getUserInfo(c *gin.Context) (map[string]string, error) {
	tokenStr, err := c.Cookie(TokenCookie)
	if err != nil {
		return nil, err
	}
	claims, err := h.deps.Token().FromToken(
		tokenStr,
		map[string]string{
			UserParam:   "",
			RoleParam:   "",
			ExpireParam: "",
		},
	)
	if err != nil {
		return nil, err
	} else if claims[UserParam] == "" {
		return nil, ErrInvalidConfig
	}

	return claims, nil
}
