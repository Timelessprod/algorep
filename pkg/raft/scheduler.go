package raft

import (
	"math/rand"
	"sync"
	"time"

	"go.uber.org/zap"
)

type State int

const (
	FollowerState = iota
	CandidateState
	LeaderState
)

func (s State) String() string {
	return [...]string{"Follower", "Candidate", "Leader"}[s]
}

type SchedulerNode struct {
	Id          uint32
	Card        NodeCard
	State       State
	CurrentTerm uint32

	VotedFor        int32
	ElectionTimeout time.Duration
	VoteCount       uint32

	Channel ChannelContainer

	IsStarted bool
	IsCrashed bool
}

/**************
 ** Function **
 **************/

func SendWithLatency[MessageType RPCType](FromNode NodeCard, ToNode NodeCard, channel chan MessageType, message MessageType) {
	latencyFromNode := 0 * time.Millisecond
	if FromNode.Type == SchedulerNodeType {
		latencyFromNode = Config.NodeSpeedList[FromNode.Id]
	}

	latencyToNode := 0 * time.Millisecond
	if ToNode.Type == SchedulerNodeType {
		latencyToNode = Config.NodeSpeedList[ToNode.Id]
	}

	latency := latencyFromNode + latencyToNode
	time.AfterFunc(latency, func() {
		channel <- message
	})
}

/*************
 ** Methods **
 *************/

func (node *SchedulerNode) Init(id uint32) SchedulerNode {
	node.Id = id
	node.Card = NodeCard{Id: id, Type: SchedulerNodeType}
	node.State = FollowerState
	node.VotedFor = -1
	node.VoteCount = 0

	// Compute random leaderIsAliveTimeout
	durationRange := int(Config.MaxElectionTimeout - Config.MinElectionTimeout)
	node.ElectionTimeout = time.Duration(rand.Intn(durationRange)) + Config.MinElectionTimeout

	node.Channel.RequestCommand = make(chan RequestCommandRPC, Config.ChannelBufferSize)
	node.Channel.ResponseCommand = make(chan ResponseCommandRPC, Config.ChannelBufferSize)
	node.Channel.RequestVote = make(chan RequestVoteRPC, Config.ChannelBufferSize)
	node.Channel.ResponseVote = make(chan ResponseVoteRPC, Config.ChannelBufferSize)

	logger.Info("Node initialized", zap.Uint32("id", id))
	return *node
}

func (node *SchedulerNode) getTimeOut() time.Duration {
	switch node.State {
	case FollowerState:
		return node.ElectionTimeout
	case CandidateState:
		return node.ElectionTimeout
	case LeaderState:
		return Config.IsAliveNotificationInterval
	}
	logger.Panic("Invalid node state", zap.String("Node", node.Card.String()), zap.Int("state", int(node.State)))
	panic("Invalid node state")
}

func (node *SchedulerNode) broadcastRequestVote() {
	for i := uint32(0); i < Config.SchedulerNodeCount; i++ {
		if i != node.Id {
			channel := Config.NodeChannelMap[SchedulerNodeType][i].RequestVote
			request := RequestVoteRPC{
				FromNode:    node.Card,
				ToNode:      NodeCard{Id: i, Type: SchedulerNodeType},
				Term:        node.CurrentTerm,
				CandidateId: node.Id,
			}
			channel <- request
		}
	}
}

func (node *SchedulerNode) startNewElection() {
	logger.Info("Start new election", zap.String("Node", node.Card.String()))
	node.State = CandidateState
	node.VoteCount = 1
	node.CurrentTerm++
	node.VotedFor = int32(node.Id)
	node.broadcastRequestVote()
}

/************
 ** Handle **
 ************/

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

func (node *SchedulerNode) handleCrashCommand() {
	if node.IsCrashed {
		logger.Debug("Node already crashed",
			zap.String("Node", node.Card.String()),
		)
		return
	} else {
		node.IsCrashed = true
	}
}

func (node *SchedulerNode) handleSynchronizeCommand(request RequestCommandRPC) {
	if node.IsCrashed {
		logger.Debug("Node is crashed. Ignore synchronize command",
			zap.String("Node", node.Card.String()),
		)
		return
	}
	// TODO
}

func (node *SchedulerNode) handleAppendEntryCommand(request RequestCommandRPC) {
	if node.IsCrashed {
		logger.Debug("Node is crashed. Ignore AppendEntry command",
			zap.String("Node", node.Card.String()),
		)
		return
	}
	// TODO
}

// Handle Request Command RPC
func (node *SchedulerNode) handleRequestCommandRPC(request RequestCommandRPC) {
	logger.Debug("Handle Request Command RPC",
		zap.String("FromNode", request.FromNode.String()),
		zap.String("ToNode", request.ToNode.String()),
		zap.String("CommandType", request.CommandType.String()),
	)
	switch request.CommandType {
	case SynchronizeCommand:
		node.handleSynchronizeCommand(request)
	case AppendEntryCommand:
		node.handleAppendEntryCommand(request)
	case StartCommand:
		node.handleStartCommand()
	case CrashCommand:
		node.handleCrashCommand()
	case RecoverCommand:
		// TODO
	}
}

