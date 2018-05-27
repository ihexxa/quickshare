package apis

import (
	"fmt"
	"strings"
	"testing"
)

import (
	"github.com/ihexxa/quickshare/server/libs/cfg"
	"github.com/ihexxa/quickshare/server/libs/encrypt"
	"github.com/ihexxa/quickshare/server/libs/httputil"
)

func TestLogin(t *testing.T) {
	conf := cfg.NewConfig()

	type testCase struct {
		Desc        string
		AdminId     string
		AdminPwd    string
		Result      interface{}
		VerifyToken bool
	}

	testCases := []testCase{
		testCase{
			Desc:        "invalid input",
			AdminId:     "",
			AdminPwd:    "",
			Result:      httputil.Err401,
			VerifyToken: false,
		},
		testCase{
			Desc:        "account not match",
			AdminId:     "unknown",
			AdminPwd:    "unknown",
			Result:      httputil.Err401,
			VerifyToken: false,
		},
		testCase{
			Desc:        "succeed to login",
			AdminId:     conf.AdminId,
			AdminPwd:    conf.AdminPwd,
			Result:      httputil.Ok200,
			VerifyToken: true,
		},
	}

	for _, testCase := range testCases {
		srv := NewSrvShare(conf)
		res := &stubWriter{Headers: map[string][]string{}}
		ret := srv.login(testCase.AdminId, testCase.AdminPwd, res)

		if ret != testCase.Result {
			t.Fatalf("login: reponse=%v testCase=%v", ret, testCase.Result)
		}

		// verify cookie (only token.adminid part))
		if testCase.VerifyToken {
			cookieVal := strings.Replace(
				res.Header().Get("Set-Cookie"),
				fmt.Sprintf("%s=", conf.KeyToken),
				"",
				1,
			)

			gotTokenStr := strings.Split(cookieVal, ";")[0]
			token := encrypt.JwtEncrypterMaker(conf.SecretKey)
			token.FromStr(gotTokenStr)
			gotToken, found := token.Get(conf.KeyAdminId)
			if !found || conf.AdminId != gotToken {
				t.Fatalf("login: token admin id unmatch got=%v expect=%v", gotToken, conf.AdminId)
			}
		}

	}
}
