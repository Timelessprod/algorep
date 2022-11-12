package raft

import (
	"github.com/Timelessprod/algorep/pkg/logging"
	"go.uber.org/zap"
)

var logger *zap.Logger = logging.Logger

type NodeType string

const (
	ClientNodeType    NodeType = "Client"
	SchedulerNodeType NodeType = "Scheduler"
	WorkerNodeType    NodeType = "Worker"
)

type ChannelContainer struct {
	RequestCommand  chan RequestCommandRPC
	ResponseCommand chan ResponseCommandRPC

	RequestVote  chan RequestVoteRPC
	ResponseVote chan ResponseVoteRPC
}
