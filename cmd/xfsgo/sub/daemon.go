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
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"xfsgo"
	"xfsgo/backend"
	"xfsgo/log"
	"xfsgo/node"
	"xfsgo/storage/badger"

	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var (
	rpcaddr        string
	p2paddr        string
	datadir        string
	importsnapshot string
	importchain    string
	bootstrap      string
	testnet        bool
	netid          int
	daemonCmd      = &cobra.Command{
		Use:   "daemon [flags]",
		Short: "Start a xfsgo daemon process",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDaemon()
		},
	}
)

func safeclose(fn func() error) {
	if err := fn(); err != nil {
		panic(err)
	}
}

func resetConfig(config *daemonConfig) {
	if datadir != "" {
		setupDataDir(&config.storageParams, datadir)
		config.nodeConfig.NodeDBPath = config.storageParams.nodesDir
	}
	if rpcaddr != "" {
		config.nodeConfig.RPCConfig.ListenAddr = rpcaddr
	}
	if p2paddr != "" {
		config.nodeConfig.P2PListenAddress = p2paddr
	}
	if netid != 0 {
		config.backendParams.NetworkID = uint32(netid)
	}
	if testnet {
		config.backendParams.NetworkID = defaultTestNetworkId
	}
	if bootstrap != "" {
		config.nodeConfig.P2PBootstraps = strings.Split(bootstrap, ",")
	}
}
func runDaemon() error {
	var (
		err   error            = nil
		stack *node.Node       = nil
		back  *backend.Backend = nil
	)
	config, err := parseDaemonConfig(cfgFile) // default config
	if err != nil {
		return err
	}
	resetConfig(&config) // input config
	loglevel, err := logrus.ParseLevel(config.loggerParams.level)
	if err != nil {
		return err
	}
	//logrus.Infof("lavel: %s", loglevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors:   true,
		TimestampFormat: time.RFC3339,
		FullTimestamp:   true,
	})
	logrus.SetFormatter(&log.Formatter{})
	logrus.SetLevel(loglevel)
	nodeConf := &config.nodeConfig
	nodeConf.RPCConfig.Logger = logrus.StandardLogger()
	if stack, err = node.New(nodeConf); err != nil {
		return err
	}
	chainDb := badger.New(config.storageParams.chainDir)
	keysDb := badger.New(config.storageParams.keysDir)
	stateDB := badger.New(config.storageParams.stateDir)
	extraDB := badger.New(config.storageParams.extraDir)
	defer func() {
		safeclose(chainDb.Close)
		safeclose(keysDb.Close)
		safeclose(stateDB.Close)
		safeclose(extraDB.Close)
	}()
	if back, err = backend.NewBackend(stack, &backend.Config{
		Params:  &config.backendParams,
		ChainDB: chainDb,
		KeysDB:  keysDb,
		StateDB: stateDB,
		ExtraDB: extraDB,
	}); err != nil {
		return err
	}
	if err = backend.StartNodeAndBackend(stack, back); err != nil {
		return err
	}

	if importsnapshot != "" {
		data, err := ioutil.ReadFile(importsnapshot)
		if err != nil {
			return err
		}
		config, err := parseClientConfig(cfgFile)
		if err != nil {
			return err
		}
		req := &GetBlocksArgs{
			Blocks: string(data),
		}
		cli := xfsgo.NewClient(config.rpcClientApiHost)
		var result string
		err = cli.CallMethod(1, "Chain.ImportBlock", req, &result)
		if err != nil {
			return err
		}
	}

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
out:
	select {
	case s := <-c:
		switch s {
		case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
			break out
		}
	}
	return nil
}

func init() {
	mFlags := daemonCmd.PersistentFlags()
	mFlags.StringVarP(&rpcaddr, "rpcaddr", "r", "", "Set JSON-RPC Service listen address")
	mFlags.StringVarP(&p2paddr, "p2paddr", "p", "", "Set P2P-Node listen address")
	mFlags.StringVarP(&datadir, "datadir", "d", "", "Set Data directory")
	mFlags.StringVarP(&importsnapshot, "importsnapshot", "", "", "Imports data from the specified snapshot file")
	mFlags.StringVarP(&importchain, "importchain", "", "", "Import data from the specified chain file and start the service")
	mFlags.StringVarP(&bootstrap, "bootstrap", "", "", "Specify boot node")
	mFlags.BoolVarP(&testnet, "testnet", "t", false, "Enable test network")
	mFlags.IntVarP(&netid, "netid", "n", 0, "Explicitly set network id")
	rootCmd.AddCommand(daemonCmd)
}
