package server

import (
	"encoding/json"

	"github.com/ihexxa/quickshare/src/db"
)

const fileIndexPath = "/fileindex.jsonl"

type DbConfig struct {
	DbPath string `json:"dbPath" yaml:"dbPath"`
}

type FSConfig struct {
	Root              string `json:"root" yaml:"root"`
	OpensLimit        int    `json:"opensLimit" yaml:"opensLimit"`
	OpenTTL           int    `json:"openTTL" yaml:"openTTL"`
	PublicPath        string `json:"publicPath" yaml:"publicPath"`
	SearchResultLimit int    `json:"searchResultLimit" yaml:"searchResultLimit"`
	InitFileIndex     bool   `json:"initFileIndex" yaml:"initFileIndex"`
}

type UsersCfg struct {
	EnableAuth         bool          `json:"enableAuth" yaml:"enableAuth"`
	DefaultAdmin       string        `json:"defaultAdmin" yaml:"defaultAdmin" cfg:"env"`
	DefaultAdminPwd    string        `json:"defaultAdminPwd" yaml:"defaultAdminPwd" cfg:"env"`
	CookieTTL          int           `json:"cookieTTL" yaml:"cookieTTL"`
	CookieSecure       bool          `json:"cookieSecure" yaml:"cookieSecure"`
	CookieHttpOnly     bool          `json:"cookieHttpOnly" yaml:"cookieHttpOnly"`
	MinUserNameLen     int           `json:"minUserNameLen" yaml:"minUserNameLen"`
	MinPwdLen          int           `json:"minPwdLen" yaml:"minPwdLen"`
	CaptchaWidth       int           `json:"captchaWidth" yaml:"captchaWidth"`
	CaptchaHeight      int           `json:"captchaHeight" yaml:"captchaHeight"`
	CaptchaEnabled     bool          `json:"captchaEnabled" yaml:"captchaEnabled"`
	UploadSpeedLimit   int           `json:"uploadSpeedLimit" yaml:"uploadSpeedLimit"`
	DownloadSpeedLimit int           `json:"downloadSpeedLimit" yaml:"downloadSpeedLimit"`
	SpaceLimit         int           `json:"spaceLimit" yaml:"spaceLimit"`
	LimiterCapacity    int           `json:"limiterCapacity" yaml:"limiterCapacity"`
	LimiterCyc         int           `json:"limiterCyc" yaml:"limiterCyc"`
	PredefinedUsers    []*db.UserCfg `json:"predefinedUsers" yaml:"predefinedUsers"`
}

type Secrets struct {
	TokenSecret string `json:"tokenSecret" yaml:"tokenSecret" cfg:"env"`
}

type ServerCfg struct {
	Debug          bool           `json:"debug" yaml:"debug"`
	Host           string         `json:"host" yaml:"host"`
	Port           int            `json:"port" yaml:"port" cfg:"env"`
	ReadTimeout    int            `json:"readTimeout" yaml:"readTimeout"`
	WriteTimeout   int            `json:"writeTimeout" yaml:"writeTimeout"`
	MaxHeaderBytes int            `json:"maxHeaderBytes" yaml:"maxHeaderBytes"`
	Dynamic        *db.SiteConfig `json:"dynamic" yaml:"dynamic"`
}

type WorkerPoolCfg struct {
	QueueSize   int `json:"queueSize" yaml:"queueSize"`
	SleepCyc    int `json:"sleepCyc" yaml:"sleepCyc"`
	WorkerCount int `json:"workerCount" yaml:"workerCount"`
}

type Config struct {
	Users   *UsersCfg      `json:"users" yaml:"users"`
	Fs      *FSConfig      `json:"fs" yaml:"fs"`
	Secrets *Secrets       `json:"secrets" yaml:"secrets"`
	Workers *WorkerPoolCfg `json:"workers" yaml:"workers"`
	Db      *DbConfig      `json:"db" yaml:"db"`
	Server  *ServerCfg     `json:"server" yaml:"server"`
}

func NewConfig() *Config {
	return &Config{}
}

func DefaultConfig() (string, error) {
	cfgBytes, err := json.Marshal(DefaultConfigStruct())
	return string(cfgBytes), err
}

func DefaultConfigStruct() *Config {
	return &Config{
		Fs: &FSConfig{
			Root:              "quickshare",
			OpensLimit:        1024,
			OpenTTL:           60, // 1 min
			PublicPath:        "static/public",
			SearchResultLimit: 16,
			InitFileIndex:     true,
		},
		Users: &UsersCfg{
			EnableAuth:         true,
			DefaultAdmin:       "",
			DefaultAdminPwd:    "",
			CookieTTL:          3600 * 24 * 7, // 1 week
			CookieSecure:       false,
			CookieHttpOnly:     true,
			MinUserNameLen:     3,
			MinPwdLen:          8,
			CaptchaWidth:       256,
			CaptchaHeight:      60,
			CaptchaEnabled:     true,
			UploadSpeedLimit:   1024 * 1024,       // B
			DownloadSpeedLimit: 1024 * 1024,       // B
			SpaceLimit:         1024 * 1024 * 100, // 100MB
			LimiterCapacity:    1000,
			LimiterCyc:         1000, // 1s
			PredefinedUsers:    []*db.UserCfg{},
		},
		Secrets: &Secrets{
			TokenSecret: "", // it will auto generated if it is left as empty
		},
		Server: &ServerCfg{
			Debug:          false,
			Host:           "0.0.0.0",
			Port:           8686,
			ReadTimeout:    2000,
			WriteTimeout:   1000 * 3600 * 24, // 1 day
			MaxHeaderBytes: 512,
			Dynamic: &db.SiteConfig{
				ClientCfg: &db.ClientConfig{
					SiteName: "Quickshare",
					SiteDesc: "Quick and simple file sharing",
					Bg: &db.BgConfig{
						Url:      "",
						Repeat:   "repeat",
						Position: "center",
						Align:    "fixed",
						BgColor:  "",
					},
					AllowSetBg: false,
					AutoTheme:  true,
				},
			},
		},
		Workers: &WorkerPoolCfg{
			QueueSize:   1024,
			SleepCyc:    1,
			WorkerCount: 2,
		},
		Db: &DbConfig{
			DbPath: "quickshare.sqlite",
		},
	}
}
