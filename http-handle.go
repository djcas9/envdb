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

	var layout []byte

	if DevMode {
		file, err = ioutil.ReadFile("web/" + filename)
		layout, err = ioutil.ReadFile("web/layout.html")
	} else {
		file, err = Asset("web/" + filename)
		layout, err = Asset("web/layout.html")
	}

	if err != nil {
		Log.Error(err)
	}

	html := string(layout) + string(file)
	return template.Must(template.New("*").Delims("<%", "%>").Funcs(funcs).Parse(html))
}

func executeTemplate(w http.ResponseWriter, name string, status int, data interface{}) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)

	return newTemplate(name).Execute(w, data)
}

func RouteIndex(w http.ResponseWriter, r *http.Request) error {

	session, err := store.Get(r, "envdb")

	if err != nil {
		Log.Debug(err)
	}

	nodes, _ := AllNodes()

	user, err := FindUserByEmail(session.Values["current_user"].(string))

	if err != nil {
		session.Values["current_user"] = nil
		http.Redirect(w, r, "/login", 301)
		return nil
	}

	return executeTemplate(w, "index.html", 200, map[string]interface{}{
		"Section": "home",
		"Name":    Name,
		"Version": Version,
		"Nodes":   nodes,
		"User":    user,
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

func RouteLogin(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return executeTemplate(w, "login.html", 200, map[string]interface{}{
			"Section": "login",
			"Name":    Name,
			"Version": Version,
		})
	} else if r.Method == "POST" {
		r.ParseForm()

		session, err := store.Get(r, "envdb")

		if err != nil {
			Log.Debug(err)
		}

		email := r.PostFormValue("email")
		password := r.PostFormValue("password")

		user, err := FindUserByEmail(email)

		if err != nil {
			Log.Warn(err)
			http.Redirect(w, r, "/", 302)
			return nil
		}

		if !user.ValidatePassword(password) {
			Log.Warnf("Authentication Failed for user: %s", email)
			http.Redirect(w, r, "/", 302)
			return nil
		}

		Log.Debug("Setting current user.")
		session.Values["current_user"] = email

		if err := session.Save(r, w); err != nil {
			Log.Fatal(err)
		}

		http.Redirect(w, r, "/", 302)
		return nil
	} else if r.Method == "DELETE" {
		session, err := store.Get(r, "envdb")

		if err != nil {
			Log.Debug(err)
		}

		session.Values["current_user"] = nil
		session.Options.MaxAge = -1

		if err := session.Save(r, w); err != nil {
			Log.Fatal(err)
		}

		return nil
	}

	http.Redirect(w, r, "/", 302)
	return nil
}
