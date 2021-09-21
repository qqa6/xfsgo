package p2p

import (
	"strings"
	"testing"
	"xfsgo/crypto"
	"xfsgo/p2p/discover"

	"github.com/sirupsen/logrus"
)

var XQ = "xfsnode://127.0.0.1:9092?id=50e3c5dec0ebcda7059cff8b8c1e623b35bd1a9d0f60dca03fc664376521d5c8f6050bd0b0986ec69c7d51bac4223cd1d7a006f47d745b65431c690a365f16dd"

type testProto struct {
	t *testing.T
}

func (tp *testProto) Run(p Peer) error {
	tp.t.Logf("join peer: %s", p.ID())
	return nil
}
func newTestProto(t *testing.T) Protocol {
	tp := &testProto{
		t: t,
	}
	return tp
}

func TestServer_Start(t *testing.T) {
	logger := logrus.StandardLogger()
	logger.SetLevel(logrus.DebugLevel)
	key, _ := crypto.GenPrvKey()
	s := NewServer(Config{
		ListenAddr:      "127.0.0.1:9092",
		Key:             key,
		Discover:        true,
		NodeDBPath:      "./d1",
		MaxPeers:        0,
	})
	s.Bind(newTestProto(t))
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	select {}
}

func TestServer_Start2(t *testing.T) {
	logger := logrus.StandardLogger()
	logger.SetLevel(logrus.DebugLevel)
	bootAddress := parseBootAddress(XQ)
	key, _ := crypto.GenPrvKey()
	s := NewServer(Config{
		ListenAddr:      "127.0.0.1:9093",
		Key:             key,
		Discover:        true,
		BootstrapNodes:  bootAddress,
		NodeDBPath:      "./d2",
		MaxPeers:        10,
	})
	s.Bind(newTestProto(t))
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	select {}
}
func TestServer_Start3(t *testing.T) {
	logger := logrus.StandardLogger()
	logger.SetLevel(logrus.DebugLevel)
	bootAddress := parseBootAddress(XQ)
	key, _ := crypto.GenPrvKey()
	s := NewServer(Config{
		ListenAddr:      "127.0.0.1:9094",
		Key:             key,
		Discover:        true,
		BootstrapNodes:  bootAddress,
		NodeDBPath:      "./d3",
		MaxPeers:        10,
	})
	s.Bind(newTestProto(t))
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	select {}
}
func TestServer_Start4(t *testing.T) {
	bootAddress := parseBootAddress(XQ)
	key, _ := crypto.GenPrvKey()
	s := NewServer(Config{
		ListenAddr:      "127.0.0.1:9095",
		Key:             key,
		Discover:        true,
		BootstrapNodes:  bootAddress,
		NodeDBPath:      "./d4",
		MaxPeers:        10,
	})
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	select {}
}
func parseBootAddress(addrs string) []*discover.Node {
	if addrs == "" {
		return nil
	}
	arr := strings.Split(addrs, ",")
	addrArr := make([]*discover.Node, 0)
	for _, addr := range arr {
		a, err := discover.ParseNode(addr)
		if err != nil {
			continue
		}
		addrArr = append(addrArr, a)
	}
	return addrArr
}

func TestABC(t *testing.T) {
	a := flagOutbound | flagStatic
	i := a&flagInbound != 0
	t.Logf("abc: %v", i)
}