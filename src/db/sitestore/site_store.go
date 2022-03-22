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

		err = st.setCfg(cfg)
		if err != nil {
			return err
		}
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
	return cfg, nil
}

func (st *SiteStore) setCfg(cfg *SiteConfig) error {
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
