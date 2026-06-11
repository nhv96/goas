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

	client client.ModelClient // ollama client server
}

// TODO: to be replaced with ollama call to list available models
var AvailableModels = []string{"gemma4:e2b"}

// NewAgent initialize a new agent
func NewAgent(modelName, systemPrompt string, think, stream bool) (*Agent, error) {

	if err := checkModelAvailability(modelName); err != nil {
		return nil, err
	}

	client := client.NewModelClient("http://localhost", "11434", "api")

	agent := &Agent{
		ModelName:    modelName,
		SystemPrompt: systemPrompt,
		Think:        think,
		Stream:       stream,
		client:       *client,
	}

	return agent, nil
}

func checkModelAvailability(modelName string) error {
	if !slices.Contains(AvailableModels, modelName) {
		return ErrModelNotSupported
	}

	return nil
}

// Reply is a way to pass data between agent and application
type Reply struct {
	From    string
	Content string
}

// Chat takes in a user input prompt, then inject the system prompt
// before sending it to the model server.
func (ag *Agent) Chat(userInput string) (<-chan Reply, error) {
	repChan := make(chan Reply)
	var chatMessages []payload.ChatMessage
	if len(ag.ChatHistory) > 0 {
		chatMessages = ag.ChatHistory
	} else {
		if ag.SystemPrompt != "" {
			ag.ChatHistory = []payload.ChatMessage{
				{
					Role:    payload.RoleSystem,
					Content: ag.SystemPrompt,
				},
			}
		}
	}

	chatMessages = append(chatMessages, payload.ChatMessage{
		Role:    payload.RoleUser,
		Content: userInput,
	})

	chatPayload, err := payload.CreateChatPayload(ag.ModelName, chatMessages, ag.Think, ag.Stream)
	if err != nil {
		return nil, err
	}

	thinking := false

	doHandleResponse := func(msg io.Reader) error {
		msgConcat := []string{}

		if !ag.Stream {
			chatResp, err := payload.DecodeChatResponse(msg)
			if err != nil {
				return err
			}

			modelResponse := payload.ChatMessage{
				Role:    payload.RoleAssistant,
				Content: chatResp.Message.Content,
			}

			ag.ChatHistory = append(ag.ChatHistory, modelResponse)

			repChan <- Reply{
				From:    ag.ModelName,
				Content: modelResponse.Content,
			}

			close(repChan)
		} else {
			chatResp, err := payload.DecodeChatStreamResponse(msg)
			if err != nil {
				return err
			}

			if chatResp.Message.Thinking != "" {
				thinking = true

				repChan <- Reply{
					From:    ag.ModelName,
					Content: chatResp.Message.Thinking,
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
					From:    ag.ModelName,
					Content: chatResp.Message.Content,
				}
			}

			if chatResp.Done {
				close(repChan)
			}

			ag.ChatHistory = append(ag.ChatHistory, payload.ChatMessage{
				Role:    payload.RoleAssistant,
				Content: strings.Join(msgConcat, ""),
			})
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
