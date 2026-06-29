package agent

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"slices"
	"strings"

	"github.com/nhv96/goas/internal/ollama"
	"github.com/nhv96/goas/internal/payload"
	"github.com/nhv96/goas/pkg/tools"
	"github.com/nhv96/goas/pkg/tools/grep"
	listdir "github.com/nhv96/goas/pkg/tools/list_dir"
	patchfile "github.com/nhv96/goas/pkg/tools/patch_file"
	readfile "github.com/nhv96/goas/pkg/tools/read_file"
	writefile "github.com/nhv96/goas/pkg/tools/write_file"
)

var (
	ErrModelNotSupported = errors.New("model not supported")
)

// Agentor is the interface for agent implementation
type Agentor[T AgentReply] interface {
	GetName() string
	Chat(userInput string, useTool bool) (<-chan T, error)
}

// AgentReply is the struct return to application callers
type AgentReply interface {
	IsThinking() bool
	GetContent() string
}

// Agent struct holds the configuration and the history of its conversation messages.
type Agent struct {
	ModelName string

	SystemPrompt string

	ChatHistory []payload.ChatMessage

	Tools ToolBag

	Think  bool
	Stream bool

	ollama ollama.BackendServer // ollama client server
}

// TODO: to be replaced with ollama call to list available models
var AvailableModels = []string{"gemma4:e2b"}

// NewAgent initialize a new agent
func NewAgent(modelName, systemPrompt string, think, stream bool) (*Agent, error) {

	if err := checkModelAvailability(modelName); err != nil {
		return nil, err
	}

	client, err := ollama.NewModelClient("http://localhost:11434/api")
	if err != nil {
		return nil, err
	}

	agent := &Agent{
		ModelName:    modelName,
		SystemPrompt: systemPrompt,
		Think:        think,
		Stream:       stream,
		ollama:       client,
	}

	agent.Tools = map[string]tools.Tooler{
		"grep":       grep.Init(),
		"list_dir":   listdir.Init(),
		"read_file":  readfile.Init(),
		"write_file": writefile.Init(),
		"patch_file": patchfile.Init(),
	}

	if agent.SystemPrompt != "" {
		agent.ChatHistory = []payload.ChatMessage{
			{
				Role:    payload.RoleSystem,
				Content: agent.SystemPrompt,
			},
		}
	}

	// get current working dir to promp to model
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	agent.ChatHistory = append(agent.ChatHistory, payload.ChatMessage{
		Role:    payload.RoleSystem,
		Content: fmt.Sprintf("Your working directory is %s, you must **only** interact within this working directory", dir),
	})

	return agent, nil
}

func checkModelAvailability(modelName string) error {
	if !slices.Contains(AvailableModels, modelName) {
		return ErrModelNotSupported
	}

	return nil
}

func (ag *Agent) GetName() string {
	return ag.ModelName
}

