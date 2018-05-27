package encrypt

import (
	"github.com/robbert229/jwt"
)

func JwtEncrypterMaker(secret string) TokenEncrypter {
	return &JwtEncrypter{
		alg:    jwt.HmacSha256(secret),
		claims: jwt.NewClaim(),
	}
}

type JwtEncrypter struct {
	alg    jwt.Algorithm
	claims *jwt.Claims
}

func (encrypter *JwtEncrypter) Add(key string, value string) bool {
	encrypter.claims.Set(key, value)
	return true
}

func (encrypter *JwtEncrypter) FromStr(token string) bool {
	claims, err := encrypter.alg.Decode(token)
	// TODO: should return error or error info will lost
	if err != nil {
		return false
	}

	encrypter.claims = claims
	return true
}

func (encrypter *JwtEncrypter) Get(key string) (string, bool) {
	iValue, err := encrypter.claims.Get(key)
	// TODO: should return error or error info will lost
	if err != nil {
		return "", false
	}

	return iValue.(string), true
}

func (encrypter *JwtEncrypter) ToStr() (string, bool) {
	token, err := encrypter.alg.Encode(encrypter.claims)

	// TODO: should return error or error info will lost
	if err != nil {
		return "", false
	}
	return token, true
}
