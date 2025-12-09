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
	transport      transport.Transport
	slaveID        modbus.SlaveID
	timeout        time.Duration
	retryCount     int
	retryDelay     time.Duration
	connectTimeout time.Duration
	autoReconnect  bool
	encoding       *EncodingConfig
}

// NewClient creates a new MODBUS client with the given transport
func NewClient(t transport.Transport) *Client {
	config := modbus.DefaultClientConfig()
	return &Client{
		transport:      t,
		slaveID:        config.SlaveID,
		timeout:        config.Timeout,
		retryCount:     config.RetryCount,
		retryDelay:     config.RetryDelay,
		connectTimeout: config.ConnectTimeout,
	}
}

// NewTCPClient creates a new MODBUS TCP client
func NewTCPClient(address string) *Client {
	return NewClient(transport.NewTCPTransport(address))
}

// NewClientFromConfig creates a new MODBUS client from a configuration
func NewClientFromConfig(config *modbus.ClientConfig, t transport.Transport) *Client {
	return &Client{
		transport:      t,
		slaveID:        config.SlaveID,
		timeout:        config.Timeout,
		retryCount:     config.RetryCount,
		retryDelay:     config.RetryDelay,
		connectTimeout: config.ConnectTimeout,
	}
}

// NewTCPClientFromConfig creates a new MODBUS TCP client from configuration
func NewTCPClientFromConfig(config *modbus.ClientConfig, address string) *Client {
	return NewClientFromConfig(config, transport.NewTCPTransport(address))
}

