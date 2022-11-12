package repl

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Timelessprod/algorep/pkg/raft"
)

type ClientNode struct {
	Id       uint32
	NodeCard raft.NodeCard

	Channel raft.ChannelContainer
}

func (client *ClientNode) parseNodeNumber(token string) (uint32, error) {
	nodeId, err := strconv.ParseUint(token, 0, 32)
	if err != nil {
		return 0, fmt.Errorf("Invalid node number: %s", token)
	}
	if nodeId >= uint64(raft.Config.SchedulerNodeCount) {
		return 0, fmt.Errorf("Node number should be between 0 and %d", raft.Config.SchedulerNodeCount-1)
	}
	return uint32(nodeId), nil
}

func (client *ClientNode) handleCrashCommand(tokenList []string) {
	if len(tokenList) != 2 {
		fmt.Println(CRASH_COMMAND_USAGE)
		return
	}
	nodeId, err := client.parseNodeNumber(tokenList[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	//TODO : handle the crash command
	fmt.Println("CRASH", nodeId)
}

func (client *ClientNode) handleSpeedCommand(tokenList []string) {
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

	nodeId, err := client.parseNodeNumber(tokenList[2])
	if err != nil {
		fmt.Println(err)
		return
	}
	//TODO : handle the speed command
	fmt.Println("SPEED", level, nodeId)
}

func (client *ClientNode) handleStartCommand() {
	fmt.Println("Starting all nodes...")
	//TODO : handle the start command
}

func (client *ClientNode) handleCommand(command string) {
	tokenList := strings.Fields(command)
	if len(tokenList) == 0 {
		//TODO
		fmt.Println("Invalid command")
	}

	commandToken := strings.ToUpper(tokenList[0])
	switch commandToken {
	case SPEED_COMMAND.String():
		client.handleSpeedCommand(tokenList)
	case CRASH_COMMAND.String():
		client.handleCrashCommand(tokenList)
	case START_COMMAND.String():

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

func (client *ClientNode) printPrompt() {
	fmt.Print("\n[Chaos Monkey REPL] $ ")
}

func (client *ClientNode) Init(id uint32) {
	client.Id = id
	client.NodeCard = raft.NodeCard{Id: id, Type: raft.ClientNodeType}
	client.Channel = raft.ChannelContainer{
		RequestCommand:  make(chan raft.RequestCommandRPC, raft.Config.ChannelBufferSize),
		ResponseCommand: make(chan raft.ResponseCommandRPC, raft.Config.ChannelBufferSize),
	}
}

func (client *ClientNode) Run() {
	fmt.Println("Welcome to the Chaos Monkey REPL !")
	fmt.Println("==================================")
	fmt.Println(HELP_MESSAGE)
	fmt.Println()

	reader := bufio.NewScanner(os.Stdin)
	client.printPrompt()
	for reader.Scan() {
		fmt.Print(">>> ")
		client.handleCommand(reader.Text())
		client.printPrompt()
	}
}
