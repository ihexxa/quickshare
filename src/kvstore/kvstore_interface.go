package kvstore

import "errors"

var ErrLocked = errors.New("already locked")
var ErrNoLock = errors.New("no lock to unlock")

type IKVStore interface {
	AddNamespace(nsName string) error
	GetBool(key string) (bool, bool)
	SetBool(key string, val bool) error
	DelBool(key string) error
	GetInt(key string) (int, bool)
	SetInt(key string, val int) error
	DelInt(key string) error
	GetInt64(key string) (int64, bool)
	SetInt64(key string, val int64) error
	DelInt64(key string) error
	GetFloat(key string) (float64, bool)
	SetFloat(key string, val float64) error
	DelFloat(key string) error
	GetString(key string) (string, bool)
	SetString(key, val string) error
	DelString(key string) error
	GetStringIn(namespace, key string) (string, bool)
	SetStringIn(namespace, key, val string) error
	TryLock(key string) error
	Unlock(key string) error
}
