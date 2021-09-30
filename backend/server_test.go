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

package backend

import (
	"fmt"
	"path"
	"testing"
	"time"
	"xfsgo"
	"xfsgo/common"
	"xfsgo/node"
	"xfsgo/storage/badger"
)

var (
	datadir          = "./d3"
	P2PListenAddress = ":9002"
	rpcaddr          = ":9001"
	clientaddr       = "http://127.0.0.1" + rpcaddr
)
var (
	datadir2          = "./d4"
	P2PListenAddress2 = ":9004"
	rpcaddr2          = ":9003"
	clientaddr2       = "http://127.0.0.1" + rpcaddr
)
var (
	netid                  = uint32(1)
	defaultProtocolVersion = uint32(1)
	//coinbasePrivateKey = "MHcCAQEEIONs7L6Y8K552y_EAJokBTQxA3ejz162pdhZzUaO9bdKoAoGCCqGSM49AwEHoUQDQgAE-KtA6lYlnWAYShJJ2aQnScLz79Iv-vmlVjq8bvxfj9VvOKIPQop87jXyQ01QfTbgZorGEserwb2hDwQAp_xzRA"
	defaultCoinbase = "1A2QiH4FYc9c4nsNjCMxygg9HKTK9EJWX5"
)

func safeclose(t *testing.T, fn func() error) {
	if err := fn(); err != nil {
		t.Fatal(err)
	}
}
func TestStartNodeAndBackend(t *testing.T) {
	var err error = nil
	var stack *node.Node = nil
	var back *Backend = nil
	// boots := parseBootNodes(bootnodes)
	if stack, err = node.New(&node.Config{
		P2PListenAddress: P2PListenAddress,
		ProtocolVersion:  uint8(1),
		RPCConfig: &xfsgo.RPCConfig{
			ListenAddr: rpcaddr,
		},
	}); err != nil {
		t.Fatal(err)
	}
	datadir = path.Join(path.Dir(datadir), datadir)
	chainDb := badger.New(path.Join(datadir, "chain"))
	keysDb := badger.New(path.Join(datadir, "keys"))
	stateDB := badger.New(path.Join(datadir, "state"))
	extraDB := badger.New(path.Join(datadir, "extraDB"))
	defer func() {
		safeclose(t, chainDb.Close)
		safeclose(t, keysDb.Close)
		safeclose(t, stateDB.Close)
		safeclose(t, extraDB.Close)
	}()
	if back, err = NewBackend(stack, &Config{
		Params: &Params{
			NetworkID:       netid,
			ProtocolVersion: defaultProtocolVersion,
			Coinbase:        common.StrB58ToAddress(defaultCoinbase),
		},
		ChainDB: chainDb,
		KeysDB:  keysDb,
		StateDB: stateDB,
		ExtraDB: extraDB,
	}); err != nil {
		t.Fatal(err)
	}
	if err = StartNodeAndBackend(stack, back); err != nil {
		t.Fatal(err)
	}
	select {}
}

func TestStartNodeAndBackendClient(t *testing.T) {
	clientConn := xfsgo.NewClient(clientaddr)
	timeout := time.After(time.Second * 10)
	finish := make(chan bool)
	count := 1
	var res *string = nil

	if err := clientConn.CallMethod(1, "Miner.Start", nil, &res); err != nil {
		t.Errorf("miner start %v\n", err.Error())
		return
	}
	fmt.Println("miner running...")
	go func() {
		for {
			select {
			case <-timeout:
				err := clientConn.CallMethod(1, "Miner.Stop", nil, &res)
				if err != nil {
					finish <- true
					return
				} else {
					t.Errorf("miner stop %v\n", err.Error())
					finish <- true
					return
				}
			default:
				fmt.Printf(" s:%d\n", count)
				count++
			}
			time.Sleep(time.Second * 1)
		}
	}()
	<-finish
	fmt.Println("miner stop...")
}

var XQ = "0x679c3afb494f2bcba6e62505580d831ad460bcfad2c5667e8371f3166c691648"

func TestStartNodeAndBackend2(t *testing.T) {
	var err error = nil
	var stack *node.Node = nil
	var back *Backend = nil

	if stack, err = node.New(&node.Config{
		P2PListenAddress: P2PListenAddress2,
		ProtocolVersion:  uint8(1),
		StaticStr:        "127.0.0.1" + P2PListenAddress + "/" + XQ,
		RPCConfig: &xfsgo.RPCConfig{
			ListenAddr: rpcaddr2,
		},
	}); err != nil {
		t.Fatal(err)
	}
	datadir = path.Join(path.Dir(datadir2), datadir2)
	chainDb := badger.New(path.Join(datadir2, "chain"))
	keysDb := badger.New(path.Join(datadir2, "keys"))
	stateDB := badger.New(path.Join(datadir2, "state"))
	extraDB := badger.New(path.Join(datadir2, "extraDB"))
	defer func() {
		safeclose(t, chainDb.Close)
		safeclose(t, keysDb.Close)
		safeclose(t, stateDB.Close)
		safeclose(t, extraDB.Close)
	}()
	if back, err = NewBackend(stack, &Config{
		Params: &Params{
			NetworkID:       netid,
			ProtocolVersion: defaultProtocolVersion,
			Coinbase:        common.StrB58ToAddress(defaultCoinbase),
		},
		ChainDB: chainDb,
		KeysDB:  keysDb,
		StateDB: stateDB,
		ExtraDB: extraDB,
	}); err != nil {
		t.Fatal(err)
	}
	if err = StartNodeAndBackend(stack, back); err != nil {
		t.Fatal(err)
	}
	select {}
}

func TestGetHeadClient(t *testing.T) {
	clientConn := xfsgo.NewClient(clientaddr2)
	timeout := time.After(time.Second * 10)
	finish := make(chan bool)
	count := 1
	block := make(map[string]interface{}, 1)
	go func() {
		for {
			select {
			case <-timeout: // //Wait 10s to get the head
				err := clientConn.CallMethod(1, "Chain.Head", nil, &block)
				if err != nil {
					height := block["header"].(map[string]interface{})["height"].(uint64)
					fmt.Printf("Get the synchronized block height %v\n", height)
					finish <- true
					return
				}
				t.Errorf("chain head err:%v\n", err.Error())
				finish <- true
				return
			default:
				fmt.Printf(" s:%d\n", count)
				count++
			}
			time.Sleep(time.Second * 1)
		}
	}()
	<-finish
}
