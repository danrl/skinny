package skinny

import (
	"context"
	"testing"
	"time"

	"github.com/danrl/skinny/proto/control"
)

func TestInstanceStatusRPC(t *testing.T) {
	in := &Instance{
		name:      "foo",
		increment: 3,
		timeout:   time.Second,
		promised:  100,
		id:        23,
		holder:    "alien",
		peers: []*peer{
			{
				name: "peer-1",
			},
			{
				name: "peer-2",
			},
		},
	}

	resp, err := in.Status(context.Background(), &control.StatusRequest{})
	if err != nil {
		t.Fatalf("expected `%v`, got `%v`", nil, err)
	}

	if resp.Name != "foo" {
		t.Fatalf("expected `%v`, got `%v`", "foo", resp.Name)
	}
	if resp.Increment != in.increment {
		t.Errorf("expected `%v`, got `%v`", in.increment, resp.Increment)
	}
	if resp.Timeout != "1s" {
		t.Errorf("expected `%v`, got `%v`", time.Second, resp.Timeout)
	}
	if resp.Promised != in.promised {
		t.Errorf("expected `%v`, got `%v`", in.promised, resp.Promised)
	}
	if resp.ID != in.id {
		t.Errorf("expected `%v`, got `%v`", in.id, resp.ID)
	}
	if resp.Holder != in.holder {
		t.Errorf("expected `%v`, got `%v`", in.holder, resp.Holder)
	}
	if len(resp.Peers) != len(in.peers) {
		t.Errorf("expected `%v` peers, got `%v`", len(in.peers), len(resp.Peers))
	}
	if resp.Peers[0].Name != in.peers[0].name {
		t.Errorf("expected `%v`, got `%v`", in.peers[0].name, resp.Peers[0].Name)
	}
	if resp.Peers[1].Name != in.peers[1].name {
		t.Errorf("expected `%v`, got `%v`", in.peers[1].name, resp.Peers[1].Name)
	}
}
