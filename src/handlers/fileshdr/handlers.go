package fileshdr

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ihexxa/gocfg"
	"github.com/ihexxa/multipart"

	"github.com/ihexxa/fsearch"
	"github.com/ihexxa/quickshare/src/db"
	"github.com/ihexxa/quickshare/src/db/userstore"
	"github.com/ihexxa/quickshare/src/depidx"
	q "github.com/ihexxa/quickshare/src/handlers"
	"github.com/ihexxa/quickshare/src/worker/localworker"
)

const (
	// queries
	FilePathQuery = "fp"
	ListDirQuery  = "dp"
	ShareIDQuery  = "shid"
	Keyword       = "k"

	// headers
	rangeHeader       = "Range"
	acceptRangeHeader = "Accept-Range"
	ifRangeHeader     = "If-Range"
	keepAliveHeader   = "Keep-Alive"
	connectionHeader  = "Connection"
)

type FileHandlers struct {
	cfg  gocfg.ICfg
	deps *depidx.Deps
}

func NewFileHandlers(cfg gocfg.ICfg, deps *depidx.Deps) (*FileHandlers, error) {
	handlers := &FileHandlers{
		cfg:  cfg,
		deps: deps,
	}
	deps.Workers().AddHandler(MsgTypeSha1, handlers.genSha1)
	deps.Workers().AddHandler(MsgTypeIndexing, handlers.indexingItems)

	return handlers, nil
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

func (lk *AutoLocker) Exec(handler func()) error {
	var err error
	kv := lk.h.deps.KV()
	locked := false

	defer func() {
		if p := recover(); p != nil {
			lk.h.deps.Log().Error(p)
		}
		if locked {
			if err = kv.Unlock(lk.key); err != nil {
				lk.h.deps.Log().Error(err)
			}
		}
	}()

	if err = kv.TryLock(lk.key); err != nil {
		return errors.New("fail to lock the file")
	}

	locked = true
	handler()
	return nil
}

// related elements: role, user, action(listing, downloading)/sharing
func (h *FileHandlers) canAccess(ctx context.Context, userId uint64, userName, role, op, sharedPath string) bool {
	if role == db.AdminRole {
		return true
	}

	// the file path must start with userName: <userName>/...
	parts := strings.Split(sharedPath, "/")
	if len(parts) < 2 { // the path must be longer than <userName>/files
		return false
	} else if parts[0] == userName && userName != "" && parts[1] != "" {
		return true
	}

	// check if it is shared
	// TODO: find a better approach
	if op != "list" && op != "download" {
		return false
	}

	isSharing := h.deps.FileInfos().IsSharing(ctx, userId, sharedPath)
	return isSharing
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

	userID, err := q.GetUserId(c)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}

	fsFilePath, err := h.getFSFilePath(fmt.Sprint(userID), req.Path)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			c.JSON(q.ErrResp(c, 400, err))
		} else {
			c.JSON(q.ErrResp(c, 500, err))
		}
		return
	}

	role := c.MustGet(q.RoleParam).(string)
	userName := c.MustGet(q.UserParam).(string)
	if !h.canAccess(c, userID, userName, role, "create", fsFilePath) {
		c.JSON(q.ErrResp(c, 403, q.ErrAccessDenied))
		return
	}

	tmpFilePath := q.UploadPath(userName, fsFilePath)
	if req.FileSize == 0 {
		// TODO: limit the number of files with 0 byte
		err = h.deps.FileInfos().AddUploadInfos(c, userID, tmpFilePath, fsFilePath, &db.FileInfo{
			Size: req.FileSize,
		})
		if err != nil {
			if errors.Is(err, db.ErrQuota) {
				c.JSON(q.ErrResp(c, 403, err))
			} else {
				c.JSON(q.ErrResp(c, 500, err))
			}
			return
		}
		err = h.deps.FileInfos().MoveFileInfos(c, userID, tmpFilePath, fsFilePath, false)
		if err != nil {
			c.JSON(q.ErrResp(c, 500, err))
			return
		}

		err = h.deps.FS().MkdirAll(filepath.Dir(fsFilePath))
		if err != nil {
			c.JSON(q.ErrResp(c, 500, err))
			return
		}

		err = h.deps.FS().Create(fsFilePath)
		if err != nil {
			if os.IsExist(err) {
				c.JSON(q.ErrResp(c, 304, fmt.Errorf("file(%s) exists", fsFilePath)))
			} else {
				c.JSON(q.ErrResp(c, 500, err))
			}
			return
		}

		msg, err := json.Marshal(Sha1Params{
			UserId:   userID,
			FilePath: fsFilePath,
		})
		if err != nil {
			c.JSON(q.ErrResp(c, 500, err))
			return
		}

		err = h.deps.Workers().TryPut(
			localworker.NewMsg(
				h.deps.ID().Gen(),
				map[string]string{localworker.MsgTypeKey: MsgTypeSha1},
				string(msg),
			),
		)
		if err != nil {
			c.JSON(q.ErrResp(c, 500, err))
			return
		}

		err = h.deps.FileIndex().AddPath(fsFilePath)
		if err != nil {
			c.JSON(q.ErrResp(c, 500, err))
			return
		}

		c.JSON(q.Resp(200))
		return
	}

	err = h.deps.FileInfos().AddUploadInfos(c, userID, tmpFilePath, fsFilePath, &db.FileInfo{
		Size: req.FileSize,
	})
	if err != nil {
		if errors.Is(err, db.ErrQuota) {
			c.JSON(q.ErrResp(c, 403, err))
		} else {
			c.JSON(q.ErrResp(c, 500, err))
		}
		return
	}

	var txErr error
	locker := h.NewAutoLocker(c, lockName(tmpFilePath))
	lockErr := locker.Exec(func() {
		err = h.deps.FS().Create(tmpFilePath)
		if err != nil {
			if os.IsExist(err) {
				createErr := fmt.Errorf("file(%s) exists", tmpFilePath)
				c.JSON(q.ErrResp(c, 304, createErr))
				txErr = createErr
			} else {
				c.JSON(q.ErrResp(c, 500, err))
				txErr = err
			}
			return
		}

		err = h.deps.FS().MkdirAll(filepath.Dir(req.Path))
		if err != nil {
			c.JSON(q.ErrResp(c, 500, err))
			txErr = err
			return
		}
	})
	if lockErr != nil {
		c.JSON(q.ErrResp(c, 500, lockErr))
		return
	}
	if txErr != nil {
		c.JSON(q.ErrResp(c, 500, txErr))
		return
	}
	c.JSON(q.Resp(200))
}

