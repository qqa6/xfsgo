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
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
	"time"
	"xfsgo/backend"
	"xfsgo/node"
	"xfsgo/storage/badger"

	"github.com/spf13/cobra"
)

var (
	daemonCmd = &cobra.Command{
		Use:   "daemon [flags]",
		Short: "background services",
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


func runDaemon() error {
	var (
		err   error            = nil
		stack *node.Node       = nil
		back  *backend.Backend = nil
	)
	config, err := parseDaemonConfig(cfgFile)
	if err != nil {
		return err
	}
	loglevel,err := logrus.ParseLevel(config.loggerParams.level)
	if err != nil {
		return err
	}
	logrus.SetLevel(loglevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors: true,
		TimestampFormat : time.RFC3339,
		FullTimestamp:true,
	})
	if stack, err = node.New(&config.nodeConfig); err != nil {
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
	rootCmd.AddCommand(daemonCmd)
}
