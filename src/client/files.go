package client

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ihexxa/quickshare/src/handlers/fileshdr"
	"github.com/parnurzeal/gorequest"
)

type FilesClient struct {
	addr  string
	r     *gorequest.SuperAgent
	token *http.Cookie
}

func NewFilesClient(addr string, token *http.Cookie) *FilesClient {
	gr := gorequest.New()
	return &FilesClient{
		addr:  addr,
		r:     gr,
		token: token,
	}
}

func (cl *FilesClient) url(urlpath string) string {
	return fmt.Sprintf("%s%s", cl.addr, urlpath)
}

func (cl *FilesClient) Create(filepath string, size int64) (*http.Response, string, []error) {
	return cl.r.Post(cl.url("/v1/fs/files")).
		AddCookie(cl.token).
		Send(fileshdr.CreateReq{
			Path:     filepath,
			FileSize: size,
		}).
		End()
}

func (cl *FilesClient) Delete(filepath string) (*http.Response, string, []error) {
	return cl.r.Delete(cl.url("/v1/fs/files")).
		AddCookie(cl.token).
		Param(fileshdr.FilePathQuery, filepath).
		End()
}

func (cl *FilesClient) Metadata(filepath string) (*http.Response, *fileshdr.MetadataResp, []error) {
	resp, body, errs := cl.r.Get(cl.url("/v1/fs/metadata")).
		AddCookie(cl.token).
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
		AddCookie(cl.token).
		Send(fileshdr.MkdirReq{Path: dirpath}).
		End()
}

func (cl *FilesClient) Move(oldpath, newpath string) (*http.Response, string, []error) {
	return cl.r.Patch(cl.url("/v1/fs/files/move")).
		AddCookie(cl.token).
		Send(fileshdr.MoveReq{
			OldPath: oldpath,
			NewPath: newpath,
		}).
		End()
}

func (cl *FilesClient) UploadChunk(filepath string, content string, offset int64) (*http.Response, string, []error) {
	return cl.r.Patch(cl.url("/v1/fs/files/chunks")).
		AddCookie(cl.token).
		Send(fileshdr.UploadChunkReq{
			Path:    filepath,
			Content: content,
			Offset:  offset,
		}).
		End()
}

func (cl *FilesClient) UploadStatus(filepath string) (*http.Response, *fileshdr.UploadStatusResp, []error) {
	resp, body, errs := cl.r.Get(cl.url("/v1/fs/files/chunks")).
		AddCookie(cl.token).
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
		AddCookie(cl.token).
		Param(fileshdr.FilePathQuery, filepath)
	for key, val := range headers {
		r = r.Set(key, val)
	}
	return r.End()
}

func (cl *FilesClient) List(dirPath string) (*http.Response, *fileshdr.ListResp, []error) {
	resp, body, errs := cl.r.Get(cl.url("/v1/fs/dirs")).
		AddCookie(cl.token).
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

func (cl *FilesClient) ListHome() (*http.Response, *fileshdr.ListResp, []error) {
	resp, body, errs := cl.r.Get(cl.url("/v1/fs/dirs/home")).
		AddCookie(cl.token).
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
		AddCookie(cl.token).
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
		AddCookie(cl.token).
		Param(fileshdr.FilePathQuery, filepath).
		End()
}

func (cl *FilesClient) AddSharing(dirpath string) (*http.Response, string, []error) {
	return cl.r.Post(cl.url("/v1/fs/sharings")).
		AddCookie(cl.token).
		Send(fileshdr.SharingReq{SharingPath: dirpath}).
		End()
}

func (cl *FilesClient) DelSharing(dirpath string) (*http.Response, string, []error) {
	return cl.r.Delete(cl.url("/v1/fs/sharings")).
		AddCookie(cl.token).
		Param(fileshdr.FilePathQuery, dirpath).
		End()
}

func (cl *FilesClient) IsSharing(dirpath string) (*http.Response, string, []error) {
	return cl.r.Get(cl.url("/v1/fs/sharings/exist")).
		AddCookie(cl.token).
		Param(fileshdr.FilePathQuery, dirpath).
		End()
}

// Deprecated: use ListSharingIDs intead
func (cl *FilesClient) ListSharings() (*http.Response, *fileshdr.SharingResp, []error) {
	resp, body, errs := cl.r.Get(cl.url("/v1/fs/sharings")).
		AddCookie(cl.token).
		End()
	if len(errs) > 0 {
		return nil, nil, errs
	}

	shResp := &fileshdr.SharingResp{}
	err := json.Unmarshal([]byte(body), shResp)
	if err != nil {
		return nil, nil, append(errs, err)
	}
	return resp, shResp, nil
}

func (cl *FilesClient) ListSharingIDs() (*http.Response, *fileshdr.SharingIDsResp, []error) {
	resp, body, errs := cl.r.Get(cl.url("/v1/fs/sharings/ids")).
		AddCookie(cl.token).
		End()
	if len(errs) > 0 {
		return nil, nil, errs
	}

	shResp := &fileshdr.SharingIDsResp{}
	err := json.Unmarshal([]byte(body), shResp)
	if err != nil {
		return nil, nil, append(errs, err)
	}
	return resp, shResp, nil
}

func (cl *FilesClient) GenerateHash(filepath string) (*http.Response, string, []error) {
	return cl.r.Post(cl.url("/v1/fs/hashes/sha1")).
		AddCookie(cl.token).
		Send(fileshdr.GenerateHashReq{
			FilePath: filepath,
		}).
		End()
}

func (cl *FilesClient) GetSharingDir(shareID string) (*http.Response, string, []error) {
	resp, body, errs := cl.r.Get(cl.url("/v1/fs/sharings/dirs")).
		AddCookie(cl.token).
		Param(fileshdr.ShareIDQuery, shareID).
		End()
	if len(errs) > 0 {
		return nil, "", errs
	}

	sdResp := &fileshdr.GetSharingDirResp{}
	err := json.Unmarshal([]byte(body), sdResp)
	if err != nil {
		return nil, "", append(errs, err)
	}
	return resp, sdResp.SharingDir, nil
}

func (cl *FilesClient) SearchItems(keyword string) (*http.Response, *fileshdr.SearchItemsResp, []error) {
	resp, body, errs := cl.r.Get(cl.url("/v1/fs/search")).
		AddCookie(cl.token).
		Param(fileshdr.Keyword, keyword).
		End()

	searchResp := &fileshdr.SearchItemsResp{}
	err := json.Unmarshal([]byte(body), searchResp)
	if err != nil {
		errs = append(errs, err)
		return nil, nil, errs
	}
	return resp, searchResp, nil
}
