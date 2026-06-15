package app

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/nhv96/goas/internal/agent"
	"github.com/nhv96/goas/internal/colors"
)

// Application controls the agent, printer,.. and the main chat loop
type Application[T AgentReply] struct {
	Agent Agentor[T]

	input  io.Reader
	output io.Writer
}

// Config contains config for app
type Config struct {
	ModelName string
	Think     bool
	Stream    bool

	SystemPrompt string
}

// Agentor is the interface for agent implementation
type Agentor[T AgentReply] interface {
	GetName() string
	Chat(userInput string) (<-chan T, error)
}

type AgentReply interface {
	IsThinking() bool
	GetContent() string
}

// NewApplication creates new app
func NewApplication(cfg *Config) (*Application[agent.Reply], error) {
	cfg.SystemPrompt = `You are a traveling agency that provide helpful suggestions and planning of travel trips.
	You must always start your response with Dear Madam/Sir.
	When in doubt, do not assume and you must ask questions to clarify what to do.`

	ag, err := agent.NewAgent(cfg.ModelName, cfg.SystemPrompt, cfg.Think, cfg.Stream)
	if err != nil {
		return nil, err
	}

	app := &Application[agent.Reply]{
		Agent:  ag,
		input:  os.Stdin,
		output: os.Stdout,
	}

	return app, nil
}

// Start starts the input scanner and chat loop
func (a *Application[T]) Start() {
	scanner := bufio.NewScanner(a.input)

	fmt.Fprintf(a.output, "You are chatting with: %s. Type 'exit' to quit.\n", a.Agent.GetName())
	fmt.Fprint(a.output, "> ")

	for scanner.Scan() {
		userInput := scanner.Text()

		if strings.TrimSpace(strings.ToLower(userInput)) == "exit" {
			fmt.Fprintln(a.output, "Goodbye!")
			break
		}

		replyChan, err := a.Agent.Chat(userInput)
		if err != nil {
			panic(err)
		}

		fmt.Fprintln(a.output)
		fmt.Fprint(a.output, colors.Red(a.Agent.GetName()), ": ")

		for {
			if rep, ok := <-replyChan; ok {
				displayOutput := ""
				if rep.IsThinking() {
					displayOutput = colors.Yellow(rep.GetContent())
				} else {
					displayOutput = colors.Bold(colors.Green)(rep.GetContent())
				}

				fmt.Fprintf(a.output, "%s", displayOutput)
			} else {
				fmt.Fprintln(a.output)
				break
			}
		}

		fmt.Fprintln(a.output)
		fmt.Fprint(a.output, "> ")
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(a.output, "error reading standard input:", err)
	}
}
