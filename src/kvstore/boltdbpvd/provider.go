package boltdbpvd

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"path"
	"time"

	"github.com/boltdb/bolt"

	"github.com/ihexxa/quickshare/src/kvstore"
)

var (
	ErrBucketNotFound = errors.New("bucket not found")
)

type BoltPvd struct {
	dbPath    string
	db        *bolt.DB
	maxStrLen int
}

func New(dbPath string, maxStrLen int) *BoltPvd {
	boltPath := path.Clean(dbPath)
	db, err := bolt.Open(boltPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		panic(err)
	}

	buckets := []string{"bools", "ints", "int64s", "floats", "strings", "locks"}
	for _, bucketName := range buckets {
		// TODO: should return err
		db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(bucketName))
			if b != nil {
				return nil
			}

			_, err := tx.CreateBucket([]byte(bucketName))
			if err != nil {
				panic(err)
			}
			return nil
		})
	}

	return &BoltPvd{
		dbPath:    dbPath,
		db:        db,
		maxStrLen: maxStrLen,
	}
}

func (bp *BoltPvd) AddNamespace(nsName string) error {
	return bp.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(nsName))
		if b != nil {
			return nil
		}

		_, err := tx.CreateBucket([]byte(nsName))
		return err
	})
}

func (bp *BoltPvd) DelNamespace(nsName string) error {
	return bp.db.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket([]byte(nsName))
	})
}

func (bp *BoltPvd) HasNamespace(nsName string) bool {
	err := bp.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(nsName))
		if b == nil {
			return ErrBucketNotFound
		}
		return nil
	})

	return err == nil
}

func (bp *BoltPvd) Close() error {
	return bp.db.Close()
}

func (bp *BoltPvd) GetBool(key string) (bool, bool) {
	return bp.GetBoolIn("bools", key)
}

func (bp *BoltPvd) GetBoolIn(ns, key string) (bool, bool) {
	buf, ok := make([]byte, 1), false

	bp.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ns))
		v := b.Get([]byte(key))
		copy(buf, v)
		ok = v != nil
		return nil
	})

	// 1 means true, 0 means false
	return buf[0] == 1, ok
}

func (bp *BoltPvd) SetBool(key string, val bool) error {
	return bp.SetBoolIn("bools", key, val)
}

func (bp *BoltPvd) SetBoolIn(ns, key string, val bool) error {
	var bVal byte = 0
	if val {
		bVal = 1
	}
	return bp.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ns))
		return b.Put([]byte(key), []byte{bVal})
	})
}

func (bp *BoltPvd) DelBool(key string) error {
	return bp.DelBoolIn("bools", key)
}

func (bp *BoltPvd) DelBoolIn(ns, key string) error {
	return bp.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ns))
		return b.Delete([]byte(key))
	})
}

func (bp *BoltPvd) ListBools() (map[string]bool, error) {
	return bp.ListBoolsIn("bools")
}

func (bp *BoltPvd) ListBoolsIn(ns string) (map[string]bool, error) {
	list := map[string]bool{}
	err := bp.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ns))
		if b == nil {
			return ErrBucketNotFound
		}

		b.ForEach(func(k, v []byte) error {
			list[string(k)] = (v[0] == 1)
			return nil
		})
		return nil
	})
	return list, err
}

func (bp *BoltPvd) ListBoolsByPrefixIn(prefix, ns string) (map[string]bool, error) {
	results := map[string]bool{}

	err := bp.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ns)).Cursor()
		if b == nil {
			return ErrBucketNotFound
		}

		prefixBytes := []byte(prefix)
		for k, _ := b.Seek(prefixBytes); k != nil && bytes.HasPrefix(k, prefixBytes); k, _ = b.Next() {
			results[string(k)] = true
		}
		return nil
	})

	return results, err
}

func (bp *BoltPvd) GetInt(key string) (int, bool) {
	x, ok := bp.GetInt64(key)
	return int(x), ok
}

func (bp *BoltPvd) SetInt(key string, val int) error {
	return bp.SetInt64(key, int64(val))
}

func (bp *BoltPvd) DelInt(key string) error {
	return bp.DelInt64(key)
}

func (bp *BoltPvd) GetInt64(key string) (int64, bool) {
	return bp.GetInt64In("int64s", key)
}

func (bp *BoltPvd) GetInt64In(ns, key string) (int64, bool) {
	buf, ok := make([]byte, binary.MaxVarintLen64), false

	bp.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ns))
		if b == nil {
			return nil
		}

		v := b.Get([]byte(key))
		copy(buf, v)
		ok = v != nil
		return nil
	})

	if !ok {
		return 0, false
	}
	x, n := binary.Varint(buf)
	if n < 0 {
		return 0, false
	}
	return x, true
}

func (bp *BoltPvd) SetInt64(key string, val int64) error {
	return bp.SetInt64In("int64s", key, val)
}

func (bp *BoltPvd) SetInt64In(ns, key string, val int64) error {
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutVarint(buf, val)

	return bp.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ns))
		if b == nil {
			return ErrBucketNotFound
		}
		return b.Put([]byte(key), buf[:n])
	})
}

