package skinny

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/danrl/skinny/proto/consensus"
	"github.com/danrl/skinny/proto/lock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

/* --- begin: test helper: mock instance ---------------------------------------------------------------------------- */

var ErrFailedRequest = errors.New("mock instances failed on purpose")

type mockInstance struct {
	t        *testing.T
	latency  time.Duration
	fail     bool
	listener *bufconn.Listener
	server   *grpc.Server
	in       *Instance
	conn     *grpc.ClientConn
}

func newMockInstance(t *testing.T, name string, increment uint64, timeout time.Duration) *mockInstance {
	var err error

	mi := mockInstance{
		t: t,
		in: &Instance{
			name:      name,
			increment: increment,
			timeout:   timeout,
		},
	}

	// listener
	mi.listener = bufconn.Listen(8 * 1024 * 1024)

	// client connection
	mi.conn, err = grpc.Dial("bufconn", grpc.WithDialer(mi.dialer), grpc.WithUnaryInterceptor(
		func(
			ctx context.Context,
			method string,
			req interface{},
			reply interface{},
			cc *grpc.ClientConn,
			invoker grpc.UnaryInvoker,
			opts ...grpc.CallOption,
		) error {
			if mi.fail {
				return ErrFailedRequest
			}
			time.Sleep(mi.latency)
			return invoker(ctx, method, req, reply, cc, opts...)
		}), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("setup connection: %v", err)
	}

	// server
	mi.server = grpc.NewServer()
	lock.RegisterLockServer(mi.server, mi.in)
	consensus.RegisterConsensusServer(mi.server, mi.in)
	go func() {
		if err := mi.server.Serve(mi.listener); err != nil {
			t.Logf("serve: %v", err)
		}
	}()

	t.Logf("mock instance `%v` created", name)
	return &mi
}

func (mi *mockInstance) dialer(string, time.Duration) (net.Conn, error) {
	return mi.listener.Dial()
}

func (mi *mockInstance) destroy() {
	mi.conn.Close()
	mi.server.Stop()
	mi.listener.Close()
	mi.t.Logf("mock instance `%v` destroyed", mi.in.name)
}

/* --- end: test helper: mock instance ------------------------------------------------------------------------------ */

func TestNew(t *testing.T) {
	in := New("foo", 3, time.Second)
	if in.name != "foo" {
		t.Errorf("expected name `%v`, got `%v`", "foo", in.name)
	}
	if in.increment != 3 {
		t.Errorf("expected increment `%v`, got `%v`", 3, in.increment)
	}
	if in.timeout != time.Second {
		t.Errorf("expected timeout `%v`, got `%v`", time.Second, in.timeout)
	}
}

func TestInstanceAddPeer(t *testing.T) {
	// fire up test instance
	leader := newMockInstance(t, "leader", 1, time.Second)
	defer leader.destroy()

	// fire up peer instance
	peer1 := newMockInstance(t, "peer-1", 2, time.Second)
	defer peer1.destroy()

	client := consensus.NewConsensusClient(peer1.conn)

	err := leader.in.AddPeer(peer1.in.name, client)
	if err != nil {
		t.Fatalf("expected `%v`, got `%v`", nil, err)
	}

	t.Run("add peer", func(t *testing.T) {
		if len(leader.in.peers) != 1 {
			t.Fatalf("expected `%v` peers, got `%v`", 1, len(leader.in.peers))
		}
		if leader.in.peers[0].name != peer1.in.name {
			t.Errorf("expected peer name `%v`, got `%v`", peer1.in.name, leader.in.peers[0].name)
		}
		if leader.in.peers[0].client != client {
			t.Errorf("expected peer client `%v`, got `%v`", client, leader.in.peers[0].client)
		}
	})

	t.Run("duplicate peer name", func(t *testing.T) {
		err := leader.in.AddPeer(peer1.in.name, nil)
		if err != ErrDuplicatePeer {
			t.Errorf("expected `%v`, got `%v`", ErrDuplicatePeer, err)
		}
	})

	t.Run("duplicate peer client", func(t *testing.T) {
		err := leader.in.AddPeer("totally-different", client)
		if err != ErrDuplicatePeer {
			t.Errorf("expected `%v`, got `%v`", ErrDuplicatePeer, err)
		}
	})
}

func TestInstanceIsMajority(t *testing.T) {
	t.Run("lonely instance", func(t *testing.T) {
		in := &Instance{}
		if got := in.isMajority(0); got  {
			t.Errorf("expected `%v`, got `%v`", false, got)
		}
		if got := in.isMajority(1); !got {
			t.Errorf("expected `%v`, got `%v`", true, got)
		}
	})

	t.Run("two instances", func(t *testing.T) {
		in := &Instance{}
		in.peers = append(in.peers, &peer{})
		if got := in.isMajority(0); got  {
			t.Errorf("expected `%v`, got `%v`", false, got)
		}
		if got := in.isMajority(1); got  {
			t.Errorf("expected `%v`, got `%v`", false, got)
		}
		if got := in.isMajority(2); !got {
			t.Errorf("expected `%v`, got `%v`", true, got)
		}
	})

	t.Run("quorum of odd instances", func(t *testing.T) {
		in := &Instance{}
		in.peers = append(in.peers, &peer{}, &peer{})
		if got := in.isMajority(0); got  {
			t.Errorf("expected `%v`, got `%v`", false, got)
		}
		if got := in.isMajority(1); got  {
			t.Errorf("expected `%v`, got `%v`", false, got)
		}
		if got := in.isMajority(2); !got {
			t.Errorf("expected `%v`, got `%v`", true, got)
		}
		if got := in.isMajority(3); !got {
			t.Errorf("expected `%v`, got `%v`", true, got)
		}
	})

	t.Run("quorum of even instances", func(t *testing.T) {
		in := &Instance{}
		in.peers = append(in.peers, &peer{}, &peer{}, &peer{})
		if got := in.isMajority(0); got  {
			t.Errorf("expected `%v`, got `%v`", false, got)
		}
		if got := in.isMajority(1); got  {
			t.Errorf("expected `%v`, got `%v`", false, got)
		}
		if got := in.isMajority(2); got  {
			t.Errorf("expected `%v`, got `%v`", false, got)
		}
		if got := in.isMajority(3); !got {
			t.Errorf("expected `%v`, got `%v`", true, got)
		}
		if got := in.isMajority(4); !got {
			t.Errorf("expected `%v`, got `%v`", true, got)
		}
	})
}
