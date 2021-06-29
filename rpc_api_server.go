package xblockchain

import (
	"fmt"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
)

type RPCAPIServer struct {

}

func NewRPCAPIServer() *RPCAPIServer {
	return &RPCAPIServer{}
}

func (rPCAPIServer *RPCAPIServer) Start(ip string, port string) {
	server := rpc.NewServer()
	address := ip + ":" + port
	listener, err := net.Listen("tcp4", address)
	if err != nil {
		fmt.Printf("Listen server errors: %v", err)
		os.Exit(1)
	}
	defer listener.Close()

	//server.Register()
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Create server connection errors: %s\n", err)
			continue
		}
		go server.ServeCodec(jsonrpc.NewServerCodec(conn))
	}
}
