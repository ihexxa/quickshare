package fsutil

import (
	"errors"
	"io"
	"os"
)

import (
	"quickshare/server/libs/errutil"
	"quickshare/server/libs/fileidx"
	"quickshare/server/libs/qtube"
)

type FsUtil interface {
	CreateFile(fullPath string) error
	CopyChunkN(fullPath string, chunk io.Reader, start int64, length int64) bool
	DelFile(fullPath string) bool
	Open(fullPath string) (qtube.ReadSeekCloser, error)
	MkdirAll(path string, mode os.FileMode) bool
	Readdir(dirName string, n int) ([]*fileidx.FileInfo, error)
}

func NewSimpleFs(errUtil errutil.ErrUtil) FsUtil {
	return &SimpleFs{
		Err: errUtil,
	}
}

type SimpleFs struct {
	Err errutil.ErrUtil
}

var (
	ErrExists  = errors.New("file exists")
	ErrUnknown = errors.New("unknown error")
)

func (sfs *SimpleFs) CreateFile(fullPath string) error {
	flag := os.O_CREATE | os.O_EXCL | os.O_RDONLY
	perm := os.FileMode(0644)
	newFile, err := os.OpenFile(fullPath, flag, perm)
	defer newFile.Close()

	if err == nil {
		return nil
	} else if os.IsExist(err) {
		return ErrExists
	} else {
		return ErrUnknown
	}
}

func (sfs *SimpleFs) CopyChunkN(fullPath string, chunk io.Reader, start int64, length int64) bool {
	flag := os.O_WRONLY
	perm := os.FileMode(0644)
	file, openErr := os.OpenFile(fullPath, flag, perm)

	defer file.Close()
	if sfs.Err.IsErr(openErr) {
		return false
	}

	if _, err := file.Seek(start, io.SeekStart); sfs.Err.IsErr(err) {
		return false
	}

	if _, err := io.CopyN(file, chunk, length); sfs.Err.IsErr(err) && err != io.EOF {
		return false
	}

	return true
}

func (sfs *SimpleFs) DelFile(fullPath string) bool {
	return !sfs.Err.IsErr(os.Remove(fullPath))
}

func (sfs *SimpleFs) MkdirAll(path string, mode os.FileMode) bool {
	err := os.MkdirAll(path, mode)
	return !sfs.Err.IsErr(err)
}

// TODO: not support read from last seek position
func (sfs *SimpleFs) Readdir(dirName string, n int) ([]*fileidx.FileInfo, error) {
	dir, openErr := os.Open(dirName)
	defer dir.Close()

	if sfs.Err.IsErr(openErr) {
		return []*fileidx.FileInfo{}, openErr
	}

	osFileInfos, readErr := dir.Readdir(n)
	if sfs.Err.IsErr(readErr) && readErr != io.EOF {
		return []*fileidx.FileInfo{}, readErr
	}

	fileInfos := make([]*fileidx.FileInfo, 0)
	for _, osFileInfo := range osFileInfos {
		if osFileInfo.Mode().IsRegular() {
			fileInfos = append(
				fileInfos,
				&fileidx.FileInfo{
					ModTime:   osFileInfo.ModTime().UnixNano(),
					PathLocal: osFileInfo.Name(),
					Uploaded:  osFileInfo.Size(),
				},
			)
		}
	}

	return fileInfos, readErr
}

// the associated file descriptor has mode O_RDONLY as using os.Open
func (sfs *SimpleFs) Open(fullPath string) (qtube.ReadSeekCloser, error) {
	return os.Open(fullPath)
}
