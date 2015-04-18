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

// Check is the daemon is running
func (self *Daemon) Running() (bool, *os.Process, error) {
	d, err := self.d.Search()

	if err != nil {
		return false, d, err
	}

	if err := d.Signal(syscall.Signal(0)); err != nil {
		return false, d, err
	}

	return true, d, nil
}

// Start the servers as a daemon process.
func (self *Daemon) StartServer(svr *Server, svrWebPort int) {

	if ok, p, _ := self.Running(); ok {
		fmt.Printf("%s server is already running. PID: %d\n", Name, p.Pid)
		return
	}

	fmt.Printf("Starting %s server in daemon mode\n", Name)
	Log.SetLevel(DebugLevel)

	d, err := self.d.Reborn()

	if err != nil {
		Log.Fatal(err)
	}

	if d != nil {
		return
	}

	defer self.d.Release()

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

// Get the current status of the daemon.
func (self *Daemon) Status() {

	if ok, p, _ := self.Running(); ok {
		fmt.Printf("%s server is running. PID: %d\n", Name, p.Pid)
	} else {
		self.d.Release()
		fmt.Printf("%s server is NOT running.\n", Name)
	}
}

// Stop the daemon
func (self *Daemon) Stop() {
	if ok, p, _ := self.Running(); ok {
		fmt.Printf("Attempting to shutdown %s server. PID: %d\n", Name, p.Pid)
		if err := p.Signal(syscall.Signal(syscall.SIGQUIT)); err != nil {
			Log.Fatal(err)
		}
	} else {
		self.d.Release()
		fmt.Printf("%s server is not running.\n", Name)
	}
}
