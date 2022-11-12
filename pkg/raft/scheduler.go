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
	State       State
	CurrentTerm uint32

	VotedFor        int32
	ElectionTimeout time.Duration
	VoteCount       uint32

	Channel ChannelContainer
}

/*************
 ** Methods **
 *************/

func (node *SchedulerNode) Init(id uint32) SchedulerNode {
	node.Id = id
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
	logger.Panic("Invalid node state", zap.Uint32("id", node.Id), zap.Int("state", int(node.State)))
	panic("Invalid node state")
}

func (node *SchedulerNode) broadcastRequestVote() {
	for i := uint32(0); i < Config.SchedulerNodeCount; i++ {
		if i != node.Id {
			channel := Config.NodeChannelMap[SchedulerNodeType][i].RequestVote
			request := RequestVoteRPC{
				FromNode:    node.Id,
				ToNode:      i,
				Term:        node.CurrentTerm,
				CandidateId: node.Id,
			}
			channel <- request
		}
	}
}

func (node *SchedulerNode) startNewElection() {
	logger.Info("Start new election", zap.Uint32("id", node.Id))
	node.State = CandidateState
	node.VoteCount = 1
	node.CurrentTerm++
	node.VotedFor = int32(node.Id)
	node.broadcastRequestVote()
}

/************
 ** Handle **
 ************/

// Handle Request Command RPC
func (node *SchedulerNode) handleRequestCommandRPC(request RequestCommandRPC) {
	logger.Debug("Handle Request Command RPC",
		zap.Uint32("FromNode", request.FromNode),
		zap.Uint32("ToNode", request.ToNode),
		zap.String("CommandType", request.CommandType.String()),
	)
	//TODO Manage command (add Entry/Sync log / conflict resolution / success / failure)
}

// Handle Response Command RPC
func (node *SchedulerNode) handleResponseCommandRPC(response ResponseCommandRPC) {
	logger.Debug("Handle Response Command RPC",
		zap.Uint32("FromNode", response.FromNode),
		zap.Uint32("ToNode", response.ToNode),
		zap.Uint32("term", response.Term),
		zap.Bool("success", response.Success),
	)
	//TODO if majority of success, commit
}

// Handle Request Vote RPC
func (node *SchedulerNode) handleRequestVoteRPC(request RequestVoteRPC) {
	logger.Debug("Handle Request Vote RPC",
		zap.Uint32("FromNode", request.FromNode),
		zap.Uint32("ToNode", request.ToNode),
		zap.Int("CandidateId", int(request.CandidateId)),
	)
	channel := Config.NodeChannelMap[SchedulerNodeType][request.FromNode].ResponseVote
	response := ResponseVoteRPC{
		FromNode:    request.ToNode,
		ToNode:      request.FromNode,
		Term:        node.CurrentTerm,
		VoteGranted: false,
	}

	if node.CurrentTerm < request.Term {
		logger.Debug("Candidate term is higher than current term. Vote granted !",
			zap.Uint32("id", node.Id),
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
	logger.Debug("Handle Response Vote RPC",
		zap.Uint32("FromNode", response.FromNode),
		zap.Uint32("ToNode", response.ToNode),
		zap.Int("CandidateId", int(response.ToNode)),
	)
	if node.State != CandidateState {
		logger.Debug("Node is not a candidate. Ignore response vote RPC",
			zap.Uint32("id", node.Id),
			zap.String("state", node.State.String()),
		)
	} else {
		if response.VoteGranted {
			node.VoteCount++
			if node.VoteCount > Config.SchedulerNodeCount/2 {
				node.State = LeaderState
				logger.Info("Leader elected", zap.Uint32("id", node.Id))
				return
			}
		} else {
			logger.Warn("Vote not granted",
				zap.Uint32("id", node.Id),
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
				FromNode:    node.Id,
				ToNode:      i,
				Term:        node.CurrentTerm,
				CommandType: Synchronize,
			}
			channel <- request
		}
	}
}

func (node *SchedulerNode) handleTimeout() {
	switch node.State {
	case FollowerState:
		logger.Warn("Leader does not respond", zap.Uint32("id", node.Id), zap.Duration("electionTimeout", node.ElectionTimeout))
		node.startNewElection()
	case CandidateState:
		logger.Warn("Too much time to get a majority vote", zap.Uint32("id", node.Id), zap.Duration("electionTimeout", node.ElectionTimeout))
		node.startNewElection()
	case LeaderState:
		logger.Info("It's time for the Leader to send an IsAlive notification to followers", zap.Uint32("id", node.Id))
		node.broadcastSynchronizeCommandRPC()
	}
}

/*********
 ** Run **
 *********/

func (node *SchedulerNode) Run(wg *sync.WaitGroup) {
	defer wg.Done()
	logger.Info("Node is waiting the START command from REPL", zap.Uint32("id", node.Id))
	// TODO wait for the start command from REPL and listen to the channel RequestCommand
	logger.Info("Node started", zap.Uint32("id", node.Id))

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
