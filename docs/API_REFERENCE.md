# ModbusGo API Reference

## Package Structure

```
github.com/adibhanna/modbus-go
├── modbus/       # Core types and constants
├── pdu/          # Protocol Data Unit handling
├── transport/    # Transport implementations
└── config/       # Configuration management
```

## Configuration API

### ClientConfig

```go
type ClientConfig struct {
    SlaveID        SlaveID
    Timeout        time.Duration
    RetryCount     int
    RetryDelay     time.Duration
    ConnectTimeout time.Duration
    TransportType  TransportType
}

// Create default configuration
config := modbus.DefaultClientConfig()

// Load from JSON file
config, err := modbus.LoadClientConfigFromJSON("config.json")

// Load from JSON string
config, err := modbus.LoadClientConfigFromJSONString(jsonString)

// Save to JSON file
err := config.SaveClientConfigToJSON("config.json")

// Convert to JSON string
jsonString, err := config.ToJSONString()
```

### JSONClientConfig

```go
type JSONClientConfig struct {
    SlaveID         int    `json:"slave_id"`
    TimeoutMs       int    `json:"timeout_ms"`
    RetryCount      int    `json:"retry_count"`
    RetryDelayMs    int    `json:"retry_delay_ms"`
    ConnectTimeoutMs int   `json:"connect_timeout_ms"`
    TransportType   string `json:"transport_type"`
}

// Convert between formats
clientConfig := jsonConfig.ToClientConfig()
jsonConfig := clientConfig.ToJSONClientConfig()
```

### Extended Configuration

```go
import "github.com/adibhanna/modbus-go/config"

// Load comprehensive configuration
cfg, err := config.LoadConfig("config.json")

// Access configuration sections
address := cfg.Connection.GetFullAddress()  // "192.168.1.102:502"
timeout := cfg.Connection.GetTimeout()      // time.Duration
slaveID := cfg.Modbus.GetSlaveID()         // modbus.SlaveID
```

## Client API

### Client Interface

```go
type Client interface {
    // Connection management
    Connect() error
    Close() error
    IsConnected() bool
    
    // Configuration management
    SetSlaveID(id SlaveID)
    GetSlaveID() SlaveID
    SetTimeout(timeout time.Duration)
    GetTimeout() time.Duration
    SetRetryCount(count int)
    GetRetryCount() int
    SetRetryDelay(delay time.Duration)
    GetRetryDelay() time.Duration
    SetConnectTimeout(timeout time.Duration)
    GetConnectTimeout() time.Duration
    GetConfig() *ClientConfig
    ApplyConfig(config *ClientConfig)
    
    // Bit access functions
    ReadCoils(address Address, quantity Quantity) ([]bool, error)
    ReadDiscreteInputs(address Address, quantity Quantity) ([]bool, error)
    WriteSingleCoil(address Address, value bool) error
    WriteMultipleCoils(address Address, values []bool) error
    
    // Register access functions
    ReadHoldingRegisters(address Address, quantity Quantity) ([]uint16, error)
    ReadInputRegisters(address Address, quantity Quantity) ([]uint16, error)
    WriteSingleRegister(address Address, value uint16) error
    WriteMultipleRegisters(address Address, values []uint16) error
    MaskWriteRegister(address Address, andMask, orMask uint16) error
    ReadWriteMultipleRegisters(readAddr Address, readQty Quantity, 
        writeAddr Address, values []uint16) ([]uint16, error)
    
    // Diagnostic functions
    ReadExceptionStatus() (uint8, error)
    Diagnostic(subFunction uint16, data []byte) ([]byte, error)
    GetCommEventCounter() (status uint16, count uint16, error)
    GetCommEventLog() (status uint16, eventCount uint16, 
        messageCount uint16, events []byte, error)
    ReportServerID() (serverID []byte, runStatus byte, 
        additionalData []byte, error)
    
    // File record access
    ReadFileRecords(records []FileRecord) ([]FileRecord, error)
    WriteFileRecords(records []FileRecord) error
    
    // FIFO queue access
    ReadFIFOQueue(address Address) ([]uint16, error)
    
    // Device identification
    ReadDeviceIdentification(readCode uint8, objectID uint8) (
        objects map[uint8][]byte, error)
}
```

