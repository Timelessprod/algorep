package main

import (
	"math/rand"
	"time"

	"go.uber.org/zap"
)

type ChannelContainer struct {
	appendEntry chan AppendEntryRPC
	requestVote chan RequestVoteRPC
}

type State int

const (
	FollowerState = iota
	CandidateState
	LeaderState
)

type Node struct {
	id              uint32
	state           State
	currentTerm     uint32
	votedFor        int32
	electionTimeout time.Duration

	channel ChannelContainer
}

func (node *Node) init(id uint32) Node {
	node.id = id
	node.state = FollowerState
	node.votedFor = -1

	// Compute random leaderIsAliveTimeout
	durationRange := int(config.maxElectionTimeout - config.minElectionTimeout)
	node.electionTimeout = time.Duration(rand.Intn(durationRange)) + config.minElectionTimeout

	node.channel.appendEntry = make(chan AppendEntryRPC, config.channelBufferSize)
	node.channel.requestVote = make(chan RequestVoteRPC, config.channelBufferSize)

	logger.Info("Node initialized", zap.Uint32("id", id))
	return *node
}

// handleAppendEntryRPC
func (node *Node) handleAppendEntryRPC(appendEntryRPC AppendEntryRPC) {
	logger.Debug("handleAppendEntryRPC",
		zap.Uint32("fromNode", appendEntryRPC.fromNode),
		zap.Uint32("toNode", appendEntryRPC.toNode),
		zap.String("message", appendEntryRPC.message),
	)
	fromNodeChannel := config.nodeChannelList[appendEntryRPC.fromNode]
	fromNodeChannel.appendEntry <- AppendEntryRPC{fromNode: node.id, toNode: appendEntryRPC.fromNode, message: "Ping !"}
}

// handleRequestVoteRPC
func (node *Node) handleRequestVoteRPC(requestVoteRPC RequestVoteRPC) {
	logger.Debug("handleRequestVoteRPC", zap.Uint32("fromNode", requestVoteRPC.fromNode), zap.Uint32("toNode", requestVoteRPC.toNode))
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

func (node *Node) handleTimeout() {
	switch node.state {
	case FollowerState:
		logger.Warn("Leader does not respond", zap.Uint32("id", node.id), zap.Duration("electionTimeout", node.electionTimeout))
	case CandidateState:
		logger.Warn("Too much time to get a majority vote", zap.Uint32("id", node.id), zap.Duration("electionTimeout", node.electionTimeout))
	case LeaderState:
		logger.Info("It's time for the Leader to send an IsAlive notification to followers", zap.Uint32("id", node.id))
	}
}

func (node *Node) run() {
	logger.Info("Node started", zap.Uint32("id", node.id))
	defer wg.Done()

	for {
		select {
		case appendEntryRPC := <-node.channel.appendEntry:
			node.handleAppendEntryRPC(appendEntryRPC)
		case requestVoteRPC := <-node.channel.requestVote:
			node.handleRequestVoteRPC(requestVoteRPC)
		case <-time.After(node.getTimeOut()):
			node.handleTimeout()
		}
	}
}
