package p2p

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"xfsgo/common"
	"xfsgo/p2p/discover"

	"github.com/sirupsen/logrus"
)

// Peer to peer connection session
type peerConn struct {
	inbound         bool
	id              discover.NodeId
	self            discover.NodeId
	server          *server
	key             *ecdsa.PrivateKey
	rw              net.Conn
	version         uint8
	handshakeStatus int
	flag int
}

func (c *peerConn) serve() {
	// Get the address and port number of the client
	fromAddr := c.rw.RemoteAddr()
	inbound := c.flag & flagInbound != 0
	if inbound {
		if err := c.serverHandshake(); err != nil {
			logrus.Warnf("handshake error from %s: %v", fromAddr, err)
			c.close()
			return
		}
	} else {
		if err := c.clientHandshake(); err != nil {
			logrus.Warnf("handshake error from %s: %v", fromAddr, err)
			c.close()
			return
		}
	}
	logrus.Infof("p2p handshake success by %s", fromAddr)
	// 加入节点p2pserver 节点
	c.server.addpeer <- c
}

//Client handshake sending method
func (c *peerConn) clientHandshake() error {

	// Whether the handshake status is based on handshake
	if c.handshakeCompiled() {
		return nil
	}
	request := &helloRequestMsg{
		version:   c.version,
		id:        c.self,
		receiveId: c.id,
	}
	logrus.Debugf("send hello request version: %d, id: %s, receiveId: %s", c.version,c.self, c.id)
	// send data 发送消息
	_, err := c.rw.Write(request.marshal())
	if err != nil {
		return err
	}
	// Read reply data 读取消息
	hello, err := c.readHelloReRequestMsg()
	if err != nil {
		return err
	}
	if hello.version != c.version {
		return fmt.Errorf("handshake check err, got version: %d, want version: %d",
			hello.version, c.version)
	}
	gotId := hello.receiveId
	wantId := c.self
	if !bytes.Equal(gotId[:], wantId[:]) {
		return fmt.Errorf("handshake check err got my name: 0x%x, my real name: 0x%x",
			gotId, wantId)
	}
	c.handshakeStatus = 1
	return nil
}

// Service handshake response method
func (c *peerConn) serverHandshake() error {
	// Whether the handshake status is based on handshake
	if c.handshakeCompiled() {
		return nil
	}

	// Read reply data
	// 获取接收到的数据
	hello, err := c.readHelloRequestMsg()
	if err != nil {
		return err
	}
	if hello.version != c.version {
		return fmt.Errorf("handshake check err, got version: %d, want version: %d",
			hello.version, c.version)
	}
	gotId := hello.receiveId
	wantId := c.self
	if !bytes.Equal(gotId[:], wantId[:])  {
		return fmt.Errorf("handshake check err got my name: 0x%x, my real name: 0x%x",
			gotId, wantId)
	}
	c.id = hello.id
	reply := &helloReRequestMsg{
		id:        c.self,
		receiveId: hello.id,
		version:   c.version,
	}
	// 回复
	logrus.Infof("send hanshake reply %v", reply.marshal())
	if _, err = c.rw.Write(reply.marshal()); err != nil {
		return err
	}
	return nil
}

// Read reply message
func (c *peerConn) readHelloReRequestMsg() (*helloReRequestMsg, error) {
	msg, err := c.readMessage()
	if err != nil {
		return nil, err
	}
	if msg.Type() != typeReHelloRequest {
		return nil, err
	}
	nMsg := new(helloReRequestMsg)
	raw, _ := ioutil.ReadAll(msg.RawReader())
	if !nMsg.unmarshal(raw) {
		return nil, errors.New("parse hello request err")
	}
	return nMsg, nil
}

// Read peer session messages
func (c *peerConn) readHelloRequestMsg() (*helloRequestMsg, error) {
	msg, err := c.readMessage()
	if err != nil {
		return nil, err
	}
	if msg.Type() != typeHelloRequest {
		return nil, err
	}
	nMsg := new(helloRequestMsg)
	raw, _ := ioutil.ReadAll(msg.RawReader())
	if !nMsg.unmarshal(raw) {
		return nil, errors.New("parse hello request err")
	}
	return nMsg, nil
}

// Write peer session messages
func (c *peerConn) writeMessage(mType uint8, data []byte) error {
	cLen := len(data)
	val := make([]byte, cLen+4)
	binary.LittleEndian.PutUint32(val, uint32(cLen))
	copy(val[4:], data)
	msg := []byte{c.version, mType}
	msg = append(msg, val...)
	_, err := c.rw.Write(msg)
	if err != nil {
		return err
	}
	return nil
}

func (c *peerConn) readMessage() (MessageReader, error) {
	return ReadMessage(c.rw)
}

func (c *peerConn) close() {
	common.Safeclose(c.rw.Close)
}

func (c *peerConn) handshakeCompiled() bool {
	return c.handshakeStatus == 1
}