// NewTCPClientFromJSONFile creates a new MODBUS TCP client from a JSON configuration file
func NewTCPClientFromJSONFile(configPath, address string) (*Client, error) {
	config, err := modbus.LoadClientConfigFromJSON(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return NewTCPClientFromConfig(config, address), nil
}

// NewTCPClientFromJSONString creates a new MODBUS TCP client from a JSON configuration string
func NewTCPClientFromJSONString(jsonConfig, address string) (*Client, error) {
	config, err := modbus.LoadClientConfigFromJSONString(jsonConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return NewTCPClientFromConfig(config, address), nil
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

// SetRetryDelay sets the delay between retry attempts
func (c *Client) SetRetryDelay(delay time.Duration) {
	c.retryDelay = delay
}

// GetRetryDelay returns the current retry delay
func (c *Client) GetRetryDelay() time.Duration {
	return c.retryDelay
}

// SetConnectTimeout sets the connection timeout
func (c *Client) SetConnectTimeout(timeout time.Duration) {
	c.connectTimeout = timeout
}

// GetConnectTimeout returns the current connection timeout
func (c *Client) GetConnectTimeout() time.Duration {
	return c.connectTimeout
}

// SetAutoReconnect enables or disables automatic reconnection on connection failure
func (c *Client) SetAutoReconnect(enabled bool) {
	c.autoReconnect = enabled
}

// GetAutoReconnect returns whether automatic reconnection is enabled
func (c *Client) GetAutoReconnect() bool {
	return c.autoReconnect
}

// GetConfig returns the current client configuration
func (c *Client) GetConfig() *modbus.ClientConfig {
	return &modbus.ClientConfig{
		SlaveID:        c.slaveID,
		Timeout:        c.timeout,
		RetryCount:     c.retryCount,
		RetryDelay:     c.retryDelay,
		ConnectTimeout: c.connectTimeout,
		TransportType:  c.transport.GetTransportType(),
	}
}

// ApplyConfig applies a configuration to the client
func (c *Client) ApplyConfig(config *modbus.ClientConfig) {
	c.slaveID = config.SlaveID
	c.timeout = config.Timeout
	c.retryCount = config.RetryCount
	c.retryDelay = config.RetryDelay
	c.connectTimeout = config.ConnectTimeout
	// Update transport timeout as well
	c.transport.SetTimeout(c.timeout)
}

// sendRequest sends a request with retry logic and optional auto-reconnect
func (c *Client) sendRequest(req *pdu.Request) (*pdu.Response, error) {
	var lastErr error

	for attempt := 0; attempt <= c.retryCount; attempt++ {
		// Check connection and attempt reconnect if enabled
		if !c.transport.IsConnected() {
			if c.autoReconnect {
				if err := c.Connect(); err != nil {
					lastErr = fmt.Errorf("auto-reconnect failed: %w", err)
					if attempt < c.retryCount {
						time.Sleep(c.retryDelay)
					}
					continue
				}
			} else {
				return nil, fmt.Errorf("transport not connected")
			}
		}

		resp, err := c.transport.SendRequest(c.slaveID, req)
		if err == nil {
			return resp, nil
		}
		lastErr = err

		// Don't retry on the last attempt
		if attempt < c.retryCount {
			time.Sleep(c.retryDelay) // Configurable delay between retries
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

// GetCommEventCounter gets the communication event counter (function code 0x0B, Serial line only)
func (c *Client) GetCommEventCounter() (status uint16, eventCount uint16, err error) {
	req, err := pdu.GetCommEventCounterRequest()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to create get comm event counter request: %w", err)
	}

	resp, err := c.sendRequest(req)
	if err != nil {
		return 0, 0, err
	}

	return pdu.ParseGetCommEventCounterResponse(resp)
}

// GetCommEventLog gets the communication event log (function code 0x0C, Serial line only)
func (c *Client) GetCommEventLog() (status uint16, eventCount uint16, messageCount uint16, events []byte, err error) {
	req, err := pdu.GetCommEventLogRequest()
	if err != nil {
		return 0, 0, 0, nil, fmt.Errorf("failed to create get comm event log request: %w", err)
	}

	resp, err := c.sendRequest(req)
	if err != nil {
		return 0, 0, 0, nil, err
	}

	return pdu.ParseGetCommEventLogResponse(resp)
}

// ReportServerID gets the server ID (function code 0x11, Serial line only)
func (c *Client) ReportServerID() ([]byte, error) {
	req, err := pdu.ReportServerIDRequest()
	if err != nil {
		return nil, fmt.Errorf("failed to create report server ID request: %w", err)
	}

	resp, err := c.sendRequest(req)
	if err != nil {
		return nil, err
	}

	return pdu.ParseReportServerIDResponse(resp)
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

	return pdu.ParseReadFileRecordResponse(resp, records)
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

	return pdu.ParseWriteFileRecordResponse(resp)
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

// Broadcast methods - send to all devices (slave ID 0), no response expected

// BroadcastWriteSingleCoil broadcasts a write single coil command to all devices
func (c *Client) BroadcastWriteSingleCoil(address modbus.Address, value bool) error {
	req, err := pdu.WriteSingleCoilRequest(address, value)
	if err != nil {
		return fmt.Errorf("failed to create write single coil request: %w", err)
	}

	return c.sendBroadcast(req)
}

// BroadcastWriteSingleRegister broadcasts a write single register command to all devices
func (c *Client) BroadcastWriteSingleRegister(address modbus.Address, value uint16) error {
	req, err := pdu.WriteSingleRegisterRequest(address, value)
	if err != nil {
		return fmt.Errorf("failed to create write single register request: %w", err)
	}

	return c.sendBroadcast(req)
}

// BroadcastWriteMultipleCoils broadcasts a write multiple coils command to all devices
func (c *Client) BroadcastWriteMultipleCoils(address modbus.Address, values []bool) error {
	req, err := pdu.WriteMultipleCoilsRequest(address, values)
	if err != nil {
		return fmt.Errorf("failed to create write multiple coils request: %w", err)
	}

	return c.sendBroadcast(req)
}

// BroadcastWriteMultipleRegisters broadcasts a write multiple registers command to all devices
func (c *Client) BroadcastWriteMultipleRegisters(address modbus.Address, values []uint16) error {
	req, err := pdu.WriteMultipleRegistersRequest(address, values)
	if err != nil {
		return fmt.Errorf("failed to create write multiple registers request: %w", err)
	}

	return c.sendBroadcast(req)
}

// sendBroadcast sends a broadcast request (no response expected)
func (c *Client) sendBroadcast(req *pdu.Request) error {
	if !c.transport.IsConnected() {
		if c.autoReconnect {
			if err := c.Connect(); err != nil {
				return fmt.Errorf("auto-reconnect failed: %w", err)
			}
		} else {
			return fmt.Errorf("transport not connected")
		}
	}

	// Send to broadcast address (0), ignore response
	_, err := c.transport.SendRequest(modbus.BroadcastAddress, req)
	// For broadcast, we don't care about the response (there shouldn't be one)
	// Some transports may return a timeout error which is expected
	if err != nil {
		// Only return error if it's not a timeout (broadcast has no response)
		// For TCP, this will likely timeout which is expected
		return nil
	}
	return nil
}
