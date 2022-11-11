package main

import "time"

type Config struct {
	nodeCount                   uint32
	channelBufferSize           uint32
	nodeChannelList             []*ChannelContainer
	minElectionTimeout          time.Duration
	maxElectionTimeout          time.Duration
	IsAliveNotificationInterval time.Duration
}

var config = Config{
	nodeCount:         5,
	channelBufferSize: 100,
	// Range of time to wait for a leader heartbeat or granting vote to candidate
	minElectionTimeout: 150 * time.Millisecond,
	maxElectionTimeout: 300 * time.Millisecond,
	// Repeat interval for leader after it has sent out heartbeat
	IsAliveNotificationInterval: 50 * time.Millisecond,
}
