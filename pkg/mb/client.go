package mb

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/goburrow/modbus"
)

// ClientWrapper holds the Modbus client and connection handler.
type ClientWrapper struct {
	connHandler *modbus.TCPClientHandler
	client      modbus.Client
}

// Init establishes the Modbus connection. It returns a pointer to the ClientWrapper
// and an error if the connection fails.
func Init() (*ClientWrapper, error) {
	handler := modbus.NewTCPClientHandler("127.0.0.1:502")
	handler.Timeout = 5 * time.Second
	handler.SlaveId = 1

	err := handler.Connect()
	if err != nil {
		// Return the error instead of calling log.Fatalf, allowing the caller to handle it.
		return nil, fmt.Errorf("failed to connect to Modbus server: %w", err)
	}

	w := &ClientWrapper{
		connHandler: handler,
		client:      modbus.NewClient(handler),
	}
	return w, nil
}

// Close closes the underlying Modbus connection.
func (cw *ClientWrapper) Close() error {
	return cw.connHandler.Close()
}

// ReadAddress reads a single holding register and converts its 16-bit value to an integer.
func (cw *ClientWrapper) ReadAddress(addr int) (int, error) {
	if addr < 0 {
		return -1, fmt.Errorf("address cannot be negative: %d", addr)
	}

	a := uint16(addr)

	// Use the client for reading
	results, err := cw.client.ReadHoldingRegisters(a, 1)
	if err != nil {
		return -1, fmt.Errorf("failed to read register %d: %w", addr, err)
	}

	// Assuming the register holds a 16-bit unsigned integer (Modbus standard)
	value := binary.BigEndian.Uint16(results)

	return int(value), nil
}

// WriteAddress writes a single holding register value.
func (cw *ClientWrapper) WriteAddress(addr, val int) error {
	if addr < 0 {
		return fmt.Errorf("address cannot be negative: %d", addr)
	}
	
	a := uint16(addr)
	v := uint16(val)

	// Use the client for writing
	_, err := cw.client.WriteSingleRegister(a, v)
	if err != nil {
		return fmt.Errorf("failed to write register %d: %w", addr, err)
	}

	return nil
}