// Handle Response Command RPC
func (node *SchedulerNode) handleResponseCommandRPC(response ResponseCommandRPC) {
	logger.Debug("Handle Response Command RPC",
		zap.String("FromNode", response.FromNode.String()),
		zap.String("ToNode", response.ToNode.String()),
		zap.Uint32("term", response.Term),
		zap.Bool("success", response.Success),
	)
	//TODO if majority of success, commit
}

// Handle Request Vote RPC
func (node *SchedulerNode) handleRequestVoteRPC(request RequestVoteRPC) {
	if node.IsCrashed {
		logger.Debug("Node is crashed. Ignore request vote RPC",
			zap.String("FromNode", request.FromNode.String()),
			zap.String("ToNode", request.ToNode.String()),
		)
		return
	}
	logger.Debug("Handle Request Vote RPC",
		zap.String("FromNode", request.FromNode.String()),
		zap.String("ToNode", request.ToNode.String()),
		zap.Int("CandidateId", int(request.CandidateId)),
	)
	channel := Config.NodeChannelMap[SchedulerNodeType][request.FromNode.Id].ResponseVote
	response := ResponseVoteRPC{
		FromNode:    request.ToNode,
		ToNode:      request.FromNode,
		Term:        node.CurrentTerm,
		VoteGranted: false,
	}

	if node.CurrentTerm < request.Term {
		logger.Debug("Candidate term is higher than current term. Vote granted !",
			zap.String("id", node.Card.String()),
			zap.Uint32("candidateTerm", request.Term),
			zap.Uint32("currentTerm", node.CurrentTerm),
		)
		node.VotedFor = int32(request.CandidateId)
		response.VoteGranted = true
		node.CurrentTerm = request.Term
		node.State = FollowerState
	}
	channel <- response
	// TODO check if the candidate is up to date (log)
}

// Handle Response Vote RPC
func (node *SchedulerNode) handleResponseVoteRPC(response ResponseVoteRPC) {
	if node.IsCrashed {
		logger.Debug("Node is crashed. Ignore request vote RPC",
			zap.String("FromNode", response.FromNode.String()),
			zap.String("ToNode", response.ToNode.String()),
		)
		return
	}
	logger.Debug("Handle Response Vote RPC",
		zap.String("FromNode", response.FromNode.String()),
		zap.String("ToNode", response.ToNode.String()),
		zap.Int("CandidateId", int(response.ToNode.Id)),
	)
	if node.State != CandidateState {
		logger.Debug("Node is not a candidate. Ignore response vote RPC",
			zap.String("Node", node.Card.String()),
			zap.String("state", node.State.String()),
		)
	} else {
		if response.VoteGranted {
			node.VoteCount++
			if node.VoteCount > Config.SchedulerNodeCount/2 {
				node.State = LeaderState
				logger.Info("Leader elected", zap.String("Node", node.Card.String()))
				return
			}
		} else {
			logger.Warn("Vote not granted",
				zap.String("Node", node.Card.String()),
				zap.Uint32("ResponseTerm", response.Term),
				zap.Uint32("NodeTerm", node.CurrentTerm),
			)
		}
	}
	if response.Term > node.CurrentTerm {
		node.CurrentTerm = response.Term
		node.State = FollowerState
	}
}

// Handle Is Alive Notification RPC
func (node *SchedulerNode) broadcastSynchronizeCommandRPC() {
	for i := uint32(0); i < Config.SchedulerNodeCount; i++ {
		if i != node.Id {
			channel := Config.NodeChannelMap[SchedulerNodeType][i].RequestCommand
			request := RequestCommandRPC{
				FromNode:    node.Card,
				ToNode:      NodeCard{Id: i, Type: SchedulerNodeType},
				Term:        node.CurrentTerm,
				CommandType: SynchronizeCommand,
			}
			channel <- request
		}
	}
}

func (node *SchedulerNode) handleTimeout() {
	if node.IsCrashed {
		logger.Debug("Node is crashed. Ignore timeout", zap.String("Node", node.Card.String()))
		return
	}
	switch node.State {
	case FollowerState:
		logger.Warn("Leader does not respond", zap.String("Node", node.Card.String()), zap.Duration("electionTimeout", node.ElectionTimeout))
		node.startNewElection()
	case CandidateState:
		logger.Warn("Too much time to get a majority vote", zap.String("Node", node.Card.String()), zap.Duration("electionTimeout", node.ElectionTimeout))
		node.startNewElection()
	case LeaderState:
		logger.Info("It's time for the Leader to send an IsAlive notification to followers", zap.String("Node", node.Card.String()))
		node.broadcastSynchronizeCommandRPC()
	}
}

/*********
 ** Run **
 *********/

func (node *SchedulerNode) Run(wg *sync.WaitGroup) {
	defer wg.Done()
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
	}
}
