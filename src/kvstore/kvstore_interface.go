package kvstore

import (
	"errors"

	"github.com/boltdb/bolt"
)

var ErrLocked = errors.New("already locked")
var ErrNoLock = errors.New("no lock to unlock")

type IKVStore interface {
	AddNamespace(nsName string) error
	DelNamespace(nsName string) error
	HasNamespace(nsName string) bool
	GetBool(key string) (bool, bool)
	GetBoolIn(ns, key string) (bool, bool)
	SetBool(key string, val bool) error
	SetBoolIn(ns, key string, val bool) error
	DelBool(key string) error
	DelBoolIn(ns, key string) error
	ListBools() (map[string]bool, error)
	ListBoolsIn(ns string) (map[string]bool, error)
	ListBoolsByPrefixIn(prefix, ns string) (map[string]bool, error)
	GetInt(key string) (int, bool)
	SetInt(key string, val int) error
	DelInt(key string) error
	GetInt64(key string) (int64, bool)
	SetInt64(key string, val int64) error
	GetInt64In(ns, key string) (int64, bool)
	SetInt64In(ns, key string, val int64) error
	ListInt64sIn(ns string) (map[string]int64, error)
	DelInt64(key string) error
	DelInt64In(ns, key string) error
	GetFloat(key string) (float64, bool)
	SetFloat(key string, val float64) error
	DelFloat(key string) error
	GetString(key string) (string, bool)
	SetString(key, val string) error
	DelString(key string) error
	DelStringIn(ns, key string) error
	GetStringIn(ns, key string) (string, bool)
	SetStringIn(ns, key, val string) error
	ListStringsIn(ns string) (map[string]string, error)
	ListStringsByPrefixIn(prefix, ns string) (map[string]string, error)
	TryLock(key string) error
	Unlock(key string) error
	Close() error
	Bolt() *bolt.DB
}
