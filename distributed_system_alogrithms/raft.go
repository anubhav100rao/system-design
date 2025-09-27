// raft_replication_sim.go
// Step 2: Raft leader election + log replication simulation.
// Run: go run raft_replication_sim.go

package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// ---------- Log entry / Apply message ----------------
type LogEntry struct {
	Term    int
	Command string
}
type ApplyMsg struct {
	Index   int
	Command string
}

// ---------- RequestVote RPC types ---------------------
type RequestVoteArgs struct {
	Term         int
	CandidateId  int
	LastLogIndex int
	LastLogTerm  int
}
type RequestVoteReply struct {
	Term        int
	VoteGranted bool
}

// ---------- AppendEntries RPC types -------------------
type AppendEntriesArgs struct {
	Term         int
	LeaderId     int
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []LogEntry
	LeaderCommit int
}
type AppendEntriesReply struct {
	Term    int
	Success bool
	// simplified conflict hints omitted for brevity in this sim
}

// ---------- Network simulation -----------------------
type Network struct {
	nodes    map[int]*Node
	mtx      sync.RWMutex
	rand     *rand.Rand
	dropRate float64
	minDelay time.Duration
	maxDelay time.Duration
}

func NewNetwork(seed int64, dropRate float64, minDelay, maxDelay time.Duration) *Network {
	return &Network{
		nodes:    make(map[int]*Node),
		rand:     rand.New(rand.NewSource(seed)),
		dropRate: dropRate,
		minDelay: minDelay,
		maxDelay: maxDelay,
	}
}

func (net *Network) Register(node *Node) {
	net.mtx.Lock()
	net.nodes[node.id] = node
	net.mtx.Unlock()
}

func (net *Network) SendRequestVote(to int, args RequestVoteArgs, replyCh chan<- RequestVoteReply) {
	// simulate drop
	if net.rand.Float64() < net.dropRate {
		return
	}
	delay := net.minDelay
	if net.maxDelay > net.minDelay {
		delay += time.Duration(net.rand.Int63n(int64(net.maxDelay - net.minDelay)))
	}
	go func() {
		time.Sleep(delay)
		net.mtx.RLock()
		target, ok := net.nodes[to]
		net.mtx.RUnlock()
		if !ok {
			return
		}
		reply := target.handleRequestVoteRemote(args)
		select {
		case replyCh <- reply:
		default:
		}
	}()
}

func (net *Network) SendAppendEntries(to int, args AppendEntriesArgs, replyCh chan<- AppendEntriesReply) {
	// simulate drop
	if net.rand.Float64() < net.dropRate {
		return
	}
	delay := net.minDelay
	if net.maxDelay > net.minDelay {
		delay += time.Duration(net.rand.Int63n(int64(net.maxDelay - net.minDelay)))
	}
	go func() {
		time.Sleep(delay)
		net.mtx.RLock()
		target, ok := net.nodes[to]
		net.mtx.RUnlock()
		if !ok {
			return
		}
		reply := target.handleAppendEntriesRemote(args)
		select {
		case replyCh <- reply:
		default:
		}
	}()
}

// ---------- Node (Raft server) -----------------------
type Node struct {
	id    int
	peers []int
	net   *Network

	mu sync.Mutex

	// persistent-ish
	currentTerm int
	votedFor    int
	log         []LogEntry // 1-based indexing: log[0] is dummy

	// volatile
	commitIndex int
	lastApplied int
	state       string // "follower","candidate","leader"

	// leader state
	nextIndex  []int
	matchIndex []int

	// channels & control
	electionResetEvt chan struct{}
	applyCh          chan ApplyMsg
	stopCh           chan struct{}
	wg               sync.WaitGroup
}

func NewNode(id int, peers []int, net *Network, applyCh chan ApplyMsg) *Node {
	// initialize log with a dummy entry at index 0 (term 0)
	return &Node{
		id:               id,
		peers:            peers,
		net:              net,
		currentTerm:      0,
		votedFor:         -1,
		log:              []LogEntry{{Term: 0, Command: ""}}, // index 0
		commitIndex:      0,
		lastApplied:      0,
		state:            "follower",
		electionResetEvt: make(chan struct{}, 1),
		applyCh:          applyCh,
		stopCh:           make(chan struct{}),
	}
}

func (n *Node) Start() {
	n.net.Register(n)
	n.wg.Add(1)
	go n.run()
	// applier
	n.wg.Add(1)
	go n.applierLoop()
}

