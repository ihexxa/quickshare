package fileindex

import (
	// "strings"

	"github.com/ihexxa/fsearch"
	"github.com/ihexxa/quickshare/src/fs"
)

type FileTreeIndex struct {
	db    fs.ISimpleFS
	index *fsearch.FSearch
}

func NewFileTreeIndex(db fs.ISimpleFS) *FileTreeIndex {
	return &FileTreeIndex{
		db: db,
		// TODO: support max result size config
		index: fsearch.New("/", 1024),
	}
}

func (idx *FileTreeIndex) Search(keyword string) ([]string, error) {
	return idx.index.Search(keyword)
}

func (idx *FileTreeIndex) AddPath(pathname string) error {
	return idx.index.AddPath(pathname)
}

func (idx *FileTreeIndex) DelPath(pathname string) error {
	return idx.index.DelPath(pathname)
}

func (idx *FileTreeIndex) RenamePath(pathname, newName string) error {
	return idx.index.RenamePath(pathname, newName)
}

func (idx *FileTreeIndex) MovePath(pathname, dstParentPath string) error {
	return idx.index.MovePath(pathname, dstParentPath)
}
