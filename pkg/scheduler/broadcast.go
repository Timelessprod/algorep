package scheduler

import "github.com/Timelessprod/algorep/pkg/core"

/*** BROADCASTING ***/

// broadcastRequestVote broadcasts a RequestVote RPC to all the nodes (except itself)
func (node *SchedulerNode) broadcastRequestVote() {
	for i := uint32(0); i < core.Config.SchedulerNodeCount; i++ {
		if i != node.Id {
			lastLogIndex := uint32(len(node.log))
			channel := core.Config.NodeChannelMap[core.SchedulerNodeType][i].RequestVote
			request := core.RequestVoteRPC{
				FromNode:     node.Card,
				ToNode:       core.NodeCard{Id: i, Type: core.SchedulerNodeType},
				Term:         node.CurrentTerm,
				CandidateId:  node.Id,
				LastLogIndex: lastLogIndex,
				LastLogTerm:  node.LogTerm(lastLogIndex),
			}
			channel <- request
		}
	}
}

// sendSynchronizeCommandRPC sends a SynchronizeCommand RPC to a node
func (node *SchedulerNode) sendSynchronizeCommandRPC(nodeId uint32) {
	channel := core.Config.NodeChannelMap[core.SchedulerNodeType][nodeId].RequestCommand
	lastIndex := uint32(len(node.log))

	request := core.RequestCommandRPC{
		FromNode:    node.Card,
		ToNode:      core.NodeCard{Id: nodeId, Type: core.SchedulerNodeType},
		CommandType: core.SynchronizeCommand,

		Term:        node.CurrentTerm,
		PrevIndex:   node.nextIndex[nodeId] - 1,
		PrevTerm:    node.LogTerm(node.nextIndex[nodeId] - 1),
		Entries:     core.ExtractListFromMap(&node.log, node.nextIndex[nodeId], lastIndex),
		CommitIndex: node.commitIndex,
	}

	channel <- request
}

// brodcastSynchronizeCommand sends a SynchronizeCommand to all nodes (except itself)
func (node *SchedulerNode) broadcastSynchronizeCommandRPC() {
	for i := uint32(0); i < core.Config.SchedulerNodeCount; i++ {
		if i != node.Id {
			node.sendSynchronizeCommandRPC(i)
		}
	}
}
