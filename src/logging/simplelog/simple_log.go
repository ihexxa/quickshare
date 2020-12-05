package simplelog

import (
	"fmt"
	"log"
)

type SimpleLogger struct {
	debug bool
}

func NewSimpleLogger() *SimpleLogger {
	return &SimpleLogger{}
}

func (l *SimpleLogger) Debug() {
	l.debug = true
}

func (l *SimpleLogger) Log(values ...interface{}) {
	log.Println(values...)
}

func (l *SimpleLogger) Logf(pattern string, values ...interface{}) {
	log.Printf(pattern, values...)
}
func (l *SimpleLogger) Error(values ...interface{}) {
	log.Println(append([]interface{}{"error:"}, values...)...)
}
func (l *SimpleLogger) Errorf(pattern string, values ...interface{}) {
	log.Printf(fmt.Sprintf("error: %s", pattern), values...)
}