func (bp *BoltPvd) DelInt64(key string) error {
	return bp.DelInt64In("int64s", key)
}

func (bp *BoltPvd) DelInt64In(ns, key string) error {
	return bp.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ns))
		if b == nil {
			return ErrBucketNotFound
		}
		return b.Delete([]byte(key))
	})
}

func (bp *BoltPvd) ListInt64sIn(ns string) (map[string]int64, error) {
	list := map[string]int64{}
	err := bp.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ns))
		if b == nil {
			return ErrBucketNotFound
		}

		b.ForEach(func(k, v []byte) error {
			x, n := binary.Varint(v)
			if n < 0 {
				return fmt.Errorf("fail to parse int64 for key (%s)", k)
			}
			list[string(k)] = x
			return nil
		})
		return nil
	})
	return list, err
}

func float64ToBytes(num float64) []byte {
	buf := make([]byte, 64)
	binary.PutUvarint(buf, math.Float64bits(num))
	return buf
}

func bytesToFloat64(buf []byte) float64 {
	uintVal, _ := binary.Uvarint(buf[:64])
	return math.Float64frombits(uintVal)
}

func (bp *BoltPvd) GetFloat(key string) (float64, bool) {
	buf, ok := make([]byte, 64), false
	bp.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("floats"))
		v := b.Get([]byte(key))
		copy(buf, v)
		ok = v != nil
		return nil
	})
	if !ok {
		return 0.0, false
	}
	return bytesToFloat64(buf), true
}

func (bp *BoltPvd) SetFloat(key string, val float64) error {
	return bp.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("floats"))
		return b.Put([]byte(key), float64ToBytes(val))
	})
}
func (bp *BoltPvd) DelFloat(key string) error {
	return bp.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("floats"))
		return b.Delete([]byte(key))
	})
}

func (bp *BoltPvd) GetString(key string) (string, bool) {
	buf, ok, length := make([]byte, bp.maxStrLen), false, bp.maxStrLen
	bp.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("strings"))
		v := b.Get([]byte(key))
		length = copy(buf, v)
		ok = v != nil
		return nil
	})
	return string(buf[:length]), ok
}

func (bp *BoltPvd) SetString(key string, val string) error {
	if len(val) > bp.maxStrLen {
		return fmt.Errorf("can not set string value longer than %d", bp.maxStrLen)
	}

	return bp.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("strings"))
		return b.Put([]byte(key), []byte(val))
	})
}

func (bp *BoltPvd) DelString(key string) error {
	return bp.DelStringIn("strings", key)
}

func (bp *BoltPvd) DelStringIn(ns, key string) error {
	return bp.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ns))
		if b == nil {
			return ErrBucketNotFound
		}
		return b.Delete([]byte(key))
	})
}

func (bp *BoltPvd) TryLock(key string) error {
	return bp.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("locks"))
		if b.Get([]byte(key)) != nil {
			return kvstore.ErrLocked
		}
		return b.Put([]byte(key), []byte{})
	})
}

func (bp *BoltPvd) Unlock(key string) error {
	return bp.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("locks"))
		if b.Get([]byte(key)) != nil {
			return b.Delete([]byte(key))
		}
		return kvstore.ErrNoLock
	})
}

func (bp *BoltPvd) GetStringIn(ns, key string) (string, bool) {
	buf, ok, length := make([]byte, bp.maxStrLen), false, bp.maxStrLen
	bp.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ns))
		if b == nil {
			return nil
		}

		v := b.Get([]byte(key))
		length = copy(buf, v)
		ok = v != nil
		return nil
	})

	return string(buf[:length]), ok
}

func (bp *BoltPvd) SetStringIn(ns, key, val string) error {
	if len(val) > bp.maxStrLen {
		return fmt.Errorf("can not set string value longer than %d", bp.maxStrLen)
	}

	return bp.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ns))
		if b == nil {
			return ErrBucketNotFound
		}

		return b.Put([]byte(key), []byte(val))
	})
}

func (bp *BoltPvd) ListStringsIn(ns string) (map[string]string, error) {
	kv := map[string]string{}
	err := bp.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ns))
		if b == nil {
			return ErrBucketNotFound
		}

		b.ForEach(func(k, v []byte) error {
			kv[string(k)] = string(v)
			return nil
		})
		return nil
	})

	return kv, err
}

func (bp *BoltPvd) ListStringsByPrefixIn(prefix, ns string) (map[string]string, error) {
	results := map[string]string{}

	err := bp.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ns)).Cursor()
		if b == nil {
			return ErrBucketNotFound
		}

		prefixBytes := []byte(prefix)
		for k, v := b.Seek(prefixBytes); k != nil && bytes.HasPrefix(k, prefixBytes); k, v = b.Next() {
			results[string(k)] = string(v)
		}
		return nil
	})

	return results, err
}

func (bp *BoltPvd) StartUpdateTxBolt(op func(*bolt.Tx) error) error {
	return bp.db.Update(op)
}
