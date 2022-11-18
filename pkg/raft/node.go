package raft

import (
	"fmt"
	"time"

	"github.com/Timelessprod/algorep/pkg/logging"
	"go.uber.org/zap"
)

var logger *zap.Logger = logging.Logger

// Node id when no node is selected
const NO_NODE = -1

// LeaderId value when we don't know who is the leader
const NO_LEADER = NO_NODE

/****************
 ** Node Types **
 ****************/

// Node Types
type NodeType string

const (
	ClientNodeType    NodeType = "Client"
	SchedulerNodeType NodeType = "Scheduler"
	WorkerNodeType    NodeType = "Worker"
)

// Convert a NodeType to a string
func (n NodeType) String() string {
	return string(n)
}

/****************
 ** Node Speed **
 ****************/

// Node Speed
const (
	LowNodeSpeed    time.Duration = 50 * time.Millisecond
	MediumNodeSpeed time.Duration = 10 * time.Millisecond
	HighNodeSpeed   time.Duration = 2 * time.Millisecond
)

/***********************
 ** Channel Container **
 ***********************/

// ChannelContainer contains all the channels used by a node to receive messages
type ChannelContainer struct {
	RequestCommand  chan RequestCommandRPC
	ResponseCommand chan ResponseCommandRPC

	RequestVote  chan RequestVoteRPC
	ResponseVote chan ResponseVoteRPC
}

/***************
 ** Node Card **
 ***************/

// NodeCard contains the information about a node. It is used to identify a node (Id and Type)
type NodeCard struct {
	Id   uint32
	Type NodeType
}

// Convert a NodeCard to a string representation
func (n NodeCard) String() string {
	return fmt.Sprint(n.Type.String(), " - ", n.Id)
}

/****************
 ** Node State **
 ****************/

// Node state (follower, candidate, leader)
type State int

const (
	FollowerState = iota
	CandidateState
	LeaderState
)

// Convert a State to a string
func (s State) String() string {
	return [...]string{"Follower", "Candidate", "Leader"}[s]
}
