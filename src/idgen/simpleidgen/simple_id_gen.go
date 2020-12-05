package simpleidgen

import (
	"sync"
	"time"
)

var lastID = uint64(0)
var mux = &sync.Mutex{}

type SimpleIDGen struct{}

func New() *SimpleIDGen {
	return &SimpleIDGen{}
}

func (id *SimpleIDGen) Gen() uint64 {
	mux.Lock()
	defer mux.Unlock()
	newID := uint64(time.Now().UnixNano())
	if newID != lastID {
		lastID = newID
		return lastID
	}
	lastID = newID + 1
	return lastID
}