func (n *Node) Stop() {
	close(n.stopCh)
	n.wg.Wait()
}

// ---------- helper accessors (must be called under lock) ----------
func (n *Node) lastIndex() int {
	return len(n.log) - 1
}
func (n *Node) termAt(index int) int {
	if index < 0 || index > n.lastIndex() {
		return -1
	}
	return n.log[index].Term
}

// ---------- RPC handlers (invoked remotely by Network) ----------
func (n *Node) handleRequestVoteRemote(args RequestVoteArgs) RequestVoteReply {
	n.mu.Lock()
	defer n.mu.Unlock()
	reply := RequestVoteReply{Term: n.currentTerm, VoteGranted: false}
	if args.Term < n.currentTerm {
		return reply
	}
	if args.Term > n.currentTerm {
		n.currentTerm = args.Term
		n.votedFor = -1
		n.state = "follower"
	}
	// simple up-to-date check: compare last log term/index
	lastIdx := n.lastIndex()
	lastTerm := n.termAt(lastIdx)
	upToDate := (args.LastLogTerm > lastTerm) || (args.LastLogTerm == lastTerm && args.LastLogIndex >= lastIdx)
	if (n.votedFor == -1 || n.votedFor == args.CandidateId) && upToDate {
		n.votedFor = args.CandidateId
		reply.VoteGranted = true
		// reset election timer
		select {
		case n.electionResetEvt <- struct{}{}:
		default:
		}
	}
	reply.Term = n.currentTerm
	return reply
}

func (n *Node) handleAppendEntriesRemote(args AppendEntriesArgs) AppendEntriesReply {
	n.mu.Lock()
	defer n.mu.Unlock()
	reply := AppendEntriesReply{Term: n.currentTerm, Success: false}
	// 1) Reply false if term < currentTerm
	if args.Term < n.currentTerm {
		return reply
	}
	// 2) If term > currentTerm, become follower
	if args.Term > n.currentTerm {
		n.currentTerm = args.Term
		n.votedFor = -1
		n.state = "follower"
	}
	// reset election timer
	select {
	case n.electionResetEvt <- struct{}{}:
	default:
	}

	// 3) Check log contains PrevLogIndex with matching term
	if args.PrevLogIndex > n.lastIndex() || n.termAt(args.PrevLogIndex) != args.PrevLogTerm {
		// mismatch, reject
		return reply
	}
	// 4) Append new entries (overwrite conflicts)
	// entries correspond to positions starting at PrevLogIndex+1
	insertIdx := args.PrevLogIndex + 1
	i := 0
	for ; i < len(args.Entries); i++ {
		if insertIdx+i <= n.lastIndex() {
			// if existing entry term differs, truncate and append rest
			if n.termAt(insertIdx+i) != args.Entries[i].Term {
				// truncate
				n.log = n.log[:insertIdx+i]
				break
			}
		} else {
			break
		}
	}
	// append remaining entries
	for ; i < len(args.Entries); i++ {
		n.log = append(n.log, args.Entries[i])
	}
	// 5) Update commitIndex
	if args.LeaderCommit > n.commitIndex {
		n.commitIndex = min(args.LeaderCommit, n.lastIndex())
	}
	reply.Success = true
	reply.Term = n.currentTerm
	return reply
}

// ---------- Election and leader routines ----------
func (n *Node) run() {
	defer n.wg.Done()
	// randomized seed per node
	r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(n.id)))
	for {
		// randomized election timeout (150-300ms)
		timeout := time.Duration(150+r.Intn(150)) * time.Millisecond
		timer := time.NewTimer(timeout)
		select {
		case <-timer.C:
			n.mu.Lock()
			if n.state != "leader" {
				n.mu.Unlock()
				n.startElection()
			} else {
				n.mu.Unlock()
			}
		case <-n.electionResetEvt:
			if !timer.Stop() {
				<-timer.C
			}
			// loop to pick a new timeout
		case <-n.stopCh:
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			return
		}
	}
}

