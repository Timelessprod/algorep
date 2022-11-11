package main

import (
	"math/rand"
	"time"

	"go.uber.org/zap"
)

type ChannelContainer struct {
	requestCommand  chan RequestCommandRPC
	responseCommand chan ResponseCommandRPC

	requestVote  chan RequestVoteRPC
	responseVote chan ResponseVoteRPC
}

type State int

const (
	FollowerState = iota
	CandidateState
	LeaderState
)

func (s State) String() string {
	return [...]string{"Follower", "Candidate", "Leader"}[s]
}

type Node struct {
	id          uint32
	state       State
	currentTerm uint32

	votedFor        int32
	electionTimeout time.Duration
	voteCount       uint32

	channel ChannelContainer
}

/*************
 ** Methods **
 *************/

func (node *Node) init(id uint32) Node {
	node.id = id
	node.state = FollowerState
	node.votedFor = -1
	node.voteCount = 0

	// Compute random leaderIsAliveTimeout
	durationRange := int(config.maxElectionTimeout - config.minElectionTimeout)
	node.electionTimeout = time.Duration(rand.Intn(durationRange)) + config.minElectionTimeout

	node.channel.requestCommand = make(chan RequestCommandRPC, config.channelBufferSize)
	node.channel.responseCommand = make(chan ResponseCommandRPC, config.channelBufferSize)
	node.channel.requestVote = make(chan RequestVoteRPC, config.channelBufferSize)
	node.channel.responseVote = make(chan ResponseVoteRPC, config.channelBufferSize)

	logger.Info("Node initialized", zap.Uint32("id", id))
	return *node
}

func (node *Node) getTimeOut() time.Duration {
	switch node.state {
	case FollowerState:
		return node.electionTimeout
	case CandidateState:
		return node.electionTimeout
	case LeaderState:
		return config.IsAliveNotificationInterval
	}
	logger.Panic("Invalid node state", zap.Uint32("id", node.id), zap.Int("state", int(node.state)))
	panic("Invalid node state")
}

func (node *Node) broadcastRequestVote() {
	for i := uint32(0); i < config.nodeCount; i++ {
		if i != node.id {
			channel := config.nodeChannelList[i].requestVote
			request := RequestVoteRPC{
				fromNode:    node.id,
				toNode:      i,
				term:        node.currentTerm,
				candidateId: node.id,
			}
			channel <- request
		}
	}
}

func (node *Node) startNewElection() {
	logger.Info("Start new election", zap.Uint32("id", node.id))
	node.state = CandidateState
	node.voteCount = 1
	node.currentTerm++
	node.votedFor = int32(node.id)
	node.broadcastRequestVote()
}

/************
 ** Handle **
 ************/

// Handle Request Command RPC
func (node *Node) handleRequestCommandRPC(request RequestCommandRPC) {
	logger.Debug("Handle Request Command RPC",
		zap.Uint32("fromNode", request.fromNode),
		zap.Uint32("toNode", request.toNode),
	)
	//TODO
}

// Handle Response Command RPC
func (node *Node) handleResponseCommandRPC(response ResponseCommandRPC) {
	logger.Debug("Handle Response Command RPC",
		zap.Uint32("fromNode", response.fromNode),
		zap.Uint32("toNode", response.toNode),
	)
	//TODO
}

// Handle Request Vote RPC
func (node *Node) handleRequestVoteRPC(request RequestVoteRPC) {
	logger.Debug("Handle Request Vote RPC",
		zap.Uint32("fromNode", request.fromNode),
		zap.Uint32("toNode", request.toNode),
		zap.Int("candidateId", int(request.candidateId)),
	)
	channel := config.nodeChannelList[request.fromNode].responseVote
	response := ResponseVoteRPC{
		fromNode:    request.toNode,
		toNode:      request.fromNode,
		term:        node.currentTerm,
		voteGranted: false,
	}

	if node.currentTerm < request.term {
		logger.Debug("Candidate term is higher than current term. Vote granted !",
			zap.Uint32("id", node.id),
			zap.Uint32("candidateTerm", request.term),
			zap.Uint32("currentTerm", node.currentTerm),
		)
		node.votedFor = int32(request.candidateId)
		response.voteGranted = true
		node.currentTerm = request.term
	}
	channel <- response
	// TODO check if the candidate is up to date (log)
}

// Handle Response Vote RPC
func (node *Node) handleResponseVoteRPC(response ResponseVoteRPC) {
	logger.Debug("Handle Response Vote RPC",
		zap.Uint32("fromNode", response.fromNode),
		zap.Uint32("toNode", response.toNode),
		zap.Int("candidateId", int(response.toNode)),
	)
	if node.state != CandidateState {
		logger.Debug("Node is not a candidate. Ignore response vote RPC",
			zap.Uint32("id", node.id),
			zap.String("state", node.state.String()),
		)
	} else {
		if response.voteGranted {
			node.voteCount++
			if node.voteCount > config.nodeCount/2 {
				node.state = LeaderState
				logger.Info("Leader elected", zap.Uint32("id", node.id))
				return
			}
		} else {
			logger.Warn("Vote not granted",
				zap.Uint32("id", node.id),
				zap.Uint32("ResponseTerm", response.term),
				zap.Uint32("NodeTerm", node.currentTerm),
			)
		}
	}
	if response.term > node.currentTerm {
		node.currentTerm = response.term
		node.state = FollowerState
	}
}

func (node *Node) handleTimeout() {
	switch node.state {
	case FollowerState:
		logger.Warn("Leader does not respond", zap.Uint32("id", node.id), zap.Duration("electionTimeout", node.electionTimeout))
		node.startNewElection()
	case CandidateState:
		logger.Warn("Too much time to get a majority vote", zap.Uint32("id", node.id), zap.Duration("electionTimeout", node.electionTimeout))
		node.startNewElection()
	case LeaderState:
		logger.Info("It's time for the Leader to send an IsAlive notification to followers", zap.Uint32("id", node.id))
	}
}

/*********
 ** Run **
 *********/

func (node *Node) run() {
	logger.Info("Node started", zap.Uint32("id", node.id))
	defer wg.Done()

	for {
		select {
		case request := <-node.channel.requestCommand:
			node.handleRequestCommandRPC(request)
		case response := <-node.channel.responseCommand:
			node.handleResponseCommandRPC(response)
		case request := <-node.channel.requestVote:
			node.handleRequestVoteRPC(request)
		case response := <-node.channel.responseVote:
			node.handleResponseVoteRPC(response)
		case <-time.After(node.getTimeOut()):
			node.handleTimeout()
		}
	}
}
