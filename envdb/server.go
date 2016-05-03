package envdb

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

// Holds node metadata. This struct is used by the server
// to find and send command to individual nodes.
type NodeData struct {
	Id                string       `json:"id"`
	EnvdbVersion      string       `json:"envdb-version"`
	Name              string       `json:"name"`
	Ip                string       `json:"ip"`
	Hostname          string       `json:"hostname"`
	Socket            *gotalk.Sock `json:"-"`
	OsQuery           bool
	OsQueryVersion    string
	OsQueryConfigPath string
	Online            bool   `json:"online"`
	PendingDelete     bool   `json:"-"`
	Os                string `json:"os"`
}

// Server holds the tcp server socket, connected nodes
// and configurations.
type Server struct {
	Config *ServerConfig
	Port   int
	Socket *gotalk.Server
	Nodes  map[*gotalk.Sock]*NodeData
	mu     sync.RWMutex
}

// Create a new server.
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

	if err := NodeUpdateOnlineStatus(); err != nil {
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
					os.Exit(0)
				}
			}
		}
	}()

	return server, nil
}

// Shutdow the server and tell all connected nodes to
// mark themselves as offline.
func (server *Server) Shutdown() {
	Log.Infof("%s shutting down.", Name)

	nodes, _ := AllNodes()

	for _, node := range nodes {
		node.Online = false

		if err := node.Update(); err != nil {
			Log.Error("unable to update node record")
			Log.Error("Error: ", err)
		}
	}
}

// When a new node connects add said node to the Server struct Nodes
// and process the node checkin.
func (server *Server) onAccept(s *gotalk.Sock) {
	server.mu.Lock()
	defer server.mu.Unlock()

	go func() {
		var resp Message

		err := s.Request("checkin", nil, &resp)

		if err != nil {
			Log.Fatalf("%s", err)
		}

		Log.Infof("New node connected. (%s / %s)", resp.Data["name"], resp.Data["id"])

		if _, ok := resp.Data["envdb-version"]; !ok {
			s.BufferNotify("die", []byte("This version of Envdb is out of date. Please upgrade."))
			return
		}

		if !VersionCheck(Version, resp.Data["envdb-version"].(string)) {
			s.BufferNotify("die", []byte("Envdb version mismatch"))
			return
		}

		node := &NodeData{
			Id:                resp.Data["id"].(string),
			Name:              resp.Data["name"].(string),
			EnvdbVersion:      resp.Data["envdb-version"].(string),
			Online:            true,
			Socket:            s,
			OsQuery:           resp.Data["osquery"].(bool),
			OsQueryVersion:    resp.Data["osquery-version"].(string),
			OsQueryConfigPath: resp.Data["osquery-config-path"].(string),
			Ip:                resp.Data["ip"].(string),
			Hostname:          resp.Data["hostname"].(string),
			PendingDelete:     false,
			Os:                resp.Data["os"].(string),
		}

		server.Nodes[s] = node

		if _, err := NodeUpdateOrCreate(node); err != nil {
			Log.Error("unable to create or update node record")
			Log.Error("Error: ", err)
		}

		WebSocketSend("node-update", node)

		s.CloseHandler = func(s *gotalk.Sock, _ int) {
			server.mu.Lock()
			defer server.mu.Unlock()

			node := server.Nodes[s]
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

			delete(server.Nodes, s)
		}
	}()

}

// Send data to all connected nodes.
func (server *Server) Broadcast(name string, in interface{}) {
	server.mu.RLock()
	defer server.mu.RUnlock()

	for s, _ := range server.Nodes {
		s.Notify(name, in)
	}
}

// Return true if the node is connected and working properly.
func (server *Server) Alive(id string) bool {
	node, err := server.GetNodeById(id)

	if err != nil {
		return false
	}

	var data []byte
	err = node.Socket.Request("ping", true, &data)

	if err != nil {
		return false
	}

	if string(data) != "pong" {
		return false
	}

	return true
}

