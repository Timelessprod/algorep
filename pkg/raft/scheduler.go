package raft

import (
	"math/rand"
	"sync"
	"time"

	"go.uber.org/zap"
)

/********************
 ** Scheduler Node **
 ********************/

// SchedulerNode is the node in charge of scheduling the database entries with the RAFT algorithm
type SchedulerNode struct {
	Id          uint32
	Card        NodeCard
	State       State
	CurrentTerm uint32
	LeaderId    int

	VotedFor        int32
	ElectionTimeout time.Duration
	VoteCount       uint32

	Channel ChannelContainer

	IsStarted bool
	IsCrashed bool
}

// Init the scheduler node
func (node *SchedulerNode) Init(id uint32) SchedulerNode {
	node.Id = id
	node.Card = NodeCard{Id: id, Type: SchedulerNodeType}
	node.State = FollowerState
	node.LeaderId = NO_LEADER
	node.VotedFor = NO_LEADER
	node.VoteCount = 0

	// Compute random leaderIsAliveTimeout
	durationRange := int(Config.MaxElectionTimeout - Config.MinElectionTimeout)
	node.ElectionTimeout = time.Duration(rand.Intn(durationRange)) + Config.MinElectionTimeout

	node.Channel.RequestCommand = make(chan RequestCommandRPC, Config.ChannelBufferSize)
	node.Channel.ResponseCommand = make(chan ResponseCommandRPC, Config.ChannelBufferSize)
	node.Channel.RequestVote = make(chan RequestVoteRPC, Config.ChannelBufferSize)
	node.Channel.ResponseVote = make(chan ResponseVoteRPC, Config.ChannelBufferSize)

	logger.Info("Node initialized", zap.Uint32("id", id))
	return *node
}

// Run the scheduler node
func (node *SchedulerNode) Run(wg *sync.WaitGroup) {
	defer wg.Done()
	logger.Info("Node is waiting the START command from REPL", zap.String("Node", node.Card.String()))
	// Wait for the start command from REPL and listen to the channel RequestCommand
	for !node.IsStarted {
		select {
		case request := <-node.Channel.RequestCommand:
			node.handleRequestCommandRPC(request)
		}
	}
	logger.Info("Node started", zap.String("Node", node.Card.String()))

	for {
		select {
		case request := <-node.Channel.RequestCommand:
			node.handleRequestCommandRPC(request)
		case response := <-node.Channel.ResponseCommand:
			node.handleResponseCommandRPC(response)
		case request := <-node.Channel.RequestVote:
			node.handleRequestVoteRPC(request)
		case response := <-node.Channel.ResponseVote:
			node.handleResponseVoteRPC(response)
		case <-time.After(node.getTimeOut()):
			node.handleTimeout()
		}
		time.Sleep(Config.NodeSpeedList[node.Id])
	}
}

/*** MANAGE TIMEOUT ***/

// getTimeOut returns the timeout duration depending on the node state
func (node *SchedulerNode) getTimeOut() time.Duration {
	switch node.State {
	case FollowerState:
		return node.ElectionTimeout
	case CandidateState:
		return node.ElectionTimeout
	case LeaderState:
		return Config.IsAliveNotificationInterval
	}
	logger.Panic("Invalid node state", zap.String("Node", node.Card.String()), zap.Int("state", int(node.State)))
	panic("Invalid node state")
}

// handleTimeout handles the timeout event
func (node *SchedulerNode) handleTimeout() {
	if node.IsCrashed {
		logger.Debug("Node is crashed. Ignore timeout", zap.String("Node", node.Card.String()))
		return
	}
	switch node.State {
	case FollowerState:
		logger.Warn("Leader does not respond", zap.String("Node", node.Card.String()), zap.Duration("electionTimeout", node.ElectionTimeout))
		node.startNewElection()
	case CandidateState:
		logger.Warn("Too much time to get a majority vote", zap.String("Node", node.Card.String()), zap.Duration("electionTimeout", node.ElectionTimeout))
		node.startNewElection()
	case LeaderState:
		logger.Info("It's time for the Leader to send an IsAlive notification to followers", zap.String("Node", node.Card.String()))
		node.broadcastSynchronizeCommandRPC()
	}
}

/*** BROADCASTING ***/

// broadcastRequestVote broadcasts a RequestVote RPC to all the nodes (except itself)
func (node *SchedulerNode) broadcastRequestVote() {
	for i := uint32(0); i < Config.SchedulerNodeCount; i++ {
		if i != node.Id {
			channel := Config.NodeChannelMap[SchedulerNodeType][i].RequestVote
			request := RequestVoteRPC{
				FromNode:    node.Card,
				ToNode:      NodeCard{Id: i, Type: SchedulerNodeType},
				Term:        node.CurrentTerm,
				CandidateId: node.Id,
			}
			channel <- request
		}
	}
}

