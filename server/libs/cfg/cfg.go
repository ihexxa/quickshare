package cfg

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"strconv"
	"strings"
)

type Config struct {
	AppName       string
	AdminId       string
	AdminPwd      string
	SecretKey     string
	SecretKeyByte []byte `json:",omitempty"`
	// server
	Production bool
	HostName   string
	Port       int
	// performance
	MaxUpBytesPerSec   int64
	MaxDownBytesPerSec int64
	MaxRangeLength     int64
	Timeout            int // millisecond
	ReadTimeout        int
	WriteTimeout       int
	IdleTimeout        int
	WorkerPoolSize     int
	TaskQueueSize      int
	QueueSize          int
	ParseFormBufSize   int64
	MaxHeaderBytes     int
	DownLimit          int
	MaxShares          int
	LocalFileLimit     int
	// Cookie
	CookieDomain   string
	CookieHttpOnly bool
	CookieMaxAge   int
	CookiePath     string
	CookieSecure   bool
	// keys
	KeyAdminId       string
	KeyAdminPwd      string
	KeyToken         string
	KeyFileName      string
	KeyFileSize      string
	KeyShareId       string
	KeyStart         string
	KeyLen           string
	KeyChunk         string
	KeyAct           string
	KeyExpires       string
	KeyDownLimit     string
	ActStartUpload   string
	ActUpload        string
	ActFinishUpload  string
	ActLogin         string
	ActLogout        string
	ActShadowId      string
	ActPublishId     string
	ActSetDownLimit  string
	ActAddLocalFiles string
	// resource id
	AllUsers string
	// opIds
	OpIdIpVisit  int16
	OpIdUpload   int16
	OpIdDownload int16
	OpIdLogin    int16
	OpIdGetFInfo int16
	OpIdDelFInfo int16
	OpIdOpFInfo  int16
	// local
	PathLocal         string
	PathLogin         string
	PathDownloadLogin string
	PathDownload      string
	PathUpload        string
	PathStartUpload   string
	PathFinishUpload  string
	PathFileInfo      string
	PathClient        string
	// rate Limiter
	LimiterCap     int64
	LimiterTtl     int32
	LimiterCyc     int32
	BucketCap      int16
	SpecialCapsStr map[string]int16
	SpecialCaps    map[int16]int16
}

func NewConfig() *Config {
	config := &Config{
		// secrets
		AppName:       "qs",
		AdminId:       "admin",
		AdminPwd:      "qs",
		SecretKey:     "qs",
		SecretKeyByte: []byte("qs"),
		// server
		Production: true,
		HostName:   "localhost",
		Port:       8888,
		// performance
		MaxUpBytesPerSec:   500 * 1000,
		MaxDownBytesPerSec: 500 * 1000,
		MaxRangeLength:     10 * 1024 * 1024,
		Timeout:            500, // millisecond,
		ReadTimeout:        500,
		WriteTimeout:       43200000,
		IdleTimeout:        10000,
		WorkerPoolSize:     2,
		TaskQueueSize:      2,
		QueueSize:          2,
		ParseFormBufSize:   600,
		MaxHeaderBytes:     1 << 15, // 32KB
		DownLimit:          -1,
		MaxShares:          1 << 31,
		LocalFileLimit:     -1,
		// Cookie
		CookieDomain:   "",
		CookieHttpOnly: false,
		CookieMaxAge:   3600 * 24 * 30, // one week,
		CookiePath:     "/",
		CookieSecure:   false,
		// keys
		KeyAdminId:       "adminid",
		KeyAdminPwd:      "adminpwd",
		KeyToken:         "token",
		KeyFileName:      "fname",
		KeyFileSize:      "size",
		KeyShareId:       "shareid",
		KeyStart:         "start",
		KeyLen:           "len",
		KeyChunk:         "chunk",
		KeyAct:           "act",
		KeyExpires:       "expires",
		KeyDownLimit:     "downlimit",
		ActStartUpload:   "startupload",
		ActUpload:        "upload",
		ActFinishUpload:  "finishupload",
		ActLogin:         "login",
		ActLogout:        "logout",
		ActShadowId:      "shadowid",
		ActPublishId:     "publishid",
		ActSetDownLimit:  "setdownlimit",
		ActAddLocalFiles: "addlocalfiles",
		AllUsers:         "allusers",
		// opIds
		OpIdIpVisit:  0,
		OpIdUpload:   1,
		OpIdDownload: 2,
		OpIdLogin:    3,
		OpIdGetFInfo: 4,
		OpIdDelFInfo: 5,
		OpIdOpFInfo:  6,
		// local
		PathLocal:         "files",
		PathLogin:         "/login",
		PathDownloadLogin: "/download-login",
		PathDownload:      "/download",
		PathUpload:        "/upload",
		PathStartUpload:   "/startupload",
		PathFinishUpload:  "/finishupload",
		PathFileInfo:      "/fileinfo",
		PathClient:        "/",
		// rate Limiter
		LimiterCap: 256,  // how many op supported for each user
		LimiterTtl: 3600, // second
		LimiterCyc: 1,    // second
		BucketCap:  3,    // how many op can do per LimiterCyc sec
		SpecialCaps: map[int16]int16{
			0: 5, // ip
			1: 1, // upload
			2: 1, // download
			3: 1, // login
		},
	}

	return config
}

func NewConfigFrom(path string) *Config {
	configBytes, readErr := ioutil.ReadFile(path)
	if readErr != nil {
		panic(fmt.Sprintf("config file not found: %s", path))
	}

	config := &Config{}
	marshalErr := json.Unmarshal(configBytes, config)

	// TODO: look for a better solution
	config.SpecialCaps = make(map[int16]int16)
	for strKey, value := range config.SpecialCapsStr {
		key, parseKeyErr := strconv.ParseInt(strKey, 10, 16)
		if parseKeyErr != nil {
			panic("fail to parse SpecialCapsStr, its type should be map[int16]int16")
		}
		config.SpecialCaps[int16(key)] = value
	}

	if marshalErr != nil {
		panic("config file format is incorrect")
	}

	config.SecretKeyByte = []byte(config.SecretKey)
	if config.HostName == "" {
		hostName, err := GetLocalAddr()
		if err != nil {
			panic(err)
		}
		config.HostName = hostName.String()
	}

	return config
}

func GetLocalAddr() (net.IP, error) {
	fmt.Println(`config.HostName is empty(""), choose one IP for listening automatically.`)
	infs, err := net.Interfaces()
	if err != nil {
		panic("fail to get net interfaces")
	}

	for _, inf := range infs {
		if inf.Flags&4 != 4 && !strings.Contains(inf.Name, "docker") {
			addrs, err := inf.Addrs()
			if err != nil {
				panic("fail to get addrs of interface")
			}
			for _, addr := range addrs {
				switch v := addr.(type) {
				case *net.IPAddr:
					if !strings.Contains(v.IP.String(), ":") {
						return v.IP, nil
					}
				case *net.IPNet:
					if !strings.Contains(v.IP.String(), ":") {
						return v.IP, nil
					}
				}
			}
		}
	}

	return nil, errors.New("no addr found")
}