// Disconnect a node from the server.
func (server *Server) Disconnect(id string) error {
	node, err := server.GetNodeById(id)

	if err != nil {
		return err
	}

	Log.Debugf("Disconnect node: %s (%s)", node.Name, node.Id)

	node.Socket.BufferNotify("die", []byte("good-bye!"))

	return nil
}

// DisconnectDead with disconnect a dead agent.
//
// This is useful when the server is killed for some reason without
// cleaning up the node connection state.
func (server *Server) DisconnectDead(id string) error {
	node, err := GetNodeByNodeId(id)

	if err != nil {
		return err
	}

	node.Online = false

	if err := node.Update(); err != nil {
		Log.Error("unable to update node record")
		Log.Error("Error: ", err)

		return err
	}

	return nil
}

// Disconnect and delete a node from the database.
//
// * NOTE: This is hacky because xorm has an issue with
// sqlite and table locking. Need to move this to a
// channel for node db writes and queue them.
func (server *Server) Delete(id string) error {

	Log.Debugf("Deleting node: %s", id)

	node, err := server.GetNodeById(id)

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

		err = server.Disconnect(id)

		if err != nil {
			return err
		}
	}

	return nil
}

// Convert the bytes returned to json and do some basic counting
// to make sure we never send more than the DefaultRowRow to the UI.
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

// Send a query request to all connected nodes and return
// a Response struct
func (server *Server) sendAll(in interface{}) *Response {
	server.mu.RLock()
	defer server.mu.RUnlock()

	var wg sync.WaitGroup

	resp := NewResponse()

	Log.Debug("Sending request to nodes.")
	Log.Debugf("Request: %s", in.(Query).Sql)

	var count int

	for s, node := range server.Nodes {
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
		resp.Error = fmt.Errorf("Results too large. (Limit: %d got %d)", DefaultRowLimit, resp.Total)
	}

	Log.Debug("Sending results back to requester.")
	return resp
}

// Send a query request to just one node and return a Response struct
func (server *Server) sendTo(id string, in interface{}) *Response {
	server.mu.RLock()
	defer server.mu.RUnlock()

	resp := NewResponse()

	node, err := server.GetNodeById(id)

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

// Send wraps the sendTo and sendAll functions and returns the Response
// struct from the requested send type.
func (server *Server) Send(id string, in interface{}) *Response {
	if id == "all" {
		return server.sendAll(in)
	}

	return server.sendTo(id, in)
}

// Ask for information from a node by id
func (server *Server) Ask(id, question string) (map[string]interface{}, error) {
	var data map[string]interface{}

	node, err := server.GetNodeById(id)

	if err != nil {
		return data, err
	}

	err = node.Socket.Request(question, nil, &data)

	return data, err
}

// Fetch a node by id from the Server.Nodes map
func (server *Server) GetNodeById(id string) (*NodeData, error) {
	for _, node := range server.Nodes {
		if node.Id == id {
			return node, nil
		}
	}

	return nil, errors.New("No node found for that id.")
}

// Start the tcp server and register all handlers.
func (server *Server) Run(webPort int) error {
	Log.Infof("Starting Server on port %d.", server.Port)

	server.Nodes = make(map[*gotalk.Sock]*NodeData)

	handlers := gotalk.NewHandlers()

	handlers.Handle("pong", func(in interface{}) (interface{}, error) {
		return in, nil
	})

	handlers.HandleBufferNotification("result", func(s *gotalk.Sock, name string, b []byte) {
		Log.Infof("Output %s: %s", name, string(b))
	})

	s, err := gotalk.Listen("tcp", fmt.Sprintf(":%d", server.Port), &tls.Config{
		Certificates: []tls.Certificate{server.Config.Cert},
		ClientAuth:   tls.NoClientCert,
	})

	if err != nil {
		return err
	}

	server.Socket = s
	server.Socket.HeartbeatInterval = 20 * time.Second

	server.Socket.OnHeartbeat = func(load int, t time.Time) {
		// Log.Debugf("Got heartbeat: Load (%d), Time: (%s)", load, t.Format(TimeFormat))
	}

	server.Socket.AcceptHandler = server.onAccept
	server.Socket.Handlers = handlers

	go NewWebServer(webPort, server)

	server.Socket.Accept()

	return nil
}
