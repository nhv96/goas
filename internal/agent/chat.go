package agent

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/nhv96/goas/internal/payload"
)

var (
	BaseURl       = "http://localhost:11434/api"
	ModelName     = "gemma4:e2b"
	Systempayload = "You are Jarvis, my personal AI assistant."
)

func Chat(think, stream bool) {
	url := fmt.Sprintf("%s/chat", BaseURl)

	client := &http.Client{}

	// Create a scanner to read from standard input
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Printf("You are chatting with: %s. Type 'exit' to quit.\n", ModelName)
	fmt.Print("> ")

	chatHist := []payload.ChatMessage{}

	// Loop indefinitely, reading input line by line
	for scanner.Scan() {
		userInput := scanner.Text()

		// Clean up whitespace and check for the exit keyword
		if strings.TrimSpace(strings.ToLower(userInput)) == "exit" {
			fmt.Println("Goodbye!")
			break
		}

		chatHist = append(chatHist, payload.ChatMessage{
			Role:    payload.RoleUser,
			Content: userInput,
		})

		reqPayload, err := payload.CreateChatPayload(ModelName, chatHist, think, stream)
		if err != nil {
			panic(err)
		}

		req, err := http.NewRequest(http.MethodPost, url, reqPayload)
		if err != nil {
			panic(err)
		}

		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		if stream {
			scanner := bufio.NewScanner(resp.Body)

			fmt.Println()
			fmt.Print(ModelName, ": ")

			thinking := false
			msgConcat := []string{}

			for scanner.Scan() {
				txt := scanner.Text()

				modelResp, err := payload.DecodeChatStreamResponse(bytes.NewBufferString(txt))
				if err != nil {
					panic(err)
				}

				if modelResp.Message.Thinking != "" {
					thinking = true
					fmt.Print(modelResp.Message.Thinking)
				} else {
					// catch the end of thinking response
					if thinking {
						thinking = false
						fmt.Print("\n\n")
					}

					msgConcat = append(msgConcat, modelResp.Message.Content)

					fmt.Print(modelResp.Message.Content)
				}

				if modelResp.Done {
					fmt.Println()
				}
			}

			if err := scanner.Err(); err != nil {
				panic(err)
			}

			chatHist = append(chatHist, payload.ChatMessage{
				Role:    payload.RoleAssistant, // always assistant because its message from the model
				Content: strings.Join(msgConcat, ""),
			})
		} else {
			modelResp, err := payload.DecodeChatResponse(resp.Body)
			if err != nil {
				panic(err)
			}

			chatHist = append(chatHist, payload.ChatMessage{
				Role:    modelResp.Message.Role,
				Content: modelResp.Message.Content,
			})

			fmt.Println(ModelName, ":", modelResp.Message.Content)
		}

		// Print the payload for the next input
		fmt.Println()
		fmt.Print("> ")
	}

	// Check for any errors during scanning
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "error reading standard input:", err)
	}
}