// brodcastSynchronizeCommand sends a SynchronizeCommand to all nodes (except itself)
func (node *SchedulerNode) broadcastSynchronizeCommandRPC() {
	for i := uint32(0); i < Config.SchedulerNodeCount; i++ {
		if i != node.Id {
			channel := Config.NodeChannelMap[SchedulerNodeType][i].RequestCommand
			request := RequestCommandRPC{
				FromNode:    node.Card,
				ToNode:      NodeCard{Id: i, Type: SchedulerNodeType},
				Term:        node.CurrentTerm,
				CommandType: SynchronizeCommand,
			}
			channel <- request
		}
	}
}

/*** CONFIGURATION COMMAND ***/

// startElection starts an new election in sending a RequestVote RPC to all the nodes
func (node *SchedulerNode) startNewElection() {
	logger.Info("Start new election", zap.String("Node", node.Card.String()))
	node.State = CandidateState
	node.VoteCount = 1
	node.CurrentTerm++
	node.VotedFor = int32(node.Id)
	node.broadcastRequestVote()
}

// handleStartCommand starts the node when it receives a StartCommand
func (node *SchedulerNode) handleStartCommand() {
	if node.IsStarted {
		logger.Debug("Node already started",
			zap.String("Node", node.Card.String()),
		)
		return
	} else {
		node.IsStarted = true
	}
}

// handleCrashCommand crashes the node when it receives a CrashCommand
func (node *SchedulerNode) handleCrashCommand() {
	if node.IsCrashed {
		logger.Debug("Node is already crashed",
			zap.String("Node", node.Card.String()),
		)
		return
	} else {
		node.IsCrashed = true
	}
}

// handleRecoversCommand recovers the node after crash when it receives a RecoverCommand
func (node *SchedulerNode) handleRecoverCommand() {
	if node.IsCrashed {
		node.IsCrashed = false
	} else {
		logger.Debug("Node is not crashed",
			zap.String("Node", node.Card.String()),
		)
		return
	}
}

/*** HANDLE RPC ***/

// handleSynchronizeCommand handles the SynchronizeCommand to synchronize the entries and check if the leader is alive
func (node *SchedulerNode) handleSynchronizeCommand(request RequestCommandRPC) {
	if node.IsCrashed {
		logger.Debug("Node is crashed. Ignore synchronize command",
			zap.String("Node", node.Card.String()),
		)
		return
	}
	response := ResponseCommandRPC{
		FromNode:    node.Card,
		ToNode:      request.FromNode,
		CommandType: request.CommandType,
	}
	channel := Config.NodeChannelMap[request.FromNode.Type][request.FromNode.Id].ResponseCommand

	// Si request term < current term (sync pas à jours), alors on ignore et on répond false
	if request.Term < node.CurrentTerm {
		logger.Debug("Ignore synchronize command because request term < current term",
			zap.String("Node", node.Card.String()),
			zap.Uint32("request term", request.Term),
			zap.Uint32("current term", node.CurrentTerm),
		)
		response.Success = false
		channel <- response
		return
	}

	// Si node.State != follower et donc implicitement request term >= current term
	// alors la node devient follower
	if node.State != FollowerState {
		logger.Info("Node become Follower",
			zap.String("Node", node.Card.String()),
		)
		node.State = FollowerState
	}

	// Si request term > current term, alors on met à jour current term
	if request.Term > node.CurrentTerm {
		logger.Info("Update current term",
			zap.String("Node", node.Card.String()),
			zap.Uint32("new (request term)", request.Term),
			zap.Uint32("old (current term)", node.CurrentTerm),
		)
		node.CurrentTerm = request.Term
	}

	// Seul le leader peut envoyer des commandes Sync donc on met à jour leaderId
	node.LeaderId = int(request.FromNode.Id)

	// TODO : Synchronisation (à revoir)
	// Si pas cohérence des derniers logs
	// ==> response.Success = false
	// ==>Si conflit avec un entry existant et un nouveau,
	// ====> on supprime l'entry existant et tous les suivants
	// ==> On répond avec le dernier index de l'entry valide

	// Sinon
	// => on ajoute la nouvel entry
	// => commit l'index indiqué par la requete
	// Reponse avec success = true
	response.Success = true // Temporaire si pas de conflit
	channel <- response
}

// handleAppendEntryCommand handles the AppendEntryCommand sent to the leader to append an entry to the log and ignore the command if the node is not the leader
func (node *SchedulerNode) handleAppendEntryCommand(request RequestCommandRPC) {
	if node.IsCrashed {
		logger.Debug("Node is crashed. Ignore AppendEntry command",
			zap.String("Node", node.Card.String()),
		)
		return
	}

	channel := Config.NodeChannelMap[request.FromNode.Type][request.FromNode.Id].ResponseCommand
	response := ResponseCommandRPC{
		FromNode:    node.Card,
		ToNode:      request.FromNode,
		CommandType: request.CommandType,
		LeaderId:    node.LeaderId,
	}

	if node.State == LeaderState {
		logger.Info("I am the leader ! Submit Job.... ",
			zap.String("Node", node.Card.String()),
			zap.String("Message", request.Message),
		)
		response.Success = true

	} else {
		logger.Debug("Node is not the leader. Ignore AppendEntry command and redirect to leader",
			zap.String("Node", node.Card.String()),
			zap.Int("Presumed leader id", node.LeaderId),
		)
		response.Success = false
	}

	channel <- response
	// TODO
}

