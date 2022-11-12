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
	for i := uint32(0); i < raft.Config.NodeCount; i++ {
		wg.Add(1)
		node := raft.Node{}
		node.Init(i)

		raft.Config.NodeChannelList = append(raft.Config.NodeChannelList, &node.Channel)

		go node.Run(&wg)
	}

	// Run interactive console
	go repl.REPL()
	// Wait for all nodes to finish
	wg.Wait()
}
