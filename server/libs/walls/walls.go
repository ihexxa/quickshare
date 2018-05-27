package walls

import (
	"net/http"
)

type Walls interface {
	PassIpLimit(remoteAddr string) bool
	PassOpLimit(resourceId string, opId int16) bool
	PassLoginCheck(tokenStr string, req *http.Request) bool
	MakeLoginToken(uid string) string
}

type LoginToken struct {
	AdminId string
	Expires int64
}
