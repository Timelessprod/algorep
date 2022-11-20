package scheduler

import (
	"github.com/Timelessprod/algorep/pkg/core"
	"go.uber.org/zap"
)

type StateMachine struct {
	JobMap map[string]core.Job
}

// Init initializes the state machine
func (sm *StateMachine) Init() {
	sm.JobMap = make(map[string]core.Job)
}

// Apply an Entry to the state machine
func (sm *StateMachine) Apply(entry core.Entry) {
	logger.Info("Applying entry to the StateMachine",
		zap.String("JobRef", entry.Job.GetReference()),
		zap.String("EntryType", entry.Type.String()),
	)
	sm.JobMap[entry.Job.GetReference()] = entry.Job
}

// Load a snapshot in the state machine
func (sm *StateMachine) Load(entryMap *map[uint32]core.Entry, maxIndex uint32) {
	sm.Init()
	if maxIndex == 0 {
		return
	}
	for i := uint32(1); i <= maxIndex; i++ {
		sm.Apply((*entryMap)[i])
	}
}
