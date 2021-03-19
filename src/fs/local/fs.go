package local

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ihexxa/quickshare/src/fs"
)

var ErrTooManyOpens = errors.New("too many opened files")

type LocalFS struct {
	root           string
	defaultPerm    os.FileMode
	defaultDirPerm os.FileMode
	opens          map[string]*fileInfo
	opensLimit     int
	opensMtx       *sync.RWMutex
	opensCleanSize int
	openTTL        time.Duration
	readers        map[string]*fileInfo
}

type fileInfo struct {
	lastAccess time.Time
	fd         *os.File
}

func NewLocalFS(root string, defaultPerm uint32, opensLimit, openTTL int) *LocalFS {
	if root == "" {
		root = "."
	}

	return &LocalFS{
		root:           root,
		defaultPerm:    os.FileMode(defaultPerm),
		defaultDirPerm: os.FileMode(0775),
		opens:          map[string]*fileInfo{},
		opensLimit:     opensLimit,
		openTTL:        time.Duration(openTTL) * time.Second,
		opensMtx:       &sync.RWMutex{},
		opensCleanSize: 10,
		readers:        map[string]*fileInfo{}, // TODO: track readers and close idles
	}
}

func (fs *LocalFS) Root() string {
	return fs.root
}

// closeOpens assumes that it is called after opensMtx.Lock()
func (fs *LocalFS) closeOpens(closeAll bool, exclude map[string]bool) error {
	batch := fs.opensCleanSize

	var err error
	for key, info := range fs.opens {
		if exclude[key] {
			continue
		}

		if !closeAll && batch <= 0 {
			break
		}
		batch--

		if info.lastAccess.Add(fs.openTTL).Before(time.Now()) {
			delete(fs.opens, key)
			if err = info.fd.Sync(); err != nil {
				return err
			}
			if err := info.fd.Close(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (fs *LocalFS) Sync() error {
	fs.opensMtx.Lock()
	defer fs.opensMtx.Unlock()
	return fs.closeOpens(true, map[string]bool{})
}

// check refers implementation of Dir.Open() in http package
func (fs *LocalFS) translate(name string) (string, error) {
	if filepath.Separator != '/' && strings.ContainsRune(name, filepath.Separator) {
		return "", errors.New("invalid character in file path")
	}
	return filepath.Join(fs.root, filepath.FromSlash(path.Clean("/"+name))), nil
}

func (fs *LocalFS) Create(path string) error {
	fs.opensMtx.Lock()
	defer fs.opensMtx.Unlock()
	if len(fs.opens) > fs.opensLimit {
		err := fs.closeOpens(true, map[string]bool{})
		if err != nil {
			return fmt.Errorf("too many opens and fail to clean: %w", err)
		}
		return ErrTooManyOpens
	}

	fullpath, err := fs.translate(path)
	if err != nil {
		return err
	}

	fd, err := os.OpenFile(fullpath, os.O_CREATE|os.O_RDWR|os.O_EXCL, fs.defaultPerm)
	if err != nil {
		return err
	}

	fs.opens[fullpath] = &fileInfo{
		lastAccess: time.Now(),
		fd:         fd,
	}
	return nil
}

func (fs *LocalFS) MkdirAll(path string) error {
	fullpath, err := fs.translate(path)
	if err != nil {
		return err
	}
	return os.MkdirAll(fullpath, fs.defaultDirPerm)
}

func (fs *LocalFS) Remove(entryPath string) error {
	fullpath, err := fs.translate(entryPath)
	if err != nil {
		return err
	}
	return os.RemoveAll(fullpath)
}

func (fs *LocalFS) Rename(oldpath, newpath string) error {
	fullOldPath, err := fs.translate(oldpath)
	if err != nil {
		return err
	}
	_, err = os.Stat(fullOldPath)
	if err != nil {
		return err
	}

	fullNewPath, err := fs.translate(newpath)
	if err != nil {
		return err
	}

	// avoid replacing existing file/folder
	_, err = os.Stat(fullNewPath)
	if err != nil {
		if os.IsNotExist(err) {
			return os.Rename(fullOldPath, fullNewPath)
		}
		return err
	}
	return os.ErrExist
}

func (fs *LocalFS) ReadAt(path string, b []byte, off int64) (int, error) {
	fullpath, err := fs.translate(path)
	if err != nil {
		return 0, err
	}

	info, err := func() (*fileInfo, error) {
		fs.opensMtx.Lock()
		defer fs.opensMtx.Unlock()

		info, ok := fs.opens[fullpath]
		if !ok {
			if len(fs.opens) > fs.opensLimit {
				return nil, ErrTooManyOpens
			}

			// because the fd may be for other usage, its flag is not set as os.O_RDONLY
			fd, err := os.OpenFile(fullpath, os.O_RDWR, fs.defaultPerm)
			if err != nil {
				return nil, err
			}
			info = &fileInfo{
				fd:         fd,
				lastAccess: time.Now(),
			}
			fs.opens[fullpath] = info
			fs.closeOpens(false, map[string]bool{fullpath: true})
		}

		return info, nil
	}()
	if err != nil {
		return 0, err
	}

	newOffset, err := info.fd.Seek(off, os.SEEK_SET)
	if err != nil {
		return 0, err
	} else if newOffset != off {
		// TODO: will this happen?
		return 0, fmt.Errorf("seek offset (%d) != required(%d)", newOffset, off)
	}

	return info.fd.ReadAt(b, off)
}

func (fs *LocalFS) WriteAt(path string, b []byte, off int64) (int, error) {
	fullpath, err := fs.translate(path)
	if err != nil {
		return 0, err
	}

	info, err := func() (*fileInfo, error) {
		fs.opensMtx.Lock()
		defer fs.opensMtx.Unlock()

		info, ok := fs.opens[fullpath]
		if !ok {
			if len(fs.opens) > fs.opensLimit {
				return nil, ErrTooManyOpens
			}

			// it does NOT create file for writing
			fd, err := os.OpenFile(fullpath, os.O_RDWR, fs.defaultPerm)
			if err != nil {
				return nil, err
			}
			info = &fileInfo{
				fd:         fd,
				lastAccess: time.Now(),
			}
			fs.opens[fullpath] = info
			fs.closeOpens(false, map[string]bool{fullpath: true})
		}

		return info, nil
	}()
	if err != nil {
		return 0, err
	}

	newOffset, err := info.fd.Seek(off, os.SEEK_SET)
	if err != nil {
		return 0, err
	} else if newOffset != off {
		// TODO: will this happen?
		return 0, fmt.Errorf("seek offset (%d) != required(%d)", newOffset, off)
	}

	return info.fd.WriteAt(b, off)
}

func (fs *LocalFS) Stat(path string) (os.FileInfo, error) {
	fullpath, err := fs.translate(path)
	if err != nil {
		return nil, err
	}

	fs.opensMtx.RLock()
	info, ok := fs.opens[fullpath]
	fs.opensMtx.RUnlock()
	if ok {
		return info.fd.Stat()
	}
	return os.Stat(fullpath)
}

func (fs *LocalFS) Close() error {
	fs.opensMtx.Lock()
	defer fs.opensMtx.Unlock()

	var err error
	for filePath, info := range fs.opens {
		err = info.fd.Sync()
		if err != nil {
			return err
		}
		err = info.fd.Close()
		if err != nil {
			return err
		}
		delete(fs.opens, filePath)
	}

	return nil
}

// readers are not tracked by opens
func (fs *LocalFS) GetFileReader(path string) (fs.ReadCloseSeeker, error) {
	fullpath, err := fs.translate(path)
	if err != nil {
		return nil, err
	}

	fd, err := os.OpenFile(fullpath, os.O_RDONLY, fs.defaultPerm)
	if err != nil {
		return nil, err
	}

	fs.readers[fullpath] = &fileInfo{
		fd:         fd,
		lastAccess: time.Now(),
	}
	return fd, nil
}

func (fs *LocalFS) ListDir(path string) ([]os.FileInfo, error) {
	fullpath, err := fs.translate(path)
	if err != nil {
		return nil, err
	}

	return ioutil.ReadDir(fullpath)
}
