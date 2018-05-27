package encrypt

type EncrypterMaker func(string) TokenEncrypter

// TODO: name should be Encrypter?
type TokenEncrypter interface {
	Add(string, string) bool
	FromStr(string) bool
	Get(string) (string, bool)
	ToStr() (string, bool)
}
