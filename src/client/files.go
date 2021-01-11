package client

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ihexxa/quickshare/src/handlers/fileshdr"
	"github.com/parnurzeal/gorequest"
)

type FilesClient struct {
	addr string
	r    *gorequest.SuperAgent
}

func NewFilesClient(addr string) *FilesClient {
	gr := gorequest.New()
	return &FilesClient{
		addr: addr,
		r:    gr,
	}
}

func (cl *FilesClient) url(urlpath string) string {
	return fmt.Sprintf("%s%s", cl.addr, urlpath)
}

func (cl *FilesClient) Create(filepath string, size int64) (*http.Response, string, []error) {
	return cl.r.Post(cl.url("/v1/fs/files")).
		Send(fileshdr.CreateReq{
			Path:     filepath,
			FileSize: size,
		}).
		End()
}

func (cl *FilesClient) Delete(filepath string) (*http.Response, string, []error) {
	return cl.r.Delete(cl.url("/v1/fs/files")).
		Param(fileshdr.FilePathQuery, filepath).
		End()
}

func (cl *FilesClient) Metadata(filepath string) (*http.Response, *fileshdr.MetadataResp, []error) {
	resp, body, errs := cl.r.Get(cl.url("/v1/fs/metadata")).
		Param(fileshdr.FilePathQuery, filepath).
		End()

	mResp := &fileshdr.MetadataResp{}
	err := json.Unmarshal([]byte(body), mResp)
	if err != nil {
		errs = append(errs, err)
		return nil, nil, errs
	}
	return resp, mResp, nil
}

func (cl *FilesClient) Mkdir(dirpath string) (*http.Response, string, []error) {
	return cl.r.Post(cl.url("/v1/fs/dirs")).
		Send(fileshdr.MkdirReq{Path: dirpath}).
		End()
}

func (cl *FilesClient) Move(oldpath, newpath string) (*http.Response, string, []error) {
	return cl.r.Patch(cl.url("/v1/fs/files/move")).
		Send(fileshdr.MoveReq{
			OldPath: oldpath,
			NewPath: newpath,
		}).
		End()
}

func (cl *FilesClient) UploadChunk(filepath string, content string, offset int64) (*http.Response, string, []error) {
	return cl.r.Patch(cl.url("/v1/fs/files/chunks")).
		Send(fileshdr.UploadChunkReq{
			Path:    filepath,
			Content: content,
			Offset:  offset,
		}).
		End()
}

func (cl *FilesClient) UploadStatus(filepath string) (*http.Response, *fileshdr.UploadStatusResp, []error) {
	resp, body, errs := cl.r.Get(cl.url("/v1/fs/files/chunks")).
		Param(fileshdr.FilePathQuery, filepath).
		End()

	uResp := &fileshdr.UploadStatusResp{}
	err := json.Unmarshal([]byte(body), uResp)
	if err != nil {
		errs = append(errs, err)
		return nil, nil, errs
	}
	return resp, uResp, nil
}

func (cl *FilesClient) Download(filepath string, headers map[string]string) (*http.Response, string, []error) {
	r := cl.r.Get(cl.url("/v1/fs/files")).
		Param(fileshdr.FilePathQuery, filepath)
	for key, val := range headers {
		r = r.Set(key, val)
	}
	return r.End()
}

func (cl *FilesClient) List(dirPath string) (*http.Response, *fileshdr.ListResp, []error) {
	resp, body, errs := cl.r.Get(cl.url("/v1/fs/dirs")).
		Param(fileshdr.ListDirQuery, dirPath).
		End()
	if len(errs) > 0 {
		return nil, nil, errs
	}

	lResp := &fileshdr.ListResp{}
	err := json.Unmarshal([]byte(body), lResp)
	if err != nil {
		return nil, nil, append(errs, err)
	}
	return resp, lResp, nil
}

func (cl *FilesClient) ListUploadings() (*http.Response, *fileshdr.ListUploadingsResp, []error) {
	resp, body, errs := cl.r.Get(cl.url("/v1/fs/uploadings")).
		End()
	if len(errs) > 0 {
		return nil, nil, errs
	}

	lResp := &fileshdr.ListUploadingsResp{}
	err := json.Unmarshal([]byte(body), lResp)
	if err != nil {
		return nil, nil, append(errs, err)
	}
	return resp, lResp, nil
}

func (cl *FilesClient) DelUploading(filepath string) (*http.Response, string, []error) {
	return cl.r.Delete(cl.url("/v1/fs/uploadings")).
		Param(fileshdr.FilePathQuery, filepath).
		End()
}
