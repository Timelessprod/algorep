package repl

import (
	"bufio"
	"fmt"
	"os"
)

func printPrompt() {
	fmt.Print("\n[Chaos Moneky REPL] $ ")
}

func REPL() {
	reader := bufio.NewScanner(os.Stdin)
	printPrompt()
	for reader.Scan() {
		fmt.Print(">>> ")
		handleCommand(reader.Text())
		printPrompt()
	}
}
