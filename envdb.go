package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/howeyc/gopass"
	"gopkg.in/alecthomas/kingpin.v1"
)

const (
	// Name application name
	Name = "envdb"

	// Version application version number
	Version = "0.3.1"

	// DefaultServerPort the default tcp server port
	DefaultServerPort = 3636

	// DefaultWebServerPort the default web server port
	DefaultWebServerPort = 8080
)

var (
	// TimeFormat global time format string
	TimeFormat = "15:04:05"

	app   = kingpin.New(Name, "The Environment Database - Ask your environment questions")
	debug = app.Flag("debug", "Enable debug logging.").Short('v').Bool()
	dev   = app.Flag("dev", "Enable dev mode.").Bool()
	quiet = app.Flag("quiet", "Remove all output logging.").Short('q').Bool()

	server        = app.Command("server", "Start the tcp server for node connections.")
	serverCommand = server.Arg("command", "Daemon command. (start,status,stop)").String()
	// serverConfig = server.Flag("config", "Server configuration file.").File()
	serverPort = server.Flag("port", "Port for the server to listen on.").
			Short('p').PlaceHolder(fmt.Sprintf("%d", DefaultServerPort)).Int()

	serverWebPort = server.Flag("http-port", "Port for the web server to listen on.").
			Short('P').PlaceHolder(fmt.Sprintf("%d", DefaultWebServerPort)).Int()

	node = app.Command("node", "Register a new node.")
	// clientConfig = client.Flag("config", "Client configuration file.").File()
	nodeName = node.Arg("node-name", "A name used to uniquely identify this node.").Required().String()

	nodeServer = node.Flag("server", "Address for server to connect to.").
			Short('s').PlaceHolder("127.0.0.1").Required().String()

	nodePort = node.Flag("port", "Port to use for connection.").Short('p').Int()

	users      = app.Command("users", "User Management (Default lists all users).")
	addUser    = users.Flag("add", "Add a new user.").Short('a').Bool()
	removeUser = users.Flag("remove", "Remove user by email.").Short('r').PlaceHolder("email").String()

	// Log Global logger
	Log *Logger

	// DevMode Development mode switch. If true
	// debug logging and serving assets from disk
	// is enabled.
	DevMode bool
)

func initLogger() {
	Log = NewLogger()

	Log.Prefix = Name

	if *debug {
		Log.SetLevel(DebugLevel)
	} else {
		Log.SetLevel(InfoLevel)
	}

	if *dev {
		DevMode = true
		Log.SetLevel(DebugLevel)
		Log.Info("DEBUG MODE ENABLED.")
	} else {
		DevMode = false
	}

	if *quiet {
		Log.SetLevel(FatalLevel)
	}

}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	app.Version(Version)
	args, err := app.Parse(os.Args[1:])

	initLogger()

	switch kingpin.MustParse(args, err) {

	case users.FullCommand():
		serverSetup(false)

		if *addUser {
			addDBUser()
			return
		}

		if len(*removeUser) > 0 {
			if user, err := FindUserByEmail(*removeUser); err != nil {
				Log.Fatal(err)
			} else {
				if err := user.Delete(); err != nil {
					Log.Fatal(err)
				}
			}

			Log.Info("User removed successfully.")
			return
		}

		users, err := FindAllUsers()

		if err != nil {
			Log.Fatal(err)
		}

		fmt.Println("Listing Users: ")
		for _, user := range users {
			fmt.Printf("  * %s (%s)\n", user.Name, user.Email)
		}

	case server.FullCommand():

		serverSetup(true)

	case node.FullCommand():

		output, err := exec.Command("whoami").Output()

		if err != nil {
			Log.Fatalf("Error: %s", err)
			os.Exit(-1)
		}

		if strings.Trim(string(output), "\n") != "root" {
			Log.Fatal("You must run the node client as root.")
			os.Exit(-1)
		}

		var clntPort int = DefaultServerPort

		if *nodePort != 0 {
			clntPort = *nodePort
		}

		var c = Node{
			Name:       *nodeName,
			Host:       *nodeServer,
			Port:       clntPort,
			RetryCount: 50,
		}

		config, err := NewNodeConfig()

		if err != nil {
			Log.Fatal(err)
		}

		c.Config = config

		if err := c.Run(); err != nil {
			Log.Error(err)
		}

	default:
		app.Usage(os.Stdout)
	}

}

func serverSetup(start bool) {
	var svrPort = DefaultServerPort
	var svrWebPort = DefaultWebServerPort

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

	if !start {
		return
	}

	if len(*serverCommand) <= 0 {
		if err := svr.Run(svrWebPort); err != nil {
			Log.Error(err)
		}
	} else {

		switch *serverCommand {
		case "start":
			svr.Config.Daemon.StartServer(svr, svrWebPort)
			break
		case "stop":
			svr.Config.Daemon.Stop()
			break
		case "status":
			svr.Config.Daemon.Status()
			break
		default:
			{
				Log.Fatalf("Error: Unknown Command %s.", *serverCommand)
			}
		}

	}

}

func ask(reader *bufio.Reader, question string) string {
	fmt.Print(question)

	value, _ := reader.ReadString('\n')
	trim := strings.Trim(value, "\n")

	if len(trim) <= 0 {
		Log.Fatalf("value cannot be blank.")
	}

	return trim
}

func addDBUser() {
	reader := bufio.NewReader(os.Stdin)

	name := ask(reader, "Name: ")

	email := ask(reader, "Email ")

	if !IsEmail(email) {
		Log.Fatalf("%s is not a valid email address.", email)
	}

	fmt.Print("Password: ")
	pass := gopass.GetPasswd()

	fmt.Print("Confirm: ")
	cpass := gopass.GetPasswd()

	if string(pass) != string(cpass) {
		Log.Fatal("Password and confirm do not match.")
	}

	user := &UserDb{
		Name:     name,
		Email:    email,
		Password: string(pass),
	}

	err := CreateUser(user)

	if err != nil {
		Log.Fatal(err)
	}

	Log.Info("User created successfully.")
}
