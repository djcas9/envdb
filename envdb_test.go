package main

import (
	"fmt"
	"os"
	"path"
	"testing"
	"time"
)

var (
	testServer      *Server
	testServerError error

	testNode            Node
	testNodeConfig      *NodeConfig
	testNodeConfigError error
)

func TestMain(m *testing.M) {
	TestMode = true
	*quiet = true
	initLogger()

	testServer, testServerError = NewServer(3636)

	if testServerError != nil {
		fmt.Println(testServerError)
		os.Exit(-1)
	}

	go func() {
		if err := testServer.Run(8080); err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
	}()

	testNode = Node{
		Name:       "test",
		Host:       "localhost",
		Port:       3636,
		RetryCount: 50,
	}

	testNodeConfig, testNodeConfigError = NewNodeConfig()

	if testNodeConfigError != nil {
		fmt.Println(testNodeConfigError)
		os.Exit(-1)
	}

	testNode.Config = testNodeConfig

	go func() {
		if err := testNode.Run(); err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
	}()

	// wait for node to get setup.
	time.Sleep(time.Second * 1)

	code := m.Run()

	testServer.Shutdown()

	os.Exit(code)
}

func TestServer(t *testing.T) {

	p, err := HomeDir()

	if err != nil {
		t.Fatal("Unable to get home path.")
	}

	if testServer.Port != 3636 {
		t.Fatal("Server has wrong port set.")
	}

	pp := path.Join(p, "."+Name)
	ps := path.Join(DefaultServerPath, "store.db")
	pl := path.Join(DefaultServerPath, "logs")

	if testServer.Config.Path != pp {
		t.Fatal("Server path is wrong.")
	}

	if testServer.Config.StorePath != ps {
		t.Fatal("Server store path is wrong.")
	}

	if testServer.Config.LogPath != pl {
		t.Fatal("Server log path is wrong.")
	}
}

func TestNode(t *testing.T) {

	if testNode.Port != 3636 {
		t.Fatal("Node has wrong port set.")
	}

	if testNode.Name != "test" {
		t.Fatal("Node name is wrong.")
	}

	if testNode.Host != "localhost" {
		t.Fatal("Node host is wrong.")
	}

	if testNode.RetryCount != 50 {
		t.Fatal("Node retry count is wrong.")
	}

	p, err := HomeDir()

	if err != nil {
		t.Fatal("Unable to get home path.")
	}

	p1 := path.Join(p, "."+Name)
	p2 := path.Join(DefaultNodePath, "node.cache")

	if testNode.Config.Path != p1 {
		t.Fatal("Node path is wrong.")
	}

	if testNode.Config.CacheFile != p2 {
		t.Fatal("Node cache file path is wrong.")
	}

	if testNode.Server() != "localhost:3636" {
		t.Fatal("Node has the wrong server to connect to.")
	}
}
