package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/mephux/gotalk"
)

type NodeData struct {
	Id             string       `json:"id"`
	Name           string       `json:"name"`
	Ip             string       `json:"ip"`
	Hostname       string       `json:"hostname"`
	Socket         *gotalk.Sock `json:"-"`
	OsQuery        bool
	OsQueryVersion string
	Online         bool   `json:"online"`
	PendingDelete  bool   `json:"-"`
	Os             string `json:"os"`
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
		Log.Error("Unable to setup sqlite database.")
		Log.Fatalf("Error: %s", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	go func() {
		for {
			select {
			case sig := <-sigChan:
				if sig.String() == "interrupt" {
					Log.Info("Received Interrupt.")
					server.Shutdown()
				}
			}
		}
	}()

	return server, nil
}

func (self *Server) Shutdown() {
	Log.Infof("%s shutting down.", Name)

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
			PendingDelete:  false,
			Os:             resp.Data["os"].(string),
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

			dbNode, err := NodeUpdateOrCreate(node)

			if err != nil {
				Log.Error("unable to update node record")
				Log.Error("Error: ", err)
			}

			if dbNode.PendingDelete {
				dbNode.Delete()
			} else {
				WebSocketSend("node-update", node)
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

func (self *Server) Disconnect(id string) error {
	node, err := self.GetNodeById(id)

	if err != nil {
		return err
	}

	Log.Debugf("Disconnect node: %s (%s)", node.Name, node.Id)

	node.Socket.BufferNotify("die", []byte("good-bye!"))

	return nil
}

// this is hacky because xorm has an issue with
// sqlite and table locking. Need to move this to a
// channel for node db writes and queue them.
func (self *Server) Delete(id string) error {

	Log.Debugf("Deleting node: %s", id)

	node, err := self.GetNodeById(id)

	if err != nil {

		n, dberr := GetNodeByNodeId(id)

		if dberr != nil {
			return err
		}

		return n.Delete()
	} else {
		node.PendingDelete = true

		if _, err := NodeUpdateOrCreate(node); err != nil {
			return err
		}

		err = self.Disconnect(id)

		if err != nil {
			return err
		}
	}

	return nil
}

func ProcessResults(data []byte) (bool, []map[string]interface{}, []byte) {
	var all = []map[string]interface{}{}
	var returnData = []map[string]interface{}{}

	if err := json.Unmarshal(data, &all); err == nil {
		if len(all) > DefaultRowLimit {

			Log.Debug("Results too large. sending first 2000.")

			for i, d := range all {

				if i >= DefaultRowLimit {
					break
				}

				returnData = append(returnData, d)
			}

			if data, err := json.Marshal(returnData); err == nil {
				return true, all, data
			}

		} else {
			return false, all, data
		}
	} else {
		return false, all, data
	}

	return false, all, data
}

func (self *Server) sendAll(in interface{}) *Response {
	self.mu.RLock()
	defer self.mu.RUnlock()

	var wg sync.WaitGroup

	resp := NewResponse()

	Log.Debug("Sending request to nodes.")
	Log.Debugf("Request: %s", in.(Query).Sql)

	var count int

	for s, node := range self.Nodes {
		wg.Add(1)

		go func(s *gotalk.Sock, node *NodeData) {
			defer wg.Done()

			start := time.Now()

			var data []byte
			err := s.Request("query", in, &data)

			_, all, _ := ProcessResults(data)

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

			count += len(all)

			Log.Debugf("  * %s (%s)", node.Name, elapsed)
			resp.Results = append(resp.Results, qr)
		}(s, node)
	}

	Log.Debug("Waiting for all requests to return.")

	wg.Wait()

	resp.Total = count

	if resp.Total > DefaultRowLimit {
		resp.Error = errors.New(fmt.Sprintf("Results too large. (Limit: %d got %d)", DefaultRowLimit, resp.Total))
	}

	Log.Debug("Sending results back to requester.")
	return resp
}

func (self *Server) sendTo(id string, in interface{}) *Response {
	self.mu.RLock()
	defer self.mu.RUnlock()

	resp := NewResponse()

	node, err := self.GetNodeById(id)

	if err != nil {
		resp.Total = 0

		return resp
	}

	start := time.Now()

	var data []byte
	err = node.Socket.Request("query", in, &data)

	var newData []byte
	over, all, cut := ProcessResults(data)

	if over {
		newData = cut
	} else {
		newData = data
	}

	resp.Total = len(all)

	qr := QueryResults{
		Id:       node.Id,
		Name:     node.Name,
		Hostname: node.Hostname,
		Results:  string(newData),
	}

	if err != nil {
		qr.Error = err.Error()
	}

	resp.Results = append(resp.Results, qr)

	elapsed := time.Since(start)

	Log.Debugf("\n  - Node: %s\n  - Request: %s\n  - Response Id: %s\n  - Total: %d\n  - Elapsed Time: %s", node.Name, in.(Query).Sql, resp.Id, resp.Total, elapsed)
	return resp
}

func (self *Server) Send(id string, in interface{}) *Response {

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

	return nil, errors.New("No node found for that id.")
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

	s, err := gotalk.Listen("tcp", fmt.Sprintf(":%d", self.Port), &tls.Config{
		Certificates: []tls.Certificate{self.Config.Cert},
		ClientAuth:   tls.NoClientCert,
	})

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

	go NewWebServer(webPort, self)

	self.Socket.Accept()

	return nil
}
