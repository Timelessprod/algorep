package repl

type CommandType string

const (
	SPEED_COMMAND CommandType = "SPEED"
	CRASH_COMMAND CommandType = "CRASH"
	START_COMMAND CommandType = "START"
	STOP_COMMAND  CommandType = "STOP"
	HELP_COMMAND  CommandType = "HELP"
)

type SpeedCommandType string

const (
	LOW_SPEED    SpeedCommandType = "low"
	MEDIUM_SPEED SpeedCommandType = "medium"
	HIGH_SPEED   SpeedCommandType = "high"
)

func (c CommandType) String() string {
	return string(c)
}

func (c SpeedCommandType) String() string {
	return string(c)
}

const (
	HELP_MESSAGE = `You can use 5 commands :
	- SPEED (low|medium|high) <node number> : change the speed of a node. For example: 'SPEED high 2' will change the speed of node 2 to high.
	- CRASH <node number> : crash a node. For example: 'CRASH 2' will crash node 2.
	- START : start the cluster. You can use this command only once.
	- STOP : stop the cluster. This command will kill the program.
	- HELP : display this message.`
	SPEED_COMMAND_USAGE         = "The SPEED command must have the following form: `SPEED (low|medium|high) <node number>`. For example: 'SPEED high 2'"
	CRASH_COMMAND_USAGE         = "The CRASH command must have the following form: `CRASH <node number>`. For example: 'CRASH 2'"
	INVALID_COMMAND_MESSAGE     = "Invalid command !"
	INVALID_SPEED_LEVEL_MESSAGE = "Invalid speed level !"
)
