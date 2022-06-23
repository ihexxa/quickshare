package fileindex

import (
	qradix "github.com/ihexxa/q-radix/v3"
)

type FileTreeIndex struct {
	*qradix.RTree
}

func NewFileTreeIndex() *FileTreeIndex {
	return &FileTreeIndex{
		RTree: qradix.NewRTree(),
	}
}

type IFileIndex interface {
	FromString(input chan string) error
	Get(key string) (interface{}, error)
	GetAllPrefixMatches(key string) map[string]interface{}
	GetBestMatch(key string) (string, interface{}, bool)
	Insert(key string, val interface{}) (interface{}, error)
	Remove(key string) bool
	Size() int
	String() chan string
}
