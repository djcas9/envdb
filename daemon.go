package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/sevlyar/go-daemon"
)

// Daemon wrapper for the daemon.Context struct
type Daemon struct {
	d *daemon.Context
}

// Running will check if the daemon is running or not
func (daemon *Daemon) Running() (bool, *os.Process, error) {
	d, err := daemon.d.Search()

	if err != nil {
		return false, d, err
	}

	if err := d.Signal(syscall.Signal(0)); err != nil {
		return false, d, err
	}

	return true, d, nil
}

// StartServer starts the servers (http/tcp) as a daemon process.
func (daemon *Daemon) StartServer(svr *Server, svrWebPort int) {

	if ok, p, _ := daemon.Running(); ok {
		fmt.Printf("%s server is already running. PID: %d\n", Name, p.Pid)
		return
	}

	fmt.Printf("Starting %s server in daemon mode\n", Name)
	Log.SetLevel(DebugLevel)

	d, err := daemon.d.Reborn()

	if err != nil {
		Log.Fatal(err)
	}

	if d != nil {
		return
	}

	defer daemon.d.Release()

	go func() {
		if err := svr.Run(svrWebPort); err != nil {
			Log.Error(err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	for {
		select {
		case sig := <-sigChan:
			Log.Debug("Go Signal: ", sig)
			svr.Shutdown()
			os.Exit(1)
		}
	}

}

// Status will get the current status of the daemon.
func (daemon *Daemon) Status() {

	if ok, p, _ := daemon.Running(); ok {
		fmt.Printf("%s server is running. PID: %d\n", Name, p.Pid)
	} else {
		daemon.d.Release()
		fmt.Printf("%s server is NOT running.\n", Name)
	}
}

// Stop will shutdown and stop the daemon process
func (daemon *Daemon) Stop() {
	if ok, p, _ := daemon.Running(); ok {
		fmt.Printf("Attempting to shutdown %s server. PID: %d\n", Name, p.Pid)
		if err := p.Signal(syscall.Signal(syscall.SIGQUIT)); err != nil {
			Log.Fatal(err)
		}
	} else {
		daemon.d.Release()
		fmt.Printf("%s server is not running.\n", Name)
	}
}
