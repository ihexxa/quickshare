package db

import (
	"errors"
	"fmt"
	"reflect"
)

const (
	UserSchemaNs = "UserSchemaNs"
	FileSchemaNs = "FileSchemaNs"
	UserIDsNs    = "UserIDsNs"
	UsersNs      = "UsersNs"
	RolesNs      = "RolesNs"
	FileInfoNs   = "FileInfoNs"
	ShareIDNs    = "ShareIDNs"

	uploadsPrefix = "uploads"

	KeyInitTime = "keyInitTime"

	AdminRole   = "admin"
	UserRole    = "user"
	VisitorRole = "visitor"
)

var (
	ErrInvalidFileInfo    = errors.New("invalid fileInfo")
	ErrInvalidUser        = errors.New("invalid user")
	ErrInvalidQuota       = errors.New("invalid quota")
	ErrInvalidPreferences = errors.New("invalid preferences")
	ErrBucketNotFound     = errors.New("bucket not found")
	ErrKeyNotFound        = errors.New("key not found")
	ErrKeyExisting        = errors.New("key is existing")
	ErrCreateExisting     = errors.New("create upload info which already exists")
	ErrQuota              = errors.New("quota limit reached")

	DefaultSiteName = "Quickshare"
	DefaultSiteDesc = "Quickshare"
	DefaultBgConfig = &BgConfig{
		Repeat:   "repeat",
		Position: "top",
		Align:    "fixed",
		BgColor:  "#ccc",
	}
	DefaultAllowSetBg = false
	DefaultAutoTheme  = true
	BgRepeatValues    = map[string]bool{
		"repeat-x":  true,
		"repeat-y":  true,
		"repeat":    true,
		"space":     true,
		"round":     true,
		"no-repeat": true,
	}
	BgAlignValues = map[string]bool{
		"scroll": true,
		"fixed":  true,
		"local":  true,
	}
	BgPositionValues = map[string]bool{
		"top":    true,
		"bottom": true,
		"left":   true,
		"right":  true,
		"center": true,
	}

	DefaultCSSURL     = ""
	DefaultLanPackURL = ""
	DefaultLan        = "en_US"
	DefaultTheme      = "light"
	DefaultAvatar     = ""
	DefaultEmail      = ""

	DefaultSpaceLimit         = int64(1024 * 1024 * 1024) // 1GB
	DefaultUploadSpeedLimit   = 50 * 1024 * 1024          // 50MB
	DefaultDownloadSpeedLimit = 50 * 1024 * 1024          // 50MB
	VisitorUploadSpeedLimit   = 10 * 1024 * 1024          // 10MB
	VisitorDownloadSpeedLimit = 10 * 1024 * 1024          // 10MB

	DefaultPreferences = Preferences{
		Bg:         DefaultBgConfig,
		CSSURL:     DefaultCSSURL,
		LanPackURL: DefaultLanPackURL,
		Lan:        DefaultLan,
		Theme:      DefaultTheme,
		Avatar:     DefaultAvatar,
		Email:      DefaultEmail,
	}
)

type FileInfo struct {
	IsDir   bool   `json:"isDir" yaml:"isDir"`
	Shared  bool   `json:"shared" yaml:"shared"`
	ShareID string `json:"shareID" yaml:"shareID"`
	Sha1    string `json:"sha1" yaml:"sha1"`
	Size    int64  `json:"size" yaml:"size"`
}

type UserCfg struct {
	Name string `json:"name" yaml:"name"`
	Role string `json:"role" yaml:"role"`
	Pwd  string `json:"pwd" yaml:"pwd"`
}

type Quota struct {
	SpaceLimit         int64 `json:"spaceLimit,string" yaml:"spaceLimit,string"`
	UploadSpeedLimit   int   `json:"uploadSpeedLimit" yaml:"uploadSpeedLimit"`
	DownloadSpeedLimit int   `json:"downloadSpeedLimit" yaml:"downloadSpeedLimit"`
}

type Preferences struct {
	Bg         *BgConfig `json:"bg" yaml:"bg"`
	CSSURL     string    `json:"cssURL" yaml:"cssURL"`
	LanPackURL string    `json:"lanPackURL" yaml:"lanPackURL"`
	Lan        string    `json:"lan" yaml:"lan"`
	Theme      string    `json:"theme" yaml:"theme"`
	Avatar     string    `json:"avatar" yaml:"avatar"`
	Email      string    `json:"email" yaml:"email"`
}

type SiteConfig struct {
	ClientCfg *ClientConfig `json:"clientCfg" yaml:"clientCfg"`
}

type ClientConfig struct {
	SiteName   string    `json:"siteName" yaml:"siteName"`
	SiteDesc   string    `json:"siteDesc" yaml:"siteDesc"`
	Bg         *BgConfig `json:"bg" yaml:"bg"`
	AllowSetBg bool      `json:"allowSetBg" yaml:"allowSetBg"`
	AutoTheme  bool      `json:"autoTheme" yaml:"autoTheme"`
}

type BgConfig struct {
	Url      string `json:"url" yaml:"url"`
	Repeat   string `json:"repeat" yaml:"repeat"`
	Position string `json:"position" yaml:"position"`
	Align    string `json:"align" yaml:"align"`
	BgColor  string `json:"bgColor" yaml:"bgColor"`
}

type User struct {
	ID          uint64       `json:"id,string" yaml:"id,string"`
	Name        string       `json:"name" yaml:"name"`
	Pwd         string       `json:"pwd" yaml:"pwd"`
	Role        string       `json:"role" yaml:"role"`
	UsedSpace   int64        `json:"usedSpace,string" yaml:"usedSpace,string"`
	Quota       *Quota       `json:"quota" yaml:"quota"`
	Preferences *Preferences `json:"preferences" yaml:"preferences"`
}

