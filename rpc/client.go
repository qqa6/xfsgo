package rpc

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"xblockchain/rpc/errors"
)

type Client struct {
	hostUrl string
}

type jsonRPCReq struct {
	JsonRPC string `json:"jsonrpc"`
	ID int `json:"id"`
	Method string `json:"method"`
	Params interface{} `json:"params"`
}

type jsonRPCResp struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
	Error   *errors.JsonRPCError `json:"error"`
	ID      int         `json:"id"`
}

func NewClient(url string) *Client {
	return &Client{
		hostUrl: url,
	}
}

func (cli *Client) CallMethod(id int,methodname string, params interface{}, out interface{}) error {
	client := resty.New()
	req := &jsonRPCReq{
		JsonRPC: "2.0",
		ID: id,
		Method: methodname,
		Params: params,
	}
	var resp *jsonRPCResp = nil
	r, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		SetResult(&resp).    // or SetResult(AuthSuccess{}).
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