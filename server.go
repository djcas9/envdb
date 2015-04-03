package main

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/rsms/gotalk"
)

type AgentData struct {
	Id             string
	Name           string
	Ip             string
	Hostname       string
	Socket         *gotalk.Sock
	OsQuery        bool
	OsQueryVersion string
	Online         bool
}

type Server struct {
	Config *ServerConfig
	Port   int
	Socket *gotalk.Server
	Agents map[*gotalk.Sock]*AgentData
	mu     sync.RWMutex
}

func NewServer(port int) (Server, error) {
	server := Server{
		Port: port,
	}

	Log.Debug("Building server configurations.")
	config, err := NewServerConfig()

	if err != nil {
		return server, err
	}

	server.Config = config

	Log.Debugf("Attempting to open server store: %s.", config.StorePath)

	if err := DBInit(config.StorePath, config.LogPath); err != nil {
		Log.Fatal("Unable to setup sqlite database.")
		Log.Fatalf("Error: %s", err)
	}

	return server, nil
}

func (self *Server) onAccept(s *gotalk.Sock) {
	self.mu.Lock()
	defer self.mu.Unlock()

	go func() {
		var resp Message

		err := s.Request("checkin", Message{}, &resp)

		if err != nil {
			Log.Fatalf("ERROR: %s", err)
		}

		Log.Infof("New agent connected. (%s / %s)", resp.Data["name"], resp.Data["id"])

		agent := &AgentData{
			Id:             resp.Data["id"].(string),
			Name:           resp.Data["name"].(string),
			Online:         true,
			Socket:         s,
			OsQuery:        resp.Data["osquery"].(bool),
			OsQueryVersion: resp.Data["osquery-version"].(string),
			Ip:             resp.Data["ip"].(string),
			Hostname:       resp.Data["hostname"].(string),
		}

		self.Agents[s] = agent

		if _, err := AgentUpdateOrCreate(agent); err != nil {
			Log.Error("unable to create or update agent record")
			Log.Error("Error: ", err)
		}

		s.CloseHandler = func(s *gotalk.Sock, _ int) {
			self.mu.Lock()
			defer self.mu.Unlock()

			agent := self.Agents[s]
			agent.Online = false

			Log.Infof("Agent disconnected. (%s / %s)", agent.Name, agent.Id)

			if _, err := AgentUpdateOrCreate(agent); err != nil {
				Log.Error("unable update agent record")
				Log.Error("Error: ", err)
			}

			delete(self.Agents, s)
		}
	}()

}

func (self *Server) Broadcast(name string, in interface{}) {
	self.mu.RLock()
	defer self.mu.RUnlock()

	for s, _ := range self.Agents {
		s.Notify(name, in)
	}
}

func (self *Server) sendAll(in interface{}) []QueryResults {
	// go func() {
	self.mu.RLock()
	defer self.mu.RUnlock()

	var results []QueryResults

	for s, agent := range self.Agents {

		var data []byte
		err := s.Request("query", in, &data)

		qr := QueryResults{
			Id:       agent.Id,
			Name:     agent.Name,
			Hostname: agent.Hostname,
			Results:  string(data),
		}

		if err != nil {
			qr.Error = err.Error()
		}

		results = append(results, qr)
	}

	return results
	// }()
}

func (self *Server) sendTo(id string, in interface{}) []QueryResults {
	self.mu.RLock()
	defer self.mu.RUnlock()

	var results []QueryResults

	agent, err := self.GetAgentById(id)

	if err != nil {
		return results
	}

	var data []byte
	err = agent.Socket.Request("query", in, &data)

	qr := QueryResults{
		Id:       agent.Id,
		Name:     agent.Name,
		Hostname: agent.Hostname,
		Results:  string(data),
	}

	if err != nil {
		qr.Error = err.Error()
	}

	results = append(results, qr)

	return results
}

func (self *Server) Send(id string, in interface{}) []QueryResults {

	if id == "all" {
		return self.sendAll(in)
	}

	return self.sendTo(id, in)
}

func (self *Server) GetAgentById(id string) (*AgentData, error) {

	for _, agent := range self.Agents {
		if agent.Id == id {
			return agent, nil
		}
	}

	return nil, errors.New("no agent found for that id.")
}

func (self *Server) Run(webPort int) error {
	Log.Infof("Starting Server on port %d.", self.Port)

	self.Agents = make(map[*gotalk.Sock]*AgentData)

	handlers := gotalk.NewHandlers()

	handlers.Handle("pong", func(in interface{}) (interface{}, error) {
		return in, nil
	})

	handlers.HandleBufferNotification("result", func(s *gotalk.Sock, name string, b []byte) {
		Log.Infof("Output %s: %s", name, string(b))
	})

	s, err := gotalk.Listen("tcp", fmt.Sprintf(":%d", self.Port))

	if err != nil {
		return err
	}

	self.Socket = s
	self.Socket.HeartbeatInterval = 20 * time.Second

	self.Socket.OnHeartbeat = func(load int, t time.Time) {
		Log.Debugf("Got heartbeat: Load (%d), Time: (%s)", load, t.Format(TimeFormat))
	}

	self.Socket.AcceptHandler = self.onAccept
	self.Socket.Handlers = handlers

	// go func() {
	// for {
	// time.Sleep(10 * time.Second)

	// self.Send(Query{
	// Sql:    "SELECT uid, name, port FROM listening_ports l, processes p WHERE l.pid=p.pid;",
	// Format: "json",
	// })
	// }
	// }()

	go NewWebServer(webPort, self)

	self.Socket.Accept()

	return nil
}
