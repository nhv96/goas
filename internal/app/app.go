package app

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/nhv96/goas/internal/agent"
	"github.com/nhv96/goas/internal/colors"
)

// Application controls the agent, printer,.. and the main chat loop
type Application struct {
	Agent *agent.Agent
}

// Config contains config for app
type Config struct {
	ModelName string
	Think     bool
	Stream    bool
}

// NewApplication creates new app
func NewApplication(cfg *Config) (*Application, error) {
	ag, err := agent.NewAgent(cfg.ModelName, "", cfg.Think, cfg.Stream)
	if err != nil {
		return nil, err
	}

	app := &Application{
		Agent: ag,
	}

	return app, nil
}

// Start starts the input scanner and chat loop
func (a *Application) Start() {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Printf("You are chatting with: %s. Type 'exit' to quit.\n", a.Agent.ModelName)
	fmt.Print("> ")

	for scanner.Scan() {
		userInput := scanner.Text()

		if strings.TrimSpace(strings.ToLower(userInput)) == "exit" {
			fmt.Println("Goodbye!")
			break
		}

		replyChan, err := a.Agent.Chat(userInput)
		if err != nil {
			panic(err)
		}

		fmt.Println()
		fmt.Print(colors.Red(a.Agent.ModelName), ": ")

		for {
			if rep, ok := <-replyChan; ok {
				fmt.Printf("%s", rep.Content)
			} else {
				fmt.Println()
				break
			}
		}

		fmt.Print("> ")
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "error reading standard input:", err)
	}
}
