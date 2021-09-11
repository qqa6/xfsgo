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
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"xfsgo"
	"xfsgo/backend"
	"xfsgo/common"
	"xfsgo/node"

	"github.com/spf13/viper"
)

const (
	defaultConfigFile        = "./config.yml"
	defaultStorageDir        = ".xfsgo"
	defaultChainDir          = "chain"
	defaultStateDir          = "state"
	defaultKeysDir           = "keys"
	defaultExtraDir          = "extra"
	defaultNodesDir          = "nodes"
	defaultRPCClientAPIHost  = "127.0.0.1:9002"
	defaultNodeRPCListenAddr = "127.0.0.1:9001"
	defaultNodeP2PListenAddr = "127.0.0.1:9002"
	defaultNetworkId         = uint32(1)
	defaultTestNetworkId     = uint32(2)
	defaultProtocolVersion   = uint32(1)
	defaultLoggerLevel   = "INFO"
)

type storageParams struct {
	dataDir  string
	chainDir string
	keysDir  string
	stateDir string
	extraDir string
	nodesDir string
}

type loggerParams struct {
	level string
}

type daemonConfig struct {
	loggerParams loggerParams
	storageParams storageParams
	nodeConfig    node.Config
	backendParams backend.Params
}

type clientConfig struct {
	rpcClientApiHost string
}

func readFromConfigPath(v *viper.Viper, customFile string) error {
	filename := filepath.Base(defaultConfigFile)
	ext := filepath.Ext(defaultConfigFile)
	configPath := filepath.Dir(defaultConfigFile)
	v.AddConfigPath("$HOME/.xfsgo")
	v.AddConfigPath("/etc/xfsgo")
	v.AddConfigPath(configPath)
	v.SetConfigType(strings.TrimPrefix(ext, "."))
	v.SetConfigName(strings.TrimSuffix(filename, ext))
	v.SetConfigFile(customFile)
	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("read config file err: %s",err)
	}
	return nil
}

func parseConfigLoggerParams(v *viper.Viper) loggerParams {
	params := loggerParams{}
	params.level = v.GetString("logger.level")
	if params.level == "" {
		params.level = defaultLoggerLevel
	}
	return params
}

func resetDataDir(params *storageParams) {
	if params.dataDir == "" {
		home := os.Getenv("HOME")
		params.dataDir = path.Join(
			home, defaultStorageDir)
	}
	if params.chainDir == "" {
		params.chainDir = path.Join(
			params.dataDir, defaultChainDir)
	}
	if params.stateDir == "" {
		params.stateDir = path.Join(
			params.dataDir, defaultStateDir)
	}
	if params.keysDir == "" {
		params.keysDir = path.Join(
			params.dataDir, defaultKeysDir)
	}
	if params.extraDir == "" {
		params.extraDir = path.Join(
			params.dataDir, defaultExtraDir)
	}
	if params.nodesDir == "" {
		params.nodesDir = path.Join(
			params.dataDir, defaultNodesDir)
	}
}

func parseConfigStorageParams(v *viper.Viper) storageParams {
	storageParams := storageParams{}
	storageParams.dataDir = v.GetString("storage.datadir")
	storageParams.chainDir = v.GetString("storage.chaindir")
	storageParams.stateDir = v.GetString("storage.statedir")
	storageParams.keysDir = v.GetString("storage.keysdir")
	storageParams.extraDir = v.GetString("storage.extradir")
	storageParams.nodesDir = v.GetString("storage.nodesdir")
	resetDataDir(&storageParams)
	return storageParams
}
func defaultBootstrapNodes(netid uint32) []string {
	// hardcoded bootstrap nodes
	if netid == 1 {
		// main net boot nodes
		return []string{}
	} else if netid == 2 {
		// test net boot nodes
		return []string{}
	}
	return make([]string, 0)
}
func parseConfigNodeParams(v *viper.Viper, netid uint32) node.Config {
	config := node.Config{
		RPCConfig: new(xfsgo.RPCConfig),
	}
	config.RPCConfig.ListenAddr = v.GetString("rpcserver.listen")
	config.P2PListenAddress = v.GetString("p2pnode.listen")
	config.P2PBootstraps = v.GetStringSlice("p2pnode.bootstrap")
	config.P2PStaticNodes = v.GetStringSlice("p2pnode.static")
	config.ProtocolVersion = uint8(v.GetUint64("protocol.version"))
	if config.RPCConfig.ListenAddr == "" {
		config.RPCConfig.ListenAddr = defaultNodeRPCListenAddr
	}
	if config.P2PListenAddress == "" {
		config.P2PListenAddress = defaultNodeP2PListenAddr
	}
	if config.P2PBootstraps == nil || len(config.P2PBootstraps) == 0 {
		config.P2PBootstraps = defaultBootstrapNodes(netid)
	}
	return config
}

func parseConfigBackendParams(v *viper.Viper) backend.Params {
	config := backend.Params{}
	mCoinbase := v.GetString("miner.coinbase")
	if mCoinbase != "" {
		config.Coinbase = common.StrB58ToAddress(mCoinbase)
	}
	config.ProtocolVersion = v.GetUint32("protocol.version")
	config.NetworkID = v.GetUint32("protocol.networkid")
	if config.ProtocolVersion == 0 {
		config.ProtocolVersion = defaultProtocolVersion
	}
	if config.NetworkID == 0 {
		config.NetworkID = defaultNetworkId
	}
	return config
}

func parseDaemonConfig(configFilePath string) (daemonConfig, error) {
	config := viper.New()
	if err := readFromConfigPath(config, configFilePath); err != nil {
		return daemonConfig{}, err
	}
	mStorageParams := parseConfigStorageParams(config)
	mBackendParams := parseConfigBackendParams(config)
	mLoggerParams := parseConfigLoggerParams(config)
	nodeParams := parseConfigNodeParams(config, mBackendParams.NetworkID)
	nodeParams.NodeDBPath = mStorageParams.nodesDir
	return daemonConfig{
		loggerParams: mLoggerParams,
		storageParams: mStorageParams,
		nodeConfig:    nodeParams,
		backendParams: mBackendParams,
	}, nil
}

func parseClientConfig(configFilePath string) (clientConfig, error) {
	config := viper.New()
	if err := readFromConfigPath(config, configFilePath); err != nil {
		return clientConfig{}, err
	}
	mRpcClientApiHost := config.GetString("rpclient.apihost")
	if mRpcClientApiHost == "" {
		mRpcClientApiHost = defaultRPCClientAPIHost
	}
	return clientConfig{
		rpcClientApiHost: mRpcClientApiHost,
	}, nil
}
