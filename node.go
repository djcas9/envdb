package main

import (
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"runtime"
	"time"

	"github.com/mephux/gotalk"
	"github.com/nu7hatch/gouuid"
)

var (
	// KillClient tells the node if it should disconnect or not
	KillClient = false

	// Connection is a channel to control
	// the nodes connection state
	Connection = make(chan bool, 1)

	// RetryCount holds the current number of
	// connection retry attempts
	RetryCount = 0
)

// Node struct holds the socket, configurations
// and other metadata.
type Node struct {
	Id         string
	Config     *NodeConfig
	Name       string
	Host       string
	Port       int
	Socket     *gotalk.Sock
	RetryCount int
}

// Message struct is used to pass data
// during the node checkin process.
type Message struct {
	Error error
	Data  map[string]interface{}
}

// Handlers will register all node hnadlers.
func (node *Node) Handlers() {
	handlers := gotalk.NewHandlers()

	handlers.HandleBufferNotification("die", func(s *gotalk.Sock, name string, b []byte) {
		KillClient = true
		node.Socket.Close()
		Log.Warn(string(b))
		Connection <- true
	})

	handlers.Handle("ping", func(_ bool) ([]byte, error) {
		return []byte("pong"), nil
	})

	handlers.Handle("system-information", func() (map[string]interface{}, error) {
		return SystemInformation()
	})

	handlers.Handle("query", func(query Query) ([]byte, error) {
		return query.Run()
	})

	handlers.Handle("tables", func(query Query) ([]byte, error) {
		return query.Run()
	})

	handlers.Handle("checkin", func() (Message, error) {
		var err error

		if node.Config.HasCache {
			Log.Debugf("Node has cache file. Using cached id %s", node.Config.Cache.Id)

			Log.Infof("Connection successful. Id: %s", node.Config.Cache.Id)
			node.Id = node.Config.Cache.Id
		} else {

			id, uuerr := uuid.NewV4()
			err = uuerr

			if err != nil {
				Log.Fatalf("Error creating id: %s", err)
			}

			Log.Debugf("No cache file found. Creating cache file and new id %s", id)

			Log.Infof("Connection successful. Id: %s", id.String())
			node.Config.Cache.Id = id.String()
			node.Id = node.Config.Cache.Id

			node.Config.WriteCache()
		}

		info := OsQueryInfo()

		Log.Infof("osquery enabled: %t", info.Enabled)

		if info.Enabled {
			Log.Infof("osquery version: %s", info.Version)
		}

		if !VersionCheck(MinOsQueryVersion, info.Version) {
			Log.Errorf("%s requires osqueryi version %s or later.", Name, MinOsQueryVersion)
			info.Enabled = false
		}

		var hostname = "n/a"
		var ip = node.Socket.Addr()

		if os, err := os.Hostname(); err == nil {
			hostname = os
		}

		addrs, _ := net.LookupIP(hostname)

		for _, addr := range addrs {
			if ipv4 := addr.To4(); ipv4 != nil {
				ip = ipv4.String()
			}
		}

		os := runtime.GOOS

		rmsg := Message{
			Error: err,
			Data: map[string]interface{}{
				"envdb-version":       Version,
				"name":                node.Name,
				"id":                  node.Id,
				"osquery":             info.Enabled,
				"osquery-version":     info.Version,
				"osquery-config-path": info.ConfigPath,
				"ip":       ip,
				"hostname": hostname,
				"os":       os,
			},
		}

		return rmsg, nil
	})

	node.Socket.Handlers = handlers
}

// Server will return the server connection string.
func (node *Node) Server() string {
	return fmt.Sprintf("%s:%d", node.Host, node.Port)
}

// Connect a node to the server.
func (node *Node) Connect() error {
	Log.Infof("Connecting to %s", node.Server())

	s, err := gotalk.Connect("tcp", node.Server(), &tls.Config{
		InsecureSkipVerify: true,
	})

	if err != nil {
		return err
	}

	node.Socket = s

	node.Socket.HeartbeatInterval = 20 * time.Second

	node.Socket.OnHeartbeat = func(load int, t time.Time) {
		Log.Debugf("Got heartbeat: Load (%d), Time: (%s)", load, t.Format(TimeFormat))
	}

	node.Socket.CloseHandler = func(s *gotalk.Sock, code int) {
		if KillClient {
			KillClient = false
			Connection <- true
		} else {
			Log.Warnf("Lost connection to server. (Error Code: %d)", code)

			RetryCount = node.RetryCount
			node.Reconnect()
		}
	}

	return nil
}

// Reconnect to the server if connection is lost.
func (node *Node) Reconnect() {
	node.Socket.Close()

	Log.Warnf("Attempting to reconnect. (Retry Count: %d)", RetryCount)

	if RetryCount <= 0 {
		Log.Info("Connection retry count exceeded. Exiting.")
		Connection <- true
	}

	time.Sleep(5 * time.Second)

	if err := node.Run(); err != nil {
		RetryCount--
		Log.Error(err)
		node.Reconnect()
		return
	}

	RetryCount = node.RetryCount
	Log.Info("Reconnect successful.")
}

// Run node, connect and register all handlers.
func (node *Node) Run() error {

	if err := node.Connect(); err != nil {
		return err
	}

	node.Handlers()

	<-Connection

	return nil
}
