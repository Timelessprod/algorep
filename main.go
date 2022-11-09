package main

import (
	"sync"
)

var wg sync.WaitGroup

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

	wg.Wait()
}