// Chat takes in a user input prompt, then inject the system prompt
// before sending it to the model server.
func (ag *Agent) Chat(userInput string, useTool bool) (<-chan Envelop, error) {
	envelopChan := make(chan Envelop, 5)

	ag.ChatHistory = append(ag.ChatHistory, payload.ChatMessage{
		Role:    payload.RoleUser,
		Content: userInput,
	})

	toolDesc := []string{}
	if useTool {
		toolDesc = ag.Tools.Descriptions()
	}

	chatPayload, err := payload.CreateChatPayload(ag.ModelName, ag.ChatHistory, ag.Think, ag.Stream, toolDesc)
	if err != nil {
		return nil, err
	}

	msgConcat := []string{}

	// a flag to know when should we conclude that the streaming chat
	// is actually finished (after tool calling we need to wait for conclude msg from model).
	waitToolCall := false

	// channel to communicate with the child goroutine that send chat message.
	// if should it send another message (for after tool calling), use this channel
	// to inform the goroutine to send another message.
	// otherwise close it.
	pushMsgChan := make(chan io.Reader, 1)
	pushMsgChan <- chatPayload

	doHandleResponse := func(msg io.Reader) error {
		if !ag.Stream {
			chatResp, err := payload.DecodeChatResponse(msg)
			if err != nil {
				return err
			}

			msgConcat = append(msgConcat, chatResp.Message.Content)

			// tool call handling
			if len(chatResp.Message.ToolCalls) > 0 {
				ag.handleToolCall(chatResp)

				chatPayload, err = payload.CreateChatPayload(ag.ModelName, ag.ChatHistory, ag.Think, ag.Stream, toolDesc)
				if err != nil {
					return err
				}

				// signal to send another final request to sync
				// the tool calling result to the model
				pushMsgChan <- chatPayload
				// close(pushMsgChan) // TODO: check bug here
			} else {
				if chatResp.Message.Content != "" || chatResp.Message.Thinking != "" {
					ag.ChatHistory = append(ag.ChatHistory,
						payload.ChatMessage{
							Role:    payload.RoleAssistant,
							Content: strings.Join(msgConcat, ""),
						},
					)

					envelopChan <- Envelop{
						From:    ag.ModelName,
						Content: chatResp.Message.Content,
					}
				}

				// TODO: careful
				close(envelopChan)
			}

		} else {
			chatResp, err := payload.DecodeChatStreamResponse(msg)
			if err != nil {
				return err
			}

			if chatResp.Message.Thinking != "" {
				envelopChan <- Envelop{
					From:     ag.ModelName,
					Content:  chatResp.Message.Thinking,
					Thinking: true,
				}
			} else {
				if chatResp.Message.Content != "" {
					msgConcat = append(msgConcat, chatResp.Message.Content)

					envelopChan <- Envelop{
						From:     ag.ModelName,
						Content:  chatResp.Message.Content,
						Thinking: false,
					}
				} else if len(chatResp.Message.ToolCalls) > 0 {
					ag.handleToolCall(chatResp)

					chatPayload, err = payload.CreateChatPayload(ag.ModelName, ag.ChatHistory, ag.Think, ag.Stream, toolDesc)
					if err != nil {
						return err
					}

					// signal to send another chat request to sync
					// the tool calling result to the model
					pushMsgChan <- chatPayload

					waitToolCall = true
				}
			}

			if chatResp.Done {
				// when all tools has been done
				if !waitToolCall {
					close(envelopChan)
					close(pushMsgChan)
				}

				if len(msgConcat) > 0 {
					ag.ChatHistory = append(ag.ChatHistory,
						payload.ChatMessage{
							Role:    payload.RoleAssistant,
							Content: strings.Join(msgConcat, ""),
						},
					)
				}

				// when a message is done, we might turn off waitToolCall.
				// but the next message in the stream could be another tool call,
				// and turn on the flag again.
				waitToolCall = false
			}
		}

		return nil
	}

	go func() {
		for {
			if pl, ok := <-pushMsgChan; ok {
				fmt.Println("\nwaiting respond...")
				err = ag.ollama.SendChat(pl, ag.Stream, doHandleResponse)
				if err != nil {
					return
				}
			} else {
				break
			}
		}
	}()

	return envelopChan, err
}

func (ag *Agent) handleToolCall(chatResp *payload.ChatResponse) {
	// cheat, ollama responds empty field type
	for _, tc := range chatResp.Message.ToolCalls {
		tc.Type = "function"
	}

	ag.ChatHistory = append(ag.ChatHistory, payload.ChatMessage{
		Role:      payload.RoleAssistant,
		ToolCalls: chatResp.Message.ToolCalls,
	})

	for _, toolCall := range chatResp.Message.ToolCalls {
		if tool, ok := ag.Tools[toolCall.Function.Name]; ok {
			fmt.Println("\ncalling tool:", toolCall.Function.Name)

			callResult, err := tool.Run(toolCall.Function.Arguments)
			if err != nil {
				fmt.Println(err)
				continue
			}

			ag.ChatHistory = append(ag.ChatHistory,
				payload.ChatMessage{
					Role:     payload.RoleTool,
					ToolName: toolCall.Function.Name,
					Content:  callResult.(string),
				},
			)
		}
	}
}
