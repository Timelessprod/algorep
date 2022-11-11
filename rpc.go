package main

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
	fromNode uint32
	toNode   uint32

	term        uint32
	commandType CommandType
	message     string
}

type ResponseCommandRPC struct {
	fromNode uint32
	toNode   uint32

	term       uint32
	success    bool
	matchIndex uint32
}

/*** Vote ***/

type RequestVoteRPC struct {
	fromNode uint32
	toNode   uint32

	term        uint32
	candidateId uint32
}

type ResponseVoteRPC struct {
	fromNode uint32
	toNode   uint32

	term        uint32
	voteGranted bool
}
