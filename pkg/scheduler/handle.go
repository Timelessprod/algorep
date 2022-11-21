package scheduler

import (
	"fmt"

	"github.com/Timelessprod/algorep/pkg/core"
	"github.com/Timelessprod/algorep/pkg/utils"
	"go.uber.org/zap"
)

/*** HANDLE RPC ***/

// handleRequestSynchronizeCommand handles the SynchronizeCommand to synchronize the entries and check if the leader is alive
func (node *SchedulerNode) handleRequestSynchronizeCommand(request core.RequestCommandRPC) {
	if node.IsCrashed {
		logger.Debug("Node is crashed. Ignore synchronize command",
			zap.String("Node", node.Card.String()),
		)
		return
	}

	response := core.ResponseCommandRPC{
		FromNode:    node.Card,
		ToNode:      request.FromNode,
		Term:        node.CurrentTerm,
		CommandType: request.CommandType,
	}
	channel := core.Config.NodeChannelMap[request.FromNode.Type][request.FromNode.Id].ResponseCommand

	node.updateTerm(request.Term)

	// Si request term < current term (sync pas à jours), alors on ignore et on répond false
	if node.CurrentTerm > request.Term {
		logger.Debug("Ignore synchronize command because request term < current term",
			zap.String("Node", node.Card.String()),
			zap.Uint32("request term", request.Term),
			zap.Uint32("current term", node.CurrentTerm),
		)
		response.Success = false
		channel <- response
		return
	}

	// Seul le leader peut envoyer des commandes Sync donc on met à jour leaderId
	node.LeaderId = int(request.FromNode.Id)

	if node.State != core.FollowerState {
		logger.Info("Node become Follower",
			zap.String("Node", node.Card.String()),
		)
		node.State = core.FollowerState
	}

	lastLogConsistency := node.LogTerm(request.PrevIndex) == request.PrevTerm &&
		request.PrevIndex <= uint32(len(node.log))
	success := request.PrevIndex == 0 || lastLogConsistency
	logger.Debug("Fields of received synchronization request",
		zap.String("Node", node.Card.String()),
		zap.Bool("lastLogConsistency", lastLogConsistency),
		zap.Uint32("request.PrevIndex", request.PrevIndex),
		zap.Uint32("len(node.log)", uint32(len(node.log))),
		zap.Uint32("request.PrevTerm", request.PrevTerm),
		zap.Uint32("node.LogTerm(request.PrevIndex)", node.LogTerm(request.PrevIndex)),
		zap.String("request.entries", fmt.Sprintf("%v", request.Entries)),
	)

	var index uint32
	if success {
		index = request.PrevIndex
		for j := 0; j < len(request.Entries); j++ {
			index++
			if node.LogTerm(index) != request.Entries[j].Term {
				node.log[index] = request.Entries[j]
			}
		}
		core.FlushAfterIndex(&node.log, index)
		node.commitIndex = utils.MinUint32(request.CommitIndex, index)
	} else {
		index = 0
	}

	response.MatchIndex = index
	response.Success = success
	channel <- response
}

// handleResponseSynchronizeCommand handles the response of the SynchronizeCommand
func (node *SchedulerNode) handleResponseSynchronizeCommand(response core.ResponseCommandRPC) {
	if node.IsCrashed {
		logger.Debug("Node is crashed. Ignore synchronize command",
			zap.String("Node", node.Card.String()),
		)
		return
	}

	logger.Debug("Receive synchronize command response",
		zap.String("FromNode", response.FromNode.String()),
		zap.String("ToNode", node.Card.String()),
		zap.Bool("Success", response.Success),
		zap.Uint32("MatchIndex", response.MatchIndex),
	)

	node.updateTerm(response.Term)
	if node.State == core.LeaderState && node.CurrentTerm == response.Term {
		fromNode := response.FromNode.Id
		if response.Success {
			node.matchIndex[fromNode] = response.MatchIndex
			node.nextIndex[fromNode] = response.MatchIndex + 1
		} else {
			node.nextIndex[fromNode] = utils.MaxUint32(1, node.nextIndex[fromNode]-1)
		}
	}
}

