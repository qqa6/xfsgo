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
	"xfsgo/node"

	"github.com/sirupsen/logrus"
)

func StartNodeAndBackend(node *node.Node, backend *Backend) error {
	logrus.Info("STARTING NODE SERVICES...")
	if err := node.Start(); err != nil {
		return err
	}
	logrus.Info("NODE SERVICE IS STARTED. STARTING DAEMON ...")
	if err := backend.Start(); err != nil {
		return err
	}
	logrus.Info("DAEMON IS STARTED !!!")
	return nil
}
