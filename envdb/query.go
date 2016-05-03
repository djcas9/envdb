package envdb

import (
	"encoding/json"
	"errors"
	"os/exec"
	"strings"

	"github.com/nu7hatch/gouuid"
)

const (
	// MinOsQueryVersion Supported osqueryi version
	MinOsQueryVersion        = "1.4.4"
	DefaultOsQueryConfigPath = "/etc/osquery/osquery.conf"
)

// Query Holds the raw sql and
// format options to be passed to osqueryi
type Query struct {
	Sql    string
	Format string
}

// QueryResults holds all results returned
// by osqueryi. This struct is used to transport
// the data from server to UI
type QueryResults struct {
	Id       string      `json:"id"`
	Name     string      `json:"name"`
	Hostname string      `json:"hostname"`
	Results  interface{} `json:"results"`
	Error    string      `json:"error"`
}

// Response wraps QueryResults.
// The Response struct also holds more request metadata
// and is used to paginate results in memory on the
// server
type Response struct {
	Id      string         `json:"id"`
	Results []QueryResults `json:"results"`
	Total   int            `json:"total"`
	Error   error          `json:"error"`
}

// OsQueryInfo holds information about osquery
type OsQueryMetadata struct {
	Enabled    bool
	Version    string
	ConfigPath string
}

// Initialize a new Response
func NewResponse() *Response {
	var id string

	if uuid, err := uuid.NewV4(); err == nil {
		id = uuid.String()
	}

	return &Response{
		Id:    id,
		Error: nil,
		Total: 0,
	}
}

// OsQueryInfo ather information about osqueryi from the node.
func OsQueryInfo() *OsQueryMetadata {
	var info = &OsQueryMetadata{
		Enabled:    false,
		Version:    "",
		ConfigPath: "",
	}

	var output []byte

	binary, lookErr := exec.LookPath("osqueryi")

	if lookErr != nil {
		return info
	}

	output, err := exec.Command(binary, "--version").CombinedOutput()

	data := string(output)

	if err != nil {
		info.Version = data

		return info
	}

	newData := strings.Trim(strings.Replace(data, "osqueryi version ", "", -1), "\n")

	info.Enabled = true
	info.Version = newData

	// Return before query exec
	if TestMode {
		return info
	}

	items := []string{binary, "--json", "select * from osquery_info;"}
	if output, err := exec.Command("/usr/bin/sudo", items...).CombinedOutput(); err != nil {
		return info
	} else {

		var jdata []map[string]interface{}

		if err := json.Unmarshal(output, &jdata); err != nil {
			return info
		}

		if data, ok := jdata[0]["config_path"]; ok {
			info.ConfigPath = data.(string)
		}

		return info
	}
}

// Run a query for the node and return its
// combinded outpout.
func (q *Query) Run() ([]byte, error) {
	var output []byte

	binary, lookErr := exec.LookPath("osqueryi")

	if lookErr != nil {
		return output, lookErr
	}

	items := []string{binary, "--" + q.Format, q.Sql}

	output, err := exec.Command("/usr/bin/sudo", items...).CombinedOutput()

	if err != nil {
		err = errors.New(string(output))
	}

	return output, err
}
