package skinny

import (
	"context"

	pb "github.com/danrl/skinny/proto/control"
)

// Status exposes internal state information of an instance
func (in *Instance) Status(ctx context.Context, req *pb.StatusRequest) (*pb.StatusResponse, error) {
	in.mu.RLock()
	defer in.mu.RUnlock()

	status := pb.StatusResponse{
		Name:      in.name,
		Increment: in.increment,
		Timeout:   in.timeout.String(),
		Promised:  in.promised,
		ID:        in.id,
		Holder:    in.holder,
	}

	for _, peer := range in.peers {
		status.Peers = append(status.Peers, &pb.StatusResponse_Peer{
			Name: peer.name,
		})
	}

	return &status, nil
}
