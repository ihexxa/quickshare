package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/ihexxa/gocfg"

	"github.com/ihexxa/quickshare/src/db/sitestore"
	"github.com/ihexxa/quickshare/src/kvstore/boltdbpvd"
)

type Args struct {
	Host    string   `short:"h" long:"host" description:"server hostname"`
	Port    int      `short:"p" long:"port" description:"server port"`
	DbPath  string   `short:"d" long:"db" description:"path of the quickshare.db"`
	Configs []string `short:"c" long:"configs" description:"config path"`
}

// LoadCfg loads the default config, the config in database, config files and arguments in order.
// All config values will be merged into one, and the latter overwrites the former.
// Each config can be part of the whole ServerCfg
func LoadCfg(ctx context.Context, args *Args) (*gocfg.Cfg, error) {
	defaultCfg, err := DefaultConfig()
	if err != nil {
		return nil, err
	}
	cfg, err := gocfg.New(NewConfig()).Load(gocfg.JSONStr(defaultCfg))
	if err != nil {
		return nil, err
	}

	tmpCfg := *cfg
	dbPath, err := getDbPath(&tmpCfg, args.Configs, args.DbPath)
	if err != nil {
		return nil, err
	}
	_, err = os.Stat(dbPath)
	if err == nil {
		cfg, err = mergeDbConfig(ctx, cfg, dbPath)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		} else {
			fmt.Printf("warning: Database does not exist in (%s), skipped\n", dbPath)
		}
	}

	cfg, err = mergeConfigFiles(ctx, cfg, args.Configs)
	if err != nil {
		return nil, err
	}

	return mergeArgs(cfg, args)
}

func mergeDbConfig(ctx context.Context, cfg *gocfg.Cfg, dbPath string) (*gocfg.Cfg, error) {
	kv := boltdbpvd.New(dbPath, 1024)
	defer kv.Close()

	siteStore, err := sitestore.NewSiteStore(kv)
	if err != nil {
		return nil, fmt.Errorf("fail to new site config store: %s", err)
	}

	clientCfg, err := siteStore.GetCfg(ctx)
	if err != nil {
		if errors.Is(err, sitestore.ErrNotFound) {
			return cfg, nil
		}
		return nil, err
	}

	clientCfgBytes, err := json.Marshal(clientCfg)
	if err != nil {
		return nil, err
	}

	cfgStr := fmt.Sprintf(`{"site": %s}`, string(clientCfgBytes))
	return cfg.Load(gocfg.JSONStr(cfgStr))
}

// getDbPath returns db path from arguments, configs, default config in order
// this is a little tricky
func getDbPath(cfg *gocfg.Cfg, configPaths []string, argDbPath string) (string, error) {
	if argDbPath != "" {
		return argDbPath, nil
	}

	var err error

	for _, configPath := range configPaths {
		if strings.HasSuffix(configPath, ".yml") || strings.HasSuffix(configPath, ".yaml") {
			cfg, err = cfg.Load(gocfg.YAML(configPath))
		} else if strings.HasSuffix(configPath, ".json") {
			cfg, err = cfg.Load(gocfg.JSON(configPath))
		} else {
			return "", fmt.Errorf("unknown config file type (.yml .yaml .json are supported): %s", configPath)
		}
		if err != nil {
			return "", err
		}
	}

	return cfg.GrabString("Db.DbPath"), nil
}

func mergeConfigFiles(ctx context.Context, cfg *gocfg.Cfg, configPaths []string) (*gocfg.Cfg, error) {
	var err error

	for _, configPath := range configPaths {
		if strings.HasSuffix(configPath, ".yml") || strings.HasSuffix(configPath, ".yaml") {
			cfg, err = cfg.Load(gocfg.YAML(configPath))
		} else if strings.HasSuffix(configPath, ".json") {
			cfg, err = cfg.Load(gocfg.JSON(configPath))
		} else {
			return nil, fmt.Errorf("unknown config file type (.yml .yaml .json are supported): %s", configPath)
		}
		if err != nil {
			return nil, err
		}
	}

	return cfg, nil
}

func mergeArgs(cfg *gocfg.Cfg, args *Args) (*gocfg.Cfg, error) {
	if args.Host != "" {
		cfg.SetString("Server.Host", args.Host)
	}
	if args.Port != 0 {
		cfg.SetInt("Server.Port", args.Port)
	}
	if args.DbPath != "" {
		cfg.SetString("Db.DbPath", args.DbPath)
	}

	return cfg, nil
}
