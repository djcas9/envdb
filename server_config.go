package main

import (
	"os"
	"path"
)

var (
	DefaultServerPath = ""
	DefaultStorePath  = ""
	DefaultLogPath    = ""
)

type ServerConfig struct {
	StorePath string
	Path      string
	LogPath   string
}

func NewServerConfig() (*ServerConfig, error) {
	config := &ServerConfig{}

	p, err := HomeDir()

	if err != nil {
		return config, err
	}

	Log.Debugf("Home path set to %s.", p)

	DefaultServerPath = path.Join(p, "."+Name)
	DefaultStorePath = path.Join(DefaultServerPath, "store.db")
	DefaultLogPath = path.Join(DefaultServerPath, "logs")

	Log.Debugf("Default Server Config Path: %s.", DefaultServerPath)
	Log.Debugf("Default Server Store Path: %s.", DefaultStorePath)
	Log.Debugf("Default Server Log Path: %s.", DefaultLogPath)

	if err := os.MkdirAll(DefaultServerPath, 0777); err != nil {
		return config, err
	}

	if err := os.MkdirAll(DefaultLogPath, 0777); err != nil {
		return config, err
	}

	config.Path = DefaultServerPath
	config.StorePath = DefaultStorePath
	config.LogPath = DefaultLogPath

	return config, err
}
