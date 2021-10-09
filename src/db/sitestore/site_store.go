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
	SiteName string    `json:"siteName"`
	SiteDesc string    `json:"siteDesc"`
	Bg       *BgConfig `json:"bg"`
}

type BgConfig struct {
	Url      string `json:"url"`
	Repeat   string `json:"repeat"`
	Position string `json:"position"`
	Align    string `json:"align"`
}

type SiteConfig struct {
	ClientCfg *ClientConfig `json:"clientCfg"`
}

type SiteStore struct {
	mtx   *sync.RWMutex
	store kvstore.IKVStore
}

func NewSiteStore(store kvstore.IKVStore) (*SiteStore, error) {
	_, ok := store.GetStringIn(InitNs, InitTimeKey)
	if !ok {
		var err error
		for _, nsName := range []string{
			InitNs,
			SiteNs,
		} {
			if err = store.AddNamespace(nsName); err != nil {
				return nil, err
			}
		}
	}

	err := store.SetStringIn(InitNs, InitTimeKey, fmt.Sprintf("%d", time.Now().Unix()))
	if err != nil {
		return nil, err
	}

	return &SiteStore{
		store: store,
		mtx:   &sync.RWMutex{},
	}, nil
}

func (fi *SiteStore) SetClientCfg(cfg *ClientConfig) error {
	fi.mtx.Lock()
	defer fi.mtx.Unlock()

	siteCfg := &SiteConfig{}
	cfgStr, ok := fi.store.GetStringIn(SiteNs, SiteCfgKey)
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
	return fi.store.SetStringIn(SiteNs, SiteCfgKey, string(cfgBytes))
}

func (fi *SiteStore) GetCfg() (*SiteConfig, error) {
	fi.mtx.RLock()
	defer fi.mtx.RUnlock()

	cfgStr, ok := fi.store.GetStringIn(SiteNs, SiteCfgKey)
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
