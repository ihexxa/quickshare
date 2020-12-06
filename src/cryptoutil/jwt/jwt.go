package jwt

import (
	"errors"

	jwtpkg "github.com/robbert229/jwt"
)

type JWTEncDec struct {
	alg jwtpkg.Algorithm
}

func NewJWTEncDec(secret string) *JWTEncDec {
	return &JWTEncDec{
		alg: jwtpkg.HmacSha256(secret),
	}
}

func (ed *JWTEncDec) FromToken(token string, kvs map[string]string) (map[string]string, error) {
	claims, err := ed.alg.Decode(token)
	if err != nil {
		return nil, err
	}

	for key := range kvs {
		iVal, err := claims.Get(key)
		if err != nil {
			return nil, err
		}
		strVal, ok := iVal.(string)
		if !ok {
			return nil, errors.New("incorrect JWT claim")
		}

		kvs[key] = strVal
	}
	return kvs, nil
}

func (ed *JWTEncDec) ToToken(kvs map[string]string) (string, error) {
	claims := jwtpkg.NewClaim()
	for key, val := range kvs {
		claims.Set(key, val)
	}

	token, err := ed.alg.Encode(claims)
	if err != nil {
		return "", err
	}
	return token, nil
}