func (h *FileHandlers) Delete(c *gin.Context) {
	filePath := c.Query(FilePathQuery)
	filePath = filepath.Clean(filePath)
	if filePath == "" {
		c.JSON(q.ErrResp(c, 400, errors.New("invalid file path")))
		return
	}

	userId, err := q.GetUserId(c)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}

	role := c.MustGet(q.RoleParam).(string)
	userName := c.MustGet(q.UserParam).(string)
	if !h.canAccess(c, userId, userName, role, "delete", filePath) {
		c.JSON(q.ErrResp(c, 403, q.ErrAccessDenied))
		return
	}

	var txErr error
	locker := h.NewAutoLocker(c, lockName(filePath))
	lockErr := locker.Exec(func() {
		err = h.deps.FS().Remove(filePath)
		if err != nil {
			txErr = err
			return
		}

		err = h.deps.FileInfos().DelFileInfo(c, userId, filePath)
		if err != nil {
			txErr = err
			return
		}

		err = h.deps.FileIndex().DelPath(filePath)
		if err != nil && !errors.Is(err, fsearch.ErrNotFound) {
			txErr = err
			return
		}
	})
	if lockErr != nil {
		c.JSON(q.ErrResp(c, 500, lockErr))
		return
	}
	if txErr != nil {
		c.JSON(q.ErrResp(c, 500, txErr))
		return
	}
	c.JSON(q.Resp(200))
}

type MetadataResp struct {
	Name    string    `json:"name"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"modTime"`
	IsDir   bool      `json:"isDir"`
	Sha1    string    `json:"sha1"`
}

