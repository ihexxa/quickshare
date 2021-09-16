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
	"github.com/ihexxa/quickshare/src/idgen"
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
	readerTTL      time.Duration
	readers        map[string]*fileInfo
	ider           idgen.IIDGen
}

type fileInfo struct {
	lastAccess time.Time
	fd         *os.File
}

func NewLocalFS(root string, defaultPerm uint32, opensLimit, openTTL, readerTTL int, ider idgen.IIDGen) *LocalFS {
	if root == "" {
		root = "."
	}

	return &LocalFS{
		ider:           ider,
		root:           root,
		defaultPerm:    os.FileMode(defaultPerm),
		defaultDirPerm: os.FileMode(0775),
		opens:          map[string]*fileInfo{},
		opensLimit:     opensLimit,
		openTTL:        time.Duration(openTTL) * time.Second,
		readerTTL:      time.Duration(readerTTL) * time.Second,
		opensMtx:       &sync.RWMutex{},
		opensCleanSize: 3,
		readers:        map[string]*fileInfo{}, // TODO: track readers and close idles
	}
}

func (fs *LocalFS) Root() string {
	return fs.root
}

// should be protected by opensMtx
func (fs *LocalFS) isTooManyOpens() bool {
	return len(fs.opens)+len(fs.readers) > fs.opensLimit
}

// closeOpens assumes that it is called after opensMtx.Lock()
func (fs *LocalFS) closeOpens(iterateAll, forced bool, exclude map[string]bool) (int, error) {
	batch := fs.opensCleanSize

	var err error
	closed := 0
	for filePath, info := range fs.opens {
		if !iterateAll && exclude[filePath] {
			continue
		}

		if !iterateAll && batch <= 0 {
			break
		}
		batch--

		if forced || info.lastAccess.Add(fs.openTTL).Before(time.Now()) {
			if err = fs.closeInfo(filePath, info); err != nil {
				return closed, err
			}
			closed++
		}
	}

	batch = fs.opensCleanSize
	for id, info := range fs.readers {
		if !iterateAll && exclude[id] {
			continue
		}

		if !iterateAll && batch <= 0 {
			break
		}
		batch--

		if forced || info.lastAccess.Add(fs.readerTTL).Before(time.Now()) {
			var err error
			if err = info.fd.Sync(); err != nil {
				return closed, err
			}
			if err := info.fd.Close(); err != nil {
				return closed, err
			}
			delete(fs.readers, id)
			closed++
		}
	}

	return closed, nil
}

func (fs *LocalFS) closeInfo(key string, info *fileInfo) error {
	var err error
	if err = info.fd.Sync(); err != nil {
		return err
	}
	if err := info.fd.Close(); err != nil {
		return err
	}
	delete(fs.opens, key)
	return nil
}

func (fs *LocalFS) Sync() error {
	fs.opensMtx.Lock()
	defer fs.opensMtx.Unlock()

	var err error
	for _, info := range fs.opens {
		if err = info.fd.Sync(); err != nil {
			return err
		}
	}

	for _, info := range fs.readers {
		if err = info.fd.Sync(); err != nil {
			return err
		}
	}

	return nil
}

func (fs *LocalFS) Close() error {
	fs.opensMtx.Lock()
	defer fs.opensMtx.Unlock()

	_, err := fs.closeOpens(true, true, map[string]bool{})
	return err
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

	if fs.isTooManyOpens() {
		closed, err := fs.closeOpens(false, false, map[string]bool{})
		if err != nil || closed == 0 {
			return fmt.Errorf("too many opens and fail to clean(%d): %w", closed, err)
		}
	}

	fullpath, err := fs.translate(path)
	if err != nil {
		return err
	}

	_, ok := fs.opens[fullpath]
	if ok {
		return os.ErrExist
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
	// TODO: if the rename will be implemented without Rename
	// we must check if the files are in reading/writing
	fullOldPath, err := fs.translate(oldpath)
	if err != nil {
		return err
	}
	fullNewPath, err := fs.translate(newpath)
	if err != nil {
		return err
	}
	if fullOldPath == fullNewPath {
		return nil
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
			if fs.isTooManyOpens() {
				closed, err := fs.closeOpens(false, false, map[string]bool{})
				if err != nil || closed == 0 {
					return nil, fmt.Errorf("too many opens and fail to clean (%d): %w", closed, err)
				}
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
		} else {
			info.lastAccess = time.Now()
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
			if fs.isTooManyOpens() {
				closed, err := fs.closeOpens(false, false, map[string]bool{})
				if err != nil || closed == 0 {
					// TODO: return Eagain and make client retry later
					return nil, fmt.Errorf("too many opens and fail to clean (%d): %w", closed, err)
				}
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
		} else {
			info.lastAccess = time.Now()
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

// readers are not tracked by opens
func (fs *LocalFS) GetFileReader(path string) (fs.ReadCloseSeeker, uint64, error) {
	fullpath, err := fs.translate(path)
	if err != nil {
		return nil, 0, err
	}

	fs.opensMtx.Lock()
	defer fs.opensMtx.Unlock()

	if fs.isTooManyOpens() {
		closed, err := fs.closeOpens(false, false, map[string]bool{})
		if err != nil || closed == 0 {
			return nil, 0, fmt.Errorf("too many opens and fail to clean (%d): %w", closed, err)
		}
	}

	fd, err := os.OpenFile(fullpath, os.O_RDONLY, fs.defaultPerm)
	if err != nil {
		return nil, 0, err
	}

	id := fs.ider.Gen()
	fs.readers[fmt.Sprint(id)] = &fileInfo{
		fd:         fd,
		lastAccess: time.Now(),
	}

	return fd, id, nil
}

func (fs *LocalFS) CloseReader(id string) error {
	fs.opensMtx.Lock()
	defer fs.opensMtx.Unlock()

	info, ok := fs.readers[id]
	if !ok {
		return fmt.Errorf("reader not found: %s %v", id, fs.readers)
	}

	var err error
	if err = info.fd.Sync(); err != nil {
		return err
	}
	if err := info.fd.Close(); err != nil {
		return err
	}
	delete(fs.readers, id)
	return nil
}

func (fs *LocalFS) ListDir(path string) ([]os.FileInfo, error) {
	fullpath, err := fs.translate(path)
	if err != nil {
		return nil, err
	}

	return ioutil.ReadDir(fullpath)
}
