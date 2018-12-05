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
			&peer{
				name: "peer-1",
			},
			&peer{
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
	if resp.Increment != 3 {
		t.Errorf("expected `%v`, got `%v`", 3, resp.Increment)
	}
	if resp.Timeout != "1s" {
		t.Errorf("expected `%v`, got `%v`", time.Second, resp.Timeout)
	}
	if resp.Promised != 100 {
		t.Errorf("expected `%v`, got `%v`", 100, resp.Promised)
	}
	if resp.ID != 23 {
		t.Errorf("expected `%v`, got `%v`", 23, resp.ID)
	}
	if resp.Holder != "alien" {
		t.Errorf("expected `%v`, got `%v`", "alien", resp.Holder)
	}
	if len(resp.Peers) != 2 {
		t.Errorf("expected `%v` peers, got `%v`", 1, len(resp.Peers))
	}
	if resp.Peers[0].Name != "peer-1" {
		t.Errorf("expected `%v`, got `%v`", "peer-1", resp.Peers[0].Name)
	}
	if resp.Peers[1].Name != "peer-2" {
		t.Errorf("expected `%v`, got `%v`", "peer-2", resp.Peers[1].Name)
	}
}
