package raft

// Generic RPC Type
type RPCType interface {
	RequestCommandRPC | ResponseCommandRPC | RequestVoteRPC | ResponseVoteRPC
}

/******************
 ** Command Type **
 ******************/

// CommandType is the type of a command used by RequestCommandRPC
type CommandType int

const (
	SynchronizeCommand = iota
	AppendEntryCommand
	StartCommand
	CrashCommand
	RecoverCommand
)

// Convert a CommandType to a string
func (c CommandType) String() string {
	return [...]string{"Synchronize", "AppendEntry", "Start", "Crash", "Recover"}[c]
}

/*****************
 ** Command RPC **
 *****************/

// RequestCommandRPC is the RPC used to send a command to a node
type RequestCommandRPC struct {
	FromNode NodeCard
	ToNode   NodeCard

	Term        uint32
	CommandType CommandType
	Entry

	PrevIndex   uint32
	PrevTerm    uint32
	Entries     []Entry
	CommitIndex uint32
}

// ResponseCommandRPC is the RPC used to send a response to a command
type ResponseCommandRPC struct {
	FromNode NodeCard
	ToNode   NodeCard

	Term     uint32
	LeaderId int

	CommandType CommandType
	Message     string

	Success    bool
	MatchIndex uint32
}

/**************
 ** Vote RPC **
 **************/

// RequestVoteRPC is the RPC used to request a vote
type RequestVoteRPC struct {
	FromNode NodeCard
	ToNode   NodeCard

	Term         uint32
	CandidateId  uint32
	LastLogTerm  uint32
	LastLogIndex uint32
}

// ResponseVoteRPC is the RPC used to send a response to a vote request
type ResponseVoteRPC struct {
	FromNode NodeCard
	ToNode   NodeCard

	Term        uint32
	VoteGranted bool
}