### TCP Client

```go
// Basic constructor
func NewTCPClient(address string) *Client

// Constructor with configuration
func NewTCPClientFromConfig(config *ClientConfig, address string) *Client
func NewClientFromConfig(config *ClientConfig, transport Transport) *Client

// Constructor from JSON
func NewTCPClientFromJSONFile(configPath, address string) (*Client, error)
func NewTCPClientFromJSONString(jsonConfig, address string) (*Client, error)

// Examples

// 1. Basic client
client := modbus.NewTCPClient("192.168.1.102:502")
client.SetSlaveID(1)
client.SetTimeout(10 * time.Second)

// 2. Client from configuration struct
config := modbus.DefaultClientConfig()
config.SlaveID = 2
config.RetryCount = 5
client := modbus.NewTCPClientFromConfig(config, "192.168.1.102:502")

// 3. Client from JSON file
client, err := modbus.NewTCPClientFromJSONFile("config.json", "192.168.1.102:502")
if err != nil {
    return err
}

// 4. Client from JSON string
jsonConfig := `{
    "slave_id": 1,
    "timeout_ms": 10000,
    "retry_count": 3,
    "retry_delay_ms": 100,
    "connect_timeout_ms": 5000,
    "transport_type": "tcp"
}`
client, err := modbus.NewTCPClientFromJSONString(jsonConfig, "192.168.1.102:502")
if err != nil {
    return err
}

defer client.Close()

// Read operations
values, err := client.ReadHoldingRegisters(100, 10)
coils, err := client.ReadCoils(0, 16)

// Write operations
err = client.WriteSingleRegister(100, 1234)
err = client.WriteMultipleRegisters(200, []uint16{1, 2, 3, 4})

// Configuration management
currentConfig := client.GetConfig()
client.SetRetryDelay(500 * time.Millisecond)
err = currentConfig.SaveClientConfigToJSON("saved-config.json")
```

### RTU Client

```go
// Constructor
func NewRTUClient(device string, slaveID SlaveID, baudRate int, 
    dataBits int, stopBits int, parity Parity) (*RTUClient, error)

// Example
client, err := modbus.NewRTUClient("/dev/ttyUSB0", 1, 9600, 8, 1, modbus.ParityNone)
if err != nil {
    return err
}
defer client.Close()
```

### ASCII Client

```go
// Constructor  
func NewASCIIClient(device string, slaveID SlaveID, baudRate int,
    dataBits int, stopBits int, parity Parity) (*ASCIIClient, error)

// Example
client, err := modbus.NewASCIIClient("/dev/ttyS0", 1, 9600, 7, 1, modbus.ParityEven)
```

## Server API

### DataStore Interface

```go
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
    GetCommEventCounter() (status uint16, eventCount uint16, error)
    
    // Communication Event Log
    GetCommEventLog() (status uint16, eventCount uint16, 
        messageCount uint16, events []byte, error)
}
```

### DefaultDataStore

```go
// Constructor
func NewDefaultDataStore(coilCount, discreteInputCount, 
    holdingRegCount, inputRegCount int) *DefaultDataStore

// Helper methods
func (ds *DefaultDataStore) SetCoil(address Address, value bool) error
func (ds *DefaultDataStore) SetDiscreteInput(address Address, value bool) error
func (ds *DefaultDataStore) SetHoldingRegister(address Address, value uint16) error
func (ds *DefaultDataStore) SetInputRegister(address Address, value uint16) error
func (ds *DefaultDataStore) SetExceptionStatus(status uint8)
func (ds *DefaultDataStore) WriteFIFOQueue(address Address, values []uint16) error
func (ds *DefaultDataStore) IncrementDiagnosticCounter(counter string)

// Example
dataStore := modbus.NewDefaultDataStore(10000, 10000, 10000, 10000)
dataStore.SetHoldingRegister(100, 1234)
dataStore.SetCoil(0, true)
```

