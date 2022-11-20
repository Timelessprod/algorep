package scheduler

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/Timelessprod/algorep/pkg/core"
	"github.com/Timelessprod/algorep/pkg/utils"
	"go.uber.org/zap"
)

var logger *zap.Logger = core.Logger

/********************
 ** Scheduler Node **
 ********************/

// SchedulerNode is the node in charge of scheduling the database entries with the RAFT algorithm
type SchedulerNode struct {
	Id          uint32
	Card        core.NodeCard
	State       core.State
	CurrentTerm uint32
	LeaderId    int

	VotedFor        int32
	ElectionTimeout time.Duration
	VoteCount       uint32

	// Each entry contains command for state machine
	// and term when entry was received by leader (first index is 1)
	log map[uint32]core.Entry
	// Job id counter
	jobIdCounter uint32
	// Index of highest log entry known to be committed (initialized to 0, increases monotonically)
	commitIndex uint32
	// Index of highest log entry known to be replicated on other nodes (initialized to 0, increases monotonically)
	matchIndex []uint32
	// Index of highest log entry available to store next entry (initialized to 1, increases monotonically)
	nextIndex []uint32
	// Index of highest log entry applied to state machine (initialized to 0, increases monotonically)
	lastApplied uint32

	// StateMachine
	StateMachine StateMachine

	Channel core.ChannelContainer

	IsStarted bool
	IsCrashed bool

	// State file to debug the node state
	StateFile *os.File
}

// Init the scheduler node
func (node *SchedulerNode) Init(id uint32) SchedulerNode {
	node.Id = id
	node.Card = core.NodeCard{Id: id, Type: core.SchedulerNodeType}
	node.State = core.FollowerState
	node.LeaderId = core.NO_NODE
	node.VotedFor = core.NO_NODE
	node.VoteCount = 0

	// Compute random leaderIsAliveTimeout
	durationRange := int(core.Config.MaxElectionTimeout - core.Config.MinElectionTimeout)
	node.ElectionTimeout = time.Duration(rand.Intn(durationRange)) + core.Config.MinElectionTimeout

	// Initialize all elements used to store and replicate the log
	node.log = make(map[uint32]core.Entry)
	node.jobIdCounter = 0
	node.commitIndex = 0
	node.matchIndex = make([]uint32, core.Config.SchedulerNodeCount)
	for i := range node.matchIndex {
		node.matchIndex[i] = 0
	}
	node.nextIndex = make([]uint32, core.Config.SchedulerNodeCount)
	for i := range node.nextIndex {
		node.nextIndex[i] = 1
	}
	node.lastApplied = 0

	// Initialize the state machine
	node.StateMachine = StateMachine{}
	node.StateMachine.Init()

	// Initialize the channel container
	node.Channel.RequestCommand = make(chan core.RequestCommandRPC, core.Config.ChannelBufferSize)
	node.Channel.ResponseCommand = make(chan core.ResponseCommandRPC, core.Config.ChannelBufferSize)
	node.Channel.RequestVote = make(chan core.RequestVoteRPC, core.Config.ChannelBufferSize)
	node.Channel.ResponseVote = make(chan core.ResponseVoteRPC, core.Config.ChannelBufferSize)

	// Initialize the state file
	node.InitStateInFile()

	logger.Info("Node initialized", zap.Uint32("id", id))
	return *node
}

// Run the scheduler node
func (node *SchedulerNode) Run(wg *sync.WaitGroup) {
	defer wg.Done()
	defer node.StateFile.Close()
	logger.Info("Node is waiting the START command from REPL", zap.String("Node", node.Card.String()))
	// Wait for the start command from REPL and listen to the channel RequestCommand
	for !node.IsStarted {
		select {
		case request := <-node.Channel.RequestCommand:
			node.handleRequestCommandRPC(request)
		}
	}
	logger.Info("Node started", zap.String("Node", node.Card.String()))

	for {
		select {
		case request := <-node.Channel.RequestCommand:
			node.handleRequestCommandRPC(request)
		case response := <-node.Channel.ResponseCommand:
			node.handleResponseCommandRPC(response)
		case request := <-node.Channel.RequestVote:
			node.handleRequestVoteRPC(request)
		case response := <-node.Channel.ResponseVote:
			node.handleResponseVoteRPC(response)
		case <-time.After(node.getTimeOut()):
			node.handleTimeout()
		}
		node.printNodeStateInFile()
		node.updateCommitIndex()
		node.updateStateMachine()
		time.Sleep(core.Config.NodeSpeedList[node.Id])
	}
}

