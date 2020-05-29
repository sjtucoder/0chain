package conductrpc

import (
	"fmt"
	"net"
	"net/rpc"
	"sync"

	"0chain.net/conductor/config"
)

// type aliases
type (
	NodeID    = config.NodeID
	Round     = config.Round
	Phase     = config.Phase
	RoundName = config.RoundName
)

// ViewChangeEvent represents view change information.
type ViewChangeEvent struct {
	Sender   NodeID   // node that sends the VC
	Round    Round    // view change round
	Miners   []NodeID // magic block miners
	Sharders []NodeID // magic block sharders
}

// PhaseEvent represents phase switching.
type PhaseEvent struct {
	Sender NodeID //
	Phase  Phase  //
}

// AddMinerEvent in miner SC.
type AddMinerEvent struct {
	Sender  NodeID // event emitter
	MinerID NodeID // the added miner
}

// AddSharderEvent in miner SC.
type AddSharderEvent struct {
	Sender    NodeID // event emitter
	SharderID NodeID // the added sharder
}

// Round proceed in pay_fees of Miner SC.
type RoundEvent struct {
	Sender NodeID // event emitter
	Round  Round  // round number
}

// ContributeMPKEvent where a miner successfully sent its contribution.
type ContributeMPKEvent struct {
	Sender  NodeID // event emitter
	MinerID NodeID // miner that contributes
}

// ShareOrSignsSharesEvent where a miner successfully sent its share or sign
type ShareOrSignsSharesEvent struct {
	Sender  NodeID // event emitter
	MinerID NodeID // miner that sends
}

// known locks
const (
	Locked   = true  // should wait
	Unlocked = false // can join
)

type nodeLock struct {
	lock    bool //
	counter int  //
}

type Server struct {
	server  *rpc.Server
	address string
	l       net.Listener

	// server events

	// onViewChange occurs where BC made VC (round == view change round)
	onViewChange chan *ViewChangeEvent
	// onPhase occurs for every phase change
	onPhase chan *PhaseEvent
	// onAddMiner occurs where miner SC proceed add_miner function
	onAddMiner chan *AddMinerEvent
	// onAddSharder occurs where miner SC proceed add_sharder function
	onAddSharder chan *AddSharderEvent

	// onNodeReady used by miner/sharder to notify the server that the node
	// has started and ready to register (if needed) in miner SC and start
	// it work. E.g. the node has started and waits the conductor to enter BC.
	onNodeReady chan NodeID

	onRoundEvent              chan *RoundEvent
	onContributeMPKEvent      chan *ContributeMPKEvent
	onShareOrSignsSharesEvent chan *ShareOrSignsSharesEvent

	// add / lock  miner / sharder
	mutex sync.Mutex
	locks map[NodeID]*nodeLock // expected miner/sharder -> locked/unlocked

	quitOnce sync.Once
	quit     chan struct{}
}

// NewServer Conductor RPC server.
func NewServer(address string) (s *Server, err error) {
	s = new(Server)
	s.quit = make(chan struct{})

	// without a buffer
	s.onViewChange = make(chan *ViewChangeEvent, 10)
	s.onPhase = make(chan *PhaseEvent, 10)
	s.onAddMiner = make(chan *AddMinerEvent, 10)
	s.onAddSharder = make(chan *AddSharderEvent, 10)
	s.onNodeReady = make(chan NodeID, 10)

	s.onRoundEvent = make(chan *RoundEvent, 100)
	s.onContributeMPKEvent = make(chan *ContributeMPKEvent, 10)
	s.onShareOrSignsSharesEvent = make(chan *ShareOrSignsSharesEvent, 10)

	s.locks = make(map[NodeID]*nodeLock)
	s.server = rpc.NewServer()
	if err = s.server.Register(s); err != nil {
		return nil, err
	}
	s.address = address
	return
}

func (s *Server) Serve() (err error) {
	var l net.Listener
	if l, err = net.Listen("tcp", s.address); err != nil {
		return
	}
	go s.server.Accept(l)
	return
}

//
// add/lock miner/sharder
//