### TCP Server

```go
// Constructor
func NewTCPServer(address string, dataStore DataStore) (*TCPServer, error)

// Methods
func (s *TCPServer) Start() error
func (s *TCPServer) Stop() error

// Example
server, err := modbus.NewTCPServer(":502", dataStore)
if err != nil {
    return err
}

// Start in blocking mode
err = server.Start()

// Or non-blocking
go server.Start()
```

### RTU Server

```go
// Constructor
func NewRTUServer(device string, slaveID SlaveID, baudRate int,
    dataBits int, stopBits int, parity Parity, 
    dataStore DataStore) (*RTUServer, error)

// Example
server, err := modbus.NewRTUServer("/dev/ttyUSB0", 1, 9600, 8, 1, 
    modbus.ParityNone, dataStore)
```

## Types

### Basic Types

```go
// SlaveID represents a MODBUS slave/unit identifier (1-247)
type SlaveID uint8

// Address represents a MODBUS register or coil address (0-65535)
type Address uint16

// Quantity represents a quantity of registers or coils (1-65535)
type Quantity uint16

// FunctionCode represents a MODBUS function code
type FunctionCode uint8

// ExceptionCode represents a MODBUS exception code
type ExceptionCode uint8
```

### Configuration Types

```go
// ClientConfig holds client configuration
type ClientConfig struct {
    SlaveID       SlaveID
    Timeout       time.Duration
    RetryCount    int
    TransportType TransportType
}

// ServerConfig holds server configuration
type ServerConfig struct {
    SlaveID       SlaveID
    TransportType TransportType
}

// SerialConfig holds serial port configuration
type SerialConfig struct {
    Device   string
    BaudRate int
    DataBits int
    StopBits int
    Parity   Parity
    Timeout  time.Duration
}
```

### Data Types

```go
// FileRecord represents a file record for extended memory access
type FileRecord struct {
    ReferenceType uint8    // Must be 0x06 (FileRecordTypeExtended)
    FileNumber    uint16   // File number (0-65535)
    RecordNumber  uint16   // Starting record number (0-9999)
    RecordLength  uint16   // Number of registers
    RecordData    []uint16 // Data (for write operations)
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

// DiagnosticData holds diagnostic counters
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
```

### Error Types

```go
// ModbusError represents a MODBUS protocol error
type ModbusError struct {
    FunctionCode  FunctionCode
    ExceptionCode ExceptionCode
    Message       string
}

func (e *ModbusError) Error() string

// Check for MODBUS errors
if modbusErr, ok := err.(*ModbusError); ok {
    switch modbusErr.ExceptionCode {
    case modbus.ExceptionCodeIllegalDataAddress:
        // Handle invalid address
    case modbus.ExceptionCodeServerDeviceBusy:
        // Retry later
    }
}
```

## Constants

### Function Codes

```go
const (
    // Bit Access
    FuncCodeReadCoils          = 0x01
    FuncCodeReadDiscreteInputs = 0x02
    FuncCodeWriteSingleCoil    = 0x05
    FuncCodeWriteMultipleCoils = 0x0F
    
    // Register Access
    FuncCodeReadHoldingRegisters   = 0x03
    FuncCodeReadInputRegisters     = 0x04
    FuncCodeWriteSingleRegister    = 0x06
    FuncCodeWriteMultipleRegisters = 0x10
    FuncCodeMaskWriteRegister      = 0x16
    FuncCodeReadWriteMultipleRegs  = 0x17
    
    // File Record Access
    FuncCodeReadFileRecord  = 0x14
    FuncCodeWriteFileRecord = 0x15
    
    // Diagnostics
    FuncCodeReadExceptionStatus = 0x07
    FuncCodeDiagnostic          = 0x08
    FuncCodeGetCommEventCounter = 0x0B
    FuncCodeGetCommEventLog     = 0x0C
    FuncCodeReportServerID      = 0x11
    
    // FIFO Queue
    FuncCodeReadFIFOQueue = 0x18
    
    // Encapsulated Interface
    FuncCodeEncapsulatedInterface = 0x2B
)
```

