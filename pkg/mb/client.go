package mb

import (
	"encoding/binary"
	"fmt"
	"log"
	"time"

	"github.com/goburrow/modbus"
)

type ClientWrapper struct {
	connHandler *modbus.TCPClientHandler
	client      modbus.Client
}

func Init() *ClientWrapper {
	handler := modbus.NewTCPClientHandler("127.0.0.1:502")
	handler.Timeout = 5 * time.Second
	handler.SlaveId = 1

	err := handler.Connect()
	if err != nil {
		log.Fatalf("Connection failed: %v", err)
	}

	w := &ClientWrapper{
		connHandler: handler,
		client:      modbus.NewClient(handler),
	}
	return w
}

func (cw *ClientWrapper) Close() error {
	return cw.connHandler.Close()
}

func (cw *ClientWrapper) ReadAddress(addr int) (int, error) {
	a := uint16(addr)

	results, err := cw.client.ReadHoldingRegisters(a, 1)
	if err != nil {
		return -1, fmt.Errorf("Failed to read register: %v", err)
	}

	value := binary.BigEndian.Uint16(results)

	return int(value), nil
}

func (cw *ClientWrapper) WriteAddress(addr, val int) error {
	a := uint16(addr)
	v := uint16(val)

	_, err := cw.client.WriteSingleRegister(a, v)
	if err != nil {
		return fmt.Errorf("Failed to write register: %v", err)
	}

	return nil
}