type UploadInfo struct {
	RealFilePath string `json:"realFilePath" yaml:"realFilePath"`
	Size         int64  `json:"size" yaml:"size"`
	Uploaded     int64  `json:"uploaded" yaml:"uploaded"`
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

func UploadNS(user string) string {
	return fmt.Sprintf("%s/%s", uploadsPrefix, user)
}

func CheckSiteCfg(cfg *SiteConfig, fillDefault bool) error {
	if cfg.ClientCfg == nil {
		if !fillDefault {
			return errors.New("cfg.ClientCfg not defined")
		}
		cfg.ClientCfg = &ClientConfig{
			SiteName:   DefaultSiteName,
			SiteDesc:   DefaultSiteDesc,
			AllowSetBg: DefaultAllowSetBg,
			AutoTheme:  DefaultAutoTheme,
			Bg: &BgConfig{
				Url:      DefaultBgConfig.Url,
				Repeat:   DefaultBgConfig.Repeat,
				Position: DefaultBgConfig.Position,
				Align:    DefaultBgConfig.Align,
				BgColor:  DefaultBgConfig.BgColor,
			},
		}

		return nil
	}

	if cfg.ClientCfg.SiteName == "" {
		cfg.ClientCfg.SiteName = DefaultSiteName
	}
	if cfg.ClientCfg.SiteDesc == "" {
		cfg.ClientCfg.SiteDesc = DefaultSiteDesc
	}

	if cfg.ClientCfg.Bg == nil {
		if !fillDefault {
			return errors.New("cfg.ClientCfg.Bg not defined")
		}
		cfg.ClientCfg.Bg = DefaultBgConfig
	}
	if err := CheckBgConfig(cfg.ClientCfg.Bg, fillDefault); err != nil {
		return err
	}

	return nil
}

// TODO: check upper and lower limit
func CheckQuota(quota *Quota) error {
	if quota.SpaceLimit < 0 {
		return ErrInvalidQuota
	}
	if quota.UploadSpeedLimit < 0 {
		return ErrInvalidQuota
	}
	if quota.DownloadSpeedLimit < 0 {
		return ErrInvalidQuota
	}
	return nil
}

func CheckPreferences(prefers *Preferences, fillDefault bool) error {
	if prefers.CSSURL == "" {
		prefers.CSSURL = DefaultCSSURL
	}
	if prefers.LanPackURL == "" {
		prefers.LanPackURL = DefaultLanPackURL
	}
	if prefers.Lan == "" {
		if !fillDefault {
			return ErrInvalidPreferences
		}
		prefers.Lan = DefaultLan
	}
	if prefers.Theme == "" {
		if !fillDefault {
			return ErrInvalidPreferences
		}
		prefers.Theme = DefaultTheme
	}
	if prefers.Avatar == "" {
		prefers.Avatar = DefaultAvatar
	}
	// TODO: add strict checking
	if prefers.Email == "" {
		prefers.Email = DefaultEmail
	}
	if prefers.Bg == nil {
		if !fillDefault {
			return ErrInvalidPreferences
		}
		prefers.Bg = DefaultBgConfig
	}
	if err := CheckBgConfig(prefers.Bg, fillDefault); err != nil {
		return err
	}

	return nil
}

func CheckBgConfig(cfg *BgConfig, fillDefault bool) error {
	if cfg.Url == "" && cfg.BgColor == "" {
		if !fillDefault {
			return errors.New("one of Bg.Url or Bg.BgColor must be defined")
		}
		cfg.BgColor = DefaultBgConfig.BgColor
	}
	if !BgRepeatValues[cfg.Repeat] {
		return fmt.Errorf("invalid repeat value (%s)", cfg.Repeat)
	}
	if !BgPositionValues[cfg.Position] {
		return fmt.Errorf("invalid position value (%s)", cfg.Position)
	}
	if !BgAlignValues[cfg.Align] {
		return fmt.Errorf("invalid align value (%s)", cfg.Align)
	}

	if cfg.Repeat == "" {
		cfg.Repeat = DefaultBgConfig.Repeat
	}
	if cfg.Position == "" {
		cfg.Position = DefaultBgConfig.Position
	}
	if cfg.Align == "" {
		cfg.Align = DefaultBgConfig.Align
	}
	if cfg.BgColor == "" {
		cfg.BgColor = DefaultBgConfig.BgColor
	}
	return nil
}

func CheckUser(user *User, fillDefault bool) error {
	if user.ID == 0 && user.Role != AdminRole {
		return fmt.Errorf("invalid ID: (%w)", ErrInvalidUser)
	}
	// TODO: add length check
	if user.Name == "" || user.Role == "" {
		return fmt.Errorf("invalid Name/pwd/role: (%w)", ErrInvalidUser)
	}
	if user.UsedSpace < 0 {
		return fmt.Errorf("invalid UsedSpace: (%w)", ErrInvalidUser)
	}
	if user.Quota == nil || user.Preferences == nil {
		return fmt.Errorf("invalid Quota: (%w)", ErrInvalidUser)
	}

	var err error
	if err = CheckQuota(user.Quota); err != nil {
		return err
	}
	if err = CheckPreferences(user.Preferences, fillDefault); err != nil {
		return err
	}

	return nil
}

// TODO: auto trigger hash generating
func CheckFileInfo(info *FileInfo, fillDefault bool) error {
	if (info.Shared && info.ShareID == "") || (!info.Shared && info.ShareID != "") {
		return fmt.Errorf("shared and ShareID are in conflict: %w", ErrInvalidFileInfo)
	}
	if !info.IsDir && (info.Shared || info.ShareID != "") {
		return fmt.Errorf("dir can not be shared: %w", ErrInvalidFileInfo)
	}
	return nil
}
