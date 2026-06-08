package agent

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/nhv96/goas/internal/prompt"
)

var (
	BaseURl      = "http://localhost:11434/api"
	ModelName    = "gemma4:e2b"
	SystemPrompt = "You are Jarvis, my personal AI assistant."
)

func Chat(think, stream bool) {
	url := fmt.Sprintf("%s/chat", BaseURl)

	client := &http.Client{}

	// Create a scanner to read from standard input
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Printf("You are chatting with: %s. Type 'exit' to quit.\n", ModelName)
	fmt.Print("> ")

	chatHist := []prompt.ChatMessage{}

	// Loop indefinitely, reading input line by line
	for scanner.Scan() {
		userInput := scanner.Text()

		// Clean up whitespace and check for the exit keyword
		if strings.TrimSpace(strings.ToLower(userInput)) == "exit" {
			fmt.Println("Goodbye!")
			break
		}

		payload, err := prompt.CreateChatPrompt(ModelName, userInput, chatHist, false, false)
		if err != nil {
			panic(err)
		}

		req, err := http.NewRequest(http.MethodPost, url, payload)
		if err != nil {
			panic(err)
		}

		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		modelResp, err := prompt.DecodeChatResponse(resp.Body)
		if err != nil {
			panic(err)
		}

		chatHist = append(chatHist, prompt.ChatMessage{
			Role:    modelResp.Message.Role,
			Content: modelResp.Message.Content,
		})

		fmt.Println(ModelName, ":", modelResp.Message.Content)

		// Print the prompt for the next input
		fmt.Print("> ")
	}

	// Check for any errors during scanning
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "error reading standard input:", err)
	}
}
