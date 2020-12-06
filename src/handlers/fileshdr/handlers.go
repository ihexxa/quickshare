package fileshdr

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ihexxa/gocfg"
	"github.com/ihexxa/multipart"

	"github.com/ihexxa/quickshare/src/depidx"
	q "github.com/ihexxa/quickshare/src/handlers"
)

var (
	// dirs
	UploadDir = "uploadings"
	FsDir     = "files"

	// queries
	FilePathQuery = "fp"
	ListDirQuery  = "dp"

	// headers
	rangeHeader       = "Range"
	acceptRangeHeader = "Accept-Range"
	ifRangeHeader     = "If-Range"
)

type FileHandlers struct {
	cfg       gocfg.ICfg
	deps      *depidx.Deps
	uploadMgr *UploadMgr
}

func NewFileHandlers(cfg gocfg.ICfg, deps *depidx.Deps) (*FileHandlers, error) {
	var err error
	if err = deps.FS().MkdirAll(UploadDir); err != nil {
		return nil, err
	}
	if err = deps.FS().MkdirAll(FsDir); err != nil {
		return nil, err
	}

	return &FileHandlers{
		cfg:       cfg,
		deps:      deps,
		uploadMgr: NewUploadMgr(deps.KV()),
	}, err
}

type AutoLocker struct {
	h   *FileHandlers
	c   *gin.Context
	key string
}

func (h *FileHandlers) NewAutoLocker(c *gin.Context, key string) *AutoLocker {
	return &AutoLocker{
		h:   h,
		c:   c,
		key: key,
	}
}

func (lk *AutoLocker) Exec(handler func()) {
	var err error
	kv := lk.h.deps.KV()
	if err = kv.TryLock(lk.key); err != nil {
		lk.c.JSON(q.Resp(500))
		return
	}

	handler()

	if err = kv.Unlock(lk.key); err != nil {
		// TODO: use logger
		fmt.Println(err)
	}
}

type CreateReq struct {
	Path     string `json:"path"`
	FileSize int64  `json:"fileSize"`
}

func (h *FileHandlers) Create(c *gin.Context) {
	req := &CreateReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}

	tmpFilePath := h.GetTmpPath(req.Path)
	locker := h.NewAutoLocker(c, tmpFilePath)
	locker.Exec(func() {
		err := h.deps.FS().Create(tmpFilePath)
		if err != nil {
			c.JSON(q.ErrResp(c, 500, err))
			return
		}
		err = h.uploadMgr.AddInfo(req.Path, tmpFilePath, req.FileSize, false)
		if err != nil {
			c.JSON(q.ErrResp(c, 500, err))
			return
		}

		fileDir := h.FsPath(filepath.Dir(req.Path))
		err = h.deps.FS().MkdirAll(fileDir)
		if err != nil {
			c.JSON(q.ErrResp(c, 500, err))
			return
		}
	})

	c.JSON(q.Resp(200))
}

func (h *FileHandlers) Delete(c *gin.Context) {
	filePath := c.Query(FilePathQuery)
	if filePath == "" {
		c.JSON(q.ErrResp(c, 400, errors.New("invalid file path")))
		return
	}

	filePath = h.FsPath(filePath)
	err := h.deps.FS().Remove(filePath)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}

	c.JSON(q.Resp(200))
}

type MetadataResp struct {
	Name    string    `json:"name"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"modTime"`
	IsDir   bool      `json:"isDir"`
}

func (h *FileHandlers) Metadata(c *gin.Context) {
	filePath := c.Query(FilePathQuery)
	if filePath == "" {
		c.JSON(q.ErrResp(c, 400, errors.New("invalid file path")))
		return
	}

	filePath = h.FsPath(filePath)
	info, err := h.deps.FS().Stat(filePath)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}

	c.JSON(200, MetadataResp{
		Name:    info.Name(),
		Size:    info.Size(),
		ModTime: info.ModTime(),
		IsDir:   info.IsDir(),
	})
}

type MkdirReq struct {
	Path string `json:"path"`
}

func (h *FileHandlers) Mkdir(c *gin.Context) {
	req := &MkdirReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(q.ErrResp(c, 400, err))
		return
	}

	dirPath := h.FsPath(req.Path)
	err := h.deps.FS().MkdirAll(dirPath)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}

	c.JSON(q.Resp(200))
}

type MoveReq struct {
	OldPath string `json:"oldPath"`
	NewPath string `json:"newPath"`
}

func (h *FileHandlers) Move(c *gin.Context) {
	req := &MoveReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(q.ErrResp(c, 400, err))
		return
	}

	oldPath := h.FsPath(req.OldPath)
	newPath := h.FsPath(req.NewPath)
	_, err := h.deps.FS().Stat(oldPath)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}
	_, err = h.deps.FS().Stat(newPath)
	if err != nil && !os.IsNotExist(err) {
		c.JSON(q.ErrResp(c, 500, err))
		return
	} else if err == nil {
		// err is nil because file exists
		c.JSON(q.ErrResp(c, 400, os.ErrExist))
		return
	}

	err = h.deps.FS().Rename(oldPath, newPath)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}

	c.JSON(q.Resp(200))
}

type UploadChunkReq struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Offset  int64  `json:"offset"`
}

