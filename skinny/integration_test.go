package skinny

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/danrl/skinny/proto/consensus"
	"github.com/danrl/skinny/proto/lock"
)

// TestIntegration runs a typical scenario to test the inter-workings of the most important components. The whole test
// fails should a single step fail. This is because a failed step points to a inconsistent state. There is little use
// in testing an inconsistent quorum.
func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// set up a quorum of five
	quorum := []*mockInstance{}
	for i := 1; i <= 5; i++ {
		mi := newMockInstance(t, fmt.Sprintf("instance-%v", i), uint64(i), 200*time.Millisecond)
		defer mi.destroy()
		quorum = append(quorum, mi)
	}
	for i, mi := range quorum {
		for j, peer := range quorum {
			if i == j {
				// do not self-peer
				continue
			}
			err := mi.in.AddPeer(peer.in.name, consensus.NewConsensusClient(peer.conn))
			if err != nil {
				t.Fatalf("add peer: %v", err)
			}
		}
	}

	// check quorum for correct initial configuration
	for i, mi := range quorum {
		position := uint64(i + 1)

		// all IDs should be 0
		if mi.in.id != 0 {
			t.Fatalf("instance-%v: expected `%v`, got `%v`", position, 0, mi.in.id)
		}
		// all holders should be empty
		if mi.in.holder != "" {
			t.Fatalf("instance-%v: expected `%v`, got `%v`", position, "", mi.in.holder)
		}
		// no promises should have been made at this point
		if mi.in.promised != 0 {
			t.Fatalf("instance-%v: expected `%v`, got `%v`", position, 0, mi.in.promised)
		}
		// all increments should match the position in the quorum
		if mi.in.increment != position {
			t.Fatalf("instance-%v: expected `%v`, got `%v`", position, position, mi.in.increment)
		}
	}

	/*
	 *       (1)
	 *
	 *   (2)     (3)
	 *
	 *    (4)   (5)
	 *             \
	 *             🐹 [hamster is mocking a beaver here]
	 */
	// beaver asks instance-5 for the lock
	{
		req := lock.AcquireRequest{
			Holder: beaver,
		}
		resp, err := quorum[4].in.Acquire(context.Background(), &req)
		if err != nil {
			t.Fatalf("expected `%v`, got `%v`", nil, err)
		}
		if !resp.Acquired {
			t.Fatalf("expected `%v`, got `%v`", true, resp.Acquired)
		}
		if resp.Holder != req.Holder {
			t.Fatalf("expected `%v`, got `%v`", req.Holder, resp.Holder)
		}
		// check quorum state
		good := 0
		for _, mi := range quorum {
			if mi.in.promised == quorum[4].in.promised &&
				mi.in.id == quorum[4].in.id &&
				mi.in.holder == req.Holder {
				good++
			}
		}
		if good < 3 {
			t.Fatal("majority in bad state")
		}
	}

	/*
	 *       (1)
	 *
	 *   (2)     (3)
	 *
	 *    (4)   (5)
	 *   /
	 *  👾
	 */
	// alien asks instance-4 for the lock, but beaver still holds it
	{
		req := lock.AcquireRequest{
			Holder: alien,
		}
		resp, err := quorum[3].in.Acquire(context.Background(), &req)
		if err != nil {
			t.Fatalf("expected `%v`, got `%v`", nil, err)
		}
		if resp.Acquired {
			t.Fatalf("expected `%v`, got `%v`", false, resp.Acquired)
		}
		if resp.Holder != beaver {
			t.Fatalf("expected `%v`, got `%v`", beaver, resp.Holder)
		}
		// check quorum state
		good := 0
		for _, mi := range quorum {
			if mi.in.promised == quorum[3].in.promised &&
				mi.in.id == quorum[3].in.promised &&
				mi.in.holder == beaver {
				good++
			}
		}
		if good < 3 {
			t.Fatal("majority in bad state")
		}
	}

	// beaver tells instance-5 to release the lock
	{
		req := lock.ReleaseRequest{}
		resp, err := quorum[4].in.Release(context.Background(), &req)
		if err != nil {
			t.Fatalf("expected `%v`, got `%v`", nil, err)
		}
		if !resp.Released {
			t.Fatalf("expected `%v`, got `%v`", true, resp.Released)
		}
		// check quorum state
		good := 0
		for _, mi := range quorum {
			if mi.in.promised == quorum[4].in.promised &&
				mi.in.id == quorum[4].in.id &&
				mi.in.holder == "" {
				good++
			}
		}
		if good < 3 {
			t.Fatal("majority in bad state")
		}
	}

	// alien asks instance-4 for the lock
	{
		req := lock.AcquireRequest{
			Holder: alien,
		}
		resp, err := quorum[3].in.Acquire(context.Background(), &req)
		if err != nil {
			t.Fatalf("expected `%v`, got `%v`", nil, err)
		}
		if !resp.Acquired {
			t.Fatalf("expected `%v`, got `%v`", true, resp.Acquired)
		}
		if resp.Holder != req.Holder {
			t.Fatalf("expected `%v`, got `%v`", req.Holder, resp.Holder)
		}
		// check quorum state
		good := 0
		for _, mi := range quorum {
			if mi.in.promised == quorum[3].in.promised &&
				mi.in.id == quorum[3].in.id &&
				mi.in.holder == req.Holder {
				good++
			}
		}
		if good < 3 {
			t.Fatal("majority in bad state")
		}
	}
}
