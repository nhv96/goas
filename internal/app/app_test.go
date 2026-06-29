package app

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockAgent struct {
	Name       string
	MockStream chan MockReply
	MockError  error
}

type MockReply struct {
	Thinking bool
	Content  string
}

func (r MockReply) IsThinking() bool {
	return r.Thinking
}

func (r MockReply) GetContent() string {
	return r.Content
}

func (m *MockAgent) Chat(userInput string, useTool bool) (<-chan MockReply, error) {
	if m.MockError != nil {
		return nil, m.MockError
	}

	return m.MockStream, nil
}

func (m *MockAgent) GetName() string {
	return m.Name
}

func Test_AppChat(t *testing.T) {
	mockChan := make(chan MockReply, 2)
	mockChan <- MockReply{Thinking: true, Content: "i am thinking"}
	mockChan <- MockReply{Thinking: false, Content: "i am done"}
	close(mockChan)

	mock := &MockAgent{Name: "mock agent", MockStream: mockChan, MockError: nil}

	in := strings.NewReader("hey")
	out := bytes.Buffer{}

	app := &Application[MockReply]{Agent: mock, input: in, output: &out}

	app.Start()

	expected := `You are chatting with: mock agent. Type 'exit' to quit.
> 
[31mmock agent[0m: [33mi am thinking[0m[1;32mi am done[0m

> `

	assert.Equal(t, expected, out.String())
}
