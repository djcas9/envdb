package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/elazarl/go-bindata-assetfs"
	"github.com/rsms/gotalk"
)

var (
	Clients = make(map[*gotalk.Sock]int)
	socksmu sync.RWMutex
)

type SqlRequest struct {
	Id  string
	Sql string
}

func onAccept(s *gotalk.Sock) {
	// Keep track of connected sockets
	socksmu.Lock()
	defer socksmu.Unlock()
	Clients[s] = 1

	s.CloseHandler = func(s *gotalk.Sock, _ int) {
		socksmu.Lock()
		defer socksmu.Unlock()
		delete(Clients, s)
	}
}

func WebSocketSend(name string, in interface{}) {
	socksmu.RLock()
	defer socksmu.RUnlock()

	Log.Debug("WebSocketSend: ", name, in)

	for s, _ := range Clients {
		s.Notify(name, in)
	}
}

// func readCsv(data string) ([]map[string]string, error) {
// var dd []map[string]string

// reader := csv.NewReader(strings.NewReader(data))
// reader.FieldsPerRecord = -1

// rawCSVdata, err := reader.ReadAll()

// if err != nil {
// return dd, err
// }

// header := rawCSVdata[0]

// for i, each := range rawCSVdata {
// if i == 0 {
// continue
// }

// data := make(map[string]string)

// for ii, e := range each {
// data[header[ii]] = e
// }

// fmt.Println(i, each)
// dd = append(dd, data)
// }

// return dd, nil
// }

func NewWebServer(webPort int, server *Server) {

	gotalk.Handle("tables", func(sql SqlRequest) (QueryResults, error) {
		var tables []string

		data := server.Send(sql.Id, Query{
			Sql:    ".tables",
			Format: "json",
		})

		if len(data) != 1 {
			d := QueryResults{}
			return d, errors.New(fmt.Sprintf("Agent not found for id (%s).", sql.Id))
		}

		tables = strings.Split(data[0].Results.(string), "\n")

		for i, t := range tables {
			newT := strings.Replace(t, "  => ", "", -1)
			tables[i] = newT
		}

		tables = tables[:len(tables)-1]
		data[0].Results = tables

		return data[0], nil
	})

	gotalk.Handle("table-info", func(sql SqlRequest) (QueryResults, error) {
		data := server.Send(sql.Id, Query{
			Sql:    sql.Sql,
			Format: "json",
		})

		if len(data) != 1 {
			d := QueryResults{}
			return d, errors.New(fmt.Sprintf("Agent not found for id (%s).", sql.Id))
		}

		return data[0], nil
	})

	gotalk.Handle("query", func(sql SqlRequest) ([]QueryResults, error) {
		data := server.Send(sql.Id, Query{
			Sql:    sql.Sql,
			Format: "json",
		})

		return data, nil
	})

	ws := gotalk.WebSocketHandler()
	ws.OnAccept = onAccept

	if DEV_MODE {
		// dev
		http.Handle("/public/", http.FileServer(http.Dir("./web/")))
	} else {
		http.Handle("/public/",
			http.FileServer(
				&assetfs.AssetFS{
					Asset:    Asset,
					AssetDir: AssetDir,
					Prefix:   "web",
				},
			),
		)
	}

	http.HandleFunc("/", RouteIndex)

	// API - Version 1
	http.HandleFunc("/api/v1/agents", RouteAgents)

	http.Handle("/gotalk/", ws)

	var port string = fmt.Sprintf(":%d", webPort)

	Log.Infof("HTTP Server: http://%s%s", "127.0.0.1", port)

	err := http.ListenAndServe(port, nil)

	if err != nil {
		Log.Error(err.Error())
	}
}
