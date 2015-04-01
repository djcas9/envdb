package main

import (
	"bytes"
	"os"
	"path"

	"github.com/BurntSushi/toml"
)

var (
	DefaultAgentPath = ""
	DefaultCacheFile = ""
)

type AgentConfig struct {
	Path      string
	CacheFile string
	Cache     *AgentCache
	HasCache  bool
}

type AgentCache struct {
	Id string
}

func NewAgentConfig() (*AgentConfig, error) {
	config := &AgentConfig{}

	p, err := HomeDir()

	if err != nil {
		return config, err
	}

	Log.Debugf("Home path set to %s.", p)

	DefaultAgentPath = path.Join(p, "."+Name)
	DefaultCacheFile = path.Join(DefaultAgentPath, "agent-cache")

	Log.Debugf("Default Agent Config Path: %s.", DefaultAgentPath)
	Log.Debugf("Default Agent Cache File: %s.", DefaultCacheFile)

	if err := os.MkdirAll(DefaultAgentPath, 0777); err != nil {
		return config, err
	}

	config.Path = DefaultAgentPath
	config.CacheFile = DefaultCacheFile
	config.HasCache = false
	config.Cache = &AgentCache{}

	if IsExist(DefaultCacheFile) {
		if _, err := toml.DecodeFile(DefaultCacheFile, config.Cache); err != nil {
			return config, err
		}

		config.HasCache = true
	}

	return config, err
}

func (self AgentConfig) WriteCache() error {
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
