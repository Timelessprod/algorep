package worker

import (
	"github.com/Timelessprod/algorep/pkg/core"
	"go.uber.org/zap"
)

var logger *zap.Logger = core.Logger

/*****************
 ** Worker Node **
 *****************/

type WorkerNode struct {
	Id           uint32
	Card         core.NodeCard
	LastLeaderId uint32

	Channel core.ChannelContainer
}

// Init initializes the worker node
func (node *WorkerNode) Init(id uint32) {
	node.Id = id
	node.Card = core.NodeCard{Id: id, Type: core.WorkerNodeType}
	node.Channel = core.ChannelContainer{
		JobQueue: make(chan core.Job, core.Config.ChannelBufferSize),
	}
	node.LastLeaderId = 0 // Valeur par défaut le temps de trouver le leader
}

// Run the worker node
func (node *WorkerNode) Run() {
	logger.Info("Node started", zap.String("Node", node.Card.String()))
	for {
		job := <-node.Channel.JobQueue
		node.processJob(job)
	}
}

// processJob processes a job
func (node *WorkerNode) processJob(job core.Job) {
	logger.Info("Processing job",
		zap.String("Node", node.Card.String()),
		zap.String("Job", job.GetReference()),
	)
	// TODO
}
