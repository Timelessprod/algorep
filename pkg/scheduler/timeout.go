package scheduler

import (
	"time"

	"github.com/Timelessprod/algorep/pkg/core"
	"go.uber.org/zap"
)

/*** MANAGE TIMEOUT ***/

// getTimeOut returns the timeout duration depending on the node state
func (node *SchedulerNode) getTimeOut() time.Duration {
	switch node.State {
	case core.FollowerState:
		return node.ElectionTimeout
	case core.CandidateState:
		return node.ElectionTimeout
	case core.LeaderState:
		return core.Config.IsAliveNotificationInterval
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
	case core.FollowerState:
		logger.Warn("Leader does not respond", zap.String("Node", node.Card.String()), zap.Duration("electionTimeout", node.ElectionTimeout))
		node.startNewElection()
	case core.CandidateState:
		logger.Warn("Too much time to get a majority vote", zap.String("Node", node.Card.String()), zap.Duration("electionTimeout", node.ElectionTimeout))
		node.startNewElection()
	case core.LeaderState:
		logger.Info("It's time for the Leader to send an IsAlive notification to followers", zap.String("Node", node.Card.String()))
		node.broadcastSynchronizeCommandRPC()
	}
}
