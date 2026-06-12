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

	SystemPrompt string
}

// NewApplication creates new app
func NewApplication(cfg *Config) (*Application, error) {
	sysPrompt := `You are a traveling agency that provide helpful suggestions and planning of travel trips.
	You must always start your response with Dear Madam/Sir.
	When in doubt, do not assume and you must ask questions to clarify what to do.`

	ag, err := agent.NewAgent(cfg.ModelName, sysPrompt, cfg.Think, cfg.Stream)
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
				displayOutput := ""
				if rep.Thinking {
					displayOutput = colors.Yellow(rep.Content)
				} else {
					displayOutput = colors.Bold(colors.Green)(rep.Content)
				}
				fmt.Printf("%s", displayOutput)
			} else {
				fmt.Println()
				break
			}
		}

		fmt.Println()
		fmt.Print("> ")
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "error reading standard input:", err)
	}
}
