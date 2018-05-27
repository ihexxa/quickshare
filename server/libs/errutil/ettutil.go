package errutil

import (
	"os"
	"runtime/debug"
)

import (
	"quickshare/server/libs/logutil"
)

type ErrUtil interface {
	IsErr(err error) bool
	IsFatalErr(err error) bool
	RecoverPanic()
}

func NewErrChecker(logStack bool, logger logutil.LogUtil) ErrUtil {
	return &ErrChecker{logStack: logStack, log: logger}
}

type ErrChecker struct {
	log      logutil.LogUtil
	logStack bool
}

// IsErr checks if error occurs
func (e *ErrChecker) IsErr(err error) bool {
	if err != nil {
		e.log.Printf("Error:%q\n", err)
		if e.logStack {
			e.log.Println(debug.Stack())
		}
		return true
	}
	return false
}

// IsFatalPanic should be used with defer
func (e *ErrChecker) IsFatalErr(fe error) bool {
	if fe != nil {
		e.log.Printf("Panic:%q", fe)
		if e.logStack {
			e.log.Println(debug.Stack())
		}
		os.Exit(1)
	}
	return false
}

// RecoverPanic catchs the panic and logs panic information
func (e *ErrChecker) RecoverPanic() {
	if r := recover(); r != nil {
		e.log.Printf("Recovered:%v", r)
		if e.logStack {
			e.log.Println(debug.Stack())
		}
	}
}
