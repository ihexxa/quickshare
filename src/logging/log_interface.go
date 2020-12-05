package logging

type ILogger interface {
	Debug()
	Log(values ...interface{})
	Logf(pattern string, values ...interface{})
	Error(values ...interface{})
	Errorf(pattern string, values ...interface{})
}