func (h *FileHandlers) Metadata(c *gin.Context) {
	filePath := c.Query(FilePathQuery)
	filePath = filepath.Clean(filePath)
	if filePath == "" {
		c.JSON(q.ErrResp(c, 400, errors.New("invalid file path")))
		return
	}

	userId, err := q.GetUserId(c)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}

	role := c.MustGet(q.RoleParam).(string)
	userName := c.MustGet(q.UserParam).(string)
	if !h.canAccess(c, userId, userName, role, "metadata", filePath) {
		c.JSON(q.ErrResp(c, 403, q.ErrAccessDenied))
		return
	}

	info, err := h.deps.FS().Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(q.ErrResp(c, 404, os.ErrNotExist))
		} else {
			c.JSON(q.ErrResp(c, 500, err))
		}
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

	userId, err := q.GetUserId(c)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}

	role := c.MustGet(q.RoleParam).(string)
	userName := c.MustGet(q.UserParam).(string)
	dirPath := filepath.Clean(req.Path)
	if !h.canAccess(c, userId, userName, role, "mkdir", dirPath) {
		c.JSON(q.ErrResp(c, 403, q.ErrAccessDenied))
		return
	}

	err = h.deps.FS().MkdirAll(dirPath)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}

	err = h.deps.FileIndex().AddPath(dirPath)
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
	userId, err := q.GetUserId(c)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}
	role := c.MustGet(q.RoleParam).(string)
	userName := c.MustGet(q.UserParam).(string)

	oldPath := filepath.Clean(req.OldPath)
	newPath := filepath.Clean(req.NewPath)
	if !h.canAccess(c, userId, userName, role, "move", oldPath) ||
		!h.canAccess(c, userId, userName, role, "move", newPath) {
		c.JSON(q.ErrResp(c, 403, q.ErrAccessDenied))
		return
	}

	itemInfo, err := h.deps.FS().Stat(oldPath)
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

	err = h.deps.FileInfos().MoveFileInfos(c, userId, oldPath, newPath, itemInfo.IsDir())
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}

	err = h.deps.FS().Rename(oldPath, newPath)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}

	newPathDir := filepath.Dir(newPath)
	err = h.deps.FileIndex().AddPath(newPathDir)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}
	err = h.deps.FileIndex().MovePath(oldPath, newPathDir)
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

	userId, err := q.GetUserId(c)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}
	role := c.MustGet(q.RoleParam).(string)
	userName := c.MustGet(q.UserParam).(string)
	filePath := filepath.Clean(req.Path)
	if !h.canAccess(c, userId, userName, role, "upload.chunk", filePath) {
		c.JSON(q.ErrResp(c, 403, q.ErrAccessDenied))
		return
	}

	ok, err := h.deps.Limiter().CanWrite(userId, len([]byte(req.Content)))
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	} else if !ok {
		c.JSON(q.ErrResp(c, 429, errors.New("retry later")))
		return
	}

	var txErr error
	var statusCode int
	tmpFilePath := q.UploadPath(userName, filePath)
	locker := h.NewAutoLocker(c, lockName(tmpFilePath))
	fsFilePath, fileSize, uploaded, wrote := "", int64(0), int64(0), 0
	lockErr := locker.Exec(func() {
		var err error

		fsFilePath, fileSize, uploaded, err = h.deps.FileInfos().GetUploadInfo(c, userId, tmpFilePath)
		if err != nil {
			txErr, statusCode = err, 500
			return
		} else if uploaded != req.Offset {
			txErr, statusCode = errors.New("offset != uploaded"), 500
			return
		}

		content, err := base64.StdEncoding.DecodeString(req.Content)
		if err != nil {
			txErr, statusCode = err, 500
			return
		}

		wrote, err = h.deps.FS().WriteAt(tmpFilePath, []byte(content), req.Offset)
		if err != nil {
			txErr, statusCode = err, 500
			return
		}

		err = h.deps.FileInfos().SetUploadInfo(c, userId, tmpFilePath, req.Offset+int64(wrote))
		if err != nil {
			txErr, statusCode = err, 500
			return
		}

		// move the file from uploading dir to uploaded dir
		if uploaded+int64(wrote) == fileSize {
			err = h.deps.FileInfos().MoveFileInfos(c, userId, tmpFilePath, fsFilePath, false)
			if err != nil {
				txErr, statusCode = err, 500
				return
			}

			err = h.deps.FS().Rename(tmpFilePath, fsFilePath)
			if err != nil {
				txErr, statusCode = fmt.Errorf("%s error: %w", fsFilePath, err), 500
				return
			}

			msg, err := json.Marshal(Sha1Params{
				UserId:   userId,
				FilePath: fsFilePath,
			})
			if err != nil {
				txErr, statusCode = err, 500
				return
			}

			err = h.deps.Workers().TryPut(
				localworker.NewMsg(
					h.deps.ID().Gen(),
					map[string]string{localworker.MsgTypeKey: MsgTypeSha1},
					string(msg),
				),
			)
			if err != nil {
				txErr, statusCode = err, 500
				return
			}

			err = h.deps.FileIndex().AddPath(fsFilePath)
			if err != nil {
				txErr, statusCode = err, 500
				return
			}
		}
	})
	if lockErr != nil {
		c.JSON(q.ErrResp(c, 500, err))
	}
	if txErr != nil {
		c.JSON(q.ErrResp(c, statusCode, txErr))
		return
	}
	c.JSON(200, &UploadStatusResp{
		Path:     fsFilePath,
		IsDir:    false,
		FileSize: fileSize,
		Uploaded: uploaded + int64(wrote),
	})
}

