package errutil

import (
	"os"

	log "github.com/sirupsen/logrus"
)

type ErrUtil interface {
	IsErr(err error) bool
	IsFatalErr(err error) bool
	RecoverPanic()
}

func NewErrChecker() ErrUtil {
	return &ErrChecker{}
}

type ErrChecker struct {
}

// IsErr checks if error occurs
func (e *ErrChecker) IsErr(err error) bool {
	if err != nil {
		log.Printf("Error:%q\n", err)
		return true
	}
	return false
}

// IsFatalPanic should be used with defer
func (e *ErrChecker) IsFatalErr(fe error) bool {
	if fe != nil {
		log.Printf("Panic:%q", fe)
		os.Exit(1)
	}
	return false
}

// RecoverPanic catchs the panic and logs panic information
func (e *ErrChecker) RecoverPanic() {
	if r := recover(); r != nil {
		log.Printf("Recovered:%v", r)
	}
}
