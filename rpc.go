package main

type AppendEntryRPC struct {
	fromNode uint32
	toNode   uint32

	// Request
	term    uint32
	message string

	// Response
	currentTerm uint32
	success     bool
	matchIndex  uint32
}

type RequestVoteRPC struct {
	fromNode uint32
	toNode   uint32

	// Request
	term        uint32
	candidateId uint32

	// Response
	currentTerm uint32
	voteGranted bool
}
