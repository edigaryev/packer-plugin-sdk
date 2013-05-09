// The plugin package provides the functionality to both expose a Packer
// plugin binary and to connect to an existing Packer plugin binary.
//
// Packer supports plugins in the form of self-contained external static
// Go binaries. These binaries behave in a certain way (enforced by this
// package) and are connected to in a certain way (also enforced by this
// package).
package plugin

import (
	"fmt"
	"github.com/mitchellh/packer/packer"
	"log"
	"net"
	"net/rpc"
	"os"
	packrpc "github.com/mitchellh/packer/packer/rpc"
	"strconv"
	"strings"
)

// This serves a single RPC connection on the given RPC server on
// a random port.
func serve(server *rpc.Server) (err error) {
	minPort, err := strconv.ParseInt(os.Getenv("PACKER_PLUGIN_MIN_PORT"), 10, 32)
	if err != nil {
		return
	}

	maxPort, err := strconv.ParseInt(os.Getenv("PACKER_PLUGIN_MAX_PORT"), 10, 32)
	if err != nil {
		return
	}

	log.Printf("Plugin minimum port: %d\n", minPort)
	log.Printf("Plugin maximum port: %d\n", maxPort)

	var address string
	var listener net.Listener
	for port := minPort; port <= maxPort; port++ {
		address = fmt.Sprintf(":%d", port)
		listener, err = net.Listen("tcp", address)
		if err != nil {
			if !strings.Contains(err.Error(), "address already in use") {
				// Not an address already in use error, return.
				return
			} else {
				// Address is in use, just try another
				err = nil
				continue
			}
		}

		break
	}

	defer listener.Close()

	// Output the address to stdout
	log.Printf("Plugin address: %s\n", address)
	fmt.Println(address)
	os.Stdout.Sync()

	// Accept a connection
	log.Println("Waiting for connection...")
	conn, err := listener.Accept()
	if err != nil {
		log.Printf("Error accepting connection: %s\n", err.Error())
		return
	}

	// Serve a single connection
	log.Println("Serving a plugin connection...")
	server.ServeConn(conn)
	return
}

// Serves a builder from a plugin.
func ServeBuilder(builder packer.Builder) {
	log.Println("Preparing to serve a builder plugin...")

	server := rpc.NewServer()
	packrpc.RegisterBuilder(server, builder)

	if err := serve(server); err != nil {
		log.Panic(err)
	}
}

// Serves a command from a plugin.
func ServeCommand(command packer.Command) {
	log.Println("Preparing to serve a command plugin...")

	server := rpc.NewServer()
	packrpc.RegisterCommand(server, command)

	if err := serve(server); err != nil {
		log.Panic(err)
	}
}