### Exception Codes

```go
const (
    ExceptionCodeIllegalFunction     = 0x01
    ExceptionCodeIllegalDataAddress  = 0x02
    ExceptionCodeIllegalDataValue    = 0x03
    ExceptionCodeServerDeviceFailure = 0x04
    ExceptionCodeAcknowledge         = 0x05
    ExceptionCodeServerDeviceBusy    = 0x06
    ExceptionCodeMemoryParityError   = 0x08
    ExceptionCodeGatewayPathUnavail  = 0x0A
    ExceptionCodeGatewayTargetFail   = 0x0B
)
```

### Diagnostic Sub-Functions

```go
const (
    DiagSubReturnQueryData           = 0x0000
    DiagSubRestartCommOption         = 0x0001
    DiagSubReturnDiagRegister        = 0x0002
    DiagSubChangeASCIIDelimiter      = 0x0003
    DiagSubForceListenOnlyMode       = 0x0004
    DiagSubClearCounters             = 0x000A
    DiagSubReturnBusMessageCount     = 0x000B
    DiagSubReturnBusCommErrorCount   = 0x000C
    DiagSubReturnBusExceptionCount   = 0x000D
    DiagSubReturnServerMessageCount  = 0x000E
    DiagSubReturnServerNoRespCount   = 0x000F
    DiagSubReturnServerNAKCount      = 0x0010
    DiagSubReturnServerBusyCount     = 0x0011
    DiagSubReturnBusCharOverrunCount = 0x0012
    DiagSubClearOverrunCounter       = 0x0014
)
```

### Limits

```go
const (
    MaxPDUSize       = 253  // Maximum PDU size in bytes
    MaxTCPADUSize    = 260  // Maximum TCP ADU size
    MaxSerialADUSize = 256  // Maximum Serial ADU size
    
    // Quantity limits
    MaxReadCoils            = 2000
    MaxReadDiscreteInputs   = 2000
    MaxReadHoldingRegs      = 125
    MaxReadInputRegs        = 125
    MaxWriteMultipleCoils   = 1968
    MaxWriteMultipleRegs    = 123
    MaxReadWriteRegs        = 125
    MaxWriteReadWriteRegs   = 121
    MaxReadFileRecordBytes  = 245
    MaxWriteFileRecordBytes = 251
    MaxFIFOCount            = 31
)
```

### Special Values

```go
const (
    // Coil values
    CoilOn  = 0xFF00
    CoilOff = 0x0000
    
    // File record type
    FileRecordTypeExtended = 0x06
    
    // MEI Types
    MEITypeCANopenGeneralReference = 0x0D
    MEITypeDeviceIdentification    = 0x0E
    
    // Device ID codes
    DeviceIDReadBasic    = 0x01
    DeviceIDReadRegular  = 0x02
    DeviceIDReadExtended = 0x03
    DeviceIDReadSpecific = 0x04
)
```

## PDU Package

### Request Type

```go
type Request struct {
    FunctionCode FunctionCode
    Data         []byte
}

// Constructor
func NewRequest(functionCode FunctionCode, data []byte) *Request

// Request builders
func ReadCoilsRequest(address Address, quantity Quantity) (*Request, error)
func ReadHoldingRegistersRequest(address Address, quantity Quantity) (*Request, error)
func WriteSingleCoilRequest(address Address, value bool) (*Request, error)
func WriteSingleRegisterRequest(address Address, value uint16) (*Request, error)
func WriteMultipleCoilsRequest(address Address, values []bool) (*Request, error)
func WriteMultipleRegistersRequest(address Address, values []uint16) (*Request, error)
// ... and more
```

### Response Type

