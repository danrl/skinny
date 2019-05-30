package skinny

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	pb "github.com/danrl/skinny/proto/lock"
)

// Acquire tries to acquire a lock
func (in *Instance) Acquire(ctx context.Context, req *pb.AcquireRequest) (*pb.AcquireResponse, error) {
	in.mu.Lock()
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
	} else if retries < 3 {
		retries++
		backoff := time.Duration(retries) * 2 * time.Millisecond
		jitter := time.Duration(rand.Int63n(1000)) * time.Microsecond
		fmt.Printf("waiting %v before retry #%v\n", backoff+jitter, retries)

		in.mu.Unlock()
		time.Sleep(backoff + jitter)
		in.mu.Lock()

		fmt.Printf("retry #%v\n", retries)
		goto retry
	}
	resp := pb.AcquireResponse{
		Acquired: in.holder == req.Holder,
		Holder:   in.holder,
	}
	in.mu.Unlock()

	return &resp, nil
}

// Release releases a previously held lock
func (in *Instance) Release(ctx context.Context, req *pb.ReleaseRequest) (*pb.ReleaseResponse, error) {
	in.mu.Lock()
	fmt.Println("client: release lock")
	retries := 0
retry:
	promised := in.propose()
	if promised {
		_ = in.commit(in.promised, "")
	} else if retries < 3 {
		retries++
		backoff := time.Duration(retries) * 2 * time.Millisecond
		jitter := time.Duration(rand.Int63n(1000)) * time.Microsecond
		fmt.Printf("waiting %v before retry #%v\n", backoff+jitter, retries)

		in.mu.Unlock()
		time.Sleep(backoff + jitter)
		in.mu.Lock()

		fmt.Printf("retry #%v\n", retries)
		goto retry
	}
	resp := pb.ReleaseResponse{
		Released: in.holder == "",
	}
	in.mu.Unlock()

	return &resp, nil
}
