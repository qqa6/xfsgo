// Copyright 2018 The xfsgo Authors
// This file is part of the xfsgo library.
//
// The xfsgo library is free software: you can redistribute it and/or modify
// it under the terms of the MIT Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The xfsgo library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// MIT Lesser General Public License for more details.
//
// You should have received a copy of the MIT Lesser General Public License
// along with the xfsgo library. If not, see <https://mit-license.org/>.

package p2p

import (
	"github.com/sirupsen/logrus"
	"strings"
	"testing"
	"xfsgo/crypto"
	"xfsgo/p2p/discover"
)

var XQ = "xfsnode://4b1d7796c65e05f2c93c3d6cb0011251e74530f413cf58cc9e546926d189601f6e74a77704cf43e878b5b8d624d87f79c4197f69fe59cb079e8595cd69c3a53f@127.0.0.1:9092"

func TestServer_Start(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	key, _ := crypto.GenPrvKey()
	s := NewServer(Config{
		ProtocolVersion: version1,
		ListenAddr: "127.0.0.1:9092",
		Key: key,
		Discover: true,
		NodeDBPath: "./d1",
		MaxPeers: 0,
	})
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	select {}
}

func TestServer_Start2(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	bootAddress := parseBootAddress(XQ)
	key, _ := crypto.GenPrvKey()
	s := NewServer(Config{
		ProtocolVersion: version1,
		ListenAddr:      "127.0.0.1:9093",
		Key:             key,
		Discover: true,
		BootstrapNodes: bootAddress,
		NodeDBPath: "./d2",
		MaxPeers: 10,
	})
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	select {}
}
func TestServer_Start3(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	bootAddress := parseBootAddress(XQ)
	key, _ := crypto.GenPrvKey()
	s := NewServer(Config{
		ProtocolVersion: version1,
		ListenAddr:      "127.0.0.1:9094",
		Key:             key,
		Discover: true,
		BootstrapNodes: bootAddress,
		NodeDBPath: "./d3",
		MaxPeers: 10,
	})
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	select {}
}
func TestServer_Start4(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	bootAddress := parseBootAddress(XQ)
	key, _ := crypto.GenPrvKey()
	s := NewServer(Config{
		ProtocolVersion: version1,
		ListenAddr:      "127.0.0.1:9095",
		Key:             key,
		Discover: true,
		BootstrapNodes: bootAddress,
		NodeDBPath: "./d4",
		MaxPeers: 10,
	})
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	select {}
}
func parseBootAddress(addrs string) []*discover.Node {
	if addrs == "" {
		return nil
	}
	arr := strings.Split(addrs, ",")
	addrArr := make([]*discover.Node, 0)
	for _, addr := range arr {
		a, err := discover.ParseNode(addr)
		if err != nil {
			continue
		}
		addrArr = append(addrArr, a)
	}
	return addrArr
}

func TestABC(t *testing.T) {
	a := flagOutbound|flagStatic
	i := a & flagInbound != 0
	t.Logf("abc: %v", i)
}