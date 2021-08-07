package multiusers

import (
	"bytes"
	"errors"

	"github.com/dchest/captcha"
	"github.com/gin-gonic/gin"

	q "github.com/ihexxa/quickshare/src/handlers"
)

type GetCaptchaIDResp struct {
	CaptchaID string `json:"id"`
}

func (h *MultiUsersSvc) GetCaptchaID(c *gin.Context) {
	captchaID := captcha.New()
	c.JSON(200, &GetCaptchaIDResp{CaptchaID: captchaID})
}

// path: /captchas/imgs?id=xxx
func (h *MultiUsersSvc) GetCaptchaImg(c *gin.Context) {
	captchaID := c.Query(q.CaptchaIDParam)
	if captchaID == "" {
		c.JSON(q.ErrResp(c, 400, errors.New("empty captcha ID")))
		return
	}

	capWidth := h.cfg.IntOr("Users.CaptchaWidth", 256)
	capHeight := h.cfg.IntOr("Users.CaptchaHeight", 64)

	// TODO: improve performance
	buf := new(bytes.Buffer)
	err := captcha.WriteImage(buf, captchaID, capWidth, capHeight)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}

	c.Data(200, "image/png", buf.Bytes())
}
