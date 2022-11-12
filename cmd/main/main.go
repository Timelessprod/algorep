package main

import (
	"sync"

	"github.com/Timelessprod/algorep/pkg/logging"
	"github.com/Timelessprod/algorep/pkg/raft"
	"github.com/Timelessprod/algorep/pkg/repl"
	"go.uber.org/zap"
)

var wg sync.WaitGroup
var logger *zap.Logger = logging.Logger

func main() {
	defer logger.Sync()

	// Create nodes
	for i := uint32(0); i < raft.Config.SchedulerNodeCount; i++ {
		wg.Add(1)
		node := raft.SchedulerNode{}
		node.Init(i)

		raft.Config.NodeChannelMap[raft.SchedulerNodeType] = append(raft.Config.NodeChannelMap[raft.SchedulerNodeType], &node.Channel)

		go node.Run(&wg)
	}

	// Run interactive console
	client := repl.ClientNode{}
	client.Init(0)
	raft.Config.NodeChannelMap[raft.ClientNodeType] = append(raft.Config.NodeChannelMap[raft.ClientNodeType], &client.Channel)
	go client.Run()

	// Wait for all nodes to finish
	wg.Wait()
}
