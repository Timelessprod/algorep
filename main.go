package main

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	nodeNUmber        uint32
	channelBufferSize uint32
	nodeChannelList   []*ChannelContainer
}

var config = Config{
	nodeNUmber:        5,
	channelBufferSize: 100,
}

var logger *zap.Logger
var wg sync.WaitGroup

/***
    Message Object
***/

type AppendEntryRPC struct {
	fromNode uint32
	toNode   uint32

	// Request
	term    uint32
	message string

	// Response
	currentTerm uint32
	success     bool
	matchIndex  uint32
}

type RequestVoteRPC struct {
	fromNode uint32
	toNode   uint32

	// Request
	term        uint32
	candidateId uint32

	// Response
	currentTerm uint32
	voteGranted bool
}

/***
	ChannelContainer Object
***/
type ChannelContainer struct {
	appendEntry chan AppendEntryRPC
	requestVote chan RequestVoteRPC
}

/***
	Node Object
***/

type Node struct {
	id uint32

	currentTerm uint32
	votedFor    int32

	channel ChannelContainer
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

func (node *Node) run() {
	logger.Info("Node started", zap.Uint32("id", node.id))
	defer wg.Done()

	for {
		select {
		case appendEntryRPC := <-node.channel.appendEntry:
			node.handleAppendEntryRPC(appendEntryRPC)
		case requestVoteRPC := <-node.channel.requestVote:
			node.handleRequestVoteRPC(requestVoteRPC)
		}
	}
}

func (node *Node) init(id uint32) Node {
	node.id = id
	node.votedFor = -1
	node.channel.appendEntry = make(chan AppendEntryRPC, config.channelBufferSize)
	node.channel.requestVote = make(chan RequestVoteRPC, config.channelBufferSize)

	logger.Info("Node initialized", zap.Uint32("id", id))
	return *node
}

/***
	Functions
***/

func initLogger() *zap.Logger {
	// Create a new logger with our custom configuration
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	config.DisableStacktrace = true // Disable stacktrace after WARN or ERROR

	logger, _ := config.Build()

	logger.Debug("Logger initialized")
	return logger
}

/***
	Main
***/

func main() {
	logger = initLogger()
	defer logger.Sync()

	// Create nodes
	for i := uint32(0); i < config.nodeNUmber; i++ {
		wg.Add(1)
		node := Node{}
		node.init(i)
		config.nodeChannelList = append(config.nodeChannelList, &node.channel)
		go node.run()
	}

	config.nodeChannelList[0].appendEntry <- AppendEntryRPC{fromNode: 1, toNode: 0}
	wg.Wait()
}
