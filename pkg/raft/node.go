package raft

import (
	"fmt"

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

func (n NodeType) String() string {
	return string(n)
}

type ChannelContainer struct {
	RequestCommand  chan RequestCommandRPC
	ResponseCommand chan ResponseCommandRPC

	RequestVote  chan RequestVoteRPC
	ResponseVote chan ResponseVoteRPC
}

type NodeCard struct {
	Id   uint32
	Type NodeType
}

func (n NodeCard) String() string {
	return fmt.Sprint(n.Type.String(), " - ", n.Id)
}
