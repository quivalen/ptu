package main

import (
	"github.com/ivanilves/gopack/net/forwarder"
	"github.com/ivanilves/gopack/ssh/client"
	"github.com/ivanilves/gopack/ssh/listener"
	"github.com/ivanilves/gopack/util/config"
	"github.com/ivanilves/gopack/util/display"
	"log"
	"time"
)

const retrySeconds = 10

func main() {
	// Some tender erotic foreplay
	if config.IsHelpRequested() {
		display.PrintHelpAndExit()
	}

	// Load defaults: built-in or from file, if it exists
	d, err := config.LoadDefaults()
	if err != nil {
		log.Fatalf("Unable to load defaults: %s", err)
	}

	// Merge default config with params passed as command line arguments
	c, err := config.ParseArguments(d)
	if err != nil {
		log.Fatalf("Error while parsing command line arguments: %s", err)
	}

	display.PrintConfig(c.SSHServer, c.SSHUsername, c.SSHUseAgent, c.TargetHost, c.ExposedPort, c.ConnectTo, c.BuildID)

RETRY:
	// Initialize SSH client (at least try to!)
	sshClient, err := client.New(c.SSHServer, c.SSHUsername, c.SSHPassword, c.SSHUseAgent)
	if err != nil {
		log.Printf("Error initializing SSH client: %s (will retry)", err)
		time.Sleep(retrySeconds * time.Second)
		goto RETRY
	} else {
		log.Printf("[OK] SSH client initialized!")
	}

	// Set up SSH listener <exposed_bind>:<exposed_port> on the SSH server <ssh_server>:<ssh_port>
	sshListener, err := listener.New(sshClient, c.ExposedHost)
	if err != nil {
		log.Fatalf("Error setting up SSH listener: %s", err)
	}

	// Check, if listener really listens to specified bind address (GatewayPorts thing)
	realExposedBind, err := client.ProbeBindByPort(sshClient, c.ExposedPort)
	if err != nil {
		log.Printf("Error probing exposed bind: %s", err)
	}
	if c.ExposedBind != realExposedBind && realExposedBind != "0.0.0.0" {
		display.PrintGatewayPortsNB()
	}

	// Vamos muchachos!
	for {
		err := forwarder.Forward(sshListener, c.TargetHost)
		if err != nil {
			log.Printf("[OMG] Critical failure on forwarder! Will re-setup SSH connection.")
			time.Sleep(retrySeconds * time.Second)
			goto RETRY
		}
	}
}
