package agent

import (
	"errors"
	"io"
	"slices"
	"strings"

	"github.com/nhv96/goas/internal/client"
	"github.com/nhv96/goas/internal/payload"
)

var (
	ErrModelNotSupported = errors.New("model not supported")
)

// Agent struct holds the configuration and the history of its conversation messages.
type Agent struct {
	ModelName string

	SystemPrompt string

	ChatHistory []payload.ChatMessage

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

	chatPayload, err := payload.CreateChatPayload(ag.ModelName, ag.ChatHistory, ag.Think, ag.Stream)
	if err != nil {
		return nil, err
	}

	thinking := false
	msgConcat := []string{}

	doHandleResponse := func(msg io.Reader) error {

		if !ag.Stream {
			chatResp, err := payload.DecodeChatResponse(msg)
			if err != nil {
				return err
			}

			msgConcat = append(msgConcat, chatResp.Message.Content)

			repChan <- Reply{
				From:    ag.ModelName,
				Content: chatResp.Message.Content,
			}

			close(repChan)

			ag.ChatHistory = append(ag.ChatHistory,
				payload.ChatMessage{
					Role:    payload.RoleAssistant,
					Content: strings.Join(msgConcat, ""),
				},
			)
		} else {
			chatResp, err := payload.DecodeChatStreamResponse(msg)
			if err != nil {
				return err
			}

			if chatResp.Message.Thinking != "" {
				thinking = true

				repChan <- Reply{
					From:     ag.ModelName,
					Content:  chatResp.Message.Thinking,
					Thinking: thinking,
				}
			} else {
				// catch the end of thinking response
				if thinking {
					thinking = false
					// inject 2 new lines
					chatResp.Message.Content = "\n\n" + chatResp.Message.Content
				}

				msgConcat = append(msgConcat, chatResp.Message.Content)

				repChan <- Reply{
					From:     ag.ModelName,
					Content:  chatResp.Message.Content,
					Thinking: thinking,
				}
			}

			if chatResp.Done {
				close(repChan)

				ag.ChatHistory = append(ag.ChatHistory,
					payload.ChatMessage{
						Role:    payload.RoleAssistant,
						Content: strings.Join(msgConcat, ""),
					},
				)
			}
		}

		return nil
	}

	go func() {
		err = ag.client.SendChat(chatPayload, ag.Stream, doHandleResponse)
		if err != nil {
			return
		}
	}()

	return repChan, err
}
