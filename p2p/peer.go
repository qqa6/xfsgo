package p2p

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"time"
	"xfsgo/p2p/discover"
	"xfsgo/p2p/log"
)

type Peer interface {
	Is(flag int) bool
	ID() discover.NodeId
	Close()
	Run()
	QuitCh() chan struct{}
	WriteMessage(mType uint8, data []byte) error
	WriteMessageObj(mType uint8, data interface{}) error
	Reader() io.Reader
	GetProtocolMsgCh() chan MessageReader
}

type peer struct {
	id       discover.NodeId
	conn     *peerConn
	rw       net.Conn
	close    chan struct{}
	lastTime int64
	readBuf  bytes.Buffer
	ps       []Protocol
	quit     chan struct{}
	psCh     chan MessageReader
	logger   log.Logger
}

// create peer [Peer to peer connection session,Network protocol]
func newPeer(conn *peerConn, ps []Protocol) Peer {
	p := &peer{
		conn:   conn,
		id:     conn.id,
		rw:     conn.rw,
		logger: conn.logger,
		ps:     ps,
		close:  make(chan struct{}),
		quit:   make(chan struct{}),
		psCh:   make(chan MessageReader),
	}
	now := time.Now()
	p.lastTime = now.Unix()
	return p
}

func (p *peer) ID() discover.NodeId {
	return p.id
}

func (p *peer) QuitCh() chan struct{} {
	return p.quit
}
func (p *peer) Is(flag int) bool {
	return p.conn.flag&flag != 0
}

// Read heartbeat message
func (p *peer) readLoop() {
	for {
		select {
		case <-p.close:
			return
		default:
		}
		msg, err := ReadMessage(p.rw)
		if err != nil {
			return
		}
		p.handle(msg)
	}
}

func (p *peer) handle(msg MessageReader) {
	data, err := msg.ReadAll()
	if err != nil {
		return
	}
	//p.logger.Infof("peer handle message type %d, data: %s", msg.Type(), string(data))
	switch msg.Type() {
	case typePingMsg:
		p.logger.Infof("receive heartbeat request")
		_ = p.conn.writeMessage(typePongMsg, []byte("hello"))
	case typePongMsg:
		p.logger.Infof("receive response of haertbeat and update alive time")
		now := time.Now()
		p.lastTime = now.Unix()
	default:
		bodyBs := msg.RawReader()
		cpy := &messageReader{
			raw:   bodyBs,
			mType: msg.Type(),
			data:  bytes.NewReader(data),
		}
		p.psCh <- cpy // copy -> ps chan
	}
}

func (p *peer) Reader() io.Reader {
	return &p.readBuf
}

func (p *peer) Read(bs []byte) (int, error) {
	return p.readBuf.Read(bs)
}

func (p *peer) GetProtocolMsgCh() chan MessageReader {
	return p.psCh
}

func (p *peer) WriteMessage(mType uint8, bs []byte) error {
	return p.conn.writeMessage(mType, bs)
}

func (p *peer) WriteMessageObj(mType uint8, obj interface{}) error {
	bs, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	p.logger.Infof("peer write message type: %d, data: %x, obj: %v", mType, bs, obj)
	return p.WriteMessage(mType, bs)
}

func (p *peer) pingLoop() {
	ping := time.NewTicker(10 * time.Second)
	defer ping.Stop()
	for {
		select {
		case <-ping.C:
			if err := p.conn.writeMessage(typePingMsg, []byte("hello")); err != nil {
				return
			}
		case <-p.close:
			return
		}
	}
}

func (p *peer) suicide(timout chan struct{}) {
	for {
		select {
		case <-p.close:
			return
		default:
		}
		now := time.Now()
		nowTime := now.Unix()
		interval := nowTime - p.lastTime
		// 10s
		if interval > 30 {
			p.logger.Infof("peer stop running because of timeout ")
			break
		}
		time.Sleep(10 * time.Second)
	}
	close(timout)
}

func (p *peer) Run() {
	go p.readLoop()
	go p.pingLoop()
	runProtocol := func() {
		for _, item := range p.ps {
			go func(p *peer, item Protocol) {
				err := item.Run(p)

				if err != nil {
					p.Close()
				}
			}(p, item)
		}
	}
	runProtocol()
	timout := make(chan struct{})
	go p.suicide(timout)
loop:
	for {
		select {
		case <-timout:
			break loop
		}
	}
	close(p.close)
}

func (p *peer) Close() {
	close(p.close)
}
