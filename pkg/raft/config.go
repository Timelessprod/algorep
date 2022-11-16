package raft

import "time"

var Config = struct {
	SchedulerNodeCount          uint32
	ChannelBufferSize           uint32
	NodeSpeedList               []time.Duration
	NodeChannelMap              map[NodeType][]*ChannelContainer
	MinElectionTimeout          time.Duration
	MaxElectionTimeout          time.Duration
	MaxFindLeaderTimeout        time.Duration
	IsAliveNotificationInterval time.Duration
	MaxRetryToFindLeader        uint32
}{
	SchedulerNodeCount: 5,
	ChannelBufferSize:  100,
	NodeChannelMap:     InitNodeChannelMap(),
	// Range of time to wait for a leader heartbeat or granting vote to candidate
	MinElectionTimeout:   150 * time.Millisecond,
	MaxElectionTimeout:   300 * time.Millisecond,
	MaxFindLeaderTimeout: 300 * time.Millisecond,
	// Repeat interval for leader after it has sent out heartbeat
	IsAliveNotificationInterval: 50 * time.Millisecond,
	MaxRetryToFindLeader:        3,
}

func InitNodeChannelMap() map[NodeType][]*ChannelContainer {
	channelMap := make(map[NodeType][]*ChannelContainer)
	channelMap[ClientNodeType] = make([]*ChannelContainer, 0)
	channelMap[SchedulerNodeType] = make([]*ChannelContainer, 0)
	channelMap[WorkerNodeType] = make([]*ChannelContainer, 0)
	return channelMap
}