// AddNode adds miner of sharder and, optionally, locks it.
func (s *Server) AddNode(nodeID NodeID, lock bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.locks[nodeID] = &nodeLock{counter: 0, lock: lock}
}

// UnlockNode unlocks a miner.
func (s *Server) UnlockNode(nodeID NodeID) (err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var nl, ok = s.locks[nodeID]
	if !ok {
		return fmt.Errorf("unexpected node: %s", nodeID)
	}

	nl.lock = false
	return
}

func (s *Server) nodeLock(nodeID NodeID) (lock bool, cnt int, ok bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var nl *nodeLock
	if nl, ok = s.locks[nodeID]; !ok {
		return
	}

	lock, cnt = nl.lock, nl.counter
	nl.counter++
	return
}

// events handling

// OnViewChange events channel. The event occurs where
// BC made VC (round == view change round).
func (s *Server) OnViewChange() chan *ViewChangeEvent {
	return s.onViewChange
}

// OnPhase events channel. The event occurs where miner SC changes its phase.
func (s *Server) OnPhase() chan *PhaseEvent {
	return s.onPhase
}

// OnAddMiner events channel. The event occurs
// where miner SC proceed add_miner function.
func (s *Server) OnAddMiner() chan *AddMinerEvent {
	return s.onAddMiner
}

// OnAddSharder events channel. The event occurs
// where miner SC proceed add_sharder function.
func (s *Server) OnAddSharder() chan *AddSharderEvent {
	return s.onAddSharder
}

// OnNodeReady used by nodes to notify the server that the node has started
// and ready to register (if needed) in miner SC and start it work. E.g.
// the node has started and waits the conductor to enter BC.
func (s *Server) OnNodeReady() chan NodeID {
	return s.onNodeReady
}

func (s *Server) OnRound() chan *RoundEvent {
	return s.onRoundEvent
}

func (s *Server) OnContributeMPK() chan *ContributeMPKEvent {
	return s.onContributeMPKEvent
}

func (s *Server) OnShareOrSignsShares() chan *ShareOrSignsSharesEvent {
	return s.onShareOrSignsSharesEvent
}

//
// handlers
//

func (s *Server) ViewChange(viewChange *ViewChangeEvent, _ *struct{}) (
	err error) {

	select {
	case s.onViewChange <- viewChange:
	case <-s.quit:
	}
	return
}

func (s *Server) Phase(phase *PhaseEvent, _ *struct{}) (err error) {
	select {
	case s.onPhase <- phase:
	case <-s.quit:
	}
	return
}

func (s *Server) AddMiner(add *AddMinerEvent, _ *struct{}) (err error) {
	select {
	case s.onAddMiner <- add:
	case <-s.quit:
	}
	return
}

func (s *Server) AddSharder(add *AddSharderEvent, _ *struct{}) (err error) {
	select {
	case s.onAddSharder <- add:
	case <-s.quit:
	}
	return
}

func (s *Server) NodeReady(nodeID NodeID, join *bool) (err error) {

	var lock, cnt, ok = s.nodeLock(nodeID)
	if !ok {
		return fmt.Errorf("unexpected node: %s", nodeID)
	}

	(*join) = !lock

	if cnt > 0 {
		return // don't trigger onNodeReady twice or more times
	}

	select {
	case s.onNodeReady <- nodeID:
	case <-s.quit:
	}

	return
}

func (s *Server) Round(rnd *RoundEvent, _ *struct{}) (err error) {
	select {
	case s.onRoundEvent <- rnd:
	case <-s.quit:
	}
	return
}

func (s *Server) ContributeMPK(cmpke *ContributeMPKEvent, _ *struct{}) (
	err error) {

	select {
	case s.onContributeMPKEvent <- cmpke:
	case <-s.quit:
	}
	return
}

func (s *Server) ShareOrSignsShares(soss *ShareOrSignsSharesEvent,
	_ *struct{}) (err error) {

	select {
	case s.onShareOrSignsSharesEvent <- soss:
	case <-s.quit:
	}
	return
}

//
// flow
//

// Close the server waiting.
func (s *Server) Close() (err error) {
	s.quitOnce.Do(func() {
		close(s.quit)
		if s.l != nil {
			err = s.l.Close()
		}
	})
	return
}
