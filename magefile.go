// +build mage

package main

import (
	"os"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

var (
	Default = Build
	targets = []string{"skinnyd", "skinnyctl"}
	protos  = []string{"lock", "consensus", "control"}
)

// Install build dependencies.
func BuildDeps() error {
	err := sh.RunV("protoc", "--version")
	if err != nil {
		return err
	}
	err = sh.RunV("go", "get", "-u", "github.com/golang/protobuf/protoc-gen-go")
	if err != nil {
		return err
	}
	err = sh.RunV("go", "get", "-u", "google.golang.org/grpc")
	if err != nil {
		return err
	}

	return nil
}

// Install dependencies.
func Deps() error {
	err := sh.RunV("go", "mod", "vendor")
	if err != nil {
		return err
	}

	return nil
}

// Generate code.
func Generate() error {
	for _, name := range protos {
		err := sh.RunV("protoc", "--go_out=plugins=grpc:.", "./proto/"+name+"/"+name+".proto")
		if err != nil {
			return err
		}
	}

	return nil
}

// Run tests.
func Test() error {
	mg.Deps(Generate)

	return sh.RunV("go", "test", "-v", "./...")
}

// Build binary executables.
func Build() error {
	mg.Deps(Test)

	for _, name := range targets {
		err := sh.RunV("go", "build", "-v", "-o", "./bin/"+name, "./cmd/"+name)
		if err != nil {
			return err
		}
	}

	return nil
}

// Remove binary executables.
func Clean() error {
	return os.RemoveAll("bin")
}
