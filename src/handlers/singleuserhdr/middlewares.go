package singleuserhdr

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	q "github.com/ihexxa/quickshare/src/handlers"
)

var exposedAPIs = map[string]bool{
	"Login-fm":  true,
	"Health-fm": true,
}

var publicRootPath = "/"
var publicStaticPath = "/static"

func IsPublicPath(accessPath string) bool {
	return accessPath == publicRootPath || strings.HasPrefix(accessPath, publicStaticPath)
}

func GetHandlerName(fullname string) (string, error) {
	parts := strings.Split(fullname, ".")
	if len(parts) == 0 {
		return "", errors.New("invalid handler name")
	}
	return parts[len(parts)-1], nil
}

func (h *SimpleUserHandlers) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		handlerName, err := GetHandlerName(c.HandlerName())
		if err != nil {
			c.JSON(q.ErrResp(c, 401, err))
			return
		}
		accessPath := c.Request.URL.String()

		enableAuth := h.cfg.GrabBool("Users.EnableAuth")
		if enableAuth && !exposedAPIs[handlerName] && !IsPublicPath(accessPath) {
			token, err := c.Cookie(TokenCookie)
			if err != nil {
				c.AbortWithStatusJSON(q.ErrResp(c, 401, err))
				return
			}

			claims := map[string]string{
				UserParam:   "",
				RoleParam:   "",
				ExpireParam: "",
			}

			_, err = h.deps.Token().FromToken(token, claims)
			if err != nil {
				c.AbortWithStatusJSON(q.ErrResp(c, 401, err))
				return
			}

			now := time.Now().Unix()
			expire, err := strconv.ParseInt(claims[ExpireParam], 10, 64)
			if err != nil || expire <= now {
				c.AbortWithStatusJSON(q.ErrResp(c, 401, err))
				return
			}

			// visitor is only allowed to download
			if claims[RoleParam] != AdminRole && handlerName != "Download-fm" {
				c.AbortWithStatusJSON(q.ErrResp(c, 401, errors.New("not allowed")))
				return
			}
		}

		c.Next()
	}
}
