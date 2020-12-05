package client

import (
	"encoding/json"
	"fmt"

	"github.com/ihexxa/quickshare/src/handlers/fileshdr"
	"github.com/parnurzeal/gorequest"
)

type QSClient struct {
	addr string
	r    *gorequest.SuperAgent
}

func NewQSClient(addr string) *QSClient {
	gr := gorequest.New()
	return &QSClient{
		addr: addr,
		r:    gr,
	}
}

func (cl *QSClient) url(urlpath string) string {
	return fmt.Sprintf("%s%s", cl.addr, urlpath)
}

func (cl *QSClient) Create(filepath string, size int64) (gorequest.Response, string, []error) {
	return cl.r.Post(cl.url("/v1/fs/files")).
		Send(fileshdr.CreateReq{
			Path:     filepath,
			FileSize: size,
		}).
		End()
}

func (cl *QSClient) Delete(filepath string) (gorequest.Response, string, []error) {
	return cl.r.Delete(cl.url("/v1/fs/files")).
		Param(fileshdr.FilePathQuery, filepath).
		End()
}

func (cl *QSClient) Metadata(filepath string) (gorequest.Response, *fileshdr.MetadataResp, []error) {
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

func (cl *QSClient) Mkdir(dirpath string) (gorequest.Response, string, []error) {
	return cl.r.Post(cl.url("/v1/fs/dirs")).
		Send(fileshdr.MkdirReq{Path: dirpath}).
		End()
}

func (cl *QSClient) Move(oldpath, newpath string) (gorequest.Response, string, []error) {
	return cl.r.Patch(cl.url("/v1/fs/files/move")).
		Send(fileshdr.MoveReq{
			OldPath: oldpath,
			NewPath: newpath,
		}).
		End()
}

func (cl *QSClient) UploadChunk(filepath string, content string, offset int64) (gorequest.Response, string, []error) {
	return cl.r.Patch(cl.url("/v1/fs/files/chunks")).
		Send(fileshdr.UploadChunkReq{
			Path:    filepath,
			Content: content,
			Offset:  offset,
		}).
		End()
}

func (cl *QSClient) UploadStatus(filepath string) (gorequest.Response, *fileshdr.UploadStatusResp, []error) {
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

func (cl *QSClient) Download(filepath string, headers map[string]string) (gorequest.Response, string, []error) {
	r := cl.r.Get(cl.url("/v1/fs/files/chunks")).
		Param(fileshdr.FilePathQuery, filepath)
	for key, val := range headers {
		r = r.Set(key, val)
	}
	return r.End()
}

func (cl *QSClient) List(dirPath string) (gorequest.Response, *fileshdr.ListResp, []error) {
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
