package main

import (
	"encoding/json"
	"io/ioutil"
	"mime"
	"net/http"
	"path/filepath"
	"text/template"
)

type Context struct {
	Name    string
	Version string
	Agents  []*DbAgent
}

func renderFile(w http.ResponseWriter, filename string) error {
	var file []byte
	var ext string
	var err error

	agents, _ := AllAgents()

	context := Context{
		Name:    Name,
		Version: Version,
		Agents:  agents,
	}

	if filename == "/" {
		filename = "/index.html"
	}

	Log.Debug("TEST:::", filename)

	file, err = ioutil.ReadFile("web" + filename)

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

func RouteAgents(w http.ResponseWriter, r *http.Request) {

	agents, _ := AllAgents()

	js, err := json.Marshal(agents)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}
