package walls

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

import (
	"quickshare/server/libs/cfg"
	"quickshare/server/libs/encrypt"
	"quickshare/server/libs/limiter"
)

type AccessWalls struct {
	cf             *cfg.Config
	IpLimiter      limiter.Limiter
	OpLimiter      limiter.Limiter
	EncrypterMaker encrypt.EncrypterMaker
}

func NewAccessWalls(
	cf *cfg.Config,
	ipLimiter limiter.Limiter,
	opLimiter limiter.Limiter,
	encrypterMaker encrypt.EncrypterMaker,
) Walls {
	return &AccessWalls{
		cf:             cf,
		IpLimiter:      ipLimiter,
		OpLimiter:      opLimiter,
		EncrypterMaker: encrypterMaker,
	}
}

func (walls *AccessWalls) PassIpLimit(remoteAddr string) bool {
	if !walls.cf.Production {
		return true
	}
	return walls.IpLimiter.Access(remoteAddr, walls.cf.OpIdIpVisit)

}

func (walls *AccessWalls) PassOpLimit(resourceId string, opId int16) bool {
	if !walls.cf.Production {
		return true
	}
	return walls.OpLimiter.Access(resourceId, opId)
}

func (walls *AccessWalls) PassLoginCheck(tokenStr string, req *http.Request) bool {
	if !walls.cf.Production {
		return true
	}

	return walls.passLoginCheck(tokenStr)
}

func (walls *AccessWalls) passLoginCheck(tokenStr string) bool {
	token, getLoginTokenOk := walls.GetLoginToken(tokenStr)
	return getLoginTokenOk && token.AdminId == walls.cf.AdminId
}

func (walls *AccessWalls) GetLoginToken(tokenStr string) (*LoginToken, bool) {
	tokenMaker := walls.EncrypterMaker(string(walls.cf.SecretKeyByte))
	if !tokenMaker.FromStr(tokenStr) {
		return nil, false
	}

	adminIdFromToken, adminIdOk := tokenMaker.Get(walls.cf.KeyAdminId)
	expiresStr, expiresStrOk := tokenMaker.Get(walls.cf.KeyExpires)
	if !adminIdOk || !expiresStrOk {
		return nil, false
	}

	expires, expiresParseErr := strconv.ParseInt(expiresStr, 10, 64)
	if expiresParseErr != nil ||
		adminIdFromToken != walls.cf.AdminId ||
		expires <= time.Now().Unix() {
		return nil, false
	}

	return &LoginToken{
		AdminId: adminIdFromToken,
		Expires: expires,
	}, true
}

func (walls *AccessWalls) MakeLoginToken(userId string) string {
	expires := time.Now().Add(time.Duration(walls.cf.CookieMaxAge) * time.Second).Unix()

	tokenMaker := walls.EncrypterMaker(string(walls.cf.SecretKeyByte))
	tokenMaker.Add(walls.cf.KeyAdminId, userId)
	tokenMaker.Add(walls.cf.KeyExpires, fmt.Sprintf("%d", expires))

	tokenStr, ok := tokenMaker.ToStr()
	if !ok {
		return ""
	}
	return tokenStr
}
