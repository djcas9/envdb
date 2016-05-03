package envdb

import (
	"bytes"
	"os"
	"path"

	"github.com/BurntSushi/toml"
)

var (
	// DefaultNodePath node path
	DefaultNodePath = ""

	// DefaultCacheFile stores the node connection id.
	DefaultCacheFile = ""
)

// NodeConfig holds all node configuration values
type NodeConfig struct {
	Path      string
	CacheFile string
	Cache     *NodeCache
	HasCache  bool
}

// NodeCache holds the node connection id.
// This could also hold a lot of other information in
// the future.
type NodeCache struct {
	Id string
}

// NewNodeConfig new node configuration.
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

// WriteCache Write node cache to disk.
func (config NodeConfig) WriteCache() error {
	var cache bytes.Buffer

	e := toml.NewEncoder(&cache)

	if err := e.Encode(config.Cache); err != nil {
		return err
	}

	f, err := os.Create(config.CacheFile)
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
