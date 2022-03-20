package db

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/ihexxa/quickshare/src/db/sitestore"
)

const (
	SchemaV2 = "v2" // add size to file info

	UsersNs   = "users"
	InfoNs    = "sharing"
	ShareIDNs = "sharingKey"

	uploadsPrefix = "uploads"
)

var (
	ErrBucketNotFound = errors.New("bucket not found")
	ErrKeyNotFound    = errors.New("key not found")
	ErrCreateExisting = errors.New("create upload info which already exists")
	ErrQuota          = errors.New("quota limit reached")
)

type FileInfo struct {
	IsDir   bool   `json:"isDir"`
	Shared  bool   `json:"shared"`
	ShareID string `json:"shareID"` // for short url
	Sha1    string `json:"sha1"`
	Size    int64  `json:"size"`
}

type Quota struct {
	SpaceLimit         int64 `json:"spaceLimit,string"`
	UploadSpeedLimit   int   `json:"uploadSpeedLimit"`
	DownloadSpeedLimit int   `json:"downloadSpeedLimit"`
}

type Preferences struct {
	Bg         *sitestore.BgConfig `json:"bg"`
	CSSURL     string              `json:"cssURL"`
	LanPackURL string              `json:"lanPackURL"`
	Lan        string              `json:"lan"`
	Theme      string              `json:"theme"`
	Avatar     string              `json:"avatar"`
	Email      string              `json:"email"`
}

func ComparePreferences(p1, p2 *Preferences) bool {
	return p1.CSSURL == p2.CSSURL &&
		p1.LanPackURL == p2.LanPackURL &&
		p1.Lan == p2.Lan &&
		p1.Theme == p2.Theme &&
		p1.Avatar == p2.Avatar &&
		p1.Email == p2.Email &&
		reflect.DeepEqual(p1.Bg, p2.Bg)
}

type User struct {
	ID          uint64       `json:"id,string"`
	Name        string       `json:"name"`
	Pwd         string       `json:"pwd"`
	Role        string       `json:"role"`
	UsedSpace   int64        `json:"usedSpace,string"`
	Quota       *Quota       `json:"quota"`
	Preferences *Preferences `json:"preferences"`
}

type UploadInfo struct {
	RealFilePath string `json:"realFilePath"`
	Size         int64  `json:"size"`
	Uploaded     int64  `json:"uploaded"`
}

func UploadNS(user string) string {
	return fmt.Sprintf("%s/%s", uploadsPrefix, user)
}
