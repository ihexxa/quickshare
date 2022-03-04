package boltdbpvd

import "github.com/boltdb/bolt"

type BoltProvider interface {
	Bolt() *bolt.DB
}
