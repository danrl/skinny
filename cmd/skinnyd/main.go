package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/danrl/skinny/config"
	"github.com/danrl/skinny/proto/consensus"
	"github.com/danrl/skinny/proto/control"
	"github.com/danrl/skinny/proto/lock"
	"github.com/danrl/skinny/skinny"
	"google.golang.org/grpc"
)

func main() {
	configFile := flag.String("config", "/etc/skinny/config.yml", "Skinny configuration file")
	flag.Parse()

	cfg, err := config.NewInstanceConfig(*configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}
	in := skinny.New(cfg.Name, cfg.Increment, cfg.Timeout)

	// add peers
	for _, peer := range cfg.Peers {
		conn, err := grpc.Dial(peer.Address, grpc.WithInsecure())
		if err != nil {
			fmt.Fprintf(os.Stderr, "dial: %v", err)
			os.Exit(1)
		}
		err = in.AddPeer(peer.Name, consensus.NewConsensusClient(conn))
		if err != nil {
			conn.Close()
			fmt.Fprintf(os.Stderr, "add peer `%v`: %v", peer.Name, err)
			os.Exit(1)
		}
	}

	// register and serve protocols
	grpcServer := grpc.NewServer()
	consensus.RegisterConsensusServer(grpcServer, in)
	lock.RegisterLockServer(grpcServer, in)
	control.RegisterControlServer(grpcServer, in)

	// start listener
	listener, err := net.Listen("tcp", cfg.Listen)
	if err != nil {
		fmt.Fprintf(os.Stderr, "listen: %v", err)
		os.Exit(1)
	}
	grpcServer.Serve(listener)
}