func (h *FileHandlers) getFSFilePath(userID, fsFilePath string) (string, error) {
	fsFilePath = filepath.Clean(fsFilePath)

	_, err := h.deps.FS().Stat(fsFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fsFilePath, nil
		}
		return "", err
	}
	return "", os.ErrExist

	// temporarily disable file auto renaming
	// the new name should be returned to the client
	// but curently files status is tracked by the original file name between worker and upload manager

	// this file exists
	// maxDetect := 1024
	// for i := 1; i < maxDetect; i++ {
	// 	dir := path.Dir(fsFilePath)
	// 	nameAndExt := path.Base(fsFilePath)
	// 	ext := path.Ext(nameAndExt)
	// 	fileName := nameAndExt[:len(nameAndExt)-len(ext)]

	// 	detectPath := path.Join(dir, fmt.Sprintf("%s_%d%s", fileName, i, ext))
	// 	_, err := h.deps.FS().Stat(detectPath)
	// 	if err != nil {
	// 		if os.IsNotExist(err) {
	// 			return detectPath, nil
	// 		} else {
	// 			return "", err
	// 		}
	// 	}
	// }

	// return "", fmt.Errorf("found more than %d duplicated files", maxDetect)
}

type UploadStatusResp struct {
	Path     string `json:"path"`
	IsDir    bool   `json:"isDir"`
	FileSize int64  `json:"fileSize"`
	Uploaded int64  `json:"uploaded"`
}

func (h *FileHandlers) UploadStatus(c *gin.Context) {
	filePath := c.Query(FilePathQuery)
	filePath = filepath.Clean(filePath)
	if filePath == "" {
		c.JSON(q.ErrResp(c, 400, errors.New("invalid file name")))
	}

	userId, err := q.GetUserId(c)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}

	role := c.MustGet(q.RoleParam).(string)
	userName := c.MustGet(q.UserParam).(string)
	if !h.canAccess(c, userId, userName, role, "upload.status", filePath) {
		c.JSON(q.ErrResp(c, 403, q.ErrAccessDenied))
		return
	}

	tmpFilePath := q.UploadPath(userName, filePath)
	locker := h.NewAutoLocker(c, lockName(tmpFilePath))
	fileSize, uploaded := int64(0), int64(0)
	var txErr error
	lockErr := locker.Exec(func() {
		var err error
		_, fileSize, uploaded, err = h.deps.FileInfos().GetUploadInfo(c, userId, tmpFilePath)
		if err != nil {
			if os.IsNotExist(err) {
				c.JSON(q.ErrResp(c, 404, err))
				txErr = err
			} else {
				c.JSON(q.ErrResp(c, 500, err))
				txErr = err
			}
			return
		}
	})
	if lockErr != nil {
		c.JSON(q.ErrResp(c, 500, lockErr))
		return
	}
	if txErr != nil {
		c.JSON(q.ErrResp(c, 500, txErr))
		return
	}
	c.JSON(200, &UploadStatusResp{
		Path:     filePath,
		IsDir:    false,
		FileSize: fileSize,
		Uploaded: uploaded,
	})
}

