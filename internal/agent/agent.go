package agent

import (
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/nhv96/goas/internal/client"
	"github.com/nhv96/goas/internal/payload"
	"github.com/nhv96/goas/pkg/tools"
	movecrane "github.com/nhv96/goas/pkg/tools/move_crane"
	"github.com/nhv96/goas/pkg/tools/wait"
)

var (
	ErrModelNotSupported = errors.New("model not supported")
)

// Agent struct holds the configuration and the history of its conversation messages.
type Agent struct {
	ModelName string

	SystemPrompt string

	ChatHistory []payload.ChatMessage

	Tools ToolBag

	Think  bool
	Stream bool

	client AIModelClient // ollama client server
}

type AIModelClient interface {
	SendChat(payload io.Reader, stream bool, h client.ResponseHandler) error
}

// TODO: to be replaced with ollama call to list available models
var AvailableModels = []string{"gemma4:e2b"}

// NewAgent initialize a new agent
func NewAgent(modelName, systemPrompt string, think, stream bool) (*Agent, error) {

	if err := checkModelAvailability(modelName); err != nil {
		return nil, err
	}

	client, err := client.NewModelClient("http://localhost:11434/api")
	if err != nil {
		return nil, err
	}

	agent := &Agent{
		ModelName:    modelName,
		SystemPrompt: systemPrompt,
		Think:        think,
		Stream:       stream,
		client:       client,
	}

	agent.Tools = map[string]tools.Tooler{
		"move_crane": movecrane.Init(),
		"wait":       wait.Init(),
	}

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
func (ag *Agent) Chat(userInput string) (<-chan Reply, error) {
	repChan := make(chan Reply, 5)
	if len(ag.ChatHistory) == 0 && ag.SystemPrompt != "" {
		ag.ChatHistory = []payload.ChatMessage{
			{
				Role:    payload.RoleSystem,
				Content: ag.SystemPrompt,
			},
		}
	}

	ag.ChatHistory = append(ag.ChatHistory, payload.ChatMessage{
		Role:    payload.RoleUser,
		Content: userInput,
	})

	chatPayload, err := payload.CreateChatPayload(ag.ModelName, ag.ChatHistory, ag.Think, ag.Stream, ag.Tools.Descriptions())
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

				chatPayload, err = payload.CreateChatPayload(ag.ModelName, ag.ChatHistory, ag.Think, ag.Stream, ag.Tools.Descriptions())
				if err != nil {
					return err
				}

				// signal to send another final request to sync
				// the tool calling result to the model
				pushMsgChan <- chatPayload
				close(pushMsgChan)
			} else {
				if chatResp.Message.Content != "" || chatResp.Message.Thinking != "" {
					ag.ChatHistory = append(ag.ChatHistory,
						payload.ChatMessage{
							Role:    payload.RoleAssistant,
							Content: strings.Join(msgConcat, ""),
						},
					)

					repChan <- Reply{
						From:    ag.ModelName,
						Content: chatResp.Message.Content,
					}
				}

				// TODO: careful
				close(repChan)
			}

		} else {
			chatResp, err := payload.DecodeChatStreamResponse(msg)
			if err != nil {
				return err
			}

			if chatResp.Message.Thinking != "" {
				repChan <- Reply{
					From:     ag.ModelName,
					Content:  chatResp.Message.Thinking,
					Thinking: true,
				}
			} else {
				if chatResp.Message.Content != "" {
					msgConcat = append(msgConcat, chatResp.Message.Content)

					repChan <- Reply{
						From:     ag.ModelName,
						Content:  chatResp.Message.Content,
						Thinking: false,
					}
				} else if len(chatResp.Message.ToolCalls) > 0 {
					ag.handleToolCall(chatResp)

					chatPayload, err = payload.CreateChatPayload(ag.ModelName, ag.ChatHistory, ag.Think, ag.Stream, ag.Tools.Descriptions())
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
					close(repChan)
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
				// fmt.Println("\nsending chat...")
				err = ag.client.SendChat(pl, ag.Stream, doHandleResponse)
				if err != nil {
					return
				}
			} else {
				break
			}
		}
	}()

	return repChan, err
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
			fmt.Println("calling tool:", toolCall.Function.Name)

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
