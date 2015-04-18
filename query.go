package main

import (
	"errors"
	"os/exec"
	"strconv"
	"strings"

	"github.com/nu7hatch/gouuid"
)

const (
	// Min OsQueryi Supported Version
	MinOsQueryVersion = "1.4.4"
)

// Query wrapper. Holds the raw sql and
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

// Check that the node has a proper osqueryi version
func CheckOsQueryVersion(version string) bool {
	if version == MinOsQueryVersion {
		return true
	}

	sv := strings.Split(version, ".")
	cv := strings.Split(MinOsQueryVersion, ".")

	if len(sv) != 3 {
		return false
	}

	svi, err := strconv.Atoi(sv[0])

	cvi, err := strconv.Atoi(cv[0])

	if err != nil {
		return false
	}

	if svi < cvi {
		return false
	}

	svi2, err := strconv.Atoi(sv[1])

	cvi2, err := strconv.Atoi(cv[1])

	if err != nil {
		return false
	}

	if svi2 < cvi2 {
		return false
	}

	svi3, err := strconv.Atoi(sv[1])

	cvi3, err := strconv.Atoi(cv[1])

	if err != nil {
		return false
	}

	if svi3 < cvi3 {
		return false
	}

	return true
}

// Gather information about osqueryi from the node.
func OsQueryInfo() (bool, string) {
	var output []byte

	binary, lookErr := exec.LookPath("osqueryi")

	if lookErr != nil {
		return false, string(output)
	}

	output, err := exec.Command(binary, "--version").CombinedOutput()

	data := string(output)

	if err != nil {
		return false, data
	}

	newData := strings.Trim(strings.Replace(data, "osqueryi version ", "", -1), "\n")

	return true, newData
}

// Run a query for the node and return its
// combinded outpout.
func (self *Query) Run() ([]byte, error) {
	var output []byte

	binary, lookErr := exec.LookPath("osqueryi")

	if lookErr != nil {
		return output, lookErr
	}

	items := []string{binary, "--" + self.Format, self.Sql}

	output, err := exec.Command("/usr/bin/sudo", items...).CombinedOutput()

	if err != nil {
		err = errors.New(string(output))
	}

	return output, err
}
