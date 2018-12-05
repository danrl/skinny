package skinny

import (
	"context"
	"fmt"

	pb "github.com/danrl/skinny/proto/lock"
)

// Acquire tries to acquire a lock
func (in *Instance) Acquire(ctx context.Context, req *pb.AcquireRequest) (*pb.AcquireResponse, error) {
	in.mu.Lock()
	defer in.mu.Unlock()

	fmt.Printf("client: acquire lock on behalf of '%v'\n", req.Holder)

	retries := 0
retry:
	promised := in.propose()
	if promised {
		if in.holder == "" {
			// The lock is available and we got promised an ID!
			_ = in.commit(in.promised, req.Holder)
		} else {
			// The lock is not available. Let's commit the learned holder.
			_ = in.commit(in.promised, in.holder)
		}
	} else {
		if retries < 3 {
			retries++
			fmt.Printf("retry #%v\n", retries)
			goto retry
		}
	}

	// Report results.
	return &pb.AcquireResponse{
		Acquired: in.holder == req.Holder,
		Holder:   in.holder,
	}, nil
}

// Release releases a previously held lock
func (in *Instance) Release(ctx context.Context, req *pb.ReleaseRequest) (*pb.ReleaseResponse, error) {
	in.mu.Lock()
	defer in.mu.Unlock()

	fmt.Println("client: release lock")

	retries := 0
retry:
	promised := in.propose()
	if promised {
		_ = in.commit(in.promised, "")
	} else {
		if retries < 3 {
			retries++
			fmt.Printf("retry #%v\n", retries)
			goto retry
		}
	}

	return &pb.ReleaseResponse{
		Released: in.holder == "",
	}, nil
}
