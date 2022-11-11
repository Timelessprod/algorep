package raft

/*** Command (AppendEntry in paper) ***/

type CommandType int

const (
	Synchronize = iota
	AppendEntry
)

func (c CommandType) String() string {
	return [...]string{"Synchronize", "AppendEntry"}[c]
}

type RequestCommandRPC struct {
	FromNode uint32
	ToNode   uint32

	Term        uint32
	CommandType CommandType
	Message     string
}

type ResponseCommandRPC struct {
	FromNode uint32
	ToNode   uint32

	Term       uint32
	Success    bool
	MatchIndex uint32
}

/*** Vote ***/

type RequestVoteRPC struct {
	FromNode uint32
	ToNode   uint32

	Term        uint32
	CandidateId uint32
}

type ResponseVoteRPC struct {
	FromNode uint32
	ToNode   uint32

	Term        uint32
	VoteGranted bool
}
