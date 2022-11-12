package repl

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Timelessprod/algorep/pkg/logging"
	"github.com/Timelessprod/algorep/pkg/raft"
	"go.uber.org/zap"
)

var logger *zap.Logger = logging.Logger

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
	fmt.Print("Crashing the node ", nodeId, "... ")
	logger.Warn("Crash a node", zap.Uint32("nodeId", nodeId))
	request := raft.RequestCommandRPC{
		FromNode:    client.NodeCard,
		ToNode:      raft.NodeCard{Id: nodeId, Type: raft.SchedulerNodeType},
		CommandType: raft.CrashCommand,
	}
	raft.Config.NodeChannelMap[raft.SchedulerNodeType][nodeId].RequestCommand <- request
	fmt.Println("Done.")
}

func (client *ClientNode) handleRecoverCommand(tokenList []string) {
	if len(tokenList) != 2 {
		fmt.Println(RECOVER_COMMAND_USAGE)
		return
	}
	nodeId, err := client.parseNodeNumber(tokenList[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Print("Recovering the node ", nodeId, "... ")
	logger.Warn("Recover a node", zap.Uint32("nodeId", nodeId))
	request := raft.RequestCommandRPC{
		FromNode:    client.NodeCard,
		ToNode:      raft.NodeCard{Id: nodeId, Type: raft.SchedulerNodeType},
		CommandType: raft.RecoverCommand,
	}
	raft.Config.NodeChannelMap[raft.SchedulerNodeType][nodeId].RequestCommand <- request
	fmt.Println("Done.")
}

func (client *ClientNode) handleSpeedCommand(tokenList []string) {
	if len(tokenList) != 3 {
		fmt.Println(SPEED_COMMAND_USAGE)
		return
	}
	levelToken := strings.ToLower(tokenList[1])

	var latency time.Duration
	switch levelToken {
	case LOW_SPEED.String():
		latency = raft.LowNodeSpeed
	case MEDIUM_SPEED.String():
		latency = raft.MediumNodeSpeed
	case HIGH_SPEED.String():
		latency = raft.HighNodeSpeed
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
	fmt.Print("Setting speed to ", levelToken, " for node ", nodeId, "... ")
	logger.Info("Change Speed for a node",
		zap.Uint32("nodeId", nodeId),
		zap.String("speed", levelToken),
		zap.Duration("latency", latency),
	)
	raft.Config.NodeSpeedList[nodeId] = latency
	fmt.Println("Done.")
}

func (client *ClientNode) handleStartCommand() {
	fmt.Print("Starting all nodes... ")
	for index, channelContainer := range raft.Config.NodeChannelMap[raft.SchedulerNodeType] {
		request := raft.RequestCommandRPC{
			FromNode:    client.NodeCard,
			ToNode:      raft.NodeCard{Id: uint32(index), Type: raft.SchedulerNodeType},
			CommandType: raft.StartCommand,
		}
		channelContainer.RequestCommand <- request
	}
	//TODO : send start command to Worker nodes
	fmt.Println("Done.")
}

func (client *ClientNode) handleCommand(command string) {
	tokenList := strings.Fields(command)
	if len(tokenList) == 0 {
		fmt.Println(INVALID_COMMAND_MESSAGE)
		fmt.Println(HELP_MESSAGE)
		return
	}

	commandToken := strings.ToUpper(tokenList[0])
	switch commandToken {
	case SPEED_COMMAND.String():
		client.handleSpeedCommand(tokenList)
	case CRASH_COMMAND.String():
		client.handleCrashCommand(tokenList)
	case RECOVER_COMMAND.String():
		client.handleRecoverCommand(tokenList)
	case START_COMMAND.String():
		client.handleStartCommand()
	case STOP_COMMAND.String():
		fmt.Println("Stopping all nodes...")
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
