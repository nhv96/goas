package ollama

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/url"
)

var (
	ErrServerCallFailed          = errors.New("failed calling model server")
	ErrStreamHandlerTypeNotMatch = errors.New("handler function not match type")
)

type BackendServer interface {
	SendChat(payload io.Reader, stream bool, h ResponseHandler) error
}

type ModelClient struct {
	clientURL *url.URL

	client *http.Client
}

func NewModelClient(baseAPI string) (*ModelClient, error) {
	client := &http.Client{}

	u, err := url.Parse(baseAPI)
	if err != nil {
		return nil, err
	}

	mc := &ModelClient{
		clientURL: u,
		client:    client,
	}

	return mc, nil
}

// ResponseHandler is the function to handle messages from server response
type ResponseHandler func(message io.Reader) error

// SendChat sends the payload to model server and return whatever response from the server.
// If run with stream mode, the function expects a handler to handle the streaming messages.
func (mc *ModelClient) SendChat(payload io.Reader, stream bool, handler ResponseHandler) error {
	endpointURL := mc.clientURL.JoinPath("chat")

	req, err := http.NewRequest(http.MethodPost, endpointURL.String(), payload)
	if err != nil {
		return err
	}

	resp, err := mc.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return ErrServerCallFailed
	}

	if stream {
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			txt := scanner.Text()

			er := handler(bytes.NewBufferString(txt))
			if er != nil {
				return er
			}
		}

		if err := scanner.Err(); err != nil {
			return err
		}

		return nil
	}

	return handler(resp.Body)
}
