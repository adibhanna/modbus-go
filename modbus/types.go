package modbus

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// SlaveID represents a MODBUS slave/unit identifier
type SlaveID uint8

// Address represents a MODBUS register or coil address
type Address uint16

// Quantity represents a quantity of registers or coils
type Quantity uint16

// FunctionCode represents a MODBUS function code
type FunctionCode uint8

// ExceptionCode represents a MODBUS exception code
type ExceptionCode uint8

// IsException returns true if the function code represents an exception
func (fc FunctionCode) IsException() bool {
	return fc&0x80 != 0
}

// ToException converts a normal function code to its exception equivalent
func (fc FunctionCode) ToException() FunctionCode {
	return fc | 0x80
}

// FromException converts an exception function code to its normal equivalent
func (fc FunctionCode) FromException() FunctionCode {
	return fc &^ 0x80
}

// String returns a string representation of the function code
func (fc FunctionCode) String() string {
	if fc.IsException() {
		return fmt.Sprintf("Exception(%02x)", uint8(fc.FromException()))
	}

	switch fc {
	case FuncCodeReadCoils:
		return "ReadCoils"
	case FuncCodeReadDiscreteInputs:
		return "ReadDiscreteInputs"
	case FuncCodeReadHoldingRegisters:
		return "ReadHoldingRegisters"
	case FuncCodeReadInputRegisters:
		return "ReadInputRegisters"
	case FuncCodeWriteSingleCoil:
		return "WriteSingleCoil"
	case FuncCodeWriteSingleRegister:
		return "WriteSingleRegister"
	case FuncCodeReadExceptionStatus:
		return "ReadExceptionStatus"
	case FuncCodeDiagnostic:
		return "Diagnostic"
	case FuncCodeGetCommEventCounter:
		return "GetCommEventCounter"
	case FuncCodeGetCommEventLog:
		return "GetCommEventLog"
	case FuncCodeWriteMultipleCoils:
		return "WriteMultipleCoils"
	case FuncCodeWriteMultipleRegisters:
		return "WriteMultipleRegisters"
	case FuncCodeReportServerID:
		return "ReportServerID"
	case FuncCodeReadFileRecord:
		return "ReadFileRecord"
	case FuncCodeWriteFileRecord:
		return "WriteFileRecord"
	case FuncCodeMaskWriteRegister:
		return "MaskWriteRegister"
	case FuncCodeReadWriteMultipleRegs:
		return "ReadWriteMultipleRegisters"
	case FuncCodeReadFIFOQueue:
		return "ReadFIFOQueue"
	case FuncCodeEncapsulatedInterface:
		return "EncapsulatedInterface"
	default:
		return fmt.Sprintf("Unknown(%02x)", uint8(fc))
	}
}

// String returns a string representation of the exception code
func (ec ExceptionCode) String() string {
	switch ec {
	case ExceptionCodeIllegalFunction:
		return "IllegalFunction"
	case ExceptionCodeIllegalDataAddress:
		return "IllegalDataAddress"
	case ExceptionCodeIllegalDataValue:
		return "IllegalDataValue"
	case ExceptionCodeServerDeviceFailure:
		return "ServerDeviceFailure"
	case ExceptionCodeAcknowledge:
		return "Acknowledge"
	case ExceptionCodeServerDeviceBusy:
		return "ServerDeviceBusy"
	case ExceptionCodeMemoryParityError:
		return "MemoryParityError"
	case ExceptionCodeGatewayPathUnavail:
		return "GatewayPathUnavailable"
	case ExceptionCodeGatewayTargetFail:
		return "GatewayTargetDeviceFailedToRespond"
	default:
		return fmt.Sprintf("Unknown(%02x)", uint8(ec))
	}
}

// Error implements the error interface for ExceptionCode
func (ec ExceptionCode) Error() string {
	return fmt.Sprintf("MODBUS Exception %02x: %s", uint8(ec), ec.String())
}

