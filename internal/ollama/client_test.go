package ollama

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewClient(t *testing.T) {
	baseURL := "http://localhost:1234/apiv2"

	mc, err := NewModelClient(baseURL)

	assert.Nil(t, err)
	assert.Equal(t, baseURL, mc.clientURL.String())
}

func Test_SendChat(t *testing.T) {
	mockHandler := func(m io.Reader) error {
		b, err := io.ReadAll(m)
		assert.Nil(t, err)
		assert.Equal(t, []byte(`{"think":"","content":"abc"}`), b)

		return nil
	}

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/chat", r.URL.Path)

		w.WriteHeader(200)
		w.Write([]byte(`{"think":"","content":"abc"}`))
	}))

	defer mockServer.Close()

	mc, err := NewModelClient(mockServer.URL)
	assert.Nil(t, err)

	err = mc.SendChat(bytes.NewReader([]byte(`abc`)), false, mockHandler)
	assert.Nil(t, err)
}

func Test_SendChatStream(t *testing.T) {
	finalMsg := "hello world !!!"

	msgConcat := []string{}

	mockHandler := func(m io.Reader) error {
		b, err := io.ReadAll(m)
		assert.Nil(t, err)

		msgConcat = append(msgConcat, string(b))

		return nil
	}

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flusher, _ := w.(http.Flusher)

		assert.Equal(t, "/chat", r.URL.Path)

		w.WriteHeader(200)

		flusher.Flush()

		msgs := strings.Split(finalMsg, " ")
		for _, m := range msgs {
			_, _ = w.Write([]byte(m + "\n"))
			flusher.Flush()
		}
	}))

	defer mockServer.Close()

	mc, err := NewModelClient(mockServer.URL)
	assert.Nil(t, err)

	err = mc.SendChat(bytes.NewReader([]byte(`abc`)), true, mockHandler)
	assert.Nil(t, err)
	assert.Equal(t, finalMsg, strings.Join(msgConcat, " "))
}
