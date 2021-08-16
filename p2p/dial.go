package p2p

import (
	"crypto/rand"
	"net"
	"reflect"
	"xfsgo/p2p/discover"

	"github.com/sirupsen/logrus"
)

type task interface {
	Do(srv *server)
}

type dialtask struct {
	flag int
	dest *discover.Node
}

func (t *dialtask) Do(srv *server) {
	tcpAddr := t.dest.TcpAddr()
	coon, err := net.Dial("tcp", tcpAddr.String())
	if err != nil {
		return
	}
	id := t.dest.ID
	c := srv.newPeerConn(coon, t.flag, &id)
	c.serve()
}
type discoverTask struct {
	bootstrap bool
	result  []*discover.Node
}


func (t *discoverTask) Do(srv *server) {
	if t.bootstrap {
		srv.table.Bootstrap(srv.config.BootstrapNodes)
		return
	}
	var target discover.NodeId
	_, _ = rand.Read(target[:])
	t.result = srv.table.Lookup(target)
}


type dialstate struct {
	static map[discover.NodeId]*discover.Node
	ntab discoverTable
	maxDynDials int
	dialing map[discover.NodeId]int
	lookupBuf []*discover.Node
	lookupRunning bool
	bootstrapped  bool
	randomNodes []*discover.Node
}
type discoverTable interface {
	Self() *discover.Node
	Close()
	Bootstrap([]*discover.Node)
	Lookup(target discover.NodeId) []*discover.Node
	ReadRandomNodes([]*discover.Node) int
}

func newDialState(static []*discover.Node, table discoverTable, maxdyn int) *dialstate {
	d := &dialstate{
		ntab: table,
		maxDynDials: maxdyn,
		static: make(map[discover.NodeId]*discover.Node),
		dialing: make(map[discover.NodeId]int),
		randomNodes: make([]*discover.Node, maxdyn/2),
	}
	for _, a := range static {
		d.static[a.ID] = a
	}
	return d
}
func btou(b bool) int {
	if b {
		return 1
	}
	return 0
}
func (d *dialstate) newTasks(peers map[discover.NodeId]Peer) []task {
	var tasks []task
	addDial := func(flag int, n *discover.Node) bool {
		//the connection established needn't to join the pool
		_, dialing := d.dialing[n.ID]
		if dialing ||  peers[n.ID] != nil {
			return false
		}
		logrus.Infof("append peer id: %s to task", n.ID)
		d.dialing[n.ID] = flag
		tasks = append(tasks, &dialtask{
			flag: flag,
			dest:   n,
		})
		return true
	}
	// 计算需要的链接数
	needDynDials := d.maxDynDials
	// 检查当前已连接是否有动态类型
	for _,p := range peers {
		if !p.Is(flagDynamic) {
			needDynDials -= 1
		}
	}
	// 检车正在执行的是否包含动态类型
	for _,i := range d.dialing {
		if i == 0 {
			needDynDials -= 1
		}
	}
	for _, n := range d.static {
		addDial(flagOutbound|flagStatic, n)
	}
	randomCandidates := needDynDials / 2
	if randomCandidates > 0 && d.bootstrapped {
		n := d.ntab.ReadRandomNodes(d.randomNodes)
		for i := 0; i < randomCandidates && i < n; i++ {
			if addDial(flagOutbound|flagDynamic, d.randomNodes[i]) {
				needDynDials--
			}
		}
	}
	i := 0
	for ; i < len(d.lookupBuf) && needDynDials > 0; i++ {
		if addDial(flagOutbound|flagDynamic, d.lookupBuf[i]) {
			needDynDials--
		}
	}
	d.lookupBuf = d.lookupBuf[:copy(d.lookupBuf, d.lookupBuf[i:])]
	logrus.Infof("must need Dyn Dials count: %d, lookupRunning: %v", needDynDials, d.lookupRunning)
	if len(d.lookupBuf) < needDynDials && !d.lookupRunning {
		d.lookupRunning = true
		tasks = append(tasks, &discoverTask{bootstrap: !d.bootstrapped})

	}
	return tasks
}


func (d *dialstate) taskDone(t task) {
	mtt := reflect.TypeOf(t)
	logrus.Debugf("task done type: %s, ", mtt)
	switch mt := t.(type) {
	case *discoverTask:
		logrus.Debugf("discover task done, bootstrap: %v, result: %v", mt.bootstrap, mt.result)
		if mt.bootstrap {
			d.bootstrapped = true
		}
		d.lookupRunning = false
		d.lookupBuf = append(d.lookupBuf, mt.result...)
	case *dialtask:
		logrus.Debugf("dial task done, node id: %s", mt.dest.ID)
		delete(d.dialing, mt.dest.ID)
	}
}


