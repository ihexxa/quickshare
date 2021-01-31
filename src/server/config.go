package server

import "encoding/json"

type FSConfig struct {
	Root       string `json:"root"`
	OpensLimit int    `json:"opensLimit"`
	OpenTTL    int    `json:"openTTL"`
}

type UsersCfg struct {
	EnableAuth      bool   `json:"enableAuth"`
	DefaultAdmin    string `json:"defaultAdmin" cfg:"env"`
	DefaultAdminPwd string `json:"defaultAdminPwd" cfg:"env"`
	CookieTTL       int    `json:"cookieTTL"`
	CookieSecure    bool   `json:"cookieSecure"`
	CookieHttpOnly  bool   `json:"cookieHttpOnly"`
}

type Secrets struct {
	TokenSecret string `json:"tokenSecret" cfg:"env"`
}

type ServerCfg struct {
	Debug          bool   `json:"debug"`
	Host           string `json:"host"`
	Port           int    `json:"port"`
	ReadTimeout    int    `json:"readTimeout"`
	WriteTimeout   int    `json:"writeTimeout"`
	MaxHeaderBytes int    `json:"maxHeaderBytes"`
	PublicPath     string `json:"publicPath"`
}

type Config struct {
	Fs      *FSConfig  `json:"fs"`
	Secrets *Secrets   `json:"secrets"`
	Server  *ServerCfg `json:"server"`
	Users   *UsersCfg  `json:"users"`
}

func NewConfig() *Config {
	return &Config{}
}

func DefaultConfig() (string, error) {
	defaultCfg := &Config{
		Fs: &FSConfig{
			Root:       ".",
			OpensLimit: 128,
			OpenTTL:    60, // 1 min
		},
		Users: &UsersCfg{
			EnableAuth:      true,
			DefaultAdmin:    "",
			DefaultAdminPwd: "",
			CookieTTL:       3600 * 24 * 7, // 1 week
			CookieSecure:    false,
			CookieHttpOnly:  true,
		},
		Secrets: &Secrets{
			TokenSecret: "",
		},
		Server: &ServerCfg{
			Debug:          false,
			Host:           "0.0.0.0",
			Port:           8686,
			ReadTimeout:    2000,
			WriteTimeout:   1000 * 3600 * 24, // 1 day
			MaxHeaderBytes: 512,
			PublicPath:     "public",
		},
	}

	cfgBytes, err := json.Marshal(defaultCfg)
	return string(cfgBytes), err
}
