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
	KillClient = false
	Connection = make(chan bool, 1)
	RetryCount = 0
)

// The node struct holds the socket, configurations
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

// Register all node hnadlers.
func (self *Node) Handlers() {
	handlers := gotalk.NewHandlers()

	handlers.HandleBufferNotification("die", func(s *gotalk.Sock, name string, b []byte) {
		KillClient = true
		self.Socket.Close()
		Connection <- true
	})

	handlers.Handle("ping", func(_ bool) ([]byte, error) {
		return []byte("pong"), nil
	})

	handlers.Handle("query", func(query Query) ([]byte, error) {
		return query.Run()
	})

	handlers.Handle("tables", func(query Query) ([]byte, error) {
		return query.Run()
	})

	handlers.Handle("checkin", func() (Message, error) {
		var err error

		if self.Config.HasCache {
			Log.Debugf("Node has cache file. Using cached id %s", self.Config.Cache.Id)

			Log.Infof("Connection successful. Id: %s", self.Config.Cache.Id)
			self.Id = self.Config.Cache.Id
		} else {

			id, uuerr := uuid.NewV4()
			err = uuerr

			if err != nil {
				Log.Fatalf("Error creating id: %s", err)
			}

			Log.Debugf("No cache file found. Creating cache file and new id %s", id)

			Log.Infof("Connection successful. Id: %s", id.String())
			self.Config.Cache.Id = id.String()
			self.Id = self.Config.Cache.Id

			self.Config.WriteCache()
		}

		has, version := OsQueryInfo()

		Log.Infof("osquery enabled: %t", has)

		if has {
			Log.Infof("osquery version: %s", version)
		}

		if !CheckOsQueryVersion(version) {
			Log.Errorf("%s requires osqueryi version %s or later.", Name, MinOsQueryVersion)
			has = false
		}

		var hostname string = "n/a"
		var ip string = self.Socket.Addr()

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
				"name":            self.Name,
				"id":              self.Id,
				"osquery":         has,
				"osquery-version": version,
				"ip":              ip,
				"hostname":        hostname,
				"os":              os,
			},
		}

		return rmsg, nil
	})

	self.Socket.Handlers = handlers
}

// Return the server connection string.
func (self *Node) Server() string {
	return fmt.Sprintf("%s:%d", self.Host, self.Port)
}

// Connect a node to the server.
func (self *Node) Connect() error {
	Log.Infof("Connecting to %s", self.Server())

	s, err := gotalk.Connect("tcp", self.Server(), &tls.Config{
		InsecureSkipVerify: true,
	})

	if err != nil {
		return err
	}

	self.Socket = s

	self.Socket.HeartbeatInterval = 20 * time.Second

	self.Socket.OnHeartbeat = func(load int, t time.Time) {
		Log.Debugf("Got heartbeat: Load (%d), Time: (%s)", load, t.Format(TimeFormat))
	}

	self.Socket.CloseHandler = func(s *gotalk.Sock, code int) {
		if KillClient {
			KillClient = false
			Connection <- true
		} else {
			Log.Warnf("Lost connection to server. (Error Code: %d)", code)

			RetryCount = self.RetryCount
			self.Reconnect()
		}
	}

	return nil
}

// Reconnect to the server if connection is lost.
func (self *Node) Reconnect() {
	self.Socket.Close()

	Log.Warnf("Attempting to reconnect. (Retry Count: %d)", RetryCount)

	if RetryCount <= 0 {
		Log.Info("Connection retry count exceeded. Exiting.")
		Connection <- true
	}

	time.Sleep(5 * time.Second)

	if err := self.Run(); err != nil {
		RetryCount -= 1
		Log.Error(err)
		self.Reconnect()
		return
	}

	RetryCount = self.RetryCount
	Log.Info("Reconnect successful.")
}

// Start the node, connect and register all handlers.
func (self *Node) Run() error {

	if err := self.Connect(); err != nil {
		return err
	}

	self.Handlers()

	<-Connection

	return nil
}
