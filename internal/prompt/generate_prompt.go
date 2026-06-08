package prompt

import (
	"bytes"
	"encoding/json"
	"io"
)

type SharedPromptParams struct {
	Think  bool `json:"think"`
	Stream bool `json:"stream"`
}

type GeneratePrompt struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	System string `json:"system"`
	SharedPromptParams
}

func CreateGeneratePrompt(model, prompt, sys_prompt string, think, stream bool) (io.Reader, error) {
	p := &GeneratePrompt{
		Model:  model,
		Prompt: prompt,
		System: sys_prompt,
		SharedPromptParams: SharedPromptParams{
			Think:  think,
			Stream: stream,
		},
	}

	payload, err := json.MarshalIndent(p, "", "	")
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(payload), nil
}
