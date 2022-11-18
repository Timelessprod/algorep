package repl

import (
	"bufio"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/Timelessprod/algorep/pkg/logging"
	"github.com/Timelessprod/algorep/pkg/raft"
	"go.uber.org/zap"
)

var logger *zap.Logger = logging.Logger

/*** UTILS ***/

// GetRandomSchedulerNodeId returns a random scheduler node id
func GetRandomSchedulerNodeId() uint32 {
	return uint32(rand.Intn(int(raft.Config.SchedulerNodeCount)))
}

// printPrompt prints the prompt before reading a command
func printPrompt() {
	fmt.Print("\n[Chaos Monkey REPL] $ ")
}

/*****************
 ** Client Node **
 *****************/

type ClientNode struct {
	Id               uint32
	NodeCard         raft.NodeCard
	LastLeaderId     uint32
	ClusterIsStarted bool

	Channel raft.ChannelContainer
}

// Init initializes the client node
func (client *ClientNode) Init(id uint32) {
	client.Id = id
	client.NodeCard = raft.NodeCard{Id: id, Type: raft.ClientNodeType}
	client.Channel = raft.ChannelContainer{
		RequestCommand:  make(chan raft.RequestCommandRPC, raft.Config.ChannelBufferSize),
		ResponseCommand: make(chan raft.ResponseCommandRPC, raft.Config.ChannelBufferSize),
	}
	client.LastLeaderId = 0 // Valeur par défaut le temps de trouver le leader
	client.ClusterIsStarted = false
}

// Run the client node
func (client *ClientNode) Run() {
	fmt.Println("Welcome to the Chaos Monkey REPL !")
	fmt.Println("==================================")
	fmt.Println(HELP_MESSAGE)
	fmt.Println()

	reader := bufio.NewScanner(os.Stdin)
	printPrompt()
	for reader.Scan() {
		fmt.Print(">>> ")
		client.handleCommand(reader.Text())
		printPrompt()
	}
}

// sendMessageToLeader sends a message to the leader
func (client *ClientNode) sendMessageToLeader(message raft.RequestCommandRPC) (*raft.ResponseCommandRPC, error) {
	for i := 0; i < int(raft.Config.MaxRetryToFindLeader); i++ {
		requestChannel := raft.Config.NodeChannelMap[raft.SchedulerNodeType][client.LastLeaderId].RequestCommand
		responseChannel := raft.Config.NodeChannelMap[client.NodeCard.Type][client.NodeCard.Id].ResponseCommand
		message.ToNode = raft.NodeCard{Id: client.LastLeaderId, Type: raft.SchedulerNodeType}
		requestChannel <- message
		select {
		case response := <-responseChannel:
			// If LeaderId given by node is -1
			// it means that node does not know who is the leader
			if response.LeaderId == raft.NO_LEADER {
				logger.Warn("Leader is unknown. Check random node !",
					zap.Uint32("tested nodeId", client.LastLeaderId),
				)
				client.LastLeaderId = GetRandomSchedulerNodeId()
				continue
			}

			// Check if node connected to is still leader
			if response.LeaderId != int(client.LastLeaderId) {
				logger.Warn("Leader has changed",
					zap.Uint32("old", client.LastLeaderId),
					zap.Uint32("new", uint32(response.LeaderId)),
				)
				client.LastLeaderId = uint32(response.LeaderId)
				continue
			}

			// Else if the tested node is the leader
			return &response, nil
		case <-time.After(raft.Config.MaxFindLeaderTimeout):
			logger.Warn("Node is not responding, trying to find new leader with random node...",
				zap.Int("try", i+1),
				zap.Uint32("tested nodeId", client.LastLeaderId),
			)
			client.LastLeaderId = GetRandomSchedulerNodeId()
		}
	}
	logger.Error("No response from leader after several tries")
	return nil, errors.New("No response from leader after several tries")
}

// handleSubmitCommand handles the submit job command
func (client *ClientNode) handleSubmitCommand(tokenList []string) {
	if len(tokenList) != 2 {
		fmt.Println(SUBMIT_COMMAND_USAGE)
		return
	}

	if !client.ClusterIsStarted {
		fmt.Println(NOT_STARTED_MESSAGE)
		return
	}

	jobFilePath := tokenList[1]
	fmt.Print("Submitting job ", jobFilePath, "... ")
	request := raft.RequestCommandRPC{
		FromNode:    client.NodeCard,
		CommandType: raft.AppendEntryCommand,
		Message:     "Submit job " + jobFilePath,
	}
	_, err := client.sendMessageToLeader(request)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	//TODO : check path and read file
	//TODO : send job to leader
	//TODO : read response
	fmt.Println("Done.")
}

// handleStartCommand handles the start cluster command
func (client *ClientNode) handleCrashCommand(tokenList []string) {
	if len(tokenList) != 2 {
		fmt.Println(CRASH_COMMAND_USAGE)
		return
	}
	nodeId, err := parseNodeNumber(tokenList[1])
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

// handleReplCommand handles the recover command
func (client *ClientNode) handleRecoverCommand(tokenList []string) {
	if len(tokenList) != 2 {
		fmt.Println(RECOVER_COMMAND_USAGE)
		return
	}
	nodeId, err := parseNodeNumber(tokenList[1])
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

// handleSpeedCommand handles the speed command
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

	nodeId, err := parseNodeNumber(tokenList[2])
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

// handleStartCommand handles the start cluster command
func (client *ClientNode) handleStartCommand() {
	fmt.Print("Starting all nodes... ")
	client.ClusterIsStarted = true
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

// handleStopCommand handles the stop cluster command
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
	case SUBMIT_COMMAND.String():
		client.handleSubmitCommand(tokenList)
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
