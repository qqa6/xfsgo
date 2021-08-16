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

package xfsgo

import "fmt"

type RPCError struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

func NewRPCError(code int, message string) *RPCError {
	return &RPCError{
		Code:    code,
		Message: message,
	}
}

func NewRPCErrorCause(code int, err error) *RPCError {
	return &RPCError{
		Code:    code,
		Message: err.Error(),
	}
}

func (e *RPCError) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}
