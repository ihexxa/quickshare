package fileindex

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ihexxa/fsearch"
	"github.com/ihexxa/quickshare/src/fs"
)

type IFileIndex interface {
	Search(keyword string) ([]string, error)
	AddPath(pathname string) error
	DelPath(pathname string) error
	RenamePath(pathname, newName string) error
	MovePath(pathname, dstParentPath string) error
	WriteTo(pathname string) error
	ReadFrom(pathname string) error
}

type FileTreeIndex struct {
	fs    fs.ISimpleFS
	index *fsearch.FSearch
}

func NewFileTreeIndex(fs fs.ISimpleFS, pathSeparator string, maxResultSize int) *FileTreeIndex {
	return &FileTreeIndex{
		fs: fs,
		// TODO: support max result size config
		index: fsearch.New(pathSeparator, maxResultSize),
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

func (idx *FileTreeIndex) WriteTo(pathname string) error {
	rowsChan := idx.index.Marshal()
	err := idx.fs.Create(pathname)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}

	var row string
	var ok bool
	var wrote int
	var offset int64
	batch := []string{}
	for {
		row, ok = <-rowsChan
		if !ok || len(batch) > 1024 {
			wrote, err = idx.fs.WriteAt(pathname, []byte(strings.Join(batch, "\n")), offset)
			if err != nil {
				return err
			}
			offset += int64(wrote)
		}
		if !ok {
			break
		}
		batch = append(batch, row)
	}

	return idx.index.Error()
}

func (idx *FileTreeIndex) ReadFrom(pathname string) error {
	f, readerId, err := idx.fs.GetFileReader(pathname)
	if err != nil {
		return err
	}
	defer idx.fs.CloseReader(fmt.Sprint(readerId))

	var row string
	rowSeparator := byte('\n')
	reader := bufio.NewReader(f)
	rowsChan := make(chan string, 1024)
	defer close(rowsChan)
	for {
		row, err = reader.ReadString(rowSeparator)
		if err != nil {
			if errors.Is(err, io.EOF) {
				rowsChan <- row // some content may still read
				break
			} else {
				return err
			}
		}
		rowsChan <- row
	}

	return idx.index.Error()
}