// ModbusError represents a MODBUS-specific error
type ModbusError struct {
	FunctionCode  FunctionCode
	ExceptionCode ExceptionCode
	Message       string
}

// Error implements the error interface
func (e *ModbusError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("MODBUS Error [%s]: %s - %s",
			e.FunctionCode.String(), e.ExceptionCode.String(), e.Message)
	}
	return fmt.Sprintf("MODBUS Error [%s]: %s",
		e.FunctionCode.String(), e.ExceptionCode.String())
}

// NewModbusError creates a new ModbusError
func NewModbusError(fc FunctionCode, ec ExceptionCode, message string) *ModbusError {
	return &ModbusError{
		FunctionCode:  fc,
		ExceptionCode: ec,
		Message:       message,
	}
}

// TransportType represents the type of MODBUS transport
type TransportType int

const (
	TransportTCP TransportType = iota
	TransportRTU
	TransportASCII
)

// String returns a string representation of the transport type
func (tt TransportType) String() string {
	switch tt {
	case TransportTCP:
		return "TCP"
	case TransportRTU:
		return "RTU"
	case TransportASCII:
		return "ASCII"
	default:
		return "Unknown"
	}
}

// ClientConfig holds configuration for a MODBUS client
type ClientConfig struct {
	SlaveID        SlaveID
	Timeout        time.Duration
	RetryCount     int
	RetryDelay     time.Duration
	ConnectTimeout time.Duration
	TransportType  TransportType
}

// DefaultClientConfig returns a default client configuration
func DefaultClientConfig() *ClientConfig {
	return &ClientConfig{
		SlaveID:        1,
		Timeout:        time.Duration(DefaultResponseTimeout) * time.Millisecond,
		RetryCount:     3,
		RetryDelay:     100 * time.Millisecond,
		ConnectTimeout: time.Duration(DefaultConnectTimeout) * time.Millisecond,
		TransportType:  TransportTCP,
	}
}

// ServerConfig holds configuration for a MODBUS server
type ServerConfig struct {
	SlaveID       SlaveID
	TransportType TransportType
}

// DefaultServerConfig returns a default server configuration
func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		SlaveID:       1,
		TransportType: TransportTCP,
	}
}

// JSONClientConfig represents client configuration in JSON format
type JSONClientConfig struct {
	SlaveID          int    `json:"slave_id"`
	TimeoutMs        int    `json:"timeout_ms"`
	RetryCount       int    `json:"retry_count"`
	RetryDelayMs     int    `json:"retry_delay_ms"`
	ConnectTimeoutMs int    `json:"connect_timeout_ms"`
	TransportType    string `json:"transport_type"`
}

// ToClientConfig converts JSONClientConfig to ClientConfig
func (jcc *JSONClientConfig) ToClientConfig() *ClientConfig {
	var transportType TransportType
	switch jcc.TransportType {
	case "tcp":
		transportType = TransportTCP
	case "rtu":
		transportType = TransportRTU
	case "ascii":
		transportType = TransportASCII
	default:
		transportType = TransportTCP
	}

	return &ClientConfig{
		SlaveID:        SlaveID(jcc.SlaveID),
		Timeout:        time.Duration(jcc.TimeoutMs) * time.Millisecond,
		RetryCount:     jcc.RetryCount,
		RetryDelay:     time.Duration(jcc.RetryDelayMs) * time.Millisecond,
		ConnectTimeout: time.Duration(jcc.ConnectTimeoutMs) * time.Millisecond,
		TransportType:  transportType,
	}
}

// LoadClientConfigFromJSON loads client configuration from a JSON file
func LoadClientConfigFromJSON(filepath string) (*ClientConfig, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var jsonConfig JSONClientConfig
	if err := json.Unmarshal(data, &jsonConfig); err != nil {
		return nil, fmt.Errorf("failed to parse JSON config: %w", err)
	}

	return jsonConfig.ToClientConfig(), nil
}

