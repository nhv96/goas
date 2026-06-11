package client

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
)

var (
	ErrServerCallFailed          = errors.New("failed calling model server")
	ErrStreamHandlerTypeNotMatch = errors.New("handler function not match type")
)

type ModelClient struct {
	host    string
	port    string
	baseURL string

	client http.Client
}

func NewModelClient(host, port, url string) *ModelClient {
	client := http.Client{}

	mc := &ModelClient{
		host:    host,
		port:    port,
		baseURL: url,
		client:  client,
	}

	return mc
}

// ResponseHandler is the function to handle messages from server response
type ResponseHandler func(message io.Reader) error

// SendChat sends the payload to model server and return whatever response from the server.
// If run with stream mode, the function expects a handler to handle the streaming messages.
func (mc *ModelClient) SendChat(payload io.Reader, stream bool, handler ResponseHandler) error {
	// url := mc.baseURL + "/chat" // TODO: use url package something
	url := fmt.Sprintf("%s:%s/%s/%s", mc.host, mc.port, mc.baseURL, "chat")
	req, err := http.NewRequest(http.MethodPost, url, payload)
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
	}

	return handler(resp.Body)
}