```go
type Response struct {
    FunctionCode FunctionCode
    Data         []byte
}

// Constructor
func NewResponse(functionCode FunctionCode, data []byte) *Response
func NewExceptionResponse(functionCode FunctionCode, exception ExceptionCode) *Response

// Methods
func (r *Response) IsException() bool
func (r *Response) GetExceptionCode() (ExceptionCode, error)
```

### Encoding/Decoding Functions

```go
// Encoding
func EncodeUint16(value uint16) []byte
func EncodeUint16Slice(values []uint16) []byte
func EncodeBoolSlice(values []bool) []byte

// Decoding
func DecodeUint16(data []byte) (uint16, error)
func DecodeUint16Slice(data []byte) ([]uint16, error)
func DecodeBoolSlice(data []byte, count int) ([]bool, error)

// Validation
func ValidateAddress(address Address, quantity Quantity) error
func ValidateQuantity(functionCode FunctionCode, quantity Quantity) error
```

## Transport Package

### Transport Interface

```go
type Transport interface {
    Connect() error
    Close() error
    SendRequest(slaveID SlaveID, request *pdu.Request) (*pdu.Response, error)
    IsConnected() bool
    SetTimeout(timeout time.Duration)
}
```

### RequestHandler Interface

```go
type RequestHandler interface {
    HandleRequest(slaveID SlaveID, request *pdu.Request) *pdu.Response
}
```

### TCP Transport

```go
type TCPTransport struct {
    // Internal fields
}

// Constructor
func NewTCPTransport(address string) *TCPTransport

// Methods implement Transport interface
```

### Serial Transport

```go
type SerialTransport struct {
    // Internal fields
}

// Constructor
func NewRTUTransport(config SerialConfig) *SerialTransport
func NewASCIITransport(config SerialConfig) *SerialTransport

// Methods implement Transport interface
```

## Usage Examples

### Basic Client Operations

```go
// Connect and read registers
client := modbus.NewTCPClient("192.168.1.100:502")
client.SetSlaveID(1)
defer client.Close()

// Read 10 holding registers starting at address 100
values, err := client.ReadHoldingRegisters(100, 10)
if err != nil {
    log.Fatal(err)
}

// Write single register
err = client.WriteSingleRegister(100, 1234)

// Write multiple registers
err = client.WriteMultipleRegisters(200, []uint16{1, 2, 3, 4, 5})

// Read and write in one operation
writeData := []uint16{10, 20, 30}
readData, err := client.ReadWriteMultipleRegisters(100, 5, 200, writeData)
```

### Basic Server Setup

```go
// Create data store
dataStore := modbus.NewDefaultDataStore(10000, 10000, 10000, 10000)

// Initialize some data
for i := 0; i < 100; i++ {
    dataStore.SetHoldingRegister(modbus.Address(i), uint16(i*10))
}

// Create and start server
server, _ := modbus.NewTCPServer(":502", dataStore)
log.Fatal(server.Start())
```

### Advanced Operations

```go
// File record operations
records := []modbus.FileRecord{
    {
        ReferenceType: modbus.FileRecordTypeExtended,
        FileNumber:    4,
        RecordNumber:  1,
        RecordLength:  3,
    },
}
result, err := client.ReadFileRecords(records)

// FIFO queue operations
fifoData, err := client.ReadFIFOQueue(1000)

// Diagnostic operations
echoData := []byte{0xAA, 0x55}
response, err := client.Diagnostic(modbus.DiagSubReturnQueryData, echoData)

// Device identification
objects, err := client.ReadDeviceIdentification(
    modbus.DeviceIDReadBasic, 0x00)
```

## Thread Safety

All client and server operations are thread-safe. The DefaultDataStore uses read-write mutexes for concurrent access protection.

```go
// Safe for concurrent use
go func() {
    for {
        client.ReadHoldingRegisters(100, 10)
        time.Sleep(100 * time.Millisecond)
    }
}()

go func() {
    for {
        client.WriteSingleRegister(200, rand.Uint16())
        time.Sleep(200 * time.Millisecond)
    }
}()
```

## Version

Current version: 1.3.0

Supports MODBUS Application Protocol Specification V1.1b3