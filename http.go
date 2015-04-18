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
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/mephux/gotalk"
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

var store = sessions.NewCookieStore(securecookie.GenerateRandomKey(32))

func NewWebServer(webPort int, server *Server) {
	r := mux.NewRouter()
	r.StrictSlash(true)

	gotalk.Handle("tables", func(sql SqlRequest) (QueryResults, error) {
		var tables []string

		data := server.Send(sql.Id, Query{
			Sql:    ".tables",
			Format: "json",
		})

		if len(data.Results) != 1 {
			d := QueryResults{}

			err := errors.New(fmt.Sprintf("Node not found for id (%s).", sql.Id))
			return d, err
		}

		tables = strings.Split(data.Results[0].Results.(string), "\n")

		for i, t := range tables {
			newT := strings.Replace(t, "  => ", "", -1)
			tables[i] = newT
		}

		tables = tables[:len(tables)-1]
		data.Results[0].Results = tables

		return data.Results[0], data.Error
	})

	gotalk.Handle("table-info", func(sql SqlRequest) (QueryResults, error) {
		data := server.Send(sql.Id, Query{
			Sql:    sql.Sql,
			Format: "json",
		})

		if len(data.Results) != 1 {
			d := QueryResults{}

			err := errors.New(fmt.Sprintf("Node not found for id (%s).", sql.Id))
			return d, err
		}

		return data.Results[0], data.Error
	})

	gotalk.Handle("query", func(sql SqlRequest) ([]QueryResults, error) {
		data := server.Send(sql.Id, Query{
			Sql:    sql.Sql,
			Format: "json",
		})

		return data.Results, data.Error
	})

	gotalk.Handle("disconnect", func(id string) error {
		return server.Disconnect(id)
	})

	gotalk.Handle("delete", func(id string) error {
		return server.Delete(id)
	})

	ws := gotalk.WebSocketHandler()
	ws.OnAccept = onAccept

	if DevMode {
		// dev
		Log.Debugf("Loading assets from disk.")
		r.PathPrefix("/public/").Handler(http.FileServer(http.Dir("./web/")))
	} else {
		Log.Debugf("Loading assets from memory.")
		r.PathPrefix("/public/").Handler(http.FileServer(
			&assetfs.AssetFS{
				Asset:    Asset,
				AssetDir: AssetDir,
				Prefix:   "web",
			},
		))
	}

	r.Handle("/", handlerFunc(RouteIndex))
	r.Handle("/login", handlerFunc(RouteLogin))
	r.Handle("/query/save", handlerFunc(RouteSaveQuery))
	r.Handle("/query/delete", handlerFunc(RouteDeleteQuery))
	r.Handle("/api/v1/nodes", handlerFunc(RouteNodes))
	r.Handle("/api/v1/queries", handlerFunc(RouteSavedQueries))

	http.Handle("/gotalk/", ws)
	http.Handle("/", r)

	var port string = fmt.Sprintf(":%d", webPort)
	var err error
	var proto string = "https"

	Log.Infof("** NOTICE ** HTTP Server: %s://%s%s", proto, "127.0.0.1", port)

	err = http.ListenAndServeTLS(port, DefaultPublicKeyPath, DefaultPrivateKeyPath, context.ClearHandler(http.DefaultServeMux))

	if err != nil {
		Log.Error(err.Error())
	}
}

func (f handlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	Log.Debugf("HTTP: %s %s", r.Method, r.URL)

	session, err := store.Get(r, "envdb")

	if err != nil {
		Log.Debug(err)
	}

	if session.IsNew {
		Log.Debug("Create New Session (cookie)")

		session.Options.Secure = true

		if err := session.Save(r, w); err != nil {
			Log.Fatal(err)
		}
	}

	if r.URL.String() != "/login" {
		if session.Values["current_user"] == nil {
			http.Redirect(w, r, "/login", 301)
			return
		}
	}

	err = f(w, r)

	if err != nil {
		Log.Fatal(err)
	}
}
