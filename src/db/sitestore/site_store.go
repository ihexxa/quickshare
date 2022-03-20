package sitestore

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ihexxa/quickshare/src/kvstore"
)

const (
	InitNs      = "SiteInit"
	SiteNs      = "Site"
	InitTimeKey = "initTime"
	SiteCfgKey  = "siteCfg"
)

var (
	ErrNotFound = errors.New("site config not found")
)

func IsNotFound(err error) bool {
	return err == ErrNotFound
}

type ISiteStore interface {
	SetClientCfg(cfg *ClientConfig) error
	GetCfg() (*SiteConfig, error)
}

type ClientConfig struct {
	SiteName string    `json:"siteName" yaml:"siteName"`
	SiteDesc string    `json:"siteDesc" yaml:"siteDesc"`
	Bg       *BgConfig `json:"bg" yaml:"bg"`
}

type BgConfig struct {
	Url      string `json:"url" yaml:"url"`
	Repeat   string `json:"repeat" yaml:"repeat"`
	Position string `json:"position" yaml:"position"`
	Align    string `json:"align" yaml:"align"`
	BgColor  string `json:"bgColor" yaml:"bgColor"`
}

type SiteConfig struct {
	ClientCfg *ClientConfig `json:"clientCfg" yaml:"clientCfg"`
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
	_, ok := st.store.GetStringIn(InitNs, InitTimeKey)
	if !ok {
		var err error
		for _, nsName := range []string{
			InitNs,
			SiteNs,
		} {
			if err = st.store.AddNamespace(nsName); err != nil {
				return err
			}
		}

		// TODO: replace following with setConfig
		err = st.SetClientCfg(cfg.ClientCfg)
		if err != nil {
			return err
		}
		err = st.store.SetStringIn(InitNs, InitTimeKey, fmt.Sprintf("%d", time.Now().Unix()))
		if err != nil {
			return err
		}
	}

	return nil
}

func (st *SiteStore) SetClientCfg(cfg *ClientConfig) error {
	st.mtx.Lock()
	defer st.mtx.Unlock()

	siteCfg := &SiteConfig{}
	cfgStr, ok := st.store.GetStringIn(SiteNs, SiteCfgKey)
	if ok {
		err := json.Unmarshal([]byte(cfgStr), siteCfg)
		if err != nil {
			return err
		}
	}
	siteCfg.ClientCfg = cfg

	cfgBytes, err := json.Marshal(siteCfg)
	if err != nil {
		return err
	}
	return st.store.SetStringIn(SiteNs, SiteCfgKey, string(cfgBytes))
}

func (st *SiteStore) GetCfg() (*SiteConfig, error) {
	st.mtx.RLock()
	defer st.mtx.RUnlock()

	cfgStr, ok := st.store.GetStringIn(SiteNs, SiteCfgKey)
	if !ok {
		return nil, ErrNotFound
	}
	siteCfg := &SiteConfig{}
	err := json.Unmarshal([]byte(cfgStr), siteCfg)
	if err != nil {
		return nil, err
	}
	return siteCfg, nil
}