// TODO: support ETag
func (h *FileHandlers) Download(c *gin.Context) {
	rangeVal := c.GetHeader(rangeHeader)
	ifRangeVal := c.GetHeader(ifRangeHeader)
	filePath := c.Query(FilePathQuery)
	filePath = filepath.Clean(filePath)
	if filePath == "" {
		c.JSON(q.ErrResp(c, 400, errors.New("invalid file name")))
		return
	}

	role := c.MustGet(q.RoleParam).(string)
	userName := c.MustGet(q.UserParam).(string)
	dirPath := filepath.Dir(filePath)

	var err error
	userId := userstore.VisitorID
	if role != db.VisitorRole {
		userId, err = q.GetUserId(c)
		if err != nil {
			c.JSON(q.ErrResp(c, 500, err))
			return
		}
	}

	if !h.canAccess(c, userId, userName, role, "download", dirPath) {
		c.JSON(q.ErrResp(c, 403, q.ErrAccessDenied))
		return
	}

	// TODO: when sharing is introduced, move following logics to a separeted method
	// concurrently file accessing is managed by os
	info, err := h.deps.FS().Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(q.ErrResp(c, 404, os.ErrNotExist))
		} else {
			c.JSON(q.ErrResp(c, 500, err))
		}
		return
	} else if info.IsDir() {
		c.JSON(q.ErrResp(c, 400, errors.New("downloading a folder is not supported")))
		return
	}

	// https://golang.google.cn/pkg/net/http/#DetectContentType
	// DetectContentType considers at most the first 512 bytes of data.
	fileHeadBuf := make([]byte, 512)
	read, err := h.deps.FS().ReadAt(filePath, fileHeadBuf, 0)
	if err != nil && err != io.EOF {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}
	contentType := http.DetectContentType(fileHeadBuf[:read])

	fd, id, err := h.deps.FS().GetFileReader(filePath)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}
	defer func() {
		err := h.deps.FS().CloseReader(fmt.Sprint(id))
		if err != nil {
			h.deps.Log().Errorf("failed to close: %s", err)
		}
	}()

	extraHeaders := map[string]string{
		"Content-Disposition": fmt.Sprintf(`attachment; filename="%s"`, info.Name()),
	}

	// respond to normal requests
	if ifRangeVal != "" || rangeVal == "" {
		limitedReader, err := h.GetStreamReader(userId, fd)
		if err != nil {
			c.JSON(q.ErrResp(c, 500, err))
			return
		}
		defer func() {
			err := limitedReader.Close()
			if err != nil {
				h.deps.Log().Errorf("failed to close limitedReader: %s", err)
			}
		}()

		c.DataFromReader(200, info.Size(), contentType, limitedReader, extraHeaders)
		return
	}

	// respond to range requests
	parts, err := multipart.RangeToParts(rangeVal, contentType, fmt.Sprintf("%d", info.Size()))
	if err != nil {
		c.JSON(q.ErrResp(c, 400, err))
		return
	}

	mr, err := multipart.NewMultipartReader(fd, parts)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}
	mr.SetOutputHeaders(false)
	contentLength := mr.ContentLength()
	defer func() {
		err := mr.Close()
		if err != nil {
			h.deps.Log().Errorf("failed to close multipart reader: %s", err)
		}
	}()

	// TODO: reader will be closed by multipart response writer？
	go mr.Start()

	limitedReader, err := h.GetStreamReader(userId, mr)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}
	defer func() {
		err := limitedReader.Close()
		if err != nil {
			h.deps.Log().Errorf("failed to close limitedReader: %s", err)
		}
	}()

	// it takes the \r\n before body into account, so contentLength+2
	c.DataFromReader(206, contentLength+2, contentType, limitedReader, extraHeaders)
}

