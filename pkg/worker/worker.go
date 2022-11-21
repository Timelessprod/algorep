package worker

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/Timelessprod/algorep/pkg/core"
	"go.uber.org/zap"
)

var logger *zap.Logger = core.Logger

/*****************
 ** Worker Node **
 *****************/

type WorkerNode struct {
	Id           uint32
	Card         core.NodeCard
	LastLeaderId uint32

	Channel core.ChannelContainer
}

// Init initializes the worker node
func (node *WorkerNode) Init(id uint32) {
	node.Id = id
	node.Card = core.NodeCard{Id: id, Type: core.WorkerNodeType}
	node.Channel = core.ChannelContainer{
		ResponseCommand: make(chan core.ResponseCommandRPC, core.Config.ChannelBufferSize),
		JobQueue:        make(chan core.Job, core.Config.ChannelBufferSize),
	}
	node.LastLeaderId = 0 // Valeur par défaut le temps de trouver le leader
}

// Run the worker node
func (node *WorkerNode) Run() {
	logger.Info("Node started", zap.String("Node", node.Card.String()))
	for {
		job := <-node.Channel.JobQueue
		node.processJob(job)
	}
}

// processJob processes a job
func (node *WorkerNode) processJob(job core.Job) {
	logger.Info("Processing job",
		zap.String("Node", node.Card.String()),
		zap.String("Job", job.GetReference()),
	)
	logger.Debug("Job Details",
		zap.String("Node", node.Card.String()),
		zap.String("Job", job.GetReference()),
		zap.String("Input", job.Input),
		zap.String("Output", job.Output),
	)

	// Execute the job
	node.ExecuteJob(&job)

	// Close the job
	job.State = core.JobDone
	node.closeJob(job)
}

// ExecuteJob executes a job
func (node *WorkerNode) ExecuteJob(job *core.Job) {
	logger.Info("Execute job",
		zap.String("Node", node.Card.String()),
		zap.String("Job", job.GetReference()),
	)

	reference := job.GetReference()
	binaryName := fmt.Sprintf("job-%s.out", reference)

	// Compile the job
	logger.Debug("Compiling job ...",
		zap.String("Node", node.Card.String()),
		zap.String("Job", job.GetReference()),
		zap.String("BinaryName", binaryName),
	)
	compileCommand := exec.Command("g++", "-o", binaryName, "-x", "c++", "-")
	compileCommand.Stdin = strings.NewReader(job.Input)
	var stdoutCompile, stderrCompile bytes.Buffer
	compileCommand.Stdout = &stdoutCompile
	compileCommand.Stderr = &stderrCompile
	err := compileCommand.Run()
	if err != nil {
		logger.Error("Error while compiling job",
			zap.String("Node", node.Card.String()),
			zap.String("Job", job.GetReference()),
			zap.String("Error", err.Error()),
		)
		errorPrompt := "--- Error while compiling job ---\n%s--- StdOut ---\n%s---StdError---%s"
		job.Output = fmt.Sprintf(errorPrompt, err.Error(), stdoutCompile.String(), stderrCompile.String())
		return
	}

	// Run the binary
	logger.Debug("Running job ...",
		zap.String("Node", node.Card.String()),
		zap.String("Job", job.GetReference()),
		zap.String("BinaryName", binaryName),
	)
	runCommandString := fmt.Sprintf("./%s", binaryName)
	runCommand := exec.Command(runCommandString)
	var stdoutRun, stderrRun bytes.Buffer
	runCommand.Stdout = &stdoutRun
	runCommand.Stderr = &stderrRun
	err = runCommand.Run()
	if err != nil {
		logger.Error("Error while running job",
			zap.String("Node", node.Card.String()),
			zap.String("Job", job.GetReference()),
			zap.String("Error", err.Error()),
		)
	}
	job.Output = fmt.Sprint(stdoutRun.String(), stderrRun.String())
	logger.Debug("Job has been executed",
		zap.String("Node", node.Card.String()),
		zap.String("Job", job.GetReference()),
		zap.String("Output", job.Output),
	)

	// Remove the binary
	logger.Debug("Removing binary ...",
		zap.String("Node", node.Card.String()),
		zap.String("Job", job.GetReference()),
		zap.String("BinaryName", binaryName),
	)
	removeComand := exec.Command("rm", "-f", binaryName)
	err = removeComand.Run()
	if err != nil {
		logger.Error("Error while removing binary",
			zap.String("Node", node.Card.String()),
			zap.String("Job", job.GetReference()),
			zap.String("Error", err.Error()),
		)
	}
}

// closeJob closes a job
func (node *WorkerNode) closeJob(job core.Job) {
	logger.Debug("Closing job ...",
		zap.String("Node", node.Card.String()),
		zap.String("Job", job.GetReference()),
		zap.String("JobStatus", job.State.String()),
	)

	entry := core.Entry{
		Type: core.CloseJob,
		Job:  job,
	}
	message := core.RequestCommandRPC{
		FromNode:    node.Card,
		CommandType: core.AppendEntryCommand,
		Entries:     []core.Entry{entry},
	}
	node.sendMessageToLeader(message)
}

// sendMessageToLeader sends a message to the leader
func (node *WorkerNode) sendMessageToLeader(message core.RequestCommandRPC) {
	for retryCounter := 1; true; retryCounter++ {
		if uint32(retryCounter) > core.Config.MaxRetryToFindLeader {
			logger.Error("Max retry reached",
				zap.String("Node", node.Card.String()),
				zap.Int("retryCounter", retryCounter),
			)
		}

		requestChannel := core.Config.NodeChannelMap[core.SchedulerNodeType][node.LastLeaderId].RequestCommand
		responseChannel := core.Config.NodeChannelMap[node.Card.Type][node.Card.Id].ResponseCommand
		message.ToNode = core.NodeCard{Id: node.LastLeaderId, Type: core.SchedulerNodeType}
		requestChannel <- message

		select {
		case response := <-responseChannel:
			// If LeaderId given by node is -1
			// it means that node does not know who is the leader
			if response.LeaderId == core.NO_NODE {
				logger.Warn("Leader is unknown. Check random node !",
					zap.Uint32("tested nodeId", node.LastLeaderId),
				)
				node.LastLeaderId = core.GetRandomSchedulerNodeId()
				continue
			}

			// Check if node connected to is still leader
			if response.LeaderId != int(node.LastLeaderId) {
				logger.Warn("Leader has changed",
					zap.Uint32("old", node.LastLeaderId),
					zap.Uint32("new", uint32(response.LeaderId)),
				)
				node.LastLeaderId = uint32(response.LeaderId)
				continue
			}

			// Else if the tested node is the leader
			return
		case <-time.After(core.Config.MaxFindLeaderTimeout):
			logger.Warn("Node is not responding, trying to find new leader with random node...",
				zap.Int("try", retryCounter),
				zap.Uint32("tested nodeId", node.LastLeaderId),
			)
			node.LastLeaderId = core.GetRandomSchedulerNodeId()
		}
	}
}
