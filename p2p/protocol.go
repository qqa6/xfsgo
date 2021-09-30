package p2p

// network protocol
type Protocol interface {
	Run(p Peer) error
}

type SimpleProtocol struct {
	Func func(p Peer) error
}

func (sp *SimpleProtocol) Run(p Peer) error {
	return sp.Func(p)
}

// func Run(p Peer, ps []Protocol) {
// 	peer := newPeer(p, ps)
// }
