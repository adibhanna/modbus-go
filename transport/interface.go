package transport

import (
	"time"

	"github.com/adibhanna/modbus-go/modbus"
	"github.com/adibhanna/modbus-go/pdu"
)

// Transport defines the interface for MODBUS transport layers
type Transport interface {
	// Connect establishes the connection
	Connect() error

	// Close closes the connection
	Close() error

	// IsConnected returns true if the transport is connected
	IsConnected() bool

	// SendRequest sends a request and returns the response
	SendRequest(slaveID modbus.SlaveID, request *pdu.Request) (*pdu.Response, error)

	// SetTimeout sets the response timeout
	SetTimeout(timeout time.Duration)

	// GetTimeout returns the current timeout
	GetTimeout() time.Duration

	// GetTransportType returns the transport type
	GetTransportType() modbus.TransportType

	// String returns a string representation
	String() string
}
