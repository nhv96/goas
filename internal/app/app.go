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
type Application[T agent.AgentReply] struct {
	Agent agent.Agentor[T]

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

// NewApplication creates new app
func NewApplication(cfg *Config) (*Application[agent.Envelop], error) {
	// cfg.SystemPrompt = `You are an AI agent that control a crane system.
	// You will receive order from the user to move the crane and you must execute that request.
	// You must look at the tool output and inform the result of your tool calling.`

	cfg.SystemPrompt = `You **must** act as an expert in software engineering.
	You provide the complete, clean code implementation, followed by a brief explanation of how it works.
	Ensure the code is properly typed and commented.
	You also provide code review with best practices and identify improvements, risks or bugs.`

	ag, err := agent.NewAgent(cfg.ModelName, cfg.SystemPrompt, cfg.Think, cfg.Stream)
	if err != nil {
		return nil, err
	}

	app := &Application[agent.Envelop]{
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
		userInput := "Hi, briefly introduce yourself."

		a.chatAndDisplay(userInput, false)
	}

	for scanner.Scan() {
		userInput := scanner.Text()

		if a.shouldEnd(strings.TrimSpace(strings.ToLower(userInput))) {
			fmt.Fprintln(a.output, "Goodbye!")
			break
		}

		useTool, userInput := a.shouldUseTools(userInput)

		a.chatAndDisplay(userInput, useTool)
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

func (a *Application[T]) shouldUseTools(in string) (bool, string) {
	toolKeyword := "/tool"
	newStr := ""
	if strings.Contains(in, toolKeyword) {
		newStr = strings.Replace(in, toolKeyword, "", -1)
		newStr = strings.TrimSpace(newStr)

		return true, newStr
	}

	return false, in
}

func (a *Application[T]) chatAndDisplay(userInput string, useTool bool) {
	// flag to catch when the thinking has started/stopped
	thinking := false

	replyChan, err := a.Agent.Chat(userInput, useTool)
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