//  handleRequestCommandRPC handles the command RPC sent to the node
func (node *SchedulerNode) handleRequestCommandRPC(request RequestCommandRPC) {
	logger.Debug("Handle Request Command RPC",
		zap.String("FromNode", request.FromNode.String()),
		zap.String("ToNode", request.ToNode.String()),
		zap.String("CommandType", request.CommandType.String()),
	)
	switch request.CommandType {
	case SynchronizeCommand:
		node.handleSynchronizeCommand(request)
	case AppendEntryCommand:
		node.handleAppendEntryCommand(request)
	case StartCommand:
		node.handleStartCommand()
	case CrashCommand:
		node.handleCrashCommand()
	case RecoverCommand:
		node.handleRecoverCommand()
	}
}

// handleResponseCommandRPC handles the response command RPC sent to the node
func (node *SchedulerNode) handleResponseCommandRPC(response ResponseCommandRPC) {
	logger.Debug("Handle Response Command RPC",
		zap.String("FromNode", response.FromNode.String()),
		zap.String("ToNode", response.ToNode.String()),
		zap.Uint32("term", response.Term),
		zap.Bool("success", response.Success),
	)
	//TODO if majority of success, commit
	//TODO Sync les autres followers si necessaire
}

// handleRequestVoteRPC handles the request vote RPC sent to the node
func (node *SchedulerNode) handleRequestVoteRPC(request RequestVoteRPC) {
	if node.IsCrashed {
		logger.Debug("Node is crashed. Ignore request vote RPC",
			zap.String("FromNode", request.FromNode.String()),
			zap.String("ToNode", request.ToNode.String()),
		)
		return
	}
	logger.Debug("Handle Request Vote RPC",
		zap.String("FromNode", request.FromNode.String()),
		zap.String("ToNode", request.ToNode.String()),
		zap.Int("CandidateId", int(request.CandidateId)),
	)

	node.updateTerm(request.Term)

	channel := Config.NodeChannelMap[SchedulerNodeType][request.FromNode.Id].ResponseVote
	response := ResponseVoteRPC{
		FromNode:    request.ToNode,
		ToNode:      request.FromNode,
		Term:        node.CurrentTerm,
		VoteGranted: false,
	}

	if node.CurrentTerm == request.Term &&
		node.checkVote(request.CandidateId) {
		// TODO : and ( m.lastLogTerm > logTerm(len(log)) or (m.lastLogTerm == logTerm(len(log)) and m.lastLogIndex >= len(log)) )
		logger.Debug("Vote granted !",
			zap.String("Node", node.Card.String()),
			zap.Uint32("CandidateId", request.CandidateId),
		)
		node.VotedFor = int32(request.CandidateId)
		response.VoteGranted = true
	} else {
		logger.Debug("Vote refused !",
			zap.String("Node", node.Card.String()),
			zap.Uint32("CandidateId", request.CandidateId),
		)
		response.VoteGranted = false
	}
	channel <- response
}

// handleResponseVoteRPC handles the response vote RPC sent to the node
func (node *SchedulerNode) handleResponseVoteRPC(response ResponseVoteRPC) {
	if node.IsCrashed {
		logger.Debug("Node is crashed. Ignore request vote RPC",
			zap.String("FromNode", response.FromNode.String()),
			zap.String("ToNode", response.ToNode.String()),
		)
		return
	}

	logger.Debug("Handle Response Vote RPC",
		zap.String("FromNode", response.FromNode.String()),
		zap.String("ToNode", response.ToNode.String()),
		zap.Int("CandidateId", int(response.ToNode.Id)),
	)

	node.updateTerm(response.Term)
	if node.State == CandidateState &&
		node.CurrentTerm == response.Term {

		if response.VoteGranted {
			node.VoteCount++
			
			// When a candidate wins an election, it becomes leader.
			if node.VoteCount > Config.SchedulerNodeCount/2 {
				node.State = LeaderState
				node.LeaderId = int(node.Card.Id)
				logger.Info("Leader elected", zap.String("Node", node.Card.String()))
				// TODO : nextIndex[for each peer] = len(log) + 1
				return
			}
		}
	} else {
		logger.Debug("Node is not a candidate. Ignore response vote RPC",
			zap.String("Node", node.Card.String()),
			zap.String("VoteFromNode", response.FromNode.String()),
			zap.String("state", node.State.String()),
		)
	}
}

// updateTerm updates the term of the node if the term is higher than the current term
func (node *SchedulerNode) updateTerm(term uint32) {
	if term > node.CurrentTerm {
		node.CurrentTerm = term
		node.State = FollowerState
		node.VotedFor = NO_NODE
	}
}

// checkVote checks if the node has already voted for the candidate
func (node *SchedulerNode) checkVote(candidateId uint32) bool {
	if node.VotedFor == NO_NODE || uint32(node.VotedFor) == candidateId {
		return true
	}
	return false
}