type ListResp struct {
	Cwd       string          `json:"cwd"`
	Metadatas []*MetadataResp `json:"metadatas"`
}

func (h *FileHandlers) MergeFileInfos(ctx *gin.Context, dirPath string, infos []os.FileInfo) ([]*MetadataResp, error) {
	filePaths := []string{}
	metadatas := []*MetadataResp{}
	for _, info := range infos {
		if !info.IsDir() {
			filePaths = append(filePaths, filepath.Join(dirPath, info.Name()))
		}
		metadatas = append(metadatas, &MetadataResp{
			Name:    info.Name(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
			IsDir:   info.IsDir(),
		})
	}

	dbInfos, err := h.deps.FileInfos().ListFileInfos(ctx, filePaths)
	if err != nil {
		return nil, err
	}
	for _, metadata := range metadatas {
		if !metadata.IsDir {
			dbInfo, ok := dbInfos[filepath.Join(dirPath, metadata.Name)]
			if ok {
				metadata.Sha1 = dbInfo.Sha1
			}
		}
	}

	return metadatas, nil
}

func (h *FileHandlers) List(c *gin.Context) {
	dirPath := c.Query(ListDirQuery)
	dirPath = filepath.Clean(dirPath)
	if dirPath == "" {
		c.JSON(q.ErrResp(c, 400, errors.New("incorrect path name")))
		return
	}

	userId, err := q.GetUserId(c)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}
	role := c.MustGet(q.RoleParam).(string)
	userName := c.MustGet(q.UserParam).(string)
	if !h.canAccess(c, userId, userName, role, "list", dirPath) {
		c.JSON(q.ErrResp(c, 403, q.ErrAccessDenied))
		return
	}

	infos, err := h.deps.FS().ListDir(dirPath)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}

	metadatas, err := h.MergeFileInfos(c, dirPath, infos)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}

	c.JSON(200, &ListResp{
		Cwd:       dirPath,
		Metadatas: metadatas,
	})
}

func (h *FileHandlers) ListHome(c *gin.Context) {
	userName := c.MustGet(q.UserParam).(string)
	fsPath := q.FsRootPath(userName, "/")

	infos, err := h.deps.FS().ListDir(fsPath)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}

	metadatas, err := h.MergeFileInfos(c, fsPath, infos)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}

	c.JSON(200, &ListResp{
		Cwd:       fsPath,
		Metadatas: metadatas,
	})
}

func (h *FileHandlers) Copy(c *gin.Context) {
	c.JSON(q.NewMsgResp(501, "Not Implemented"))
}

func (h *FileHandlers) CopyDir(c *gin.Context) {
	c.JSON(q.NewMsgResp(501, "Not Implemented"))
}

func lockName(filePath string) string {
	return filePath
}

type ListUploadingsResp struct {
	UploadInfos []*db.UploadInfo `json:"uploadInfos"`
}

func (h *FileHandlers) ListUploadings(c *gin.Context) {
	userId, err := q.GetUserId(c)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}

	infos, err := h.deps.FileInfos().ListUploadInfos(c, userId)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}
	if infos == nil {
		infos = []*db.UploadInfo{}
	}
	c.JSON(200, &ListUploadingsResp{UploadInfos: infos})
}

func (h *FileHandlers) DelUploading(c *gin.Context) {
	filePath := c.Query(FilePathQuery)
	filePath = filepath.Clean(filePath)
	if filePath == "" {
		c.JSON(q.ErrResp(c, 400, errors.New("invalid file path")))
		return
	}

	userId, err := q.GetUserId(c)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}
	role := c.MustGet(q.RoleParam).(string)
	userName := c.MustGet(q.UserParam).(string)
	// op is empty, because users must be admin, or the path belongs to this user
	if !h.canAccess(c, userId, userName, role, "", filePath) {
		c.JSON(q.ErrResp(c, 403, errors.New("forbidden")))
		return
	}

	var txErr error
	var statusCode int
	tmpFilePath := q.UploadPath(userName, filePath)
	locker := h.NewAutoLocker(c, lockName(tmpFilePath))
	lockErr := locker.Exec(func() {
		_, err = h.deps.FS().Stat(tmpFilePath)
		if err != nil {
			if os.IsNotExist(err) {
				// no op
			} else {
				txErr, statusCode = err, 500
				return
			}
		} else {
			err = h.deps.FS().Remove(tmpFilePath)
			if err != nil {
				txErr, statusCode = err, 500
				return
			}
		}
	})
	if lockErr != nil {
		c.JSON(q.ErrResp(c, 500, lockErr))
		return
	}
	if txErr != nil {
		c.JSON(q.ErrResp(c, statusCode, txErr))
		return
	}
	err = h.deps.FileInfos().DelUploadingInfos(c, userId, filePath)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}
	c.JSON(q.Resp(200))
}

