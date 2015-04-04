package main

import (
	"bytes"
	"os"
	"path"

	"github.com/BurntSushi/toml"
)

var (
	DefaultNodePath  = ""
	DefaultCacheFile = ""
)

type NodeConfig struct {
	Path      string
	CacheFile string
	Cache     *NodeCache
	HasCache  bool
}

type NodeCache struct {
	Id string
}

func NewNodeConfig() (*NodeConfig, error) {
	config := &NodeConfig{}

	p, err := HomeDir()

	if err != nil {
		return config, err
	}

	Log.Debugf("Home path set to %s.", p)

	DefaultNodePath = path.Join(p, "."+Name)
	DefaultCacheFile = path.Join(DefaultNodePath, "node.cache")

	Log.Debugf("Default Node Config Path: %s.", DefaultNodePath)
	Log.Debugf("Default Node Cache File: %s.", DefaultCacheFile)

	if err := os.MkdirAll(DefaultNodePath, 0777); err != nil {
		return config, err
	}

	config.Path = DefaultNodePath
	config.CacheFile = DefaultCacheFile
	config.HasCache = false
	config.Cache = &NodeCache{}

	if IsExist(DefaultCacheFile) {
		if _, err := toml.DecodeFile(DefaultCacheFile, config.Cache); err != nil {
			return config, err
		}

		config.HasCache = true
	}

	return config, err
}

func (self NodeConfig) WriteCache() error {
	var cache bytes.Buffer

	e := toml.NewEncoder(&cache)

	if err := e.Encode(self.Cache); err != nil {
		return err
	}

	f, err := os.Create(self.CacheFile)
	defer f.Close()

	if err != nil {
		return err
	}

	if _, err := f.WriteString(cache.String()); err != nil {
		return err
	}

	f.Sync()

	return err
}
