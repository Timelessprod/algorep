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
	FromNode NodeCard
	ToNode   NodeCard

	Term        uint32
	CommandType CommandType
	Message     string
}

type ResponseCommandRPC struct {
	FromNode NodeCard
	ToNode   NodeCard

	Term       uint32
	Success    bool
	MatchIndex uint32
}

/*** Vote ***/

type RequestVoteRPC struct {
	FromNode NodeCard
	ToNode   NodeCard

	Term        uint32
	CandidateId uint32
}

type ResponseVoteRPC struct {
	FromNode NodeCard
	ToNode   NodeCard

	Term        uint32
	VoteGranted bool
}