type SharingReq struct {
	SharingPath string `json:"sharingPath"`
}

func (h *FileHandlers) AddSharing(c *gin.Context) {
	req := &SharingReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(q.ErrResp(c, 400, err))
		return
	}

	userId, err := q.GetUserId(c)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}

	sharingPath := filepath.Clean(req.SharingPath)
	// TODO: move canAccess to authedFS
	role := c.MustGet(q.RoleParam).(string)
	userName := c.MustGet(q.UserParam).(string)
	// op is empty, because users must be admin, or the path belongs to this user
	if !h.canAccess(c, userId, userName, role, "", sharingPath) {
		c.JSON(q.ErrResp(c, 403, errors.New("forbidden")))
		return
	}

	if sharingPath == "" || sharingPath == "/" {
		c.JSON(q.ErrResp(c, 403, errors.New("forbidden")))
		return
	}

	info, err := h.deps.FS().Stat(sharingPath)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	} else if !info.IsDir() {
		c.JSON(q.ErrResp(c, 400, errors.New("can not sharing a file")))
		return
	}

	err = h.deps.FileInfos().AddSharing(c, userId, sharingPath)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}
	c.JSON(q.Resp(200))
}

func (h *FileHandlers) DelSharing(c *gin.Context) {
	dirPath := c.Query(FilePathQuery)
	dirPath = filepath.Clean(dirPath)
	if dirPath == "" {
		c.JSON(q.ErrResp(c, 400, errors.New("invalid file path")))
		return
	}

	userId, err := q.GetUserId(c)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}

	// TODO: move canAccess to authedFS
	userName := c.MustGet(q.UserParam).(string)
	role := c.MustGet(q.RoleParam).(string)
	if !h.canAccess(c, userId, userName, role, "", dirPath) {
		c.JSON(q.ErrResp(c, 403, errors.New("forbidden")))
		return
	}

	err = h.deps.FileInfos().DelSharing(c, userId, dirPath)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}
	c.JSON(q.Resp(200))
}

func (h *FileHandlers) IsSharing(c *gin.Context) {
	dirPath := c.Query(FilePathQuery)
	dirPath = filepath.Clean(dirPath)
	if dirPath == "" {
		c.JSON(q.ErrResp(c, 400, errors.New("invalid file path")))
		return
	}

	userId, err := q.GetUserId(c)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}

	exist := h.deps.FileInfos().IsSharing(c, userId, dirPath)
	if exist {
		c.JSON(q.Resp(200))
	} else {
		c.JSON(q.Resp(404))
	}
}

type SharingResp struct {
	SharingDirs []string `json:"sharingDirs"`
}

// Deprecated: use ListSharingIDs instead
func (h *FileHandlers) ListSharings(c *gin.Context) {
	// TODO: move canAccess to authedFS
	userId, err := q.GetUserId(c)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}

	sharingDirs, err := h.deps.FileInfos().ListUserSharings(c, userId)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}

	dirs := []string{}
	for sharingDir := range sharingDirs {
		dirs = append(dirs, sharingDir)
	}
	c.JSON(200, &SharingResp{SharingDirs: dirs})
}

type SharingIDsResp struct {
	IDs map[string]string `json:"IDs"`
}

func (h *FileHandlers) ListSharingIDs(c *gin.Context) {
	userId, err := q.GetUserId(c)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}

	dirToID, err := h.deps.FileInfos().ListUserSharings(c, userId)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}
	c.JSON(200, &SharingIDsResp{IDs: dirToID})
}

