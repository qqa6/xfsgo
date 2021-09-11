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

package sub

import (
	"os"
	"path"
	"testing"
	"xfsgo/assert"
)

var testConfigText = `
rpclient:
  apihost: "http://127.0.0.1:9001"
  timeout: "30s"

rpcserver:
  listen: "127.0.0.1:9091"

p2pnode:
  listen: "0.0.0.0:9002"
  bootstrap: ["192.168.2.6:9002"]

protocol:
  version: 1
  networkid: 1

miner:
  coinbase: ""
  numworkers: 10

storage:
  datadir: "./d0"
  chaindir: ""
  statedir: ""
  keysdir: ""
  extradir: ""
  nodesdir: ""

logger:
  level: "INFO"
`
func Test_resetDataDir(t *testing.T) {
	home := os.Getenv("HOME")
	sp := storageParams{}
	joinPath := func(dir string, file string) string {
		if dir == "" {
			dir = path.Join(
				home, defaultStorageDir)
		}
		return path.Join(dir, file)
	}
	sp.dataDir = path.Join(
		home, defaultStorageDir)
	setupDataDir(&sp, sp.dataDir)
	assert.Equal(t, sp.chainDir, joinPath(path.Join(
		home, defaultStorageDir), defaultChainDir))
	assert.Equal(t, sp.stateDir, joinPath(path.Join(
		home, defaultStorageDir), defaultStateDir))
	assert.Equal(t, sp.keysDir, joinPath(path.Join(
		home, defaultStorageDir), defaultKeysDir))
	assert.Equal(t, sp.extraDir, joinPath(path.Join(
		home, defaultStorageDir), defaultExtraDir))
	assert.Equal(t, sp.nodesDir, joinPath(path.Join(
		home, defaultStorageDir), defaultNodesDir))
	paths := [5]string{"/a/b", "/c/d", "/e/f", "/g/h", "/i/j"}
	setupDataDir(&sp, paths[0])
	assert.Equal(t, sp.chainDir, joinPath(paths[0], defaultChainDir))
	assert.Equal(t, sp.stateDir, joinPath(paths[0], defaultStateDir))
	assert.Equal(t, sp.keysDir, joinPath(paths[0], defaultKeysDir))
	assert.Equal(t, sp.extraDir, joinPath(paths[0], defaultExtraDir))
	assert.Equal(t, sp.nodesDir, joinPath(paths[0], defaultNodesDir))
	sp.chainDir = paths[0]; sp.stateDir = paths[1]; sp.keysDir = paths[2]; sp.extraDir = paths[3]
	sp.nodesDir = paths[4]
	setupDataDir(&sp, paths[0])
	assert.Equal(t, sp.chainDir, paths[0]); assert.Equal(t, sp.stateDir, paths[1])
	assert.Equal(t, sp.keysDir, paths[2]); assert.Equal(t, sp.extraDir, paths[3])
	assert.Equal(t, sp.nodesDir, paths[4])
	sp.chainDir = ""
	setupDataDir(&sp, paths[0])
	assert.Equal(t, sp.chainDir, joinPath(paths[0], defaultChainDir))
}
