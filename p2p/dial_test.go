package p2p

import (
	"testing"
	"time"
	"xfsgo/assert"
	"xfsgo/crypto"
	"xfsgo/p2p/discover"
)

var boots = []string{
	"xfsnode://ff929a9b96a52935b3b65808c73e645b7bbf13dc53f6e0ce140079f106b67fa48d4690500db818c28618f70a04125d6707e38578904c7187fd755b8184f0375f@127.0.0.1:9092",
	"xfsnode://ff729a9b96a52935b3b65808c73e649b7bbf13dc53f6e0ce140079f106b67fa48d4690500db818c28618f70a04125d6707e38578904c7187fd755b8184f0375f@127.0.0.1:9092",
	"xfsnode://ff929a9b96a52935b3b65809c73e645b7bbf13dc53f6e0ce140079f106b67fa48d4690500db818c28618f70a04125d6707e38578904c7187fd755b8184f0375f@127.0.0.1:9092",
}
func Test_dialtask_newDialState(t *testing.T) {
	key,err := crypto.GenPrvKey()
	assert.Error(t, err)
	tab,err := discover.ListenUDP(key,"127.0.0.1:9001","./d0")
	assert.Error(t, err)
	bootNs := make([]*discover.Node, 0)
	for _, nRaw := range boots {
		n, err := discover.ParseNode(nRaw)
		assert.Error(t, err)
		bootNs = append(bootNs, n)
	}
	dynPeers := 10 / 2
	ds := newDialState(bootNs,tab,dynPeers)
	ps := make(map[discover.NodeId]Peer)
	for {
		ts := ds.newTasks(ps)
		for _, tt := range ts {
			tt.Do(nil)
			ds.taskDone(tt)
		}
		time.Sleep(10 * time.Second)
	}


}