func (node *SchedulerNode) InitStateInFile() {
	path := fmt.Sprintf("state/%d.node", node.Id)
	os.MkdirAll(filepath.Dir(path), os.ModePerm)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.Error("Error while creating state file",
			zap.String("Node", node.Card.String()),
			zap.Error(err),
		)
		return
	}
	node.StateFile = f
}

func (node *SchedulerNode) printNodeStateInFile() {
	if node.StateFile == nil {
		return
	}
	f := node.StateFile
	f.Truncate(0)
	f.Seek(0, 0)
	fmt.Fprintln(f, "--- ", node.Card.String(), " ---")
	fmt.Fprintln(f, ">>> State: ", node.State)
	fmt.Fprintln(f, ">>> IsCrashed: ", node.IsCrashed)
	fmt.Fprintln(f, ">>> CurrentTerm: ", node.CurrentTerm)
	fmt.Fprintln(f, ">>> LeaderId: ", node.LeaderId)
	fmt.Fprintln(f, ">>> VotedFor: ", node.VotedFor)
	fmt.Fprintln(f, ">>> ElectionTimeout: ", node.ElectionTimeout)
	fmt.Fprintln(f, ">>> VoteCount: ", node.VoteCount)
	fmt.Fprintln(f, ">>> CommitIndex: ", node.commitIndex)
	fmt.Fprintln(f, ">>> MatchIndex: ", node.matchIndex)
	fmt.Fprintln(f, ">>> NextIndex: ", node.nextIndex)
	fmt.Fprintln(f, "### Log ###")
	for i := 1; i <= len(node.log); i++ {
		entry := node.log[uint32(i)]
		fmt.Fprintf(f, "[%v] Job %v | Worker %v | %v\n", i, entry.Job.GetReference(), entry.Job.WorkerId, entry.Job.Status.String())
	}
	fmt.Fprintln(f, "----------------")
}

// Add a new entry to the log
func (node *SchedulerNode) addEntryToLog(entry core.Entry) {
	index := node.nextIndex[node.Card.Id]
	node.log[index] = entry
	node.nextIndex[node.Card.Id] = index + 1
}

// startElection starts an new election in sending a RequestVote RPC to all the nodes
func (node *SchedulerNode) startNewElection() {
	logger.Info("Start new election", zap.String("Node", node.Card.String()))
	node.State = core.CandidateState
	node.VoteCount = 1
	node.CurrentTerm++
	node.VotedFor = int32(node.Id)
	node.broadcastRequestVote()
}

// handleStartCommand starts the node when it receives a StartCommand
func (node *SchedulerNode) handleStartCommand() {
	if node.IsStarted {
		logger.Debug("Node already started",
			zap.String("Node", node.Card.String()),
		)
		return
	} else {
		node.IsStarted = true
	}
}

// handleCrashCommand crashes the node when it receives a CrashCommand
func (node *SchedulerNode) handleCrashCommand() {
	if node.IsCrashed {
		logger.Debug("Node is already crashed",
			zap.String("Node", node.Card.String()),
		)
		return
	} else {
		node.IsCrashed = true
	}
}

// handleRecoversCommand recovers the node after crash when it receives a RecoverCommand
func (node *SchedulerNode) handleRecoverCommand() {
	if node.IsCrashed {
		node.IsCrashed = false
	} else {
		logger.Debug("Node is not crashed",
			zap.String("Node", node.Card.String()),
		)
		return
	}
}