type GenerateHashReq struct {
	FilePath string `json:"filePath"`
}

func (h *FileHandlers) GenerateHash(c *gin.Context) {
	req := &GenerateHashReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(q.ErrResp(c, 400, err))
		return
	}

	filePath := filepath.Clean(req.FilePath)
	if filePath == "" {
		c.JSON(q.ErrResp(c, 400, errors.New("invalid file path")))
		return
	}

	userId, err := q.GetUserId(c)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}

	role := c.MustGet(q.RoleParam).(string)
	userName := c.MustGet(q.UserParam).(string)
	if !h.canAccess(c, userId, userName, role, "hash.gen", filePath) {
		c.JSON(q.ErrResp(c, 403, q.ErrAccessDenied))
		return
	}

	msg, err := json.Marshal(Sha1Params{
		UserId:   userId,
		FilePath: filePath,
	})
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}

	err = h.deps.Workers().TryPut(
		localworker.NewMsg(
			h.deps.ID().Gen(),
			map[string]string{localworker.MsgTypeKey: MsgTypeSha1},
			string(msg),
		),
	)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}
	c.JSON(q.Resp(200))
}

type GetSharingDirResp struct {
	SharingDir string `json:"sharingDir"`
}

func (h *FileHandlers) GetSharingDir(c *gin.Context) {
	shareID := c.Query(ShareIDQuery)
	if shareID == "" {
		c.JSON(q.ErrResp(c, 400, errors.New("invalid share ID")))
		return
	}

	dirPath, err := h.deps.FileInfos().GetSharingDir(c, shareID)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}
	c.JSON(200, &GetSharingDirResp{SharingDir: dirPath})
}

type SearchItemsResp struct {
	Results []string `json:"results"`
}

func (h *FileHandlers) SearchItems(c *gin.Context) {
	keywords := c.QueryArray(Keyword)
	if len(keywords) == 0 {
		c.JSON(q.ErrResp(c, 400, errors.New("empty keyword")))
		return
	}

	resultsMap := map[string]int{}
	for _, keyword := range keywords {
		searchResults, err := h.deps.FileIndex().Search(keyword)
		if err != nil {
			c.JSON(q.ErrResp(c, 500, err))
			return
		}

		for _, searchResult := range searchResults {
			if _, ok := resultsMap[searchResult]; !ok {
				resultsMap[searchResult] = 0
			}
			resultsMap[searchResult] += 1
		}
	}

	role := c.MustGet(q.RoleParam).(string)
	userName := c.MustGet(q.UserParam).(string)
	results := []string{}
	for pathname, count := range resultsMap {
		if count >= len(keywords) {
			if role == db.AdminRole ||
				(role != db.AdminRole && strings.HasPrefix(pathname, userName)) {
				results = append(results, pathname)
			}
		}
	}

	c.JSON(200, &SearchItemsResp{Results: results})
}

func (h *FileHandlers) Reindex(c *gin.Context) {
	msg, err := json.Marshal(IndexingParams{})
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}

	err = h.deps.Workers().TryPut(
		localworker.NewMsg(
			h.deps.ID().Gen(),
			map[string]string{localworker.MsgTypeKey: MsgTypeIndexing},
			string(msg),
		),
	)
	if err != nil {
		c.JSON(q.ErrResp(c, 500, err))
		return
	}

	c.JSON(q.Resp(200))
	return
}

func (h *FileHandlers) GetStreamReader(userID uint64, fd io.Reader) (io.ReadCloser, error) {
	pr, pw := io.Pipe()

	go func() {
		for {
			ok, err := h.deps.Limiter().CanRead(userID, q.DownloadChunkSize)
			if err != nil {
				pw.CloseWithError(err)
				break
			} else if !ok {
				time.Sleep(time.Duration(1) * time.Second)
				continue
			}

			_, err = io.CopyN(pw, fd, int64(q.DownloadChunkSize))
			if err != nil {
				if err != io.EOF {
					pw.CloseWithError(err)
				} else {
					pw.CloseWithError(nil)
				}
				break
			}
		}
	}()

	return pr, nil
}
