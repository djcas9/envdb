package main

import (
	"fmt"
	"os"
	"time"

	"github.com/nu7hatch/gouuid"
	"github.com/rsms/gotalk"
)

var (
	KillClient = false
	Connection = make(chan bool, 1)
	RetryCount = 0
)

type Agent struct {
	Id         string
	Config     *AgentConfig
	Name       string
	Host       string
	Port       int
	Socket     *gotalk.Sock
	RetryCount int
}

type Message struct {
	Error error
	Data  map[string]interface{}
}

func (self *Agent) Handlers() {
	handlers := gotalk.NewHandlers()

	handlers.HandleBufferNotification("die", func(s *gotalk.Sock, name string, b []byte) {
		KillClient = true
		self.Socket.Close()
		Connection <- true
	})

	handlers.HandleBufferNotification("ping", func(_ *gotalk.Sock, _ string, b []byte) {
		fmt.Printf("client: handling 'ping' request: %q\n", string(b))
	})

	handlers.Handle("query", func(query Query) ([]byte, error) {
		return query.Run()
	})

	handlers.Handle("tables", func(query Query) ([]byte, error) {
		return query.Run()
	})

	handlers.Handle("checkin", func(_ Message) (Message, error) {
		var err error

		if self.Config.HasCache {
			Log.Infof("Connection successful. Id: %s", self.Config.Cache.Id)
			self.Id = self.Config.Cache.Id
		} else {

			id, uuerr := uuid.NewV4()
			err = uuerr

			if err != nil {
				Log.Fatalf("Error creating id: ", err)
			}

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
			Log.Errorf("oswrap requires osqueryi version %s or later.", MinOsQueryVersion)
			has = false
		}

		var hostname string = "n/a"

		if os, err := os.Hostname(); err == nil {
			hostname = os
		}

		rmsg := Message{
			Error: err,
			Data: map[string]interface{}{
				"name":            self.Name,
				"id":              self.Id,
				"osquery":         has,
				"osquery-version": version,
				"ip":              self.Socket.Addr(),
				"hostname":        hostname,
			},
		}

		return rmsg, nil
	})

	self.Socket.Handlers = handlers
}

func (self *Agent) Server() string {
	return fmt.Sprintf("%s:%d", self.Host, self.Port)
}

func (self *Agent) Connect() error {
	Log.Infof("Connecting to %s", self.Server())

	s, err := gotalk.Connect("tcp", self.Server())

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

func (self *Agent) Reconnect() {
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

func (self *Agent) Run() error {

	if err := self.Connect(); err != nil {
		return err
	}

	self.Handlers()

	<-Connection

	return nil
}