func (n *Node) startElection() {
	n.mu.Lock()
	n.state = "candidate"
	n.currentTerm += 1
	termStarted := n.currentTerm
	n.votedFor = n.id
	lastIdx := n.lastIndex()
	lastTerm := n.termAt(lastIdx)
	n.mu.Unlock()

	fmt.Printf("[Node %d] CANDIDATE term=%d\n", n.id, termStarted)

	votes := 1
	voteMu := sync.Mutex{}
	majority := (len(n.peers)+1)/2 + 1

	replyCh := make(chan RequestVoteReply, len(n.peers))
	for _, peer := range n.peers {
		args := RequestVoteArgs{
			Term:         termStarted,
			CandidateId:  n.id,
			LastLogIndex: lastIdx,
			LastLogTerm:  lastTerm,
		}
		n.net.SendRequestVote(peer, args, replyCh)
	}

	electionTimeout := 200 * time.Millisecond
	timer := time.NewTimer(electionTimeout)
	defer timer.Stop()

	for {
		select {
		case rep := <-replyCh:
			// if term higher, step down
			n.mu.Lock()
			if rep.Term > n.currentTerm {
				n.currentTerm = rep.Term
				n.votedFor = -1
				n.state = "follower"
				// reset election timer
				select {
				case n.electionResetEvt <- struct{}{}:
				default:
				}
				n.mu.Unlock()
				return
			}
			n.mu.Unlock()

			if rep.VoteGranted {
				voteMu.Lock()
				votes++
				fmt.Printf("[Node %d] got vote (%d/%d) term=%d\n", n.id, votes, majority, termStarted)
				n.mu.Lock()
				isCandidate := (n.state == "candidate" && n.currentTerm == termStarted)
				n.mu.Unlock()
				if votes >= majority && isCandidate {
					n.mu.Lock()
					n.becomeLeader()
					n.mu.Unlock()
					voteMu.Unlock()
					return
				}
				voteMu.Unlock()
			}
		case <-timer.C:
			// election failed, will retry in run loop
			return
		case <-n.stopCh:
			return
		}
	}
}

func (n *Node) becomeLeader() {
	n.state = "leader"
	// initialize leader state
	N := len(n.peers) + 1
	n.nextIndex = make([]int, N) // indexed by peer index in peers array (0..len(peers)-1)
	n.matchIndex = make([]int, N)
	last := n.lastIndex()
	// note: we map peers array index -> node id. We will store arrays length len(peers)
	for i := range n.nextIndex {
		n.nextIndex[i] = last + 1
		n.matchIndex[i] = 0
	}
	fmt.Printf("[Node %d] BECOMES LEADER term=%d\n", n.id, n.currentTerm)
	// start leader replication loop (heartbeats + replication)
	go n.leaderLoop()
}

// leaderLoop periodically sends AppendEntries to followers
func (n *Node) leaderLoop() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			n.mu.Lock()
			if n.state != "leader" {
				n.mu.Unlock()
				return
			}
			// send AppendEntries to all peers
			for idx, peer := range n.peers {
				// compute prevIndex and entries slice
				next := n.nextIndex[idx]
				prevIndex := next - 1
				prevTerm := n.termAt(prevIndex)
				entries := make([]LogEntry, 0)
				if next <= n.lastIndex() {
					// copy entries from next..end
					entries = append(entries, n.log[next:]...)
				}
				args := AppendEntriesArgs{
					Term:         n.currentTerm,
					LeaderId:     n.id,
					PrevLogIndex: prevIndex,
					PrevLogTerm:  prevTerm,
					Entries:      entries,
					LeaderCommit: n.commitIndex,
				}
				// create reply channel sized 1 to avoid blocking
				rc := make(chan AppendEntriesReply, 1)
				n.net.SendAppendEntries(peer, args, rc)
				// handle reply asynchronously
				go func(peerIdx int, peerId int, args AppendEntriesArgs, rc chan AppendEntriesReply) {
					select {
					case rep := <-rc:
						n.mu.Lock()
						// if term larger, step down
						if rep.Term > n.currentTerm {
							n.currentTerm = rep.Term
							n.votedFor = -1
							n.state = "follower"
							// reset election timer
							select {
							case n.electionResetEvt <- struct{}{}:
							default:
							}
							n.mu.Unlock()
							return
						}
						if rep.Success {
							// update nextIndex and matchIndex for this follower
							newMatch := args.PrevLogIndex + len(args.Entries)
							n.matchIndex[peerIdx] = newMatch
							n.nextIndex[peerIdx] = newMatch + 1
							// try to advance commitIndex
							n.advanceCommitIndex()
						} else {
							// simple backoff: decrement nextIndex for follower (naive but works for sim)
							if n.nextIndex[peerIdx] > 1 {
								n.nextIndex[peerIdx]--
							}
						}
						n.mu.Unlock()
					case <-time.After(200 * time.Millisecond):
						// peer didn't reply in time; ignore this round
					}
				}(idx, peer, args, rc)
			}
			n.mu.Unlock()
		case <-n.stopCh:
			return
		}
	}
}

