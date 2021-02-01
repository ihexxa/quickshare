package server

import "encoding/json"

type FSConfig struct {
	Root       string `json:"root"`
	OpensLimit int    `json:"opensLimit"`
	OpenTTL    int    `json:"openTTL"`
}

type UsersCfg struct {
	EnableAuth      bool   `json:"enableAuth" yaml:"enableAuth"`
	DefaultAdmin    string `json:"defaultAdmin" yaml:"defaultAdmin" cfg:"env"`
	DefaultAdminPwd string `json:"defaultAdminPwd" yaml:"defaultAdminPwd" cfg:"env"`
	CookieTTL       int    `json:"cookieTTL" yaml:"cookieTTL"`
	CookieSecure    bool   `json:"cookieSecure" yaml:"cookieSecure"`
	CookieHttpOnly  bool   `json:"cookieHttpOnly" yaml:"cookieHttpOnly"`
}

type Secrets struct {
	TokenSecret string `json:"tokenSecret" yaml:"tokenSecret" cfg:"env"`
}

type ServerCfg struct {
	Debug          bool   `json:"debug" yaml:"debug"`
	Host           string `json:"host" yaml:"host"`
	Port           int    `json:"port" yaml:"port"`
	ReadTimeout    int    `json:"readTimeout" yaml:"readTimeout"`
	WriteTimeout   int    `json:"writeTimeout" yaml:"writeTimeout"`
	MaxHeaderBytes int    `json:"maxHeaderBytes" yaml:"maxHeaderBytes"`
	PublicPath     string `json:"publicPath" yaml:"publicPath"`
}

type Config struct {
	Fs      *FSConfig  `json:"fs" yaml:"fs"`
	Secrets *Secrets   `json:"secrets" yaml:"secrets"`
	Server  *ServerCfg `json:"server" yaml:"server"`
	Users   *UsersCfg  `json:"users" yaml:"users"`
}

func NewConfig() *Config {
	return &Config{}
}

func DefaultConfig() (string, error) {
	defaultCfg := &Config{
		Fs: &FSConfig{
			Root:       "root",
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
