package config

import (
	"testing"
	"time"
)

func TestNewInstanceConfig(t *testing.T) {
	t.Run("invalid filename", func(t *testing.T) {
		_, err := NewInstanceConfig("testdata/nonexistent.yml")
		if err == nil {
			t.Errorf("expected error, got `%v`", err)
		}
	})

	t.Run("invalid format", func(t *testing.T) {
		_, err := NewInstanceConfig("testdata/instance/bad-format.yml")
		if err == nil {
			t.Errorf("expected error, got `%v`", nil)
		}
	})

	t.Run("invalid timeout", func(t *testing.T) {
		_, err := NewInstanceConfig("testdata/instance/bad-timeout.yml")
		if err != ErrInvalidTimeout {
			t.Errorf("expected error, got `%v`", err)
		}
	})

	t.Run("duplicate peer", func(t *testing.T) {
		_, err := NewInstanceConfig("testdata/instance/duplicate-peer.yml")
		if err != ErrDuplicateInstance {
			t.Errorf("expected error, got `%v`", err)
		}
	})

	t.Run("self in peers list", func(t *testing.T) {
		_, err := NewInstanceConfig("testdata/instance/duplicate-self.yml")
		if err != ErrDuplicateInstance {
			t.Errorf("expected error, got `%v`", err)
		}
	})

	t.Run("invalid increment", func(t *testing.T) {
		_, err := NewInstanceConfig("testdata/instance/bad-increment.yml")
		if err != ErrInvalidIncrement {
			t.Errorf("expected error, got `%v`", err)
		}
	})

	t.Run("valid configuration", func(t *testing.T) {
		cfg, err := NewInstanceConfig("testdata/instance/good.yml")
		if err != nil {
			t.Fatalf("expected `nil`, got `%v`", err)
		}
		if cfg.Name != "london" {
			t.Errorf("expected name `london`, got `%v`", cfg.Name)
		}
		if cfg.Increment != 1 {
			t.Errorf("expected increment `1`, got `%v`", cfg.Increment)
		}
		if cfg.Timeout != 500*time.Millisecond {
			t.Errorf("expected timeout `500ms`, got `%v`", cfg.Timeout)
		}
		if cfg.Listen != "0.0.0.0:9000" {
			t.Errorf("expected listen `0.0.0.0:9000`, got `%v`", cfg.Listen)
		}
		if len(cfg.Peers) != 4 {
			t.Errorf("expected %v peers, got %v", 5, len(cfg.Peers))
		}
		if cfg.Peers[0].Name != "oregon" {
			t.Errorf("expected name `oregon, got `%v`", cfg.Peers[0].Name)
		}
		if cfg.Peers[0].Address != "oregon.skinny.cakelie.net:9000" {
			t.Errorf("expected address `oregon.skinny.cakelie.net:9000`, got `%v`", cfg.Peers[0].Address)
		}
	})
}

func TestNewQuorumConfig(t *testing.T) {
	t.Run("invalid filename", func(t *testing.T) {
		_, err := NewQuorumConfig("testdata/nonexistent.yml")
		if err == nil {
			t.Errorf("expected error, got `%v`", err)
		}
	})

	t.Run("invalid format", func(t *testing.T) {
		_, err := NewQuorumConfig("testdata/quorum/bad-format.yml")
		if err == nil {
			t.Errorf("expected error, got `%v`", nil)
		}
	})

	t.Run("invalid timeout", func(t *testing.T) {
		_, err := NewQuorumConfig("testdata/quorum/bad-timeout.yml")
		if err != ErrInvalidTimeout {
			t.Errorf("expected error, got `%v`", err)
		}
	})

	t.Run("duplicate instances", func(t *testing.T) {
		_, err := NewQuorumConfig("testdata/quorum/duplicates.yml")
		if err != ErrDuplicateInstance {
			t.Errorf("expected error, got `%v`", err)
		}
	})

	t.Run("valid configuration", func(t *testing.T) {
		cfg, err := NewQuorumConfig("testdata/quorum/good.yml")
		if err != nil {
			t.Fatalf("expected `nil`, got `%v`", err)
		}
		if cfg.Timeout != 5*time.Second {
			t.Errorf("expected timeout `5s`, got `%v`", cfg.Timeout)
		}
		if len(cfg.Instances) != 5 {
			t.Errorf("expected %v instances, got %v", 5, len(cfg.Instances))
		}
		if cfg.Instances[0].Name != "london" {
			t.Errorf("expected name `london, got `%v`", cfg.Instances[0].Name)
		}
		if cfg.Instances[0].Address != "london.skinny.cakelie.net:9000" {
			t.Errorf("expected address `london.skinny.cakelie.net:9000`, got `%v`", cfg.Instances[0].Address)
		}
	})
}

func TestCheckTimeout(t *testing.T) {
	t.Run("zero timeout", func(t *testing.T) {
		expected := ErrInvalidTimeout
		got := checkTimeout(time.Duration(0))
		if got != expected {
			t.Errorf("expected `%v`, got `%v`", expected, got)
		}
	})

	t.Run("good timeout", func(t *testing.T) {
		var expected error
		got := checkTimeout(100 * time.Millisecond)
		if got != expected {
			t.Errorf("expected `%v`, got `%v`", expected, got)
		}
	})
}

func TestCheckInstanceList(t *testing.T) {
	instanceEmptyName := Instance{
		Name:    "",
		Address: "localhost:9000",
	}
	instanceEmptyAddress := Instance{
		Name:    "manhattan",
		Address: "",
	}
	instanceNorth := Instance{
		Name:    "north",
		Address: "north.example.com:9000",
	}
	instanceNorth2 := Instance{
		Name:    "north",
		Address: "north-2.example.com:9000",
	}
	instanceSouth := Instance{
		Name:    "south",
		Address: "south.example.com:9000",
	}
	instanceSouth2 := Instance{
		Name:    "south-2",
		Address: "south.example.com:9000",
	}

	t.Run("empty list", func(t *testing.T) {
		expected := ErrNoInstance
		got := checkInstanceList()
		if got != expected {
			t.Errorf("expected `%v`, got `%v`", expected, got)
		}
	})

	t.Run("empty name", func(t *testing.T) {
		expected := ErrInvalidInstanceDefinition
		got := checkInstanceList(instanceEmptyName)
		if got != expected {
			t.Errorf("expected `%v`, got `%v`", expected, got)
		}
	})

	t.Run("empty address", func(t *testing.T) {
		expected := ErrInvalidInstanceDefinition
		got := checkInstanceList(instanceEmptyAddress)
		if got != expected {
			t.Errorf("expected `%v`, got `%v`", expected, got)
		}
	})

	t.Run("single instance list", func(t *testing.T) {
		var expected error
		got := checkInstanceList(instanceNorth)
		if got != expected {
			t.Errorf("expected `%v`, got `%v`", expected, got)
		}
	})

	t.Run("duplicate name", func(t *testing.T) {
		expected := ErrDuplicateInstance
		got := checkInstanceList(instanceNorth, instanceNorth2)
		if got != expected {
			t.Errorf("expected `%v`, got `%v`", expected, got)
		}
	})

	t.Run("duplicate address", func(t *testing.T) {
		expected := ErrDuplicateInstance
		got := checkInstanceList(instanceSouth, instanceSouth2)
		if got != expected {
			t.Errorf("expected `%v`, got `%v`", expected, got)
		}
	})

	t.Run("good list", func(t *testing.T) {
		var expected error
		got := checkInstanceList(instanceNorth, instanceSouth)
		if got != expected {
			t.Errorf("expected `%v`, got `%v`", expected, got)
		}
	})
}
