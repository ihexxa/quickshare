package fs

import (
	"io"
	"os"
)

type ReadCloseSeeker interface {
	io.Reader
	io.ReaderFrom
	io.Closer
	io.Seeker
}

type ISimpleFS interface {
	Create(path string) error
	MkdirAll(path string) error
	Remove(path string) error
	Rename(oldpath, newpath string) error
	ReadAt(path string, b []byte, off int64) (n int, err error)
	WriteAt(path string, b []byte, off int64) (n int, err error)
	Stat(path string) (os.FileInfo, error)
	Close() error
	Sync() error
	GetFileReader(path string) (ReadCloseSeeker, error)
	CloseReader(path string) error
	Root() string
	ListDir(path string) ([]os.FileInfo, error)
}
