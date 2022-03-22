package sitestore

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/ihexxa/quickshare/src/db"
	"github.com/ihexxa/quickshare/src/kvstore"
)

const (
	NsSite     = "NsSite"
	KeySiteCfg = "KeySiteCfg"
)

var (
	DefaultSiteName = "Quickshare"
	DefaultSiteDesc = "Quickshare"
	DefaultBgConfig = &db.BgConfig{
		Repeat:   "repeated",
		Position: "top",
		Align:    "fixed",
		BgColor:  "#ccc",
	}
)

var (
	ErrNotFound = errors.New("site config not found")
)

type ISiteStore interface {
	SetClientCfg(cfg *db.ClientConfig) error
	GetCfg() (*SiteConfig, error)
}

type SiteConfig struct {
	ClientCfg *db.ClientConfig `json:"clientCfg" yaml:"clientCfg"`
}

type SiteStore struct {
	mtx   *sync.RWMutex
	store kvstore.IKVStore
}

func NewSiteStore(store kvstore.IKVStore) (*SiteStore, error) {
	return &SiteStore{
		store: store,
		mtx:   &sync.RWMutex{},
	}, nil
}

func (st *SiteStore) Init(cfg *SiteConfig) error {
	_, ok := st.store.GetStringIn(NsSite, KeySiteCfg)
	if !ok {
		var err error
		if err = st.store.AddNamespace(NsSite); err != nil {
			return err
		}

		return st.setCfg(cfg)
	}
	return nil
}

func (st *SiteStore) getCfg() (*SiteConfig, error) {
	cfgStr, ok := st.store.GetStringIn(NsSite, KeySiteCfg)
	if !ok {
		return nil, ErrNotFound
	}

	cfg := &SiteConfig{}
	err := json.Unmarshal([]byte(cfgStr), cfg)
	if err != nil {
		return nil, err
	}

	if err = CheckCfg(cfg, true); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (st *SiteStore) setCfg(cfg *SiteConfig) error {
	if err := CheckCfg(cfg, false); err != nil {
		return err
	}

	cfgBytes, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	return st.store.SetStringIn(NsSite, KeySiteCfg, string(cfgBytes))
}

func (st *SiteStore) SetClientCfg(cfg *db.ClientConfig) error {
	st.mtx.Lock()
	defer st.mtx.Unlock()

	siteCfg, err := st.getCfg()
	if err != nil {
		return err
	}
	siteCfg.ClientCfg = cfg

	return st.setCfg(siteCfg)
}

func (st *SiteStore) GetCfg() (*SiteConfig, error) {
	st.mtx.RLock()
	defer st.mtx.RUnlock()

	return st.getCfg()
}

func CheckCfg(cfg *SiteConfig, autoFill bool) error {
	if cfg.ClientCfg == nil {
		if !autoFill {
			return errors.New("cfg.ClientCfg not defined")
		}
		cfg.ClientCfg = &db.ClientConfig{
			SiteName: DefaultSiteName,
			SiteDesc: DefaultSiteDesc,
			Bg: &db.BgConfig{
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
		if !autoFill {
			return errors.New("cfg.ClientCfg.Bg not defined")
		}
		cfg.ClientCfg.Bg = DefaultBgConfig
	}
	if cfg.ClientCfg.Bg.Url == "" && cfg.ClientCfg.Bg.BgColor == "" {
		if !autoFill {
			return errors.New("one of Bg.Url or Bg.BgColor must be defined")
		}
		cfg.ClientCfg.Bg.BgColor = DefaultBgConfig.BgColor
	}
	if cfg.ClientCfg.Bg.Repeat == "" {
		cfg.ClientCfg.Bg.Repeat = DefaultBgConfig.Repeat
	}
	if cfg.ClientCfg.Bg.Position == "" {
		cfg.ClientCfg.Bg.Position = DefaultBgConfig.Position
	}
	if cfg.ClientCfg.Bg.Align == "" {
		cfg.ClientCfg.Bg.Align = DefaultBgConfig.Align
	}
	if cfg.ClientCfg.Bg.BgColor == "" {
		cfg.ClientCfg.Bg.BgColor = DefaultBgConfig.BgColor
	}

	return nil
}
