package cryptoutil

type ITokenEncDec interface {
	FromToken(token string, kvs map[string]string) (map[string]string, error)
	ToToken(kvs map[string]string) (string, error)
}
