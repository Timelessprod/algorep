package main

import (
	"sync"

	"github.com/Timelessprod/algorep/pkg/core"
	"github.com/Timelessprod/algorep/pkg/raft"
	"github.com/Timelessprod/algorep/pkg/repl"
	"go.uber.org/zap"
)

var wg sync.WaitGroup
var logger *zap.Logger = core.Logger

func main() {
	// To flush the last log in the buffer
	defer logger.Sync()

	// Create nodes and start them
	for i := uint32(0); i < raft.Config.SchedulerNodeCount; i++ {
		wg.Add(1)
		node := raft.SchedulerNode{}
		node.Init(i)

		// Append channel to the map and speed to a list
		// We use global variable to avoid passing them to each node
		// This is a configuration of the cluster and depends on the hadware
		raft.Config.NodeChannelMap[raft.SchedulerNodeType] = append(raft.Config.NodeChannelMap[raft.SchedulerNodeType], &node.Channel)
		raft.Config.NodeSpeedList = append(raft.Config.NodeSpeedList, raft.HighNodeSpeed)

		// Start node in a goroutine to smimulate an independant core
		go node.Run(&wg)
	}

	// Run interactive console
	client := repl.ClientNode{}
	// We can image use several clients in different terminals
	// Here, for simplicity, we use only one client
	client.Init(0)
	// Add client channels to the map
	raft.Config.NodeChannelMap[raft.ClientNodeType] = append(raft.Config.NodeChannelMap[raft.ClientNodeType], &client.Channel)
	go client.Run()

	// Wait for all nodes to finish before exiting the main function
	// If we don't wait, the program will exit before the nodes have time to finish
	// and kill all goroutines
	wg.Wait()
}
