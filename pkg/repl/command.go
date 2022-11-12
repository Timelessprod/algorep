package repl

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Timelessprod/algorep/pkg/raft"
)

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

/***********************
 * Handle the commands *
 ***********************/

func parseNodeNumber(token string) (uint32, error) {
	nodeId, err := strconv.ParseUint(token, 0, 32)
	if err != nil {
		return 0, fmt.Errorf("Invalid node number: %s", token)
	}
	if nodeId >= uint64(raft.Config.NodeCount) {
		return 0, fmt.Errorf("Node number should be between 0 and %d", raft.Config.NodeCount-1)
	}
	return uint32(nodeId), nil
}

func handleCrashCommand(tokenList []string) {
	if len(tokenList) != 2 {
		fmt.Println(CRASH_COMMAND_USAGE)
		return
	}
	nodeId, err := parseNodeNumber(tokenList[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	//TODO : handle the crash command
	fmt.Println("CRASH", nodeId)
}

func handleSpeedCommand(tokenList []string) {
	if len(tokenList) != 3 {
		fmt.Println(SPEED_COMMAND_USAGE)
		return
	}
	levelToken := strings.ToLower(tokenList[1])

	var level SpeedCommandType
	switch levelToken {
	case LOW_SPEED.String():
		level = LOW_SPEED
	case MEDIUM_SPEED.String():
		level = MEDIUM_SPEED
	case HIGH_SPEED.String():
		level = HIGH_SPEED
	default:
		fmt.Println(INVALID_SPEED_LEVEL_MESSAGE)
		fmt.Println(SPEED_COMMAND_USAGE)
		return
	}

	nodeId, err := parseNodeNumber(tokenList[2])
	if err != nil {
		fmt.Println(err)
		return
	}
	//TODO : handle the speed command
	fmt.Println("SPEED", level, nodeId)
}

func handleCommand(command string) {
	tokenList := strings.Fields(command)
	if len(tokenList) == 0 {
		//TODO
		fmt.Println("Invalid command")
	}

	commandToken := strings.ToUpper(tokenList[0])
	switch commandToken {
	case SPEED_COMMAND.String():
		handleSpeedCommand(tokenList)
	case CRASH_COMMAND.String():
		handleCrashCommand(tokenList)
	case START_COMMAND.String():
		fmt.Println("START")
		// TODO : handle the start command
	case STOP_COMMAND.String():
		fmt.Println("STOP")
		os.Exit(0)
	case HELP_COMMAND.String():
		fmt.Println(HELP_MESSAGE)
	default:
		fmt.Println(INVALID_COMMAND_MESSAGE)
		fmt.Println(HELP_MESSAGE)
	}
}
