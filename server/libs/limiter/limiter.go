package limiter

type Limiter interface {
	Access(string, int16) bool
}
