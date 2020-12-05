package server

type FSConfig struct {
	Root       string `json:"root"`
	OpensLimit int    `json:"opensLimit"`
	OpenTTL    int    `json:"openTTL"`
}

type Secrets struct {
	TokenSecret string `json:"tokenSecret" cfg:"env"`
}

type ServerCfg struct {
	Debug          bool   `json:"debug"`
	Addr           string `json:"addr"`
	ReadTimeout    int    `json:"readTimeout"`
	WriteTimeout   int    `json:"writeTimeout"`
	MaxHeaderBytes int    `json:"maxHeaderBytes"`
}

type Config struct {
	Fs      *FSConfig  `json:"fs"`
	Secrets *Secrets   `json:"secrets"`
	Server  *ServerCfg `json:"server"`
}

func NewEmptyConfig() *Config {
	return &Config{}
}

func NewDefaultConfig() *Config {
	return &Config{
		Fs: &FSConfig{
			Root:       ".",
			OpensLimit: 128,
			OpenTTL:    60, // 1 min
		},
		Secrets: &Secrets{
			TokenSecret: "",
		},
		Server: &ServerCfg{
			Debug:          false,
			Addr:           "127.0.0.1:8888",
			ReadTimeout:    2000,
			WriteTimeout:   2000,
			MaxHeaderBytes: 512,
		},
	}
}
