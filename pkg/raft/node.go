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

/***************
 ** Log Entry **
 ***************/

// LogEntry is an entry in the log
type LogEntry struct {
	Term    uint32
	Command string
}

/***************************
 ** Map LogEntry function **
 ***************************/

// Select range of log entries and return a new slice of log entries (start inclusive, end inclusive)
func ExtractListFromMap(m *map[uint32]LogEntry, start uint32, end uint32) []LogEntry {
	var list []LogEntry
	for i := start; i <= end; i++ {
		list = append(list, (*m)[i])
	}
	return list
}

// Flush the log entries after the given index. index + 1 and more are flushed but index is kept.
func FlushAfterIndex(m *map[uint32]LogEntry, index uint32) {
	for i := range *m {
		if i > index {
			delete(*m, i)
		}
	}
}
