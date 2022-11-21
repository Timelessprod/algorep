package client

import (
	"fmt"
	"strconv"

	"github.com/Timelessprod/algorep/pkg/core"
)

/******************
 ** Command Type **
 ******************/

// CommandType is the type of a command used by the REPL
type CommandType string

const (
	SPEED_COMMAND   CommandType = "SPEED"
	CRASH_COMMAND   CommandType = "CRASH"
	START_COMMAND   CommandType = "START"
	SUBMIT_COMMAND  CommandType = "SUBMIT"
	STATUS_COMMAND  CommandType = "STATUS"
	STOP_COMMAND    CommandType = "STOP"
	RECOVER_COMMAND CommandType = "RECOVER"
	HELP_COMMAND    CommandType = "HELP"
)

// Convert a CommandType to a string
func (c CommandType) String() string {
	return string(c)
}

/************************
 ** Speed Command Type **
 ************************/

// SpeedCommandType is the different speed level command used by the REPL
type SpeedCommandType string

const (
	LOW_SPEED    SpeedCommandType = "low"
	MEDIUM_SPEED SpeedCommandType = "medium"
	HIGH_SPEED   SpeedCommandType = "high"
)

// Convert a SpeedCommandType to a string
func (c SpeedCommandType) String() string {
	return string(c)
}

/*******************
 ** REPL Messages **
 *******************/

const (
	HELP_MESSAGE = `You can use 8 commands :
	- SPEED (low|medium|high) <node number> : change the speed of a node. For example: 'SPEED high 2' will change the speed of node 2 to high.
	- CRASH <node number> : crash a node. For example: 'CRASH 2' will crash node 2.
	- RECOVER <node number> : recover a crashed node. For example: 'RECOVER 2' will recover node 2.
	- START : start the cluster. You can use this command only once.
	- SUBMIT <job file> : submit a job to the cluster. The cluster must be STARTed before. For example: 'SUBMIT path/job.cpp' will submit the job described in the file job.cpp.
	- STATUS [<job reference>] : display the status of the cluster or of a specific job. For example: 'STATUS' will display the status of the cluster. 'STATUS 1-2' will display the status of the job with reference 1-2.
	- STOP : stop the cluster. This command will kill the program.
	- HELP : display this message.`
	SPEED_COMMAND_USAGE           = "The SPEED command must have the following form: `SPEED (low|medium|high) <node number>`. For example: 'SPEED high 2'"
	CRASH_COMMAND_USAGE           = "The CRASH command must have the following form: `CRASH <node number>`. For example: 'CRASH 2'"
	SUBMIT_COMMAND_USAGE          = "The SUBMIT command must have the following form: `SUBMIT <job file>`. For example: 'SUBMIT path/job.cpp'"
	RECOVER_COMMAND_USAGE         = "The RECOVER command must have the following form: `RECOVER <node number>`. For example: 'RECOVER 2'"
	STATUS_COMMAND_USAGE          = "The STATUS command must have the following form: `STATUS` or `STATUS <JobReference>`. For example: 'STATUS' or 'STATUS 1-2'"
	INVALID_JOB_REFERENCE_MESSAGE = "Job not found ! Please make sure you have provided a valid reference. The job reference must have the following form: `<JobId>-<Term>`. For example: '1-2'"
	INVALID_COMMAND_MESSAGE       = "Invalid command !"
	INVALID_SPEED_LEVEL_MESSAGE   = "Invalid speed level !"
	NOT_STARTED_MESSAGE           = "Cluster is not started yet ! Run the START command first."
)

// ParseNodeNumber parses the node number from a command
func parseNodeNumber(token string) (uint32, error) {
	nodeId, err := strconv.ParseUint(token, 0, 32)
	if err != nil {
		return 0, fmt.Errorf("Invalid node number: %s", token)
	}
	if nodeId >= uint64(core.Config.SchedulerNodeCount) {
		return 0, fmt.Errorf("Node number should be between 0 and %d", core.Config.SchedulerNodeCount-1)
	}
	return uint32(nodeId), nil
}
