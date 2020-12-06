package singleuserhdr

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	q "github.com/ihexxa/quickshare/src/handlers"
)

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

		// TODO: may also check the path
		enableAuth := h.cfg.GrabBool("Users.EnableAuth")
		if enableAuth && handlerName != "Login-fm" {
			token, err := c.Cookie(TokenCookie)
			if err != nil {
				c.JSON(q.ErrResp(c, 401, err))
				return
			}

			claims := map[string]string{
				UserParam:   "",
				RoleParam:   "",
				ExpireParam: "",
			}

			_, err = h.deps.Token().FromToken(token, claims)
			if err != nil {
				c.JSON(q.ErrResp(c, 401, err))
				return
			}

			now := time.Now().Unix()
			expire, err := strconv.ParseInt(claims[ExpireParam], 10, 64)
			if err != nil || expire <= now {
				c.JSON(q.ErrResp(c, 401, err))
				return
			}

			// visitor is only allowed to download
			if claims[RoleParam] != AdminRole && handlerName != "Download-fm" {
				c.JSON(q.Resp(401))
				return
			}
		}

		c.Next()
	}
}
