package skinny

import (
	"context"
	"testing"
	"time"

	"github.com/danrl/skinny/proto/consensus"
)

const (
	beaver = "beaver"
	alien  = "alien"
)

func TestInstancePromiseRPC(t *testing.T) {
	t.Run("simple promise", func(t *testing.T) {
		var in Instance

		resp, err := in.Promise(context.Background(), &consensus.PromiseRequest{
			ID: 1,
		})
		if err != nil {
			t.Fatalf("expected `%v`, got `%v`", nil, err)
		}
		if !resp.Promised {
			t.Errorf("expected `%v`, got `%v`", true, resp.Promised)
		}
		if resp.ID != 0 {
			t.Errorf("expected `%v`, got `%v`", 0, resp.ID)
		}
		if resp.Holder != "" {
			t.Errorf("expected `%v`, got `%v`", "", resp.Holder)
		}
	})

	t.Run("simple promise refusal", func(t *testing.T) {
		in := Instance{
			promised: 23,
			id:       1,
			holder:   beaver,
		}

		resp, err := in.Promise(context.Background(), &consensus.PromiseRequest{
			ID: 5,
		})
		if err != nil {
			t.Fatalf("expected `%v`, got `%v`", nil, err)
		}
		if resp.Promised {
			t.Errorf("expected `%v`, got `%v`", false, resp.Promised)
		}
		if resp.ID != 1 {
			t.Errorf("expected `%v`, got `%v`", 0, resp.ID)
		}
		if resp.Holder != beaver {
			t.Errorf("expected `%v`, got `%v`", beaver, resp.Holder)
		}

		// instances must not have changed its internal state
		if in.promised != 23 {
			t.Errorf("expected `%v`, got `%v`", 23, in.promised)
		}
		if in.id != 1 {
			t.Errorf("expected `%v`, got `%v`", 1, in.id)
		}
		if in.holder != beaver {
			t.Errorf("expected `%v`, got `%v`", beaver, in.holder)
		}
	})
}

func TestInstanceCommitRPC(t *testing.T) {
	t.Run("simple commit", func(t *testing.T) {
		in := Instance{
			promised: 1,
		}

		resp, err := in.Commit(context.Background(), &consensus.CommitRequest{
			ID:     1,
			Holder: alien,
		})
		if err != nil {
			t.Fatalf("expected `%v`, got `%v`", nil, err)
		}
		if !resp.Committed {
			t.Errorf("expected `%v`, got `%v`", true, resp.Committed)
		}
	})

	t.Run("simple commit refusal", func(t *testing.T) {
		in := Instance{
			promised: 23,
			id:       5,
			holder:   beaver,
		}

		resp, err := in.Commit(context.Background(), &consensus.CommitRequest{
			ID:     2,
			Holder: "aloen",
		})
		if err != nil {
			t.Fatalf("expected `%v`, got `%v`", nil, err)
		}
		if resp.Committed {
			t.Errorf("expected `%v`, got `%v`", false, resp.Committed)
		}

		// instance must not have changed its internal state
		if in.promised != 23 {
			t.Errorf("expected `%v`, got `%v`", 23, in.promised)
		}
		if in.id != 5 {
			t.Errorf("expected `%v`, got `%v`", 5, in.id)
		}
		if in.holder != beaver {
			t.Errorf("expected `%v`, got `%v`", beaver, in.holder)
		}
	})
}

func TestInstancePropose(t *testing.T) {
	t.Run("successful propose", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping test in short mode.")
		}

		leader := newMockInstance(t, "leader", 1, time.Second)
		defer leader.destroy()

		peer1 := newMockInstance(t, "peer-1", 2, time.Second)
		defer peer1.destroy()
		err := leader.in.AddPeer(peer1.in.name, consensus.NewConsensusClient(peer1.conn))
		if err != nil {
			t.Fatalf("add peer: %v", err)
		}

		peer2 := newMockInstance(t, "peer-2", 3, time.Second)
		defer peer2.destroy()
		err = leader.in.AddPeer(peer2.in.name, consensus.NewConsensusClient(peer2.conn))
		if err != nil {
			t.Fatalf("add peer: %v", err)
		}

		got := leader.in.propose()
		if !got {
			t.Errorf("expected `%v`, got `%v`", true, got)
		}
	})

	t.Run("cancel requests", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping test in short mode.")
		}

		leader := newMockInstance(t, "leader", 1, 100*time.Millisecond)
		defer leader.destroy()

		peer1 := newMockInstance(t, "peer-1", 2, time.Second)
		defer peer1.destroy()
		err := leader.in.AddPeer(peer1.in.name, consensus.NewConsensusClient(peer1.conn))
		if err != nil {
			t.Fatalf("add peer: %v", err)
		}

		peer2 := newMockInstance(t, "peer-2", 3, time.Second)
		defer peer2.destroy()
		peer2.latency = 500 * time.Millisecond
		err = leader.in.AddPeer(peer2.in.name, consensus.NewConsensusClient(peer2.conn))
		if err != nil {
			t.Fatalf("add peer: %v", err)
		}

		got := leader.in.propose()
		if !got {
			t.Errorf("expected `%v`, got `%v`", true, got)
		}
	})

	t.Run("learn value", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping test in short mode.")
		}

		leader := newMockInstance(t, "leader", 1, time.Second)
		defer leader.destroy()

		peer1 := newMockInstance(t, "peer-1", 2, time.Second)
		defer peer1.destroy()
		err := leader.in.AddPeer(peer1.in.name, consensus.NewConsensusClient(peer1.conn))
		if err != nil {
			t.Fatalf("add peer: %v", err)
		}

		peer2 := newMockInstance(t, "peer-2", 3, time.Second)
		defer peer2.destroy()
		peer2.latency = 500 * time.Millisecond // to make sure we learn from peer1 before reaching a majority
		err = leader.in.AddPeer(peer2.in.name, consensus.NewConsensusClient(peer2.conn))
		if err != nil {
			t.Fatalf("add peer: %v", err)
		}

		peer1.in.promised = 23
		peer1.in.id = 23
		peer1.in.holder = beaver

		got := leader.in.propose()
		if !got {
			t.Errorf("expected `%v`, got `%v`", true, got)
		}

		// leader must have learned new value
		if leader.in.promised != 23 {
			t.Errorf("expected `%v`, got `%v`", 23, leader.in.promised)
		}
		if leader.in.id != 23 {
			t.Errorf("expected `%v`, got `%v`", 23, leader.in.id)
		}
		if leader.in.holder != beaver {
			t.Errorf("expected `%v`, got `%v`", beaver, leader.in.holder)
		}
	})
}

