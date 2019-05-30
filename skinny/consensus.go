package skinny

import (
	"context"
	"fmt"
	"sync"

	pb "github.com/danrl/skinny/proto/consensus"
)

// Promise promises an ID for future use in a commit
func (in *Instance) Promise(ctx context.Context, req *pb.PromiseRequest) (*pb.PromiseResponse, error) {
	in.mu.Lock()
	defer in.mu.Unlock()

	var promise pb.PromiseResponse
	attachment := ""

	// attach previously committed values if there has been consensus in the past
	if in.id > 0 {
		promise.ID = in.id
		promise.Holder = in.holder
		attachment = fmt.Sprintf(" (attached previously committed ID %v and holder `%v`)", in.id, in.holder)
	}

	if req.ID > in.promised {
		promise.Promised = true
		in.promised = req.ID
		fmt.Printf("promised ID %v%v\n", req.ID, attachment)
	} else {
		fmt.Printf("did not promise ID %v%v\n", req.ID, attachment)
	}

	return &promise, nil
}

// Commit commits the value for a previously promised ID
func (in *Instance) Commit(ctx context.Context, req *pb.CommitRequest) (*pb.CommitResponse, error) {
	in.mu.Lock()
	defer in.mu.Unlock()

	if req.ID >= in.promised {
		in.id = req.ID
		in.holder = req.Holder
		fmt.Printf("committed ID %v and holder `%v`\n", in.id, in.holder)
	} else {
		fmt.Printf("did not commit ID %v and holder `%v`\n", req.ID, req.Holder)
	}

	return &pb.CommitResponse{
		Committed: req.ID == in.id,
	}, nil
}

// propose asks the quorum to promise a round number (ID). It learns previous consensus if there is any.
func (in *Instance) propose() bool {
	type response struct {
		from     string
		promised bool
		id       uint64
		holder   string
	}

	in.promised += in.increment

	responses := make(chan *response)
	ctx, cancel := context.WithTimeout(context.Background(), in.timeout)
	// We cancel as soon as we have a majority to speed things up.
	// We always cancel before leaving the function to prevent a context leak.
	defer cancel()

	wg := sync.WaitGroup{}
	for _, p := range in.peers {
		wg.Add(1)

		// send proposal
		go func(p peer) {
			defer wg.Done()

			resp, err := p.client.Promise(ctx, &pb.PromiseRequest{
				ID: in.promised,
			})
			fmt.Printf("propose ID %v to %v: sent\n", in.promised, p.name)
			if err != nil {
				if ctx.Err() == context.Canceled {
					fmt.Printf("propose ID %v to %v: canceled\n", in.promised, p.name)
					return
				}
				// We want errors which are not the result of a canceled
				// proposal to be counted as a negative answer (nay) later.
				// For that we emit an empty response into the channel in those
				// cases.
				responses <- &response{from: p.name}
				fmt.Printf("propose ID %v to %v: %v\n", in.promised, p.name, err)
				return
			}
			responses <- &response{
				from:     p.name,
				promised: resp.Promised,
				id:       resp.ID,
				holder:   resp.Holder,
			}
		}(p)
	}

	// close responses channel once all responses have been received, failed, or
	// canceled
	go func() {
		wg.Wait()
		close(responses)
	}()

	// count the votes
	yea, nay := 1, 0
	canceled := false
	for r := range responses {
		// count the promises
		if r.promised {
			yea++
			fmt.Printf("propose ID %v to %v: got yea\n", in.promised, r.from)
		} else {
			nay++
			fmt.Printf("propose ID %v to %v: got nay\n", in.promised, r.from)
		}

		// learn previously committed ID and holder from other instances
		if r.id > in.id {
			in.id = r.id
			in.holder = r.holder
			fmt.Printf("propose ID %v to %v: learned ID %v and holder `%v`\n", in.promised, r.from, r.id, r.holder)
		}

		// stop counting as soon as we have a majority
		if !canceled {
			// Cancel all in-flight proposal if we have reached a majority.
			if in.isMajority(yea) || in.isMajority(nay) {
				cancel()
				canceled = true
			}
		}
	}

	// if we learned a ID than our initial proposal suggested, then we also promise this higher ID
	if in.id > in.promised {
		in.promised = in.id
		fmt.Printf("jumped to promise ID %v\n", in.promised)
	}

	return in.isMajority(yea)
}

// commit asks the quorum to accept the acquisition or release of a lock
func (in *Instance) commit(id uint64, holder string) bool {
	type response struct {
		from      string
		committed bool
	}

	fmt.Printf("committing ID %v and holder `%v`\n", id, holder)

	responses := make(chan *response)
	ctx, cancel := context.WithTimeout(context.Background(), in.timeout)
	defer cancel()

	wg := sync.WaitGroup{}
	for _, p := range in.peers {
		wg.Add(1)

		// send commit requests
		go func(p peer) {
			defer wg.Done()

			resp, err := p.client.Commit(ctx, &pb.CommitRequest{
				ID:     id,
				Holder: holder,
			})
			fmt.Printf("commit ID %v and holder `%v` to %v: sent\n", id, holder, p.name)

			if err != nil {
				if ctx.Err() == context.DeadlineExceeded {
					fmt.Printf("commit ID %v and holder `%v` to %v: deadline exceeded\n", id, holder, p.name)
					return
				}
				// We want errors which are not the result of a canceled commit to be counted as a negative answer (nay)
				// later. For that we emit an empty response into the channel in those cases.
				responses <- &response{from: p.name}
				fmt.Printf("commit ID %v and holder `%v` to %v: %v\n", id, holder, p.name, err)
				return
			}
			responses <- &response{
				from:      p.name,
				committed: resp.Committed,
			}
		}(p)
	}

	// close responses channel once all reponses have been received, failed, or canceled
	go func() {
		wg.Wait()
		close(responses)
	}()

	// we have to commit our own data
	in.id = id
	in.holder = holder

	// count the vote
	yea := 1 // we just committed our own data. make it count.
	for r := range responses {
		if r.committed {
			yea++
			fmt.Printf("commit ID %v and holder `%v` to %v: got yea\n", id, holder, r.from)
			continue
		}
		fmt.Printf("commit ID %v and holder `%v` to %v: got nay\n", id, holder, r.from)
	}

	return in.isMajority(yea)
}