// handleAppendEntryCommand handles the AppendEntryCommand sent to the leader to append an entry to the log and ignore the command if the node is not the leader
func (node *SchedulerNode) handleAppendEntryCommand(request core.RequestCommandRPC) {
	if node.IsCrashed {
		logger.Debug("Node is crashed. Ignore AppendEntry command",
			zap.String("Node", node.Card.String()),
		)
		return
	}

	channel := core.Config.NodeChannelMap[request.FromNode.Type][request.FromNode.Id].ResponseCommand
	response := core.ResponseCommandRPC{
		FromNode:    node.Card,
		ToNode:      request.FromNode,
		Term:        node.CurrentTerm,
		CommandType: request.CommandType,
		LeaderId:    node.LeaderId,
	}

	if node.State == core.LeaderState {
		entry := request.Entries[0] // Append only one entry at a time

		logger.Info("I am the leader ! Submit Job.... ",
			zap.String("Node", node.Card.String()),
			zap.String("JobRef", entry.Job.GetReference()),
		)

		entry.Term = node.CurrentTerm
		if entry.Type == core.OpenJob {
			entry.Job.WorkerId = int(node.GetWorkerId())
			entry.Job.Id = node.GetJobId()
			entry.Job.Term = node.CurrentTerm
			entry.Job.State = core.JobWaiting
		}

		node.addEntryToLog(entry)
		response.Success = true
		response.Message = fmt.Sprintf("Job %s submitted.", entry.Job.GetReference())

	} else {
		logger.Debug("Node is not the leader. Ignore AppendEntry command and redirect to leader",
			zap.String("Node", node.Card.String()),
			zap.Int("Presumed leader id", node.LeaderId),
		)
		response.Success = false
	}

	channel <- response
}

// handleStatusCommand handles the StatusCommand sent to the leader to get the status of jobs
func (node *SchedulerNode) handleStatusCommand(request core.RequestCommandRPC) {
	if node.IsCrashed {
		logger.Debug("Node is crashed. Ignore Status command",
			zap.String("Node", node.Card.String()),
		)
		return
	}

	channel := core.Config.NodeChannelMap[request.FromNode.Type][request.FromNode.Id].ResponseCommand
	response := core.ResponseCommandRPC{
		FromNode:    node.Card,
		ToNode:      request.FromNode,
		Term:        node.CurrentTerm,
		CommandType: request.CommandType,
		LeaderId:    node.LeaderId,
	}

	if node.State == core.LeaderState {
		logger.Info("I am the leader ! Giving the status.... ",
			zap.String("Node", node.Card.String()),
		)

		response.JobMap = node.StateMachine.JobMap
		response.Success = true

	} else {
		logger.Debug("Node is not the leader. Ignore Status command and redirect to leader",
			zap.String("Node", node.Card.String()),
			zap.Int("Presumed leader id", node.LeaderId),
		)
		response.Success = false
	}

	channel <- response
}

//  handleRequestCommandRPC handles the command RPC sent to the node
func (node *SchedulerNode) handleRequestCommandRPC(request core.RequestCommandRPC) {
	logger.Debug("Handle Request Command RPC",
		zap.String("FromNode", request.FromNode.String()),
		zap.String("ToNode", request.ToNode.String()),
		zap.String("CommandType", request.CommandType.String()),
	)
	switch request.CommandType {
	case core.SynchronizeCommand:
		node.handleRequestSynchronizeCommand(request)
	case core.AppendEntryCommand:
		node.handleAppendEntryCommand(request)
	case core.StartCommand:
		node.handleStartCommand()
	case core.CrashCommand:
		node.handleCrashCommand()
	case core.RecoverCommand:
		node.handleRecoverCommand()
	case core.StatusCommand:
		node.handleStatusCommand(request)
	}
}

// handleResponseCommandRPC handles the response command RPC sent to the node
func (node *SchedulerNode) handleResponseCommandRPC(response core.ResponseCommandRPC) {
	logger.Debug("Handle Response Command RPC",
		zap.String("FromNode", response.FromNode.String()),
		zap.String("ToNode", response.ToNode.String()),
		zap.Uint32("term", response.Term),
		zap.Bool("success", response.Success),
	)
	switch response.CommandType {
	case core.SynchronizeCommand:
		node.handleResponseSynchronizeCommand(response)
	default:
		logger.Error("Unknown response command type",
			zap.String("CommandType", response.CommandType.String()),
		)
	}
}

// handleRequestVoteRPC handles the request vote RPC sent to the node
func (node *SchedulerNode) handleRequestVoteRPC(request core.RequestVoteRPC) {
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

	channel := core.Config.NodeChannelMap[core.SchedulerNodeType][request.FromNode.Id].ResponseVote
	response := core.ResponseVoteRPC{
		FromNode:    request.ToNode,
		ToNode:      request.FromNode,
		Term:        node.CurrentTerm,
		VoteGranted: false,
	}

	lastLogIndex := uint32(len(node.log))
	lastLogTerm := node.LogTerm(lastLogIndex)
	logConsistency := request.LastLogTerm > lastLogTerm ||
		(request.LastLogTerm == lastLogTerm && request.LastLogIndex >= lastLogIndex)

	if node.CurrentTerm == request.Term &&
		node.checkVote(request.CandidateId) &&
		logConsistency {

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
func (node *SchedulerNode) handleResponseVoteRPC(response core.ResponseVoteRPC) {
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
	if node.State == core.CandidateState &&
		node.CurrentTerm == response.Term {

		if response.VoteGranted {
			node.VoteCount++

			// When a candidate wins an election, it becomes leader.
			if node.VoteCount > core.Config.SchedulerNodeCount/2 {
				node.becomeLeader()
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
