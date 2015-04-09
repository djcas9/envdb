package main

import (
	"path"
	"testing"
)

func TestNewServer(t *testing.T) {
	initLogger()

	var svrPort int = 3636
	// var svrWebPort int = 8080

	svr, err := NewServer(svrPort)

	if err != nil {
		t.Fatal(err)
	}

	p, err := HomeDir()

	if err != nil {
		t.Fatal("Unable to get home path.")
	}

	if svr.Port != svrPort {
		t.Fatal("Server has wrong port set.")
	}

	pp := path.Join(p, "."+Name)
	ps := path.Join(DefaultServerPath, "store.db")
	pl := path.Join(DefaultServerPath, "logs")

	if svr.Config.Path != pp {
		t.Fatal("Server path is wrong.")
	}

	if svr.Config.StorePath != ps {
		t.Fatal("Server store path is wrong.")
	}

	if svr.Config.LogPath != pl {
		t.Fatal("Server log path is wrong.")
	}
}

func TestNewNode(t *testing.T) {
	initLogger()

	var c = Node{
		Name:       "test",
		Host:       "test-server",
		Port:       3636,
		RetryCount: 50,
	}

	config, err := NewNodeConfig()

	if err != nil {
		t.Fatal(err)
	}

	c.Config = config

	if c.Port != 3636 {
		t.Fatal("Node has wrong port set.")
	}

	if c.Name != "test" {
		t.Fatal("Node name is wrong.")
	}

	if c.Host != "test-server" {
		t.Fatal("Node host is wrong.")
	}

	if c.RetryCount != 50 {
		t.Fatal("Node retry count is wrong.")
	}

	p, err := HomeDir()

	if err != nil {
		t.Fatal("Unable to get home path.")
	}

	p1 := path.Join(p, "."+Name)
	p2 := path.Join(DefaultNodePath, "node.cache")

	if c.Config.Path != p1 {
		t.Fatal("Node path is wrong.")
	}

	if c.Config.CacheFile != p2 {
		t.Fatal("Node cache file path is wrong.")
	}

	// if !c.Config.HasCache {
	// t.Fatal("Node couldn't create a cache file.")
	// }
}
