package util

import (
	"testing"
	"xblockchain/backend"
	"xblockchain/node"
)

func TestStartNodeAndBackend(t *testing.T) {
	var err error = nil
	var stack *node.Node = nil
	var back *backend.Backend = nil
	if stack, err = node.New(); err != nil {
		t.Fatal(err)
	}
	if back, err = backend.NewBackend(stack); err != nil {
		t.Fatal(err)
	}
	if err = StartNodeAndBackend(stack, back); err != nil {
		t.Fatal(err)
	}
	select {}
}

func TestStartNodeAndBackend2(t *testing.T) {
	var err error = nil
	var stack *node.Node = nil
	var back *backend.Backend = nil
	if stack, err = node.New(); err != nil {
		t.Fatal(err)
	}
	if back, err = backend.NewBackend(stack); err != nil {
		t.Fatal(err)
	}
	if err = StartNodeAndBackend(stack, back); err != nil {
		t.Fatal(err)
	}
	select {}
}

func TestStartNodeAndBackend3(t *testing.T) {
	var err error = nil
	var stack *node.Node = nil
	var back *backend.Backend = nil
	if stack, err = node.New(); err != nil {
		t.Fatal(err)
	}
	if back, err = backend.NewBackend(stack); err != nil {
		t.Fatal(err)
	}
	if err = StartNodeAndBackend(stack, back); err != nil {
		t.Fatal(err)
	}
	select {}
}
