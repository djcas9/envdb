package main

import (
	"encoding/json"
	"io/ioutil"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"text/template"
)

type Context struct {
	Name    string
	Version string
	Nodes   []*NodeDb
}

func renderFile(w http.ResponseWriter, filename string) error {
	var file []byte
	var ext string
	var err error

	nodes, _ := AllNodes()

	context := Context{
		Name:    Name,
		Version: Version,
		Nodes:   nodes,
	}

	if filename == "/" {
		filename = "/index.html"
	} else if filename == "/favicon.ico" || filename == "/favicon.png" {
		//..
	} else {
		return nil
	}

	if DEV_MODE {
		// dev
		file, err = ioutil.ReadFile("web" + filename)
	} else {
		file, err = Asset("web" + filename)
	}

	if err != nil {
		return err
	}

	ext = filepath.Ext("web" + filename)

	if ext != "" {
		w.Header().Set("Content-Type", mime.TypeByExtension(ext))
	}
	if file != nil {
		// w.Write(file)
		t, _ := template.New("index").Delims("<%", "%>").Parse(string(file))

		Log.Debug("HTTP Rendering file: ", filename)
		t.Execute(w, context)
	}

	return nil
}

func RouteIndex(w http.ResponseWriter, r *http.Request) {
	err := renderFile(w, r.URL.Path)

	if err != nil {
		Log.Error(err)
	}
}

func RouteDeleteQuery(w http.ResponseWriter, r *http.Request) {
	var errorMsg string = ""

	r.ParseForm()

	if r.Method == "POST" {
		id, err := strconv.ParseInt(r.PostFormValue("id"), 10, 64)

		Log.Debug("got id: ", id)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		query, err := FindSavedQueryById(id)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := query.Delete(); err != nil {
			errorMsg = err.Error()
		}

		data := map[string]interface{}{
			"error": errorMsg,
			"query": query,
		}

		js, err := json.Marshal(data)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}
}

func RouteSavedQueries(w http.ResponseWriter, r *http.Request) {
	data, _ := AllSavedQueries()

	js, err := json.Marshal(data)

	if err != nil {
		Log.Error("Error: ", err)

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func RouteSaveQuery(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	if r.Method == "POST" {
		query := QueryDb{
			Name:  r.PostFormValue("name"),
			Query: r.PostFormValue("query"),
			Type:  r.PostFormValue("type"),
		}

		err := NewSavedQuery(query)
		var errorMsg string = ""

		if err != nil {
			errorMsg = err.Error()
		}

		data := map[string]interface{}{
			"error": errorMsg,
			"query": query,
		}

		js, err := json.Marshal(data)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}

}

func RouteNodes(w http.ResponseWriter, r *http.Request) {

	nodes, _ := AllNodes()

	js, err := json.Marshal(nodes)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}
