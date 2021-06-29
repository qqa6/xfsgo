package util

import (
	"github.com/sirupsen/logrus"
	"xblockchain/backend"
	"xblockchain/node"
)

func StartNodeAndBackend(node *node.Node, backend *backend.Backend) error {
	logrus.Info("start node...")
	if err := node.Start(); err != nil {
		return err
	}
	if err := backend.Start(); err != nil {
		return err
	}
	return nil
}