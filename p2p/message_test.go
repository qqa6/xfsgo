package p2p

import "testing"

func TestMessage_MarshalMsg(t *testing.T) {
	msg := &Message{
		Header: &Header{
			Version: 0,
			MsgCode: 0,
		},
		Body: nil,
	}
	var bs []byte = nil
	var err error = nil
	if bs, err = msg.MarshalMsg(); err != nil {
		t.Fatal(err)
	}
	t.Logf("msg: %s\n", string(bs))
}

func TestUnmarshalMsg(t *testing.T) {
	bs := []byte("{\"header\":{\"length\":10,\"checksum\":4068781620,\"version\":0,\"msg_code\":0},\"body\":\"YWJjZGVmc2dzZw==\"}")
	var m = &Message{}
	var err error = nil
	if err = UnmarshalMsg(bs, m); err != nil {
		t.Fatal(err)
	}
	t.Logf("msgBody: %s\n", string(m.Body))
}
