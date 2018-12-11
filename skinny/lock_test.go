package skinny

import (
	"context"
	"testing"
	"time"

	"github.com/danrl/skinny/proto/consensus"

	"github.com/danrl/skinny/proto/lock"
)

func TestInstanceAcquireRPC(t *testing.T) {
	t.Run("lock available", func(t *testing.T) {
		in := &Instance{}

		resp, err := in.Acquire(context.Background(), &lock.AcquireRequest{
			Holder: "alien",
		})
		if err != nil {
			t.Fatalf("expected `%v`, got `%v`", nil, err)
		}
		if !resp.Acquired {
			t.Errorf("expected `%v`, got `%v`", true, resp.Acquired)
		}
		if resp.Holder != "alien" {
			t.Errorf("expected `%v`, got `%v`", "alien", resp.Holder)
		}
	})

	t.Run("lock already taken", func(t *testing.T) {
		in := &Instance{
			promised: 23,
			id:       23,
			holder:   "beaver",
		}

		resp, err := in.Acquire(context.Background(), &lock.AcquireRequest{
			Holder: "alien",
		})
		if err != nil {
			t.Fatalf("expected `%v`, got `%v`", nil, err)
		}
		if resp.Acquired {
			t.Errorf("expected `%v`, got `%v`", false, resp.Acquired)
		}
		if resp.Holder != "beaver" {
			t.Errorf("expected `%v`, got `%v`", "beaver", resp.Holder)
		}
	})

	t.Run("with retry", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping test in short mode.")
		}

		leader := newMockInstance(t, "leader", 1, time.Second)
		defer leader.destroy()

		peer1 := newMockInstance(t, "peer-1", 2, time.Second)
		defer peer1.destroy()
		peer1.latency = 2 * time.Second
		err := leader.in.AddPeer(peer1.in.name, consensus.NewConsensusClient(peer1.conn))
		if err != nil {
			t.Fatalf("add peer: %v", err)
		}

		peer2 := newMockInstance(t, "peer-2", 3, time.Second)
		defer peer2.destroy()
		peer2.latency = 2 * time.Second
		err = leader.in.AddPeer(peer2.in.name, consensus.NewConsensusClient(peer2.conn))
		if err != nil {
			t.Fatalf("add peer: %v", err)
		}

		resp, err := leader.in.Acquire(context.Background(), &lock.AcquireRequest{
			Holder: "alien",
		})
		if err != nil {
			t.Fatalf("expected `%v`, got `%v`", nil, err)
		}
		if resp.Acquired {
			t.Errorf("expected `%v`, got `%v`", false, resp.Acquired)
		}
		if resp.Holder != "" {
			t.Errorf("expected `%v`, got `%v`", "", resp.Holder)
		}
	})
}

func TestInstanceReleaseRPC(t *testing.T) {
	t.Run("lock not taken", func(t *testing.T) {
		in := &Instance{}

		resp, err := in.Release(context.Background(), &lock.ReleaseRequest{})
		if err != nil {
			t.Fatalf("expected `%v`, got `%v`", nil, err)
		}
		if !resp.Released {
			t.Errorf("expected `%v`, got `%v`", true, resp.Released)
		}
	})

	t.Run("lock taken", func(t *testing.T) {
		in := &Instance{
			promised: 23,
			id:       23,
			holder:   "beaver",
		}

		resp, err := in.Release(context.Background(), &lock.ReleaseRequest{})
		if err != nil {
			t.Fatalf("expected `%v`, got `%v`", nil, err)
		}
		if !resp.Released {
			t.Errorf("expected `%v`, got `%v`", true, resp.Released)
		}
	})

	t.Run("with retry", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping test in short mode.")
		}

		leader := newMockInstance(t, "leader", 1, time.Second)
		defer leader.destroy()

		peer1 := newMockInstance(t, "peer-1", 2, time.Second)
		defer peer1.destroy()
		peer1.latency = 2 * time.Second
		err := leader.in.AddPeer(peer1.in.name, consensus.NewConsensusClient(peer1.conn))
		if err != nil {
			t.Fatalf("add peer: %v", err)
		}

		peer2 := newMockInstance(t, "peer-2", 3, time.Second)
		defer peer2.destroy()
		peer2.latency = 2 * time.Second
		err = leader.in.AddPeer(peer2.in.name, consensus.NewConsensusClient(peer2.conn))
		if err != nil {
			t.Fatalf("add peer: %v", err)
		}

		leader.in.holder = "beaver"

		resp, err := leader.in.Release(context.Background(), &lock.ReleaseRequest{})
		if err != nil {
			t.Fatalf("expected `%v`, got `%v`", nil, err)
		}
		if resp.Released {
			t.Errorf("expected `%v`, got `%v`", false, resp.Released)
		}
	})
}
