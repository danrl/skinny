// Package skinny implements an educational distributed lock management service
package skinny

import (
	"errors"
	"fmt"
	"sync"
	"time"

	pb "github.com/danrl/skinny/proto/consensus"
)

// Instance represents a skinny distributed lock management service instance
type Instance struct {
	mu sync.RWMutex
	// begin protected fields
	name      string
	increment uint64
	timeout   time.Duration
	promised  uint64
	id        uint64
	holder    string
	peers     []peer
	// end protected fields
}

type peer struct {
	name   string
	client pb.ConsensusClient
}

var (
	// ErrDuplicatePeer is returned when peer already exists in the peer list
	ErrDuplicatePeer = errors.New("duplicate peer")
)

// New initializes a new skinny instance
func New(name string, increment uint64, timeout time.Duration) *Instance {
	in := Instance{
		name:      name,
		increment: increment,
		timeout:   timeout,
	}

	fmt.Println("initialized")
	return &in
}

// AddPeer adds a new peer to the peer list
func (in *Instance) AddPeer(name string, client pb.ConsensusClient) error {
	in.mu.Lock()
	defer in.mu.Unlock()

	// check for duplicate peers
	for _, p := range in.peers {
		if p.name == name || p.client == client {
			return ErrDuplicatePeer
		}
	}

	// add peer to the peer list
	in.peers = append(in.peers, peer{
		name:   name,
		client: client,
	})
	fmt.Printf("added peer %v\n", name)

	return nil
}

// isMajority returns true if the n represents a majority in the configured
// quorum. Caller must hold a (read) lock on i (Instance).
func (in *Instance) isMajority(n int) bool {
	return n > ((len(in.peers) + 1) / 2)
}
