package p2p

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/json"
	"errors"
)

var MsgCodePing = uint32(1)
var MsgCodePong = uint32(2)



var DefaultVersion = uint32(1)

type Header struct {
	Length  uint32 `json:"length"`
	Checksum uint32 `json:"checksum"`
	Version uint32 `json:"version"`
	MsgCode uint32 `json:"msg_code"`
}

type Message struct {
	Header *Header `json:"header"`
	Body   []byte `json:"body"`
}

func UnmarshalMsg(bs []byte, msgp *Message) error {
	if err := json.Unmarshal(bs, &msgp); err != nil {
		return err
	}
	header := msgp.Header
	checksum := msgp.calcChecksum()
	if header.Checksum != checksum {
		return errors.New("check err")
	}
	return nil
}

func (m Message) String() string {
	var bs []byte = nil
	var err error = nil
	if bs, err = m.MarshalMsg(); err != nil {
		return ""
	}
	return string(bs)
}

func (m *Message) MarshalMsg() ([]byte,error) {
	m.reSetupHeader()
	return json.Marshal(m)
}
func (m *Message) reSetupHeader() {
	m.Header.Length = uint32(len(m.Body))
	m.Header.Checksum = m.calcChecksum()
}

func (m *Message) calcChecksum() uint32 {
	if m.Body == nil {
		return uint32(0)
	}
	bs0 := md5.Sum(m.Body)
	bs := md5.Sum(bs0[:])
	return binary.BigEndian.Uint32(bs[0:4])
}

func SendMsg(peer *Peer, msgCode uint32, data []byte) error {
	msg := &Message{
		Header: &Header{
			Version: DefaultVersion,
			MsgCode: msgCode,
		},
		Body: data,
	}
	var err error = nil
	var bs []byte = nil
	if bs, err = msg.MarshalMsg(); err != nil {
		return err
	}
	peer.sendData(bs)
	return nil
}
func SendMsgJSONData(peer *Peer, msgCode uint32, data interface{}) error {
	var err error = nil
	var bs []byte = nil
	if bs, err = json.Marshal(data); err != nil {
		return err
	}
	msg := &Message{
		Header: &Header{
			Version: DefaultVersion,
			MsgCode: msgCode,
		},
		Body: bs,
	}
	var msgBs []byte = nil
	if msgBs, err = msg.MarshalMsg(); err != nil {
		return err
	}
	peer.sendData(msgBs)
	return nil
}
