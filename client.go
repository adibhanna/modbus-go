package modbus

import (
	"fmt"
	"time"

	"github.com/adibhanna/modbus-go/modbus"
	"github.com/adibhanna/modbus-go/pdu"
	"github.com/adibhanna/modbus-go/transport"
)

// Client represents a MODBUS client
type Client struct {
	transport  transport.Transport
	slaveID    modbus.SlaveID
	timeout    time.Duration
	retryCount int
}

// NewClient creates a new MODBUS client with the given transport
func NewClient(t transport.Transport) *Client {
	config := modbus.DefaultClientConfig()
	return &Client{
		transport:  t,
		slaveID:    config.SlaveID,
		timeout:    config.Timeout,
		retryCount: config.RetryCount,
	}
}

// NewTCPClient creates a new MODBUS TCP client
func NewTCPClient(address string) *Client {
	return NewClient(transport.NewTCPTransport(address))
}

// Connect establishes the connection
func (c *Client) Connect() error {
	c.transport.SetTimeout(c.timeout)
	return c.transport.Connect()
}

// Close closes the connection
func (c *Client) Close() error {
	return c.transport.Close()
}

// IsConnected returns true if the client is connected
func (c *Client) IsConnected() bool {
	return c.transport.IsConnected()
}

// SetSlaveID sets the slave/unit ID
func (c *Client) SetSlaveID(slaveID modbus.SlaveID) {
	c.slaveID = slaveID
}

// GetSlaveID returns the current slave/unit ID
func (c *Client) GetSlaveID() modbus.SlaveID {
	return c.slaveID
}

// SetTimeout sets the response timeout
func (c *Client) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
	c.transport.SetTimeout(timeout)
}

// GetTimeout returns the current timeout
func (c *Client) GetTimeout() time.Duration {
	return c.timeout
}

// SetRetryCount sets the number of retries on failure
func (c *Client) SetRetryCount(count int) {
	c.retryCount = count
}

// GetRetryCount returns the current retry count
func (c *Client) GetRetryCount() int {
	return c.retryCount
}

// sendRequest sends a request with retry logic
func (c *Client) sendRequest(req *pdu.Request) (*pdu.Response, error) {
	var lastErr error

	for attempt := 0; attempt <= c.retryCount; attempt++ {
		resp, err := c.transport.SendRequest(c.slaveID, req)
		if err == nil {
			return resp, nil
		}
		lastErr = err

		// Don't retry on the last attempt
		if attempt < c.retryCount {
			time.Sleep(time.Millisecond * 100) // Small delay between retries
		}
	}

	return nil, fmt.Errorf("request failed after %d attempts: %w", c.retryCount+1, lastErr)
}

// ReadCoils reads coils (function code 0x01)
func (c *Client) ReadCoils(address modbus.Address, quantity modbus.Quantity) ([]bool, error) {
	req, err := pdu.ReadCoilsRequest(address, quantity)
	if err != nil {
		return nil, fmt.Errorf("failed to create read coils request: %w", err)
	}

	resp, err := c.sendRequest(req)
	if err != nil {
		return nil, err
	}

	return pdu.ParseReadCoilsResponse(resp, quantity)
}

// ReadDiscreteInputs reads discrete inputs (function code 0x02)
func (c *Client) ReadDiscreteInputs(address modbus.Address, quantity modbus.Quantity) ([]bool, error) {
	req, err := pdu.ReadDiscreteInputsRequest(address, quantity)
	if err != nil {
		return nil, fmt.Errorf("failed to create read discrete inputs request: %w", err)
	}

	resp, err := c.sendRequest(req)
	if err != nil {
		return nil, err
	}

	return pdu.ParseReadDiscreteInputsResponse(resp, quantity)
}

// ReadHoldingRegisters reads holding registers (function code 0x03)
func (c *Client) ReadHoldingRegisters(address modbus.Address, quantity modbus.Quantity) ([]uint16, error) {
	req, err := pdu.ReadHoldingRegistersRequest(address, quantity)
	if err != nil {
		return nil, fmt.Errorf("failed to create read holding registers request: %w", err)
	}

	resp, err := c.sendRequest(req)
	if err != nil {
		return nil, err
	}

	return pdu.ParseReadHoldingRegistersResponse(resp, quantity)
}

// ReadInputRegisters reads input registers (function code 0x04)
func (c *Client) ReadInputRegisters(address modbus.Address, quantity modbus.Quantity) ([]uint16, error) {
	req, err := pdu.ReadInputRegistersRequest(address, quantity)
	if err != nil {
		return nil, fmt.Errorf("failed to create read input registers request: %w", err)
	}

	resp, err := c.sendRequest(req)
	if err != nil {
		return nil, err
	}

	return pdu.ParseReadInputRegistersResponse(resp, quantity)
}

// WriteSingleCoil writes a single coil (function code 0x05)
func (c *Client) WriteSingleCoil(address modbus.Address, value bool) error {
	req, err := pdu.WriteSingleCoilRequest(address, value)
	if err != nil {
		return fmt.Errorf("failed to create write single coil request: %w", err)
	}

	resp, err := c.sendRequest(req)
	if err != nil {
		return err
	}

	return pdu.ParseWriteSingleCoilResponse(resp, address, value)
}

// WriteSingleRegister writes a single register (function code 0x06)
func (c *Client) WriteSingleRegister(address modbus.Address, value uint16) error {
	req, err := pdu.WriteSingleRegisterRequest(address, value)
	if err != nil {
		return fmt.Errorf("failed to create write single register request: %w", err)
	}

	resp, err := c.sendRequest(req)
	if err != nil {
		return err
	}

	return pdu.ParseWriteSingleRegisterResponse(resp, address, value)
}

