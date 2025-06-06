package multiusers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/ihexxa/quickshare/src/db"
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
			q.RoleParam:   db.VisitorRole,
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
			claims[q.RoleParam] = db.AdminRole
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

		if role == db.BannedRole {
			c.AbortWithStatusJSON(q.ErrResp(c, 403, q.ErrAccessDenied))
		}

		// v2 ac control
		matches := h.routeRules.GetAllPrefixMatches(accessPath)
		key := fmt.Sprintf("%s:%s", role, method)
		matched := false
		for _, matchedRules := range matches {
			matchedRuleMap := matchedRules.(map[string]bool)
			if matchedRuleMap[key] {
				matched = true
				break
			}
		}

		// TODO: listDir and download are exceptions: for sharing
		if accessPath == "/v2/my/fs/dirs" {
			matched = true
		}

		if matched {
			c.Next()
			return
		}

		if h.apiACRules[apiRuleCname(role, method, accessPath)] {
			c.Next()
			return
		} else if accessPath == "/" || // TODO: temporarily allow accessing static resources
			accessPath == "/favicon.ico" ||
			strings.HasPrefix(accessPath, "/css") ||
			strings.HasPrefix(accessPath, "/font") ||
			strings.HasPrefix(accessPath, "/img") ||
			strings.HasPrefix(accessPath, "/js") {
			c.Next()
			return
		}
		c.AbortWithStatusJSON(q.ErrResp(c, 403, q.ErrAccessDenied))
	}
}
