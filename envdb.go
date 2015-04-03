package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/mephux/envdb/log"
	"gopkg.in/alecthomas/kingpin.v1"
)

const (
	Name    = "envdb"
	Version = "0.1.0.beta"

	DefaultServerPort    = 3636
	DefaultWebServerPort = 8080
)

var (
	TimeFormat = "15:04:05"

	app   = kingpin.New(Name, "Environment Database")
	debug = app.Flag("debug", "Enable debug logging.").Bool()
	dev   = app.Flag("dev", "Enable dev mode. (read assets from disk, enable debug output)").Bool()
	quiet = app.Flag("quiet", "Remove all output logging.").Short('q').Bool()

	server = app.Command("server", "Start the tcp server for node connections.")
	// serverConfig = server.Flag("config", "Server configuration file.").File()
	serverPort = server.Flag("port", "Port for the server to listen on.").PlaceHolder(fmt.Sprintf("%d", DefaultServerPort)).Int()

	serverWebPort = server.Flag("http-port", "Port for the web server to listen on.").PlaceHolder(fmt.Sprintf("%d", DefaultWebServerPort)).Int()

	agent = app.Command("node", "Register a new node.")
	// clientConfig = client.Flag("config", "Client configuration file.").File()
	agentName   = agent.Arg("node-name", "A name used to uniquely identify this node.").Required().String()
	agentServer = agent.Flag("server", "Address for server to connect to.").PlaceHolder("127.0.0.1").Required().String()
	agentPort   = agent.Flag("port", "Port to use for connection.").Int()

	Log *log.Logger

	DEV_MODE bool
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	kingpin.Version(Version)
	args, err := app.Parse(os.Args[1:])

	Log = log.New()

	Log.Prefix = "envdb"

	if *dev {
		DEV_MODE = true
		Log.Info("DEBUG MODE ENABLED.")
	} else {
		DEV_MODE = false
	}

	if *debug {
		Log.SetLevel(log.DebugLevel)
	} else {
		Log.SetLevel(log.InfoLevel)
	}

	if *quiet {
		Log.SetLevel(log.FatalLevel)
	}

	switch kingpin.MustParse(args, err) {

	case server.FullCommand():

		var svrPort int = DefaultServerPort
		var svrWebPort int = DefaultWebServerPort

		if *serverPort != 0 {
			svrPort = *serverPort
		}

		if *serverWebPort != 0 {
			svrWebPort = *serverWebPort
		}

		svr, err := NewServer(svrPort)

		if err != nil {
			Log.Fatal(err)
		}

		if err := svr.Run(svrWebPort); err != nil {
			Log.Error(err)
		}

	case agent.FullCommand():

		var clntPort int = DefaultServerPort

		if *agentPort != 0 {
			clntPort = *agentPort
		}

		var c = Agent{
			Name:       *agentName,
			Host:       *agentServer,
			Port:       clntPort,
			RetryCount: 50,
		}

		config, err := NewAgentConfig()

		if err != nil {
			Log.Fatal(err)
		}

		c.Config = config

		if err := c.Run(); err != nil {
			Log.Error(err)
		}

	default:
		kingpin.Usage()
	}

}
