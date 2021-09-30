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

package api

import (
	"xfsgo/miner"
)

type MinerAPIHandler struct {
	Miner *miner.Miner
}

func (handler *MinerAPIHandler) Start(_ EmptyArgs, resp *string) error {
	handler.Miner.Start()
	*resp = ""
	return nil
}

func (handler *MinerAPIHandler) Stop(_ EmptyArgs, resp *string) error {
	handler.Miner.Stop()
	*resp = ""
	return nil
}

func (handler *MinerAPIHandler) WorkersAdd(_ EmptyArgs, resp *string) error {
	NumWorkers := int(handler.Miner.GetNumWorkers()) + 1
	handler.Miner.SetNumWorkers(uint32(NumWorkers))
	return nil
}

func (handler *MinerAPIHandler) WorkersDown(_ EmptyArgs, resp *string) error {
	NumWorkers := int(handler.Miner.GetNumWorkers()) - 1
	handler.Miner.SetNumWorkers(uint32(NumWorkers))
	return nil
}
