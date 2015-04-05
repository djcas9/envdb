package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/rsms/gotalk"
)

type NodeData struct {
	Id             string       `json:"id"`
	Name           string       `json:"name"`
	Ip             string       `json:"ip"`
	Hostname       string       `json:"hostname"`
	Socket         *gotalk.Sock `json:"-"`
	OsQuery        bool
	OsQueryVersion string
	Online         bool `json:"online"`
}

type Server struct {
	Config *ServerConfig
	Port   int
	Socket *gotalk.Server
	Nodes  map[*gotalk.Sock]*NodeData
	mu     sync.RWMutex
}

func NewServer(port int) (*Server, error) {
	server := &Server{
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

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	go func() {
		for {
			select {
			case sig := <-sigChan:
				if sig.String() == "interrupt" {
					Log.Info("Received Interrupt. Exiting.")
					nodes, _ := AllNodes()

					for _, node := range nodes {
						node.Online = false

						if err := node.Update(); err != nil {
							Log.Error("unable to update node record")
							Log.Error("Error: ", err)
						}
					}

					os.Exit(1)
				}
			}
		}
	}()

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

		Log.Infof("New node connected. (%s / %s)", resp.Data["name"], resp.Data["id"])

		node := &NodeData{
			Id:             resp.Data["id"].(string),
			Name:           resp.Data["name"].(string),
			Online:         true,
			Socket:         s,
			OsQuery:        resp.Data["osquery"].(bool),
			OsQueryVersion: resp.Data["osquery-version"].(string),
			Ip:             resp.Data["ip"].(string),
			Hostname:       resp.Data["hostname"].(string),
		}

		self.Nodes[s] = node

		if _, err := NodeUpdateOrCreate(node); err != nil {
			Log.Error("unable to create or update node record")
			Log.Error("Error: ", err)
		}

		WebSocketSend("node-update", node)

		s.CloseHandler = func(s *gotalk.Sock, _ int) {
			self.mu.Lock()
			defer self.mu.Unlock()

			node := self.Nodes[s]
			node.Online = false

			Log.Infof("Node disconnected. (%s / %s)", node.Name, node.Id)
			WebSocketSend("node-update", node)

			if _, err := NodeUpdateOrCreate(node); err != nil {
				Log.Error("unable update node record")
				Log.Error("Error: ", err)
			}

			delete(self.Nodes, s)
		}
	}()

}

func (self *Server) Broadcast(name string, in interface{}) {
	self.mu.RLock()
	defer self.mu.RUnlock()

	for s, _ := range self.Nodes {
		s.Notify(name, in)
	}
}

func (self *Server) sendAll(in interface{}) []QueryResults {
	// go func() {
	self.mu.RLock()
	defer self.mu.RUnlock()

	var wg sync.WaitGroup

	var results []QueryResults

	Log.Debug("Sending request to nodes.")

	Log.Debugf("Request: %s", in.(Query).Sql)

	for s, node := range self.Nodes {
		wg.Add(1)

		go func(s *gotalk.Sock, node *NodeData) {
			defer wg.Done()

			start := time.Now()

			var data []byte
			err := s.Request("query", in, &data)

			qr := QueryResults{
				Id:       node.Id,
				Name:     node.Name,
				Hostname: node.Hostname,
				Results:  string(data),
			}

			if err != nil {
				qr.Error = err.Error()
			}

			elapsed := time.Since(start)
			Log.Debugf(" * %s (%s)", node.Name, elapsed)

			results = append(results, qr)
		}(s, node)
	}

	Log.Debug("Waiting for all requests to return.")

	wg.Wait()

	Log.Debug("Sending results back to requester.")

	return results
	// }()
}

func (self *Server) sendTo(id string, in interface{}) []QueryResults {
	self.mu.RLock()
	defer self.mu.RUnlock()

	var results []QueryResults

	node, err := self.GetNodeById(id)

	if err != nil {
		return results
	}

	var data []byte
	err = node.Socket.Request("query", in, &data)

	qr := QueryResults{
		Id:       node.Id,
		Name:     node.Name,
		Hostname: node.Hostname,
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

func (self *Server) GetNodeById(id string) (*NodeData, error) {

	for _, node := range self.Nodes {
		if node.Id == id {
			return node, nil
		}
	}

	return nil, errors.New("no node found for that id.")
}

func (self *Server) Run(webPort int) error {
	Log.Infof("Starting Server on port %d.", self.Port)

	self.Nodes = make(map[*gotalk.Sock]*NodeData)

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
		// Log.Debugf("Got heartbeat: Load (%d), Time: (%s)", load, t.Format(TimeFormat))
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