func TestInstanceCommit(t *testing.T) {
	t.Run("successful commit", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping test in short mode.")
		}

		leader := newMockInstance(t, "leader", 1, time.Second)
		defer leader.destroy()

		peer1 := newMockInstance(t, "peer-1", 2, time.Second)
		defer peer1.destroy()
		err := leader.in.AddPeer(peer1.in.name, consensus.NewConsensusClient(peer1.conn))
		if err != nil {
			t.Fatalf("add peer: %v", err)
		}

		peer2 := newMockInstance(t, "peer-2", 3, time.Second)
		defer peer2.destroy()
		err = leader.in.AddPeer(peer2.in.name, consensus.NewConsensusClient(peer2.conn))
		if err != nil {
			t.Fatalf("add peer: %v", err)
		}

		got := leader.in.commit(5, alien)
		if !got {
			t.Errorf("expected `%v`, got `%v`", true, got)
		}

		// leader must have committed new values to itself
		if leader.in.id != 5 {
			t.Errorf("expected `%v`, got `%v`", 5, leader.in.id)
		}
		if leader.in.holder != alien {
			t.Errorf("expected `%v`, got `%v`", alien, leader.in.holder)
		}
	})

	t.Run("cancel requests", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping test in short mode.")
		}

		leader := newMockInstance(t, "leader", 1, 100*time.Millisecond)
		defer leader.destroy()

		peer1 := newMockInstance(t, "peer-1", 2, time.Second)
		defer peer1.destroy()
		peer1.latency = 500 * time.Millisecond
		err := leader.in.AddPeer(peer1.in.name, consensus.NewConsensusClient(peer1.conn))
		if err != nil {
			t.Fatalf("add peer: %v", err)
		}

		peer2 := newMockInstance(t, "peer-2", 3, time.Second)
		defer peer2.destroy()
		peer2.latency = 500 * time.Millisecond
		err = leader.in.AddPeer(peer2.in.name, consensus.NewConsensusClient(peer2.conn))
		if err != nil {
			t.Fatalf("add peer: %v", err)
		}

		got := leader.in.commit(5, alien)
		if got {
			t.Errorf("expected `%v`, got `%v`", false, got)
		}

		// leader must have committed new values to itself
		if leader.in.id != 5 {
			t.Errorf("expected `%v`, got `%v`", 5, leader.in.id)
		}
		if leader.in.holder != alien {
			t.Errorf("expected `%v`, got `%v`", alien, leader.in.holder)
		}
	})

	t.Run("failing instance", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping test in short mode.")
		}

		leader := newMockInstance(t, "leader", 1, 100*time.Millisecond)
		defer leader.destroy()

		peer1 := newMockInstance(t, "peer-1", 2, time.Second)
		defer peer1.destroy()
		err := leader.in.AddPeer(peer1.in.name, consensus.NewConsensusClient(peer1.conn))
		if err != nil {
			t.Fatalf("add peer: %v", err)
		}

		peer2 := newMockInstance(t, "peer-2", 3, time.Second)
		defer peer2.destroy()
		peer2.fail = true
		err = leader.in.AddPeer(peer2.in.name, consensus.NewConsensusClient(peer2.conn))
		if err != nil {
			t.Fatalf("add peer: %v", err)
		}

		got := leader.in.commit(5, alien)
		if !got {
			t.Errorf("expected `%v`, got `%v`", true, got)
		}

		// leader must have committed new values to itself)
		if leader.in.id != 5 {
			t.Errorf("expected `%v`, got `%v`", 5, leader.in.id)
		}
		if leader.in.holder != alien {
			t.Errorf("expected `%v`, got `%v`", alien, leader.in.holder)
		}
	})
}
