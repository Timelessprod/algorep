package raft

import "time"

/************
 ** Config **
 ************/

// Config contains the configuration of the raft algorithm
var Config = struct {
	SchedulerNodeCount uint32
	WorkerNodeCount    uint32
	ChannelBufferSize  uint32
	// NodeChannelMap contains all the channels used by the nodes to communicate with each other
	NodeSpeedList []time.Duration
	// NodeSpeedList contains the speed of each node to simulate different hardware
	NodeChannelMap map[NodeType][]*ChannelContainer

	// TIMEOUTS
	// Range of time to wait for a leader heartbeat or granting vote to candidate
	MinElectionTimeout   time.Duration
	MaxElectionTimeout   time.Duration
	MaxFindLeaderTimeout time.Duration

	// INTERVALS
	// Repeat interval for leader after it has sent out heartbeat
	IsAliveNotificationInterval time.Duration

	// RETRY
	MaxRetryToFindLeader uint32
}{
	SchedulerNodeCount: 5,
	WorkerNodeCount:    2,
	ChannelBufferSize:  100,
	NodeChannelMap:     InitNodeChannelMap(),

	MinElectionTimeout:   150 * time.Millisecond,
	MaxElectionTimeout:   300 * time.Millisecond,
	MaxFindLeaderTimeout: 300 * time.Millisecond,

	IsAliveNotificationInterval: 50 * time.Millisecond,

	MaxRetryToFindLeader: 3,
}

// InitNodeChannelMap initializes the NodeChannelMap
func InitNodeChannelMap() map[NodeType][]*ChannelContainer {
	channelMap := make(map[NodeType][]*ChannelContainer)
	channelMap[ClientNodeType] = make([]*ChannelContainer, 0)
	channelMap[SchedulerNodeType] = make([]*ChannelContainer, 0)
	channelMap[WorkerNodeType] = make([]*ChannelContainer, 0)
	return channelMap
}
