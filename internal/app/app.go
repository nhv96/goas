package app

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/nhv96/goas/internal/agent"
	"github.com/nhv96/goas/internal/colors"
)

// Application controls the agent, printer,.. and the main chat loop
type Application[T AgentReply] struct {
	Agent Agentor[T]

	input  io.Reader
	output io.Writer

	introduced bool
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
	cfg.SystemPrompt = `You are an AI agent that control a crane system.
	You will receive order from the user to move the crane and you must execute that request.
	You must look at the tool output and inform the result of your tool calling.`

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
	// fmt.Fprint(a.output, "> ")

	if a.shouldIntroduce() {
		userInput := "introduce yourself"

		a.chatAndDisplay(userInput)
	}

	for scanner.Scan() {
		userInput := scanner.Text()

		if a.shouldEnd(strings.TrimSpace(strings.ToLower(userInput))) {
			fmt.Fprintln(a.output, "Goodbye!")
			break
		}

		a.chatAndDisplay(userInput)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(a.output, "error reading standard input:", err)
	}
}

func (a *Application[T]) shouldEnd(in string) bool {
	return slices.Contains([]string{"exit", "bye"}, in)
}

func (a *Application[T]) shouldIntroduce() bool {
	return !a.introduced
}

func (a *Application[T]) chatAndDisplay(userInput string) {
	// flag to catch when the thinking has started/stopped
	thinking := false

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
				if !thinking {
					thinking = true // thinking started
				}

				displayOutput = colors.Yellow(rep.GetContent())
			} else {
				displayOutput = colors.Bold(colors.Green)(rep.GetContent())

				if thinking {
					thinking = false // thinking stopped
					displayOutput = "\n\n" + displayOutput
				}
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
