package config

import (
	"errors"
	"io/ioutil"
	"time"

	yaml "gopkg.in/yaml.v2"
)

var (
	// ErrInvalidTimeout is returned when the timeout is not valid, e.g. zero or negative
	ErrInvalidTimeout = errors.New("invalid timeout")

	// ErrInvalidIncrement is returned when the increment is not valid, e.g. zero or negative
	ErrInvalidIncrement = errors.New("invalid increment")

	// ErrNoInstance is returned when an instance definition is expected but missing
	ErrNoInstance = errors.New("missing instance definition")

	// ErrInvalidInstanceDefinition is returned when an invalid instance definition is discovered, e.g. an empty name or
	// address
	ErrInvalidInstanceDefinition = errors.New("invalid instance definition")

	// ErrDuplicateInstance is returned when there are multiple definitions for the same instance
	ErrDuplicateInstance = errors.New("duplicate instance")
)

// Instance describes a single Skinny instance connection information
type Instance struct {
	Name    string `yaml:"name"`
	Address string `yaml:"address"`
}

// InstanceConfig describes a Skinny instance configuration
type InstanceConfig struct {
	Name      string        `yaml:"name"`
	Increment uint64        `yaml:"increment"`
	Timeout   time.Duration `yaml:"timeout"`
	Listen    string        `yaml:"listen"`
	Peers     []Instance    `yaml:"peers"`
}

// QuorumConfig describes a Skinny quorum configuration file
type QuorumConfig struct {
	Timeout   time.Duration `yaml:"timeout"`
	Instances []Instance    `yaml:"instances"`
}

// NewInstanceConfig loads a Skinny instance configuration from given file
func NewInstanceConfig(fname string) (*InstanceConfig, error) {
	cfg := &InstanceConfig{}
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		return nil, err
	}
	err = yaml.UnmarshalStrict(data, cfg)
	if err != nil {
		return nil, err
	}

	// sanity checks
	if cfg.Increment < 1 {
		return nil, ErrInvalidIncrement
	}
	if err := checkTimeout(cfg.Timeout); err != nil {
		return nil, err
	}
	instances := append(cfg.Peers, Instance{Name: cfg.Name, Address: cfg.Listen})
	if err := checkInstanceList(instances...); err != nil {
		return nil, err
	}

	return cfg, nil
}

// NewQuorumConfig loads a Skinny quorum configuration from given file
func NewQuorumConfig(fname string) (*QuorumConfig, error) {
	cfg := &QuorumConfig{}
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		return nil, err
	}
	err = yaml.UnmarshalStrict(data, cfg)
	if err != nil {
		return nil, err
	}

	// sanity checks
	if err := checkTimeout(cfg.Timeout); err != nil {
		return nil, err
	}
	if err := checkInstanceList(cfg.Instances...); err != nil {
		return nil, err
	}

	return cfg, nil
}

// checkTimeout performs a sanity check for a timeout value
func checkTimeout(timeout time.Duration) error {
	if timeout <= time.Duration(0) {
		return ErrInvalidTimeout
	}
	return nil
}

// checkInstanceList performs a sanity check for a list of instances
func checkInstanceList(instances ...Instance) error {
	if len(instances) == 0 {
		return ErrNoInstance
	}

	seenNames := make(map[string]bool)
	seenAddresses := make(map[string]bool)
	for _, in := range instances {
		// Instance name must be unique in the configuration file.
		if len(in.Name) == 0 {
			return ErrInvalidInstanceDefinition
		}
		if seenNames[in.Name] {
			return ErrDuplicateInstance
		}
		seenNames[in.Name] = true

		// Instance address must be unique in the configuration file.
		if len(in.Address) == 0 {
			return ErrInvalidInstanceDefinition
		}
		if seenAddresses[in.Address] {
			return ErrDuplicateInstance
		}
		seenAddresses[in.Address] = true
	}
	return nil
}