func (h *FileHandlers) UploadChunk(c *gin.Context) {
	req := &UploadChunkReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}

	tmpFilePath := h.GetTmpPath(req.Path)
	locker := h.NewAutoLocker(c, tmpFilePath)
	locker.Exec(func() {
		var err error

		_, fileSize, uploaded, err := h.uploadMgr.GetInfo(tmpFilePath)
		if err != nil {
			c.JSON(q.ErrResp(c, 500, err))
			return
		} else if uploaded != req.Offset {
			c.JSON(q.ErrResp(c, 500, errors.New("offset != uploaded")))
			return
		}

		wrote, err := h.deps.FS().WriteAt(tmpFilePath, []byte(req.Content), req.Offset)
		if err != nil {
			c.JSON(q.ErrResp(c, 500, err))
			return
		}
		err = h.uploadMgr.IncreUploaded(tmpFilePath, int64(wrote))
		if err != nil {
			c.JSON(q.ErrResp(c, 500, err))
			return
		}

		// move the file from uploading dir to uploaded dir
		if uploaded+int64(wrote) == fileSize {
			fsFilePath := h.FsPath(req.Path)
			err = h.deps.FS().Rename(tmpFilePath, fsFilePath)
			if err != nil {
				c.JSON(q.ErrResp(c, 500, err))
				return
			}
			err = h.uploadMgr.DelInfo(tmpFilePath)
			if err != nil {
				c.JSON(q.ErrResp(c, 500, err))
				return
			}
		}

		c.JSON(200, &UploadStatusResp{
			Path:     req.Path,
			IsDir:    false,
			FileSize: fileSize,
			Uploaded: uploaded + int64(wrote),
		})
	})
}

type UploadStatusResp struct {
	Path     string `json:"path"`
	IsDir    bool   `json:"isDir"`
	FileSize int64  `json:"fileSize"`
	Uploaded int64  `json:"uploaded"`
}

func (h *FileHandlers) UploadStatus(c *gin.Context) {
	filePath := c.Query(FilePathQuery)
	if filePath == "" {
		c.JSON(q.ErrResp(c, 400, errors.New("invalid file name")))
	}

	tmpFilePath := h.GetTmpPath(filePath)
	locker := h.NewAutoLocker(c, tmpFilePath)
	locker.Exec(func() {
		_, fileSize, uploaded, err := h.uploadMgr.GetInfo(tmpFilePath)
		if err != nil {
			c.JSON(q.ErrResp(c, 500, err))
			return
		}

		c.JSON(200, &UploadStatusResp{
			Path:     filePath,
			IsDir:    false,
			FileSize: fileSize,
			Uploaded: uploaded,
		})
	})
}

// TODO: support ETag
// TODO: use correct content type
func (h *FileHandlers) Download(c *gin.Context) {
	rangeVal := c.GetHeader(rangeHeader)
	ifRangeVal := c.GetHeader(ifRangeHeader)
	filePath := c.Query(FilePathQuery)
	if filePath == "" {
		c.JSON(q.ErrResp(c, 400, errors.New("invalid file name")))
	}

	// concurrency relies on os's mechanism
	filePath = h.FsPath(filePath)
	info, err := h.deps.FS().Stat(filePath)
	if err != nil {
		c.JSON(q.ErrResp(c, 400, err))
		return
	} else if info.IsDir() {
		c.JSON(q.ErrResp(c, 501, errors.New("downloading a folder is not supported")))
	}

	r, err := h.deps.FS().GetFileReader(filePath)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
	}

	// respond to normal requests
	if ifRangeVal != "" || rangeVal == "" {
		c.DataFromReader(200, info.Size(), "application/octet-stream", r, map[string]string{})
		return
	}

	// respond to range requests
	parts, err := multipart.RangeToParts(rangeVal, "application/octet-stream", fmt.Sprintf("%d", info.Size()))
	if err != nil {
		c.JSON(q.ErrResp(c, 400, err))
	}
	pr, pw := io.Pipe()
	err = multipart.WriteResponse(r, pw, filePath, parts)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
	}
	extraHeaders := map[string]string{
		"Content-Disposition": fmt.Sprintf(`attachment; filename="%s"`, filePath),
	}
	c.DataFromReader(206, info.Size(), "application/octet-stream", pr, extraHeaders)
}

type ListResp struct {
	Metadatas []*MetadataResp `json:"metadatas"`
}

func (h *FileHandlers) List(c *gin.Context) {
	dirPath := c.Query(ListDirQuery)
	if dirPath == "" {
		c.JSON(q.ErrResp(c, 400, errors.New("incorrect path name")))
		return
	}

	dirPath = h.FsPath(dirPath)
	infos, err := h.deps.FS().ListDir(dirPath)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}
	metadatas := []*MetadataResp{}
	for _, info := range infos {
		metadatas = append(metadatas, &MetadataResp{
			Name:    info.Name(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
			IsDir:   info.IsDir(),
		})
	}

	c.JSON(200, &ListResp{Metadatas: metadatas})
}

func (h *FileHandlers) Copy(c *gin.Context) {
	c.JSON(q.NewMsgResp(501, "Not Implemented"))
}

func (h *FileHandlers) CopyDir(c *gin.Context) {
	c.JSON(q.NewMsgResp(501, "Not Implemented"))
}

func (h *FileHandlers) GetTmpPath(filePath string) string {
	return path.Join(UploadDir, fmt.Sprintf("%x", sha1.Sum([]byte(filePath))))
}

func (h *FileHandlers) FsPath(filePath string) string {
	return path.Join(FsDir, filePath)
}
