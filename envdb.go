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
	Name    = "envdb"
	Version = "0.2.1"

	DefaultServerPort    = 3636
	DefaultWebServerPort = 8080
)

var (
	TimeFormat = "15:04:05"

	app   = kingpin.New(Name, "The Environment Database - Ask your environment questions")
	debug = app.Flag("debug", "Enable debug logging.").Bool()
	dev   = app.Flag("dev", "Enable dev mode. (read assets from disk and enable debug output)").Bool()
	quiet = app.Flag("quiet", "Remove all output logging.").Short('q').Bool()

	server = app.Command("server", "Start the tcp server for node connections.")
	// serverConfig = server.Flag("config", "Server configuration file.").File()
	serverPort = server.Flag("port", "Port for the server to listen on.").PlaceHolder(fmt.Sprintf("%d", DefaultServerPort)).Int()

	serverWebPort = server.Flag("http-port", "Port for the web server to listen on.").PlaceHolder(fmt.Sprintf("%d", DefaultWebServerPort)).Int()

	node = app.Command("node", "Register a new node.")
	// clientConfig = client.Flag("config", "Client configuration file.").File()
	nodeName   = node.Arg("node-name", "A name used to uniquely identify this node.").Required().String()
	nodeServer = node.Flag("server", "Address for server to connect to.").PlaceHolder("127.0.0.1").Required().String()
	nodePort   = node.Flag("port", "Port to use for connection.").Int()

	users      = app.Command("users", "User Management (Default lists all users).")
	addUser    = users.Flag("add", "Add a new user.").Bool()
	removeUser = users.Flag("remove", "Remove user by email.").PlaceHolder("email").String()

	Log *Logger

	DEV_MODE bool
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
		DEV_MODE = true
		Log.SetLevel(DebugLevel)
		Log.Info("DEBUG MODE ENABLED.")
	} else {
		DEV_MODE = false
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

	if start {
		if err := svr.Run(svrWebPort); err != nil {
			Log.Error(err)
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
