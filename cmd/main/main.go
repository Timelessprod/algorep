package main

import (
	"sync"

	"github.com/Timelessprod/algorep/pkg/client"
	"github.com/Timelessprod/algorep/pkg/core"
	"github.com/Timelessprod/algorep/pkg/scheduler"
	"go.uber.org/zap"
)

var wg sync.WaitGroup
var logger *zap.Logger = core.Logger

func main() {
	// To flush the last log in the buffer
	defer logger.Sync()

	// Init the channel map
	core.Config.NodeChannelMap = InitNodeChannelMap()

	// Create nodes and start them
	for i := uint32(0); i < core.Config.SchedulerNodeCount; i++ {
		wg.Add(1)
		node := scheduler.SchedulerNode{}
		node.Init(i)

		// Append channel to the map and speed to a list
		// We use global variable to avoid passing them to each node
		// This is a configuration of the cluster and depends on the hadware
		core.Config.NodeChannelMap[core.SchedulerNodeType] = append(core.Config.NodeChannelMap[core.SchedulerNodeType], &node.Channel)
		core.Config.NodeSpeedList = append(core.Config.NodeSpeedList, core.HighNodeSpeed)

		// Start node in a goroutine to smimulate an independant core
		go node.Run(&wg)
	}

	// Run interactive console
	client := client.ClientNode{}
	// We can image use several clients in different terminals
	// Here, for simplicity, we use only one client
	client.Init(0)
	// Add client channels to the map
	core.Config.NodeChannelMap[core.ClientNodeType] = append(core.Config.NodeChannelMap[core.ClientNodeType], &client.Channel)
	go client.Run()

	// Wait for all nodes to finish before exiting the main function
	// If we don't wait, the program will exit before the nodes have time to finish
	// and kill all goroutines
	wg.Wait()
}

// InitNodeChannelMap initializes the NodeChannelMap
func InitNodeChannelMap() map[core.NodeType][]*core.ChannelContainer {
	channelMap := make(map[core.NodeType][]*core.ChannelContainer)
	channelMap[core.ClientNodeType] = make([]*core.ChannelContainer, 0)
	channelMap[core.SchedulerNodeType] = make([]*core.ChannelContainer, 0)
	channelMap[core.WorkerNodeType] = make([]*core.ChannelContainer, 0)
	return channelMap
}