// becomeLeader sets the node as leader
func (node *SchedulerNode) becomeLeader() {
	node.State = core.LeaderState
	node.LeaderId = int(node.Card.Id)
	logger.Info("Leader elected", zap.String("Node", node.Card.String()))
	for nodeId := uint32(0); nodeId < core.Config.SchedulerNodeCount; nodeId++ {
		node.nextIndex[nodeId] = uint32(len(node.log)) + 1
	}
	node.jobIdCounter = 0
}

// updateTerm updates the term of the node if the term is higher than the current term
func (node *SchedulerNode) updateTerm(term uint32) {
	if term > node.CurrentTerm {
		logger.Debug("Update term, reset vote and change state to follower",
			zap.String("Node", node.Card.String()),
			zap.Uint32("CurrentTerm", node.CurrentTerm),
			zap.Uint32("NewTerm", term),
			zap.String("OldState", node.State.String()),
		)
		node.CurrentTerm = term
		node.State = core.FollowerState
		node.VotedFor = core.NO_NODE
	}
}

// checkVote checks if the node has already voted for the candidate
func (node *SchedulerNode) checkVote(candidateId uint32) bool {
	if node.VotedFor == core.NO_NODE || uint32(node.VotedFor) == candidateId {
		return true
	}
	return false
}

// LogTerm returns the term of the log entry at index i, or 0 if no such entry exists
func (node *SchedulerNode) LogTerm(i uint32) uint32 {
	if i < 1 || i > uint32(len(node.log)) {
		return 0
	}
	return node.log[i].Term
}

// updateCommitIndex updates the commit index of the node
func (node *SchedulerNode) updateCommitIndex() {
	if node.State != core.LeaderState {
		return
	}

	// Find the largest number M such that a majority of nodes has matchIndex[i] ≥ M
	matchIndexMedianList := make([]uint32, len(node.matchIndex)+1)
	copy(matchIndexMedianList, node.matchIndex)
	matchIndexMedianList = append(matchIndexMedianList, uint32(len(node.log)))
	sort.Slice(matchIndexMedianList, func(i, j int) bool { return matchIndexMedianList[i] < matchIndexMedianList[j] })
	median := matchIndexMedianList[core.Config.SchedulerNodeCount/2]

	if node.LogTerm(median) == node.CurrentTerm {
		node.commitIndex = median
	}
}

// updateStateMachine updates the state machine of the node
func (node *SchedulerNode) updateStateMachine() {
	if node.lastApplied >= node.commitIndex {
		return
	}

	logger.Debug("Update state machine",
		zap.String("Node", node.Card.String()),
		zap.Uint32("LastApplied", node.lastApplied),
		zap.Uint32("CommitIndex", node.commitIndex),
	)

	for i := node.lastApplied + 1; i <= node.commitIndex; i++ {
		entry := node.log[i]
		node.StateMachine.Apply(entry)

		// Propagate the job to the worker if the job is new
		if node.State == core.LeaderState && entry.Type == core.OpenJob {
			node.sendJobToWorker(&entry.Job)
		}
	}
	node.lastApplied = node.commitIndex
}

// Send a job to the worker
func (node *SchedulerNode) sendJobToWorker(job *core.Job) {
	workerId := job.WorkerId
	queue := core.Config.NodeChannelMap[core.WorkerNodeType][workerId].JobQueue
	queue <- *job
}

// GetJobId generates a new job id and increments the job id counter
func (node *SchedulerNode) GetJobId() uint32 {
	node.jobIdCounter++
	return node.jobIdCounter
}

// GetWorkerId finds the appropriate worker id for the job (the worker with the lowest load)
func (node *SchedulerNode) GetWorkerId() uint32 {
	// the number of jobs in the queue for each worker
	jobCount := make([]uint32, core.Config.WorkerNodeCount)
	for i, container := range core.Config.NodeChannelMap[core.WorkerNodeType] {
		jobCount[i] = uint32(len(container.JobQueue))
	}
	// get the worker id with the lowest number of jobs in the queue
	return utils.IndexMinUint32(jobCount)
}
