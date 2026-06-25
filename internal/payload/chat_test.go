package payload

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateChatPayload(t *testing.T) {
	model := "model_a"
	promptMsg := "msg"
	chatMessages := []ChatMessage{
		{
			Role:    RoleUser,
			Content: promptMsg,
		},
	}

	rdr, err := CreateChatPayload(model, chatMessages, true, false, []string{})
	assert.Nil(t, err)

	actualBytes, err := io.ReadAll(rdr)
	assert.Nil(t, err)

	actual := &ChatPayload{}
	err = json.Unmarshal(actualBytes, actual)
	assert.Nil(t, err)

	expectedJSON := fmt.Sprintf(`{"model":"%s","messages":[{"role":"user","content":"%s"}],"think":true,"stream":false}`, model, promptMsg)
	expect := &ChatPayload{}
	err = json.Unmarshal([]byte(expectedJSON), expect)
	assert.Nil(t, err)

	assert.Equal(t, expect, actual)
}

func TestDecodeChatStreamResponse(t *testing.T) {

	t.Run("not thinking", func(t *testing.T) {
		msg := `{"model":"gemma4:e2b","created_at":"2026-06-08T18:57:40.370507938Z","message":{"role":"assistant","content":"Hey"},"done":true,"done_reason":"stop","total_duration":18142616540,"load_duration":330261605,"prompt_eval_count":17,"prompt_eval_duration":415154000,"eval_count":222,"eval_duration":17394391000}`

		actualResp, err := DecodeChatStreamResponse(bytes.NewBufferString(msg))
		assert.Nil(t, err)

		tm, err := time.Parse(time.RFC3339, "2026-06-08T18:57:40.370507938Z")
		assert.Nil(t, err)

		var (
			ttd time.Duration = 18142616540
			ld  time.Duration = 330261605
			ped time.Duration = 415154000
			ed  time.Duration = 17394391000
		)

		expectResp := &ChatResponse{
			Model:     "gemma4:e2b",
			CreatedAt: tm,
			Message: ChatMessageResponse{
				Role:     RoleAssistant,
				Content:  "Hey",
				Thinking: "",
			},
			Done:               true,
			DoneReason:         "stop",
			TotalDuration:      ttd,
			LoadDuration:       ld,
			PromptEvalCount:    17,
			PromptEvalDuration: ped,
			EvalCount:          222,
			EvalDuration:       ed,
		}

		assert.Equal(t, expectResp, actualResp)
	})

	t.Run("thinking", func(t *testing.T) {
		msg := `{"model":"gemma4:e2b","created_at":"2026-06-08T18:57:40.370507938Z","message":{"role":"assistant","content":"","thinking":"thinkinggg"},"done":true,"done_reason":"stop","total_duration":18142616540,"load_duration":330261605,"prompt_eval_count":17,"prompt_eval_duration":415154000,"eval_count":222,"eval_duration":17394391000}`

		actualResp, err := DecodeChatStreamResponse(bytes.NewBufferString(msg))
		assert.Nil(t, err)

		tm, err := time.Parse(time.RFC3339, "2026-06-08T18:57:40.370507938Z")
		assert.Nil(t, err)

		var (
			ttd time.Duration = 18142616540
			ld  time.Duration = 330261605
			ped time.Duration = 415154000
			ed  time.Duration = 17394391000
		)

		expectResp := &ChatResponse{
			Model:     "gemma4:e2b",
			CreatedAt: tm,
			Message: ChatMessageResponse{
				Role:     RoleAssistant,
				Content:  "",
				Thinking: "thinkinggg",
			},
			Done:               true,
			DoneReason:         "stop",
			TotalDuration:      ttd,
			LoadDuration:       ld,
			PromptEvalCount:    17,
			PromptEvalDuration: ped,
			EvalCount:          222,
			EvalDuration:       ed,
		}

		assert.Equal(t, expectResp, actualResp)

	})
}
