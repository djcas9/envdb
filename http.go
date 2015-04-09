package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
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

	// Log.Debug("WebSocketSend: ", name, in)

	for s, _ := range Clients {
		s.Notify(name, in)
	}
}

// handlerFunc adapts a function to an http.Handler.
type handlerFunc func(http.ResponseWriter, *http.Request) error

var store = sessions.NewCookieStore([]byte("something-very-secret"))

func NewWebServer(webPort int, server *Server) {
	r := mux.NewRouter()
	r.StrictSlash(true)

	gotalk.Handle("tables", func(sql SqlRequest) (QueryResults, error) {
		var tables []string

		data := server.Send(sql.Id, Query{
			Sql:    ".tables",
			Format: "json",
		})

		if len(data) != 1 {
			d := QueryResults{}

			return d, errors.New(fmt.Sprintf("Node not found for id (%s).", sql.Id))
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

			return d, errors.New(fmt.Sprintf("Node not found for id (%s).", sql.Id))
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

	gotalk.Handle("disconnect", func(id string) error {
		return server.Disconnect(id)
	})

	gotalk.Handle("delete", func(id string) error {
		return server.Delete(id)
	})

	ws := gotalk.WebSocketHandler()
	ws.OnAccept = onAccept

	if DEV_MODE {
		// dev
		r.PathPrefix("/public/").Handler(http.FileServer(http.Dir("./web/")))
	} else {
		r.PathPrefix("/public/").Handler(http.FileServer(
			&assetfs.AssetFS{
				Asset:    Asset,
				AssetDir: AssetDir,
				Prefix:   "web",
			},
		))
	}

	r.Handle("/", handlerFunc(RouteIndex))
	r.Handle("/query/save", handlerFunc(RouteSaveQuery))
	r.Handle("/query/delete", handlerFunc(RouteDeleteQuery))
	r.Handle("/api/v1/nodes", handlerFunc(RouteNodes))
	r.Handle("/api/v1/queries", handlerFunc(RouteSavedQueries))

	http.Handle("/gotalk/", ws)
	http.Handle("/", r)

	var port string = fmt.Sprintf(":%d", webPort)
	var err error
	var proto string = "http"

	if SSL {
		proto = "https"
	}

	Log.Infof("HTTP Server: %s://%s%s", proto, "127.0.0.1", port)

	if SSL {
		err = http.ListenAndServeTLS(port, DefaultPublicKeyPath, DefaultPrivateKeyPath, context.ClearHandler(http.DefaultServeMux))
	} else {
		err = http.ListenAndServe(port, context.ClearHandler(http.DefaultServeMux))
	}

	if err != nil {
		Log.Error(err.Error())
	}
}

func (f handlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// session, _ := store.Get(r, "envdb")

	// if session.Values["current_user"] == nil {
	// http.Redirect(w, r, "/login", 301)
	// return
	// }

	// if err := session.Save(r, w); err != nil {
	// Log.Fatal(err)
	// }

	err := f(w, r)

	if err != nil {
		Log.Fatal(err)
	}
}