// advanceCommitIndex: leader updates commitIndex when a log index is replicated on majority
func (n *Node) advanceCommitIndex() {
	// must be called under lock
	N := n.lastIndex()
	for idx := n.commitIndex + 1; idx <= N; idx++ {
		// only consider entries from current term (Raft safety)
		if n.termAt(idx) != n.currentTerm {
			continue
		}
		// count replicas with matchIndex >= idx (include leader)
		count := 1
		for i := range n.matchIndex {
			if n.matchIndex[i] >= idx {
				count++
			}
		}
		majority := (len(n.peers)+1)/2 + 1
		if count >= majority {
			n.commitIndex = idx
		}
	}
}

// ---------- Client-facing: start a command (sent to any node) ----------
func (n *Node) StartCommand(cmd string) bool {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.state != "leader" {
		return false
	}
	// leader appends command to its log
	entry := LogEntry{Term: n.currentTerm, Command: cmd}
	n.log = append(n.log, entry)
	// immediately update own matchIndex (leader has the entry)
	// leader index mapping: peers array indices correspond to followers only;
	// we don't store self's matchIndex in same array; count leader implicitly.
	// replication loop will send to followers and update their matchIndex/nextIndex.
	return true
}

// ---------- Applier loop: applies committed but not-yet-applied entries ----------
func (n *Node) applierLoop() {
	defer n.wg.Done()
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			n.mu.Lock()
			for n.lastApplied < n.commitIndex {
				n.lastApplied++
				entry := n.log[n.lastApplied]
				// apply to state machine (here: send to applyCh to print)
				n.applyCh <- ApplyMsg{Index: n.lastApplied, Command: entry.Command}
			}
			n.mu.Unlock()
		case <-n.stopCh:
			return
		}
	}
}

// ---------- Utility: minimal helpers ----------
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ---------- main: set up simulation -----------------
func main() {
	seed := time.Now().UnixNano()
	net := NewNetwork(seed, 0.08, 15*time.Millisecond, 120*time.Millisecond) // 8% drop

	numNodes := 5
	ids := make([]int, 0, numNodes)
	for i := 0; i < numNodes; i++ {
		ids = append(ids, i)
	}

	applyCh := make(chan ApplyMsg, 100)
	nodes := make([]*Node, 0, numNodes)
	for i := 0; i < numNodes; i++ {
		peers := make([]int, 0, numNodes-1)
		for _, id := range ids {
			if id == i {
				continue
			}
			peers = append(peers, id)
		}
		node := NewNode(i, peers, net, applyCh)
		nodes = append(nodes, node)
	}

	// start nodes
	for _, n := range nodes {
		n.Start()
	}

	// applier printer goroutine
	stopPrint := make(chan struct{})
	go func() {
		for {
			select {
			case ap := <-applyCh:
				fmt.Printf("APPLY idx=%d cmd=%s\n", ap.Index, ap.Command)
			case <-stopPrint:
				return
			}
		}
	}()

	// simulated clients: every 500ms pick a random node and try to submit a command
	clientRand := rand.New(rand.NewSource(seed + 12345))
	clientTicker := time.NewTicker(500 * time.Millisecond)
	clientStop := time.After(12 * time.Second)
	cmdId := 1
loop:
	for {
		select {
		case <-clientTicker.C:
			target := nodes[clientRand.Intn(len(nodes))]
			cmd := fmt.Sprintf("set-x-%d", cmdId)
			ok := target.StartCommand(cmd)
			if ok {
				fmt.Printf("[Client] command %s accepted by Node %d (term=%d)\n", cmd, target.id, target.currentTerm)
			} else {
				// not leader; client retry is naive (we just dropped it)
				// in production client would discover leader and retry
				fmt.Printf("[Client] command %s rejected by Node %d (not leader)\n", cmd, target.id)
			}
			cmdId++
		case <-clientStop:
			break loop
		}
	}

	// shutdown
	clientTicker.Stop()
	close(stopPrint)
	for _, n := range nodes {
		n.Stop()
	}
	fmt.Println("Simulation ended.")
}
