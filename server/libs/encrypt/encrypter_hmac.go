package encrypt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

type HmacEncryptor struct {
	Key []byte
}

func (encryptor *HmacEncryptor) Encrypt(content []byte) string {
	mac := hmac.New(sha256.New, encryptor.Key)
	mac.Write(content)
	return hex.EncodeToString(mac.Sum(nil))
}
