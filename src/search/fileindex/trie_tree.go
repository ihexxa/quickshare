package fileindex

import (
	"strings"

	qradix "github.com/ihexxa/q-radix/v3"

	"github.com/ihexxa/quickshare/src/fs"
)

// type IFileIndex interface {
// 	FromString(input chan string) error
// 	Get(key string) (interface{}, error)
// 	GetAllPrefixMatches(key string) map[string]interface{}
// 	GetBestMatch(key string) (string, interface{}, bool)
// 	Insert(key string, val interface{}) (interface{}, error)
// 	Remove(key string) bool
// 	Size() int
// 	String() chan string
// }

type FileTreeIndex struct {
	db   fs.ISimpleFS
	trie *qradix.RTree
}

func NewFileTreeIndex(db fs.ISimpleFS) *FileTreeIndex {
	return &FileTreeIndex{
		db:   db,
		trie: qradix.NewRTree(),
	}
}

func (idx *FileTreeIndex) Search(segment string) []string {
	results := idx.trie.GetAllPrefixMatches(segment)
	paths := []string{}
	for _, iPaths := range results {
		paths = append(paths, iPaths.([]string)...)
	}
	return paths
}

func (idx *FileTreeIndex) Add(path string) error {
	segments := strings.Split(path, "/")
	for _, segment := range segments {
		_, err := idx.trie.Insert(segment, path)
		if err != nil {
			return err
		}
	}
	return nil
}
