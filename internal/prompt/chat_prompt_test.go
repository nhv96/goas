package prompt

import (
	"encoding/json"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateChatPrompt(t *testing.T) {
	model := "model_a"
	promptMsg := "msg"
	chatHistory := []ChatMessage{}

	rdr, err := CreateChatPrompt(model, promptMsg, chatHistory, true, false)
	assert.Nil(t, err)

	actualBytes, err := io.ReadAll(rdr)
	assert.Nil(t, err)

	actual := &ChatPrompt{}
	err = json.Unmarshal(actualBytes, actual)
	assert.Nil(t, err)

	expectedJSON := fmt.Sprintf(`{"model":"%s","messages":[{"role":"user","content":"%s"}],"think":true,"stream":false}`, model, promptMsg)
	expect := &ChatPrompt{}
	err = json.Unmarshal([]byte(expectedJSON), expect)
	assert.Nil(t, err)

	assert.Equal(t, expect, actual)
}
