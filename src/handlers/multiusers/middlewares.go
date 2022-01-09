package multiusers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ihexxa/quickshare/src/db/userstore"
	q "github.com/ihexxa/quickshare/src/handlers"
)

var ErrExpired = errors.New("token is expired")

func apiRuleCname(role, method, path string) string {
	return fmt.Sprintf("%s-%s-%s", role, method, path)
}

func (h *MultiUsersSvc) AuthN() gin.HandlerFunc {
	return func(c *gin.Context) {
		enableAuth := h.cfg.GrabBool("Users.EnableAuth")
		claims := map[string]string{
			q.UserIDParam: "",
			q.UserParam:   "",
			q.RoleParam:   userstore.VisitorRole,
			q.ExpireParam: "",
		}

		if enableAuth {
			token, err := c.Cookie(q.TokenCookie)
			if err != nil {
				if err != http.ErrNoCookie {
					c.AbortWithStatusJSON(q.ErrResp(c, 401, err))
					return
				}
				// set default values if no cookie is found
			} else if token != "" {
				claims, err = h.deps.Token().FromToken(token, claims)
				if err != nil {
					c.AbortWithStatusJSON(q.ErrResp(c, 401, err))
					return
				}

				now := time.Now().Unix()
				expire, err := strconv.ParseInt(claims[q.ExpireParam], 10, 64)
				if err != nil {
					c.AbortWithStatusJSON(q.ErrResp(c, 401, err))
					return
				} else if expire <= now {
					c.AbortWithStatusJSON(q.ErrResp(c, 401, ErrExpired))
					return
				}
			}
			// set default values if token is empty
		} else {
			claims[q.UserIDParam] = "0"
			claims[q.UserParam] = "admin"
			claims[q.RoleParam] = userstore.AdminRole
			claims[q.ExpireParam] = ""
		}

		for key, val := range claims {
			c.Set(key, val)
		}
		c.Next()
	}
}

func (h *MultiUsersSvc) APIAccessControl() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.MustGet(q.RoleParam).(string)
		method := c.Request.Method
		accessPath := c.Request.URL.Path

		// we don't lock the map because we only read it
		if h.apiACRules[apiRuleCname(role, method, accessPath)] {
			c.Next()
			return
		} else if accessPath == "/" || // TODO: temporarily allow accessing static resources
			accessPath == "/favicon.ico" ||
			strings.HasPrefix(accessPath, "/static") {
			c.Next()
			return
		}
		c.AbortWithStatusJSON(q.ErrResp(c, 403, q.ErrAccessDenied))
	}
}
