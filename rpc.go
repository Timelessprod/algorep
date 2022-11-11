package main

/*** Command (AppendEntry in paper) ***/

type RequestCommandRPC struct {
	fromNode uint32
	toNode   uint32

	term    uint32
	message string
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
