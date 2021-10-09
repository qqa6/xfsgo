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

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
)

type Client struct {
	hostUrl string
	timeOut string
}

type jsonRPCReq struct {
	JsonRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

type jsonRPCResp struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
	Error   *RPCError   `json:"error"`
	ID      int         `json:"id"`
}

func NewClient(url, timeOut string) *Client {
	return &Client{
		hostUrl: url,
		timeOut: timeOut,
	}
}

// CallMethod executes a JSON-RPC call with the given psrameters,which is important to the rpc server.
func (cli *Client) CallMethod(id int, methodname string, params interface{}, out interface{}) error {
	client := resty.New()
	outTime, err := strconv.Atoi(cli.timeOut)
	if err != nil {
		return err
	}
	client = client.SetTimeout(time.Duration(time.Second * time.Duration(outTime)))
	req := &jsonRPCReq{
		JsonRPC: "2.0",
		ID:      id,
		Method:  methodname,
		Params:  params,
	}
	// The result must be a pointer so that response json can unmarshal into it.
	var resp *jsonRPCResp = nil
	r, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		SetResult(&resp). // or SetResult(AuthSuccess{}).
		Post(cli.hostUrl)
	if err != nil {
		return err
	}
	if resp == nil {
		return fmt.Errorf("resp null")
	}
	e := resp.Error
	if e != nil {
		return e
	}
	js, err := json.Marshal(resp.Result)
	if err != nil {
		return err
	}
	err = json.Unmarshal(js, out)
	if err != nil {
		return err
	}
	_ = r
	return nil
}
