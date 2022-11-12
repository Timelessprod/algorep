package raft

import (
	"fmt"
	"time"

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

const (
	LowNodeSpeed   time.Duration = 100 * time.Millisecond
	MediumNodeSpeed time.Duration = 25 * time.Millisecond
	HighNodeSpeed   time.Duration = 5 * time.Millisecond
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