// WriteMultipleCoils writes multiple coils (function code 0x0F)
func (c *Client) WriteMultipleCoils(address modbus.Address, values []bool) error {
	req, err := pdu.WriteMultipleCoilsRequest(address, values)
	if err != nil {
		return fmt.Errorf("failed to create write multiple coils request: %w", err)
	}

	resp, err := c.sendRequest(req)
	if err != nil {
		return err
	}

	return pdu.ParseWriteMultipleCoilsResponse(resp, address, modbus.Quantity(len(values)))
}

// WriteMultipleRegisters writes multiple registers (function code 0x10)
func (c *Client) WriteMultipleRegisters(address modbus.Address, values []uint16) error {
	req, err := pdu.WriteMultipleRegistersRequest(address, values)
	if err != nil {
		return fmt.Errorf("failed to create write multiple registers request: %w", err)
	}

	resp, err := c.sendRequest(req)
	if err != nil {
		return err
	}

	return pdu.ParseWriteMultipleRegistersResponse(resp, address, modbus.Quantity(len(values)))
}

// MaskWriteRegister performs a mask write on a register (function code 0x16)
func (c *Client) MaskWriteRegister(address modbus.Address, andMask, orMask uint16) error {
	req, err := pdu.MaskWriteRegisterRequest(address, andMask, orMask)
	if err != nil {
		return fmt.Errorf("failed to create mask write register request: %w", err)
	}

	resp, err := c.sendRequest(req)
	if err != nil {
		return err
	}

	return pdu.ParseMaskWriteRegisterResponse(resp, address, andMask, orMask)
}

// ReadWriteMultipleRegisters reads and writes registers in one transaction (function code 0x17)
func (c *Client) ReadWriteMultipleRegisters(readAddress modbus.Address, readQuantity modbus.Quantity,
	writeAddress modbus.Address, writeValues []uint16) ([]uint16, error) {
	req, err := pdu.ReadWriteMultipleRegistersRequest(readAddress, readQuantity, writeAddress, writeValues)
	if err != nil {
		return nil, fmt.Errorf("failed to create read/write multiple registers request: %w", err)
	}

	resp, err := c.sendRequest(req)
	if err != nil {
		return nil, err
	}

	return pdu.ParseReadWriteMultipleRegistersResponse(resp, readQuantity)
}

// ReadFIFOQueue reads a FIFO queue (function code 0x18)
func (c *Client) ReadFIFOQueue(address modbus.Address) ([]uint16, error) {
	req, err := pdu.ReadFIFOQueueRequest(address)
	if err != nil {
		return nil, fmt.Errorf("failed to create read FIFO queue request: %w", err)
	}

	resp, err := c.sendRequest(req)
	if err != nil {
		return nil, err
	}

	return pdu.ParseReadFIFOQueueResponse(resp)
}

// ReadExceptionStatus reads exception status (function code 0x07, Serial line only)
func (c *Client) ReadExceptionStatus() (uint8, error) {
	req, err := pdu.ReadExceptionStatusRequest()
	if err != nil {
		return 0, fmt.Errorf("failed to create read exception status request: %w", err)
	}

	resp, err := c.sendRequest(req)
	if err != nil {
		return 0, err
	}

	return pdu.ParseReadExceptionStatusResponse(resp)
}

// Diagnostic performs a diagnostic function (function code 0x08, Serial line only)
func (c *Client) Diagnostic(subFunction uint16, data []byte) (uint16, []byte, error) {
	req, err := pdu.DiagnosticRequest(subFunction, data)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create diagnostic request: %w", err)
	}

	resp, err := c.sendRequest(req)
	if err != nil {
		return 0, nil, err
	}

	return pdu.ParseDiagnosticResponse(resp)
}

// ReadFileRecord reads file records (function code 0x14)
func (c *Client) ReadFileRecord(records []modbus.FileRecord) ([]modbus.FileRecord, error) {
	req, err := pdu.ReadFileRecordRequest(records)
	if err != nil {
		return nil, fmt.Errorf("failed to create read file record request: %w", err)
	}

	resp, err := c.sendRequest(req)
	if err != nil {
		return nil, err
	}

	// TODO: Implement ParseReadFileRecordResponse
	_ = resp
	return nil, fmt.Errorf("ReadFileRecord response parsing not yet implemented")
}

// WriteFileRecord writes file records (function code 0x15)
func (c *Client) WriteFileRecord(records []modbus.FileRecord) error {
	req, err := pdu.WriteFileRecordRequest(records)
	if err != nil {
		return fmt.Errorf("failed to create write file record request: %w", err)
	}

	resp, err := c.sendRequest(req)
	if err != nil {
		return err
	}

	// TODO: Implement ParseWriteFileRecordResponse
	_ = resp
	return fmt.Errorf("WriteFileRecord response parsing not yet implemented")
}

// ReadDeviceIdentification reads device identification (function code 0x2B/0x0E)
func (c *Client) ReadDeviceIdentification(readCode uint8, objectID uint8) (*modbus.DeviceIdentification, bool, uint8, error) {
	req, err := pdu.ReadDeviceIdentificationRequest(readCode, objectID)
	if err != nil {
		return nil, false, 0, fmt.Errorf("failed to create read device identification request: %w", err)
	}

	resp, err := c.sendRequest(req)
	if err != nil {
		return nil, false, 0, err
	}

	return pdu.ParseReadDeviceIdentificationResponse(resp)
}

// String returns a string representation of the client
func (c *Client) String() string {
	return fmt.Sprintf("ModbusClient(slave=%d, transport=%s)", c.slaveID, c.transport.String())
}
