package encrypt

type Encryptor interface {
	Encrypt(content []byte) string
}
