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
	"xfsgo"
)

type TxPoolHandler struct {
	TxPool *xfsgo.TxPool
}

func (tx *TxPoolHandler) GetPending(_ EmptyArgs, resp *transactions) error {
	data := tx.TxPool.GetTransactions()
	*resp = data
	return nil
}

func (tx *TxPoolHandler) GetPendingSize(_ EmptyArgs, resp *int) error {
	data := tx.TxPool.GetTransactionsSize()
	*resp = data
	return nil
}
