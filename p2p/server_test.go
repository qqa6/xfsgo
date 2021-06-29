package p2p

import (
	"testing"
)


func TestServer_Start(t *testing.T) {
	s := &Server{
		ListenAddr: "0.0.0.0:9001",
	}
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	select {}
}

func TestServer_Start2(t *testing.T) {
	s := &Server{
		ListenAddr: "0.0.0.0:9002",
		BootstrapNodes: []string{
			"127.0.0.1:9001",
		},
	}
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	select {}
}