// LoadClientConfigFromJSONString loads client configuration from a JSON string
func LoadClientConfigFromJSONString(jsonStr string) (*ClientConfig, error) {
	var jsonConfig JSONClientConfig
	if err := json.Unmarshal([]byte(jsonStr), &jsonConfig); err != nil {
		return nil, fmt.Errorf("failed to parse JSON config: %w", err)
	}

	return jsonConfig.ToClientConfig(), nil
}

// ToJSONClientConfig converts ClientConfig to JSONClientConfig
func (cc *ClientConfig) ToJSONClientConfig() *JSONClientConfig {
	var transportType string
	switch cc.TransportType {
	case TransportTCP:
		transportType = "tcp"
	case TransportRTU:
		transportType = "rtu"
	case TransportASCII:
		transportType = "ascii"
	default:
		transportType = "tcp"
	}

	return &JSONClientConfig{
		SlaveID:          int(cc.SlaveID),
		TimeoutMs:        int(cc.Timeout / time.Millisecond),
		RetryCount:       cc.RetryCount,
		RetryDelayMs:     int(cc.RetryDelay / time.Millisecond),
		ConnectTimeoutMs: int(cc.ConnectTimeout / time.Millisecond),
		TransportType:    transportType,
	}
}

// SaveClientConfigToJSON saves client configuration to a JSON file
func (cc *ClientConfig) SaveClientConfigToJSON(filepath string) error {
	jsonConfig := cc.ToJSONClientConfig()
	data, err := json.MarshalIndent(jsonConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config to JSON: %w", err)
	}

	if err := os.WriteFile(filepath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// ToJSONString converts ClientConfig to a JSON string
func (cc *ClientConfig) ToJSONString() (string, error) {
	jsonConfig := cc.ToJSONClientConfig()
	data, err := json.MarshalIndent(jsonConfig, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal config to JSON: %w", err)
	}
	return string(data), nil
}

// DataStore interface defines the methods for accessing MODBUS data
type DataStore interface {
	// Coils (discrete outputs)
	ReadCoils(address Address, quantity Quantity) ([]bool, error)
	WriteCoils(address Address, values []bool) error

	// Discrete Inputs
	ReadDiscreteInputs(address Address, quantity Quantity) ([]bool, error)

	// Holding Registers
	ReadHoldingRegisters(address Address, quantity Quantity) ([]uint16, error)
	WriteHoldingRegisters(address Address, values []uint16) error

	// Input Registers
	ReadInputRegisters(address Address, quantity Quantity) ([]uint16, error)

	// File Records
	ReadFileRecords(records []FileRecord) ([]FileRecord, error)
	WriteFileRecords(records []FileRecord) error

	// FIFO Queue
	ReadFIFOQueue(address Address) ([]uint16, error)

	// Exception Status
	ReadExceptionStatus() (uint8, error)

	// Diagnostic Data
	GetDiagnosticData(subFunction uint16, data []byte) ([]byte, error)

	// Communication Event Counter
	GetCommEventCounter() (uint16, uint16, error) // status, eventCount

	// Communication Event Log
	GetCommEventLog() (uint16, uint16, uint16, []byte, error) // status, eventCount, messageCount, events
}

// DeviceIdentification holds device identification information
type DeviceIdentification struct {
	VendorName          string
	ProductCode         string
	MajorMinorRevision  string
	VendorURL           string
	ProductName         string
	ModelName           string
	UserApplicationName string
	ConformityLevel     uint8
}

// FileRecord represents a file record sub-request
type FileRecord struct {
	ReferenceType uint8
	FileNumber    uint16
	RecordNumber  uint16
	RecordLength  uint16
	RecordData    []uint16
}

// DiagnosticData holds diagnostic information
type DiagnosticData struct {
	BusMessageCount     uint16
	BusCommErrorCount   uint16
	BusExceptionCount   uint16
	ServerMessageCount  uint16
	ServerNoRespCount   uint16
	ServerNAKCount      uint16
	ServerBusyCount     uint16
	BusCharOverrunCount uint16
}
