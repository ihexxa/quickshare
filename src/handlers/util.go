package handlers

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"github.com/ihexxa/quickshare/src/cryptoutil"
)

var (
	// dirs
	UploadDir = "uploadings"
	FsDir     = "files"
	FsRootDir = "files"

	UserIDParam    = "uid"
	UserParam      = "user"
	PwdParam       = "pwd"
	NewPwdParam    = "newpwd"
	RoleParam      = "role"
	ExpireParam    = "expire"
	CaptchaIDParam = "capid"
	TokenCookie    = "tk"
	LastID         = "lid"

	// DownloadChunkSize can not be greater than limiter's token count
	// downloadSpeedLimit can not be lower than DownloadChunkSize
	DownloadChunkSize = 100 * 1024

	ErrAccessDenied = errors.New("access denied")
	ErrUnauthorized = errors.New("unauthorized")
)

var statusCodes = map[int]string{
	100: "Continue",                      // RFC 7231, 6.2.1
	101: "SwitchingProtocols",            // RFC 7231, 6.2.2
	102: "Processing",                    // RFC 2518, 10.1
	103: "EarlyHints",                    // RFC 8297
	200: "OK",                            // RFC 7231, 6.3.1
	201: "Created",                       // RFC 7231, 6.3.2
	202: "Accepted",                      // RFC 7231, 6.3.3
	203: "NonAuthoritativeInfo",          // RFC 7231, 6.3.4
	204: "NoContent",                     // RFC 7231, 6.3.5
	205: "ResetContent",                  // RFC 7231, 6.3.6
	206: "PartialContent",                // RFC 7233, 4.1
	207: "MultiStatus",                   // RFC 4918, 11.1
	208: "AlreadyReported",               // RFC 5842, 7.1
	226: "IMUsed",                        // RFC 3229, 10.4.1
	300: "MultipleChoices",               // RFC 7231, 6.4.1
	301: "MovedPermanently",              // RFC 7231, 6.4.2
	302: "Found",                         // RFC 7231, 6.4.3
	303: "SeeOther",                      // RFC 7231, 6.4.4
	304: "NotModified",                   // RFC 7232, 4.1
	305: "UseProxy",                      // RFC 7231, 6.4.5
	307: "TemporaryRedirect",             // RFC 7231, 6.4.7
	308: "PermanentRedirect",             // RFC 7538, 3
	400: "BadRequest",                    // RFC 7231, 6.5.1
	401: "Unauthorized",                  // RFC 7235, 3.1
	402: "PaymentRequired",               // RFC 7231, 6.5.2
	403: "Forbidden",                     // RFC 7231, 6.5.3
	404: "NotFound",                      // RFC 7231, 6.5.4
	405: "MethodNotAllowed",              // RFC 7231, 6.5.5
	406: "NotAcceptable",                 // RFC 7231, 6.5.6
	407: "ProxyAuthRequired",             // RFC 7235, 3.2
	408: "RequestTimeout",                // RFC 7231, 6.5.7
	409: "Conflict",                      // RFC 7231, 6.5.8
	410: "Gone",                          // RFC 7231, 6.5.9
	411: "LengthRequired",                // RFC 7231, 6.5.10
	412: "PreconditionFailed",            // RFC 7232, 4.2
	413: "RequestEntityTooLarge",         // RFC 7231, 6.5.11
	414: "RequestURITooLong",             // RFC 7231, 6.5.12
	415: "UnsupportedMediaType",          // RFC 7231, 6.5.13
	416: "RequestedRangeNotSatisfiable",  // RFC 7233, 4.4
	417: "ExpectationFailed",             // RFC 7231, 6.5.14
	418: "Teapot",                        // RFC 7168, 2.3.3
	421: "MisdirectedRequest",            // RFC 7540, 9.1.2
	422: "UnprocessableEntity",           // RFC 4918, 11.2
	423: "Locked",                        // RFC 4918, 11.3
	424: "FailedDependency",              // RFC 4918, 11.4
	425: "TooEarly",                      // RFC 8470, 5.2.
	426: "UpgradeRequired",               // RFC 7231, 6.5.15
	428: "PreconditionRequired",          // RFC 6585, 3
	429: "TooManyRequests",               // RFC 6585, 4
	431: "RequestHeaderFieldsTooLarge",   // RFC 6585, 5
	451: "UnavailableForLegalReasons",    // RFC 7725, 3
	500: "InternalServerError",           // RFC 7231, 6.6.1
	501: "NotImplemented",                // RFC 7231, 6.6.2
	502: "BadGateway",                    // RFC 7231, 6.6.3
	503: "ServiceUnavailable",            // RFC 7231, 6.6.4
	504: "GatewayTimeout",                // RFC 7231, 6.6.5
	505: "HTTPVersionNotSupported",       // RFC 7231, 6.6.6
	506: "VariantAlsoNegotiates",         // RFC 2295, 8.1
	507: "InsufficientStorage",           // RFC 4918, 11.5
	508: "LoopDetected",                  // RFC 5842, 7.2
	510: "NotExtended",                   // RFC 2774, 7
	511: "NetworkAuthenticationRequired", // RFC 6585, 6
}

type MsgResp struct {
	Msg string `json:"msg"`
}

func NewMsgResp(code int, msg string) (int, interface{}) {
	_, ok := statusCodes[code]
	if !ok {
		panic(fmt.Sprintf("status code not found %d", code))
	}
	return code, &MsgResp{Msg: msg}

}

func Resp(code int) (int, interface{}) {
	msg, ok := statusCodes[code]
	if !ok {
		panic(fmt.Sprintf("status code not found %d", code))
	}
	return code, &MsgResp{Msg: msg}

}

func ErrResp(c *gin.Context, code int, err error) (int, interface{}) {
	_, ok := statusCodes[code]
	if !ok {
		panic(fmt.Sprintf("status code not found %d", code))
	}
	gErr := c.Error(err)
	return code, gErr.JSON()

}

func FsRootPath(userName, relFilePath string) string {
	return filepath.Join(userName, FsRootDir, relFilePath)
}

func UploadPath(userName, relFilePath string) string {
	return filepath.Join(UploadFolder(userName), fmt.Sprintf("%x", sha1.Sum([]byte(relFilePath))))
}

func UploadFolder(userName string) string {
	return filepath.Join(userName, UploadDir)
}

func GetUserInfo(tokenStr string, tokenEncDec cryptoutil.ITokenEncDec) (map[string]string, error) {
	claims, err := tokenEncDec.FromToken(
		tokenStr,
		map[string]string{
			UserIDParam: "",
			UserParam:   "",
			RoleParam:   "",
			ExpireParam: "",
		},
	)
	if err != nil {
		return nil, err
	} else if claims[UserIDParam] == "" || claims[UserParam] == "" {
		return nil, errors.New("empty user id or name")
	}

	return claims, nil
}
