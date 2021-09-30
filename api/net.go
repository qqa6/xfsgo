package api

import (
	"xfsgo"
	"xfsgo/p2p"
	"xfsgo/p2p/discover"
)

type NetAPIHandler struct {
	NetServer p2p.Server
}

type AddPeerArgs struct {
	Url string `json:"url"`
}

type DelPeerArgs struct {
	Id string `json:"id"`
}

func (net *NetAPIHandler) GetPeers(_ EmptyArgs, resp *[]string) error {
	peer := net.NetServer.Peers()
	peerid := make([]string, 1)
	for _, item := range peer {
		peerid = append(peerid, item.ID().String())
	}
	*resp = peerid
	return nil
}

func (net *NetAPIHandler) AddPeer(args AddPeerArgs, resp *string) error {
	if args.Url == "" {
		return xfsgo.NewRPCError(-1006, "Parameter cannot be empty")
	}
	node, err := discover.ParseNode(args.Url)
	if err != nil {
		return xfsgo.NewRPCErrorCause(-32001, err)
	}
	net.NetServer.AddPeer(node)
	return nil
}

func (net *NetAPIHandler) DelPeer(args DelPeerArgs, resp *string) error {
	if args.Id == "" {
		return xfsgo.NewRPCError(-1006, "Parameter cannot be empty")
	}
	nodeid, err := discover.Hex2NodeId(args.Id)
	if err != nil {
		return xfsgo.NewRPCErrorCause(-32001, err)
	}
	net.NetServer.RemovePeer(nodeid)
	return nil
}

func (net *NetAPIHandler) GetNodeId(_ EmptyArgs, resp *string) error {
	nodeid := net.NetServer.NodeId().String()
	*resp = nodeid
	return nil
}
