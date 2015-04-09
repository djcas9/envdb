package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"text/template"
)

var funcs = template.FuncMap{}

func newTemplate(filename string) *template.Template {
	var file []byte
	var err error

	if DEV_MODE {
		// dev
		file, err = ioutil.ReadFile("web/" + filename)
	} else {
		file, err = Asset("web/" + filename)
	}

	if err != nil {
		Log.Error(err)
	}

	// Log.Debug("HTTP Rendering file: ", filename)
	return template.Must(template.New("*").Delims("<%", "%>").Funcs(funcs).Parse(string(file)))
}

var tpls = map[string]*template.Template{
	"home": newTemplate("index.html"),
}

func executeTemplate(w http.ResponseWriter, name string, status int, data interface{}) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)

	return tpls[name].Execute(w, data)
}

func RouteIndex(w http.ResponseWriter, r *http.Request) error {

	nodes, _ := AllNodes()

	return executeTemplate(w, "home", 200, map[string]interface{}{
		"Section": "home",
		"Name":    Name,
		"Version": Version,
		"Nodes":   nodes,
	})

}

func RouteDeleteQuery(w http.ResponseWriter, r *http.Request) error {
	var errorMsg string = ""

	r.ParseForm()

	if r.Method == "POST" {
		id, err := strconv.ParseInt(r.PostFormValue("id"), 10, 64)

		Log.Debug("got id: ", id)

		if err != nil {
			return err
		}

		query, err := FindSavedQueryById(id)

		if err != nil {
			return err
		}

		if err := query.Delete(); err != nil {
			return err
		}

		data := map[string]interface{}{
			"error": errorMsg,
			"query": query,
		}

		js, err := json.Marshal(data)

		if err != nil {
			return err
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}

	return nil
}

func RouteSavedQueries(w http.ResponseWriter, r *http.Request) error {
	data, _ := AllSavedQueries()

	js, err := json.Marshal(data)

	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
	return nil
}

func RouteSaveQuery(w http.ResponseWriter, r *http.Request) error {
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
			return err
		}

		data := map[string]interface{}{
			"error": errorMsg,
			"query": query,
		}

		js, err := json.Marshal(data)

		if err != nil {
			return err
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}

	return nil
}

func RouteNodes(w http.ResponseWriter, r *http.Request) error {

	nodes, _ := AllNodes()

	js, err := json.Marshal(nodes)

	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
	return nil
}
