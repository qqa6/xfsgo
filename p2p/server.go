package p2p

import (
	"context"
	"errors"
	"github.com/perlin-network/noise"
	"github.com/perlin-network/noise/kademlia"
	"github.com/sirupsen/logrus"
	"net"
	"sync"
	"time"
)

type Server struct {
	ListenAddr string
	BootstrapNodes []string
	p2pNode *noise.Node
	p2pOverlay *kademlia.Protocol
	lock  sync.Mutex
	running bool
	addPeerCh chan noise.ID
	runPeerCh chan noise.ID
	delPeerCh chan noise.ID
	targetWriteCh chan targetWrite
	PeerHandlerFn func(peer *Peer) error
	peers map[noise.PublicKey]*Peer
}

type targetWrite struct {
	target noise.ID
	data []byte
}

func (srv *Server) Start() error {
	srv.lock.Lock()
	defer srv.lock.Unlock()
	if srv.running {
		return errors.New("server already running")
	}
	srv.running = true
	srv.addPeerCh = make(chan noise.ID)
	srv.delPeerCh = make(chan noise.ID)
	srv.targetWriteCh = make(chan targetWrite)
	srv.peers = make(map[noise.PublicKey]*Peer)
	if err := srv.setupLocalNode(); err != nil {
		return err
	}
	if err := srv.p2pNode.Listen() ; err != nil {
		return err
	}
	go srv.bootstrap()
	go srv.run()
	srv.running = true
	return nil
}

func (srv *Server) setupLocalNode() error {
	var err error = nil
	var addr *net.TCPAddr = nil
	if srv.ListenAddr != "" {
		if addr,err = net.ResolveTCPAddr(
			"tcp", srv.ListenAddr); err != nil {
			return err
		}
	}
	if addr == nil {
		if addr,err = net.ResolveTCPAddr(
			"tcp", "0.0.0.0:9066"); err != nil {
			return err
		}
	}
	if srv.p2pNode, err = noise.NewNode(
		noise.WithNodeBindHost(addr.IP),
		noise.WithNodeBindPort(uint16(addr.Port)),
	); err != nil {
		return err
	}

	srv.p2pNode.Handle(srv.handleP2PMessage)
	srv.p2pOverlay = srv.createOverlay()
	srv.p2pNode.Bind(srv.p2pOverlay.Protocol())
	return nil
}

func (srv *Server) handleOverlayOnPeerAdmitted(id noise.ID) {
	srv.addPeerCh <- id
}

func (srv *Server) handleOverlayOnPeerEvicted(id noise.ID) {
	srv.delPeerCh <- id
}

func (srv *Server) handleP2PMessage(ctx noise.HandlerContext) error  {
	if ctx.IsRequest() {
		return nil
	}
	id := ctx.ID()
	p := srv.peers[id.ID]
	if p == nil {
		logrus.Warnf("peer not found1")
		return nil
	}
	if err := p.handleData(ctx.Data()); err != nil {
		return err
	}
	return nil
}

func (srv *Server) createOverlay() *kademlia.Protocol {
	events := kademlia.Events{
		OnPeerAdmitted: srv.handleOverlayOnPeerAdmitted,
		OnPeerEvicted: srv.handleOverlayOnPeerEvicted,
	}
	return kademlia.New(kademlia.WithProtocolEvents(events))
}

func (srv *Server) bootstrap() {
	for _, addr := range srv.BootstrapNodes {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		if _, err := srv.p2pNode.Ping(ctx, addr); err != nil {
			logrus.Warnf("find bootstrap node err: %s\n", err)
		}
		cancel()
	}
}


func (srv *Server) run() {
	for {
		select {
		case p := <-srv.addPeerCh:
			np := newPeer(srv.targetWriteCh, p)
			srv.peers[p.ID] = np
			if srv.PeerHandlerFn == nil {
				return
			}
			go func() {
				if err := np.run(srv.PeerHandlerFn); err != nil {
					srv.delPeerCh <- p
				}
			}()
		case tw := <-srv.targetWriteCh:
			target := tw.target
			if srv.peers[target.ID] == nil {
				logrus.Warnf("peer not found2")
				break
			}
			if err := srv.p2pNode.Send(context.Background(), target.Address, tw.data); err!=nil {
				srv.delPeerCh <- tw.target
			}
		case id := <-srv.delPeerCh:
			logrus.Infof("node exit: %s", id.Address)
			delete(srv.peers, id.ID)
		}
	}
}
