package raft

import "time"

var Config = struct {
	NodeCount                   uint32
	ChannelBufferSize           uint32
	NodeChannelList             []*ChannelContainer
	MinElectionTimeout          time.Duration
	MaxElectionTimeout          time.Duration
	IsAliveNotificationInterval time.Duration
}{
	NodeCount:         5,
	ChannelBufferSize: 100,
	// Range of time to wait for a leader heartbeat or granting vote to candidate
	MinElectionTimeout: 150 * time.Millisecond,
	MaxElectionTimeout: 300 * time.Millisecond,
	// Repeat interval for leader after it has sent out heartbeat
	IsAliveNotificationInterval: 50 * time.Millisecond,
}
