# ModbusGo - Complete Documentation

## Table of Contents
1. [Introduction](#introduction)
2. [Architecture Overview](#architecture-overview)
3. [Quick Start](#quick-start)
4. [Configuration](#configuration)
5. [New Features](#new-features)
6. [Core Concepts](#core-concepts)
7. [Function Codes Reference](#function-codes-reference)
8. [Client Usage](#client-usage)
9. [Server Usage](#server-usage)
10. [Advanced Features](#advanced-features)
11. [Transport Layers](#transport-layers)
12. [Error Handling](#error-handling)
13. [Testing](#testing)
14. [Performance Considerations](#performance-considerations)
15. [Troubleshooting](#troubleshooting)

## Introduction

ModbusGo is a comprehensive, production-ready MODBUS protocol implementation in Go that supports both client and server operations. It implements the complete MODBUS Application Protocol Specification V1.1b3, including all standard function codes and advanced features.

### Key Features
- **Complete Protocol Support**: All 19 standard MODBUS function codes
- **Multiple Transports**: TCP/IP, RTU (Serial), and ASCII
- **Concurrent Safe**: Thread-safe operations with proper synchronization
- **Extensible Architecture**: Clean interfaces for custom implementations
- **Auto-Reconnect**: Automatic connection recovery on failure
- **Broadcast Support**: Send commands to all devices (slave ID 0)
- **Graceful Shutdown**: Server shutdown with timeout and proper cleanup
- **Production Ready**: Comprehensive error handling and recovery
- **Well Tested**: Extensive test coverage for all components

### Installation

```bash
go get github.com/adibhanna/modbus-go
```

## Architecture Overview

The library is organized into several key packages:

```
modbusgo/
├── modbus/          # Core types, interfaces, and constants
│   ├── constants.go # MODBUS protocol constants
│   └── types.go     # Core type definitions
├── pdu/             # Protocol Data Unit handling
│   ├── pdu.go       # PDU structure and methods
│   ├── requests.go  # Request builders
│   └── responses.go # Response parsers
├── transport/       # Transport layer implementations
│   ├── tcp.go       # TCP/IP transport
│   ├── serial.go    # RTU/ASCII serial transport
│   └── interface.go # Transport interfaces
├── client.go        # MODBUS client implementation
├── server.go        # MODBUS server implementation
└── types.go         # Package-level type exports
```

### Design Principles

1. **Separation of Concerns**: Clear separation between protocol logic, transport, and business logic
2. **Interface-Based**: Core components use interfaces for flexibility
3. **Zero Dependencies**: Minimal external dependencies (only standard library where possible)
4. **Type Safety**: Strong typing throughout the codebase
5. **Error Propagation**: Explicit error handling at all levels

## Quick Start

### Simple TCP Client

```go
package main

import (
    "fmt"
    "log"
    modbus "github.com/adibhanna/modbus-go"
)

func main() {
    // Connect to MODBUS TCP server
    client, err := modbus.NewTCPClient("localhost:502", 1)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Read holding registers
    values, err := client.ReadHoldingRegisters(0, 10)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Register values: %v\n", values)
}
```

### Simple TCP Server

```go
package main

import (
    "log"
    modbus "github.com/adibhanna/modbus-go"
)

func main() {
    // Create data store
    dataStore := modbus.NewDefaultDataStore(10000, 10000, 10000, 10000)
    
    // Initialize some data
    dataStore.SetHoldingRegister(0, 100)
    dataStore.SetCoil(0, true)
    
    // Start server
    server, err := modbus.NewTCPServer(":502", dataStore)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Println("Server starting on :502")
    if err := server.Start(); err != nil {
        log.Fatal(err)
    }
}
```

## Configuration

ModbusGo provides flexible configuration options for clients, supporting both programmatic configuration and JSON-based configuration files. This allows you to adapt the library to different devices, environments, and operational requirements.

### Configuration Overview

The library supports multiple levels of configuration:

1. **Core Library Configuration** - Simple, lightweight JSON configuration built into the core library
2. **Extended Configuration System** - Comprehensive configuration with device profiles, testing parameters, and advanced options
3. **Runtime Configuration** - Dynamic configuration changes during execution
4. **Configuration Persistence** - Save and load configuration settings

### Core Library Configuration

#### JSON Configuration Structure

The core library supports a simple JSON configuration format:

```json
{
  "slave_id": 1,
  "timeout_ms": 10000,
  "retry_count": 3,
  "retry_delay_ms": 100,
  "connect_timeout_ms": 5000,
  "transport_type": "tcp"
}
```

#### Configuration Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `slave_id` | int | 1 | MODBUS slave/unit identifier |
| `timeout_ms` | int | 1000 | Response timeout in milliseconds |
| `retry_count` | int | 3 | Number of retry attempts on failure |
| `retry_delay_ms` | int | 100 | Delay between retry attempts in milliseconds |
| `connect_timeout_ms` | int | 5000 | Connection timeout in milliseconds |
| `transport_type` | string | "tcp" | Transport protocol ("tcp", "rtu", "ascii") |

#### Client Creation Methods

##### 1. From JSON File

```go
// Load configuration from file
client, err := modbus.NewTCPClientFromJSONFile("config.json", "192.168.1.102:502")
if err != nil {
    log.Fatal(err)
}
```

##### 2. From JSON String

```go
// Load configuration from JSON string
jsonConfig := `{
    "slave_id": 2,
    "timeout_ms": 15000,
    "retry_count": 5,
    "retry_delay_ms": 250,
    "connect_timeout_ms": 8000,
    "transport_type": "tcp"
}`

client, err := modbus.NewTCPClientFromJSONString(jsonConfig, "192.168.1.102:502")
if err != nil {
    log.Fatal(err)
}
```

##### 3. From ClientConfig Struct

```go
// Create configuration struct
config := modbus.DefaultClientConfig()
config.SlaveID = 3
config.Timeout = 20 * time.Second
config.RetryCount = 1
config.RetryDelay = 500 * time.Millisecond

// Create client from config
client := modbus.NewTCPClientFromConfig(config, "192.168.1.102:502")
```

#### Runtime Configuration Management

##### Reading Current Configuration

```go
client := modbus.NewTCPClient("192.168.1.102:502")

// Get current configuration
config := client.GetConfig()
fmt.Printf("Current SlaveID: %d\n", config.SlaveID)
fmt.Printf("Current Timeout: %v\n", config.Timeout)
```

##### Modifying Configuration

```go
// Individual parameter changes
client.SetSlaveID(5)
client.SetTimeout(15 * time.Second)
client.SetRetryCount(2)
client.SetRetryDelay(200 * time.Millisecond)
client.SetConnectTimeout(10 * time.Second)

// Bulk configuration change
newConfig := modbus.DefaultClientConfig()
newConfig.SlaveID = 10
newConfig.RetryCount = 5
client.ApplyConfig(newConfig)
```

##### Configuration Persistence

```go
// Save configuration to JSON file
config := client.GetConfig()
err := config.SaveClientConfigToJSON("saved-config.json")
if err != nil {
    log.Printf("Failed to save config: %v", err)
}

// Convert configuration to JSON string
jsonString, err := config.ToJSONString()
if err != nil {
    log.Printf("Failed to convert config: %v", err)
} else {
    fmt.Printf("Configuration:\n%s\n", jsonString)
}
```

### Extended Configuration System

For comprehensive testing, device profiles, and advanced configuration management, use the extended configuration system:

```go
import "github.com/adibhanna/modbus-go/config"

// Load extended configuration
cfg, err := config.LoadConfig("extended-config.json")
if err != nil {
    log.Fatal(err)
}

// Create client using extended configuration
client := modbus.NewTCPClient(cfg.Connection.GetFullAddress())
client.SetSlaveID(cfg.Modbus.GetSlaveID())
client.SetTimeout(cfg.Connection.GetTimeout())
client.SetRetryCount(cfg.Connection.RetryCount)
```

#### Extended Configuration Structure

```json
{
  "connection": {
    "address": "192.168.1.102",
    "port": 502,
    "timeout_ms": 10000,
    "connect_timeout_ms": 5000,
    "retry_count": 3,
    "transport_type": "tcp"
  },
  "modbus": {
    "slave_id": 1,
    "unit_id": 1,
    "protocol_id": 0
  },
  "testing": {
    "enabled_tests": [
      "read_holding_registers",
      "read_coils",
      "write_single_register"
    ],
    "test_addresses": {
      "holding_registers": {
        "start_address": 0,
        "quantity": 5
      }
    }
  },
  "device_profiles": {
    "schneider_electric": {
      "slave_id": 1,
      "holding_registers_start": 1,
      "supported_functions": [1, 2, 3, 4, 5, 6, 15, 16]
    }
  },
  "current_profile": "schneider_electric"
}
```

### Device-Specific Configuration Examples

#### Schneider Electric Configuration

```json
{
  "slave_id": 1,
  "timeout_ms": 10000,
  "retry_count": 3,
  "retry_delay_ms": 100,
  "connect_timeout_ms": 5000,
  "transport_type": "tcp"
}
```

#### Siemens Configuration

```json
{
  "slave_id": 1,
  "timeout_ms": 8000,
  "retry_count": 2,
  "retry_delay_ms": 150,
  "connect_timeout_ms": 4000,
  "transport_type": "tcp"
}
```

#### Diagnostic Configuration

```json
{
  "slave_id": 1,
  "timeout_ms": 5000,
  "retry_count": 1,
  "retry_delay_ms": 200,
  "connect_timeout_ms": 3000,
  "transport_type": "tcp"
}
```

### Configuration Best Practices

#### 1. Environment-Specific Configurations

```go
// Development environment
devConfig := `{
    "slave_id": 1,
    "timeout_ms": 30000,
    "retry_count": 5,
    "retry_delay_ms": 1000
}`

// Production environment
prodConfig := `{
    "slave_id": 1,
    "timeout_ms": 5000,
    "retry_count": 2,
    "retry_delay_ms": 100
}`

var configStr string
if os.Getenv("ENV") == "production" {
    configStr = prodConfig
} else {
    configStr = devConfig
}

client, err := modbus.NewTCPClientFromJSONString(configStr, address)
```

#### 2. Configuration Validation

```go
client, err := modbus.NewTCPClientFromJSONFile("config.json", address)
if err != nil {
    // Fallback to default configuration
    log.Printf("Failed to load config, using defaults: %v", err)
    client = modbus.NewTCPClient(address)
}

// Validate configuration
config := client.GetConfig()
if config.Timeout < 1*time.Second {
    log.Printf("Warning: Timeout very low (%v), consider increasing", config.Timeout)
}
if config.RetryCount > 10 {
    log.Printf("Warning: High retry count (%d) may cause delays", config.RetryCount)
}
```

#### 3. Configuration Templates

```go
// Base configuration template
func NewIndustrialConfig() *modbus.ClientConfig {
    config := modbus.DefaultClientConfig()
    config.Timeout = 10 * time.Second
    config.RetryCount = 3
    config.RetryDelay = 500 * time.Millisecond
    config.ConnectTimeout = 15 * time.Second
    return config
}

// Device-specific adjustments
func NewSchneiderzConfig() *modbus.ClientConfig {
    config := NewIndustrialConfig()
    config.SlaveID = 1
    config.RetryDelay = 100 * time.Millisecond
    return config
}
```

### Configuration Migration

If you're upgrading from a previous version that didn't support configuration:

#### Before (Manual Configuration)

```go
client := modbus.NewTCPClient("192.168.1.102:502")
client.SetSlaveID(1)
client.SetTimeout(10 * time.Second)
client.SetRetryCount(3)
```

#### After (JSON Configuration)

```go
// Save existing configuration as JSON template
config := &modbus.ClientConfig{
    SlaveID:        1,
    Timeout:        10 * time.Second,
    RetryCount:     3,
    RetryDelay:     100 * time.Millisecond,
    ConnectTimeout: 5 * time.Second,
    TransportType:  modbus.TransportTCP,
}

// Save as template for future use
err := config.SaveClientConfigToJSON("my-device-config.json")

// Load from file in future
client, err := modbus.NewTCPClientFromJSONFile("my-device-config.json", "192.168.1.102:502")
```

## New Features

### Auto-Reconnect

The client supports automatic reconnection when the connection is lost:

```go
client, _ := modbus.NewTCPClient("192.168.1.100:502", 1)
defer client.Close()

// Enable auto-reconnect
client.SetAutoReconnect(true)

// Now if the connection is lost during operations,
// the client will automatically attempt to reconnect
values, err := client.ReadHoldingRegisters(0, 10)
if err != nil {
    // Error only occurs if reconnect also fails
    log.Printf("Failed after reconnect attempt: %v", err)
}

// Check auto-reconnect status
if client.GetAutoReconnect() {
    log.Println("Auto-reconnect is enabled")
}
```

### Broadcast Support

MODBUS supports broadcast messages (slave ID 0) for write operations where no response is expected:

```go
client, _ := modbus.NewTCPClient("192.168.1.100:502", 1)
defer client.Close()

// Broadcast write single coil to all devices
err := client.BroadcastWriteSingleCoil(0, true)

// Broadcast write single register to all devices
err = client.BroadcastWriteSingleRegister(100, 1234)

// Broadcast write multiple coils to all devices
err = client.BroadcastWriteMultipleCoils(0, []bool{true, false, true, true})

// Broadcast write multiple registers to all devices
err = client.BroadcastWriteMultipleRegisters(100, []uint16{1, 2, 3, 4})
```

**Note**: Broadcast messages do not receive responses. They are useful for:
- Synchronizing multiple devices simultaneously
- Emergency stops across all devices
- System-wide configuration changes
- Resetting all devices to a known state

### Graceful Server Shutdown

The TCP server supports graceful shutdown with configurable timeout:

```go
// Create and start server
server, _ := modbus.NewTCPServer(":502", dataStore)
go func() {
    if err := server.Start(); err != nil {
        log.Printf("Server error: %v", err)
    }
}()

// ... server runs ...

// Option 1: Simple stop (waits for all connections to close)
err := server.Stop()

// Option 2: Stop with timeout (returns error if timeout exceeded)
err = server.StopWithTimeout(5 * time.Second)
if err != nil {
    log.Printf("Server shutdown timed out: %v", err)
    // Connections were forcibly closed after timeout
}
```

### Additional Diagnostic Client Methods

Complete support for all MODBUS diagnostic functions:

```go
// Get communication event counter (FC 0x0B)
status, eventCount, err := client.GetCommEventCounter()
// status: 0xFFFF = ready, 0x0000 = not ready
// eventCount: number of successful message completions
fmt.Printf("Status: 0x%04X, Events: %d\n", status, eventCount)

// Get communication event log (FC 0x0C)
status, eventCount, messageCount, events, err := client.GetCommEventLog()
// events: byte array of event codes
fmt.Printf("Messages: %d, Events: %d\n", messageCount, eventCount)
for i, ev := range events {
    fmt.Printf("Event %d: 0x%02X\n", i, ev)
}

// Report server ID (FC 0x11)
serverData, err := client.ReportServerID()
// serverData includes run indicator and server-specific identification
fmt.Printf("Server ID data: %v\n", serverData)
if len(serverData) > 0 {
    runIndicator := serverData[0] // 0xFF = ON, 0x00 = OFF
    fmt.Printf("Run indicator: 0x%02X\n", runIndicator)
}
```

### Thread-Safe Serial Transports

RTU and ASCII transports now include mutex protection for thread safety:

```go
// RTU transport is safe for concurrent use
rtuTransport := transport.NewRTUTransport(config)

// Multiple goroutines can safely use the same transport
var wg sync.WaitGroup
for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        // Safe concurrent access
        resp, err := rtuTransport.SendRequest(1, request)
        // ...
    }(i)
}
wg.Wait()
```

### High-Level Data Type Helpers

The library provides convenient methods for reading and writing multi-register data types:

```go
client, _ := modbus.NewTCPClient("192.168.1.100:502", 1)
defer client.Close()

// Read/Write 32-bit integers (uses 2 registers)
val32, err := client.ReadUint32(100)
err = client.WriteUint32(100, 12345678)

// Read/Write 64-bit integers (uses 4 registers)
val64, err := client.ReadUint64(200)
err = client.WriteUint64(200, 123456789012345)

// Read/Write 32-bit floats
floatVal, err := client.ReadFloat32(300)
err = client.WriteFloat32(300, 3.14159)

// Read/Write 64-bit floats
float64Val, err := client.ReadFloat64(400)
err = client.WriteFloat64(400, 3.141592653589793)

// Read/Write signed integers
int32Val, err := client.ReadInt32(500)
err = client.WriteInt32(500, -12345)

// Read/Write multiple values at once
uint32s, err := client.ReadUint32s(600, 5) // Read 5 uint32 values
err = client.WriteFloat32s(700, []float32{1.1, 2.2, 3.3})

// Read/Write raw bytes
bytes, err := client.ReadBytes(800, 20) // Read 20 bytes
err = client.WriteBytes(800, []byte{0x01, 0x02, 0x03, 0x04})

// Read/Write strings
str, err := client.ReadString(900, 32) // Read up to 32 chars
err = client.WriteString(900, "Hello MODBUS", 32)

// Single register/coil read helpers
coilVal, err := client.ReadCoil(0)           // Read single coil
discreteVal, err := client.ReadDiscreteInput(0) // Read single discrete input
regVal, err := client.ReadHoldingRegister(0)    // Read single holding register
inputVal, err := client.ReadInputRegister(0)    // Read single input register

// Read from input registers
inputFloat, err := client.ReadInputFloat32(1000)
inputUint32, err := client.ReadInputUint32(1000)
```

### Endianness and Word Order Configuration

Different devices use different byte/word ordering. Configure the encoding to match your device:

```go
client, _ := modbus.NewTCPClient("192.168.1.100:502", 1)

// Default is Big Endian, High Word First (most common in MODBUS)
// This is the standard MODBUS byte order

// For devices using different byte ordering:
client.SetEncoding(modbus.LittleEndian, modbus.HighWordFirst)

// Or for different word ordering:
client.SetEncoding(modbus.BigEndian, modbus.LowWordFirst)

// Check current encoding
enc := client.GetEncoding()
fmt.Printf("Byte Order: %v, Word Order: %v\n", enc.ByteOrder, enc.WordOrder)

// Encoding affects all multi-byte operations:
// - ReadUint32/WriteUint32
// - ReadFloat32/WriteFloat32
// - ReadUint64/WriteUint64
// - ReadFloat64/WriteFloat64
// - ReadBytes/WriteBytes
// - ReadString/WriteString
```

**Common device encoding configurations:**

| Device/Manufacturer | Byte Order | Word Order |
|---------------------|------------|------------|
| Most MODBUS devices | BigEndian | HighWordFirst |
| Some Siemens PLCs | BigEndian | LowWordFirst |
| Some ABB devices | LittleEndian | HighWordFirst |

### TLS Support for TCP Transport

Secure your MODBUS TCP communications with TLS encryption:

```go
import (
    "crypto/tls"
    "crypto/x509"
    "io/ioutil"
    "github.com/adibhanna/modbus-go/transport"
)

// Basic TLS with server certificate verification
tlsConfig := &tls.Config{
    MinVersion: tls.VersionTLS12,
}

tcpTransport := transport.NewTLSTransport("192.168.1.100:802", tlsConfig)
if err := tcpTransport.Connect(); err != nil {
    log.Fatal(err)
}
defer tcpTransport.Close()

// With client certificate authentication (mutual TLS)
cert, err := tls.LoadX509KeyPair("client-cert.pem", "client-key.pem")
if err != nil {
    log.Fatal(err)
}

caCert, err := ioutil.ReadFile("ca-cert.pem")
if err != nil {
    log.Fatal(err)
}
caCertPool := x509.NewCertPool()
caCertPool.AppendCertsFromPEM(caCert)

tlsConfig := &tls.Config{
    Certificates: []tls.Certificate{cert},
    RootCAs:      caCertPool,
    MinVersion:   tls.VersionTLS12,
}

secureTransport := transport.NewTLSTransport("192.168.1.100:802", tlsConfig)

// Skip certificate verification (NOT recommended for production)
insecureConfig := &tls.Config{
    InsecureSkipVerify: true,
}
```

### RTU over TCP Transport

For serial-to-Ethernet converters that use RTU framing over TCP (not standard MODBUS TCP):

```go
import "github.com/adibhanna/modbus-go/transport"

// RTU over TCP uses RTU framing (slave ID + PDU + CRC) over TCP connection
// This is common with serial-to-Ethernet gateways that don't translate to MODBUS TCP

rtuOverTCP := transport.NewRTUOverTCPTransport("192.168.1.100:4001")

if err := rtuOverTCP.Connect(); err != nil {
    log.Fatal(err)
}
defer rtuOverTCP.Close()

// Create client with RTU over TCP transport
client := modbus.NewClient(rtuOverTCP)
client.SetSlaveID(1)

// Use normally - the transport handles RTU framing over TCP
values, err := client.ReadHoldingRegisters(0, 10)
```

**When to use RTU over TCP:**
- Serial-to-Ethernet converters in "raw" or "transparent" mode
- Devices that expect RTU framing regardless of transport
- Legacy systems upgraded with Ethernet adapters

### UDP Transport

For MODBUS over UDP (less reliable but lower latency):

```go
import "github.com/adibhanna/modbus-go/transport"

// MODBUS over UDP - uses MBAP header like TCP but over UDP
udpTransport := transport.NewUDPTransport("192.168.1.100:502")

if err := udpTransport.Connect(); err != nil {
    log.Fatal(err)
}
defer udpTransport.Close()

// Create client with UDP transport
client := modbus.NewClient(udpTransport)
client.SetSlaveID(1)

// Use normally
values, err := client.ReadHoldingRegisters(0, 10)
```

**UDP considerations:**
- No automatic retransmission of lost packets
- No ordering guarantee
- Lower latency than TCP
- Useful for non-critical, high-frequency polling
- Configure appropriate timeout and retry settings

### Idle Timeout for Connections

Configure automatic connection cleanup after periods of inactivity:

```go
import "github.com/adibhanna/modbus-go/transport"

tcpTransport := transport.NewTCPTransport("192.168.1.100:502")

// Set idle timeout - connection will be closed after 5 minutes of inactivity
tcpTransport.SetIdleTimeout(5 * time.Minute)

// Check current idle timeout
idleTimeout := tcpTransport.GetIdleTimeout()

// Set connect timeout separately
tcpTransport.SetConnectTimeout(10 * time.Second)
```

### Custom Logger Support

Add custom logging for debugging and monitoring:

```go
import "github.com/adibhanna/modbus-go/transport"

// Logger interface - implement Printf method
type MyLogger struct{}

func (l *MyLogger) Printf(format string, v ...interface{}) {
    log.Printf("[MODBUS] "+format, v...)
}

// Set logger on transport
tcpTransport := transport.NewTCPTransport("192.168.1.100:502")
tcpTransport.SetLogger(&MyLogger{})

// Also works with RTU over TCP and UDP transports
rtuOverTCP := transport.NewRTUOverTCPTransport("192.168.1.100:4001")
rtuOverTCP.SetLogger(&MyLogger{})

udpTransport := transport.NewUDPTransport("192.168.1.100:502")
udpTransport.SetLogger(&MyLogger{})

// Use standard log package
type StdLogger struct{}
func (l *StdLogger) Printf(format string, v ...interface{}) {
    log.Printf(format, v...)
}
```

## Core Concepts

### Data Model

MODBUS defines four primary data types:

| Data Type | Access | Size | Address Range | Description |
|-----------|---------|------|---------------|-------------|
| **Coils** | Read/Write | 1 bit | 0-65535 | Discrete outputs |
| **Discrete Inputs** | Read Only | 1 bit | 0-65535 | Discrete inputs |
| **Holding Registers** | Read/Write | 16 bits | 0-65535 | General purpose registers |
| **Input Registers** | Read Only | 16 bits | 0-65535 | Input data registers |

### Addressing

MODBUS uses 16-bit addressing (0-65535). The library uses zero-based addressing internally:

```go
// Address type ensures type safety
var addr modbus.Address = 100  // Register at address 100

// Quantity type for counts
var qty modbus.Quantity = 10   // Read 10 registers
```

### Protocol Data Unit (PDU)

The PDU is the core protocol message structure:

```
[Function Code: 1 byte] [Data: 0-252 bytes]
```

Maximum PDU size is 253 bytes.

### Application Data Unit (ADU)

The ADU includes transport-specific framing:

**TCP/IP ADU:**
```
[MBAP Header: 7 bytes] [PDU: up to 253 bytes]
```

**Serial RTU ADU:**
```
[Address: 1 byte] [PDU: up to 253 bytes] [CRC: 2 bytes]
```

## Function Codes Reference

### Data Access Functions

#### Read Coils (0x01)
Reads the status of discrete outputs (coils).

```go
// Read 10 coils starting at address 100
values, err := client.ReadCoils(100, 10)
// values is []bool
```

#### Read Discrete Inputs (0x02)
Reads the status of discrete inputs.

```go
// Read 8 discrete inputs starting at address 0
values, err := client.ReadDiscreteInputs(0, 8)
// values is []bool
```

#### Read Holding Registers (0x03)
Reads multiple holding registers.

```go
// Read 5 holding registers starting at address 1000
values, err := client.ReadHoldingRegisters(1000, 5)
// values is []uint16
```

#### Read Input Registers (0x04)
Reads multiple input registers.

```go
// Read 3 input registers starting at address 500
values, err := client.ReadInputRegisters(500, 3)
// values is []uint16
```

#### Write Single Coil (0x05)
Writes a single coil.

```go
// Turn on coil at address 10
err := client.WriteSingleCoil(10, true)
```

#### Write Single Register (0x06)
Writes a single holding register.

```go
// Write value 12345 to register 100
err := client.WriteSingleRegister(100, 12345)
```

#### Write Multiple Coils (0x0F)
Writes multiple coils.

```go
// Write coil values starting at address 20
values := []bool{true, false, true, true, false}
err := client.WriteMultipleCoils(20, values)
```

#### Write Multiple Registers (0x10)
Writes multiple holding registers.

```go
// Write register values starting at address 200
values := []uint16{100, 200, 300, 400}
err := client.WriteMultipleRegisters(200, values)
```

#### Mask Write Register (0x16)
Modifies a register using AND/OR masks.

```go
// address: 100, AND mask: 0x00F0, OR mask: 0x0025
// Result = (current_value AND 0x00F0) OR 0x0025
err := client.MaskWriteRegister(100, 0x00F0, 0x0025)
```

#### Read/Write Multiple Registers (0x17)
Atomic read and write operation.

```go
// Read 5 registers from address 100, write 3 values to address 200
writeValues := []uint16{111, 222, 333}
readValues, err := client.ReadWriteMultipleRegisters(100, 5, 200, writeValues)
```

### Diagnostic Functions

#### Read Exception Status (0x07)
Reads device exception status (8 bits).

```go
status, err := client.ReadExceptionStatus()
// status is uint8, each bit indicates an exception
```

#### Diagnostics (0x08)
Various diagnostic subfunctions for serial line diagnostics.

```go
// Echo test - returns the same data
result, err := client.Diagnostic(modbus.DiagSubReturnQueryData, []byte{0x12, 0x34})

// Clear counters
_, err := client.Diagnostic(modbus.DiagSubClearCounters, nil)

// Get bus message count
result, err := client.Diagnostic(modbus.DiagSubReturnBusMessageCount, nil)
```

Available diagnostic subfunctions:
- `DiagSubReturnQueryData` (0x0000): Echo test
- `DiagSubRestartCommOption` (0x0001): Restart communications
- `DiagSubReturnDiagRegister` (0x0002): Return diagnostic register
- `DiagSubForceListenOnlyMode` (0x0004): Force listen-only mode
- `DiagSubClearCounters` (0x000A): Clear all counters
- `DiagSubReturnBusMessageCount` (0x000B): Get bus message count
- `DiagSubReturnBusCommErrorCount` (0x000C): Get communication errors
- `DiagSubReturnBusExceptionCount` (0x000D): Get exception errors
- `DiagSubReturnServerMessageCount` (0x000E): Get server messages
- `DiagSubReturnServerNoRespCount` (0x000F): Get no response count

#### Get Comm Event Counter (0x0B)
Retrieves communication event counter.

```go
status, eventCount, err := client.GetCommEventCounter()
// status: 0xFFFF = ready, 0x0000 = not ready
// eventCount: number of successful message interactions
```

#### Get Comm Event Log (0x0C)
Retrieves communication event log.

```go
status, eventCount, messageCount, events, err := client.GetCommEventLog()
// events is []byte containing the event log
```

#### Report Server ID (0x11)
Retrieves server identification and status.

```go
serverID, runIndicator, additionalData, err := client.ReportServerID()
// serverID: []byte with server identification string
// runIndicator: 0xFF = ON, 0x00 = OFF
```

### File Record Access

#### Read File Record (0x14)
Reads file records from extended memory.

```go
// Define records to read
records := []modbus.FileRecord{
    {
        ReferenceType: modbus.FileRecordTypeExtended, // Must be 0x06
        FileNumber:    4,      // File number (0-65535)
        RecordNumber:  1,      // Starting record (0-9999)
        RecordLength:  2,      // Number of registers to read
    },
}

// Read the records
result, err := client.ReadFileRecords(records)
// result[0].RecordData contains the register values
```

#### Write File Record (0x15)
Writes file records to extended memory.

```go
// Define records to write
records := []modbus.FileRecord{
    {
        ReferenceType: modbus.FileRecordTypeExtended,
        FileNumber:    4,
        RecordNumber:  1,
        RecordLength:  2,
        RecordData:    []uint16{0x1234, 0x5678}, // Data to write
    },
}

err := client.WriteFileRecords(records)
```

### FIFO Queue Access

#### Read FIFO Queue (0x18)
Reads FIFO queue contents.

```go
// Read FIFO queue at pointer address 1000
values, err := client.ReadFIFOQueue(1000)
// values is []uint16 with up to 31 values
```

### Device Identification

#### Encapsulated Interface Transport (0x2B)
Used for device identification and other encapsulated protocols.

```go
// Read basic device identification
vendorName, productCode, version, err := client.ReadDeviceIdentification(
    modbus.DeviceIDReadBasic,  // Read level
    0x00,                       // Object ID
)
```

## Client Usage

### Creating Clients

#### TCP Client

```go
// Basic TCP client
client, err := modbus.NewTCPClient("192.168.1.100:502", 1)

// With custom configuration
config := modbus.ClientConfig{
    SlaveID:    1,
    Timeout:    5 * time.Second,
    RetryCount: 3,
}
client, err := modbus.NewTCPClientWithConfig("192.168.1.100:502", config)
```

#### Serial RTU Client

```go
// RTU client
client, err := modbus.NewRTUClient("/dev/ttyUSB0", 1, 9600, 8, 1, modbus.ParityNone)

// With custom config
serialConfig := modbus.SerialConfig{
    Device:   "/dev/ttyUSB0",
    BaudRate: 19200,
    DataBits: 8,
    StopBits: 1,
    Parity:   modbus.ParityEven,
    Timeout:  3 * time.Second,
}
client, err := modbus.NewRTUClientWithConfig(serialConfig, 1)
```

### Client Best Practices

1. **Connection Management**
```go
// Always close connections
defer client.Close()

// Check connection before operations
if !client.IsConnected() {
    err := client.Connect()
}
```

2. **Error Handling**
```go
values, err := client.ReadHoldingRegisters(100, 10)
if err != nil {
    if modbusErr, ok := err.(*modbus.ModbusError); ok {
        // Handle MODBUS-specific errors
        switch modbusErr.ExceptionCode {
        case modbus.ExceptionCodeIllegalDataAddress:
            // Handle invalid address
        case modbus.ExceptionCodeServerDeviceBusy:
            // Retry later
        }
    }
    // Handle other errors
}
```

3. **Bulk Operations**
```go
// Efficient: Read all at once
values, err := client.ReadHoldingRegisters(0, 100)

// Inefficient: Multiple small reads
for i := 0; i < 100; i++ {
    value, err := client.ReadHoldingRegisters(i, 1) // Avoid this
}
```

## Server Usage

### Creating Servers

#### TCP Server

```go
// Create data store
dataStore := modbus.NewDefaultDataStore(
    10000,  // Number of coils
    10000,  // Number of discrete inputs
    10000,  // Number of holding registers
    10000,  // Number of input registers
)

// Create and start server
server, err := modbus.NewTCPServer(":502", dataStore)
if err != nil {
    log.Fatal(err)
}

// Start in blocking mode
err = server.Start()

// Or start in background
go func() {
    if err := server.Start(); err != nil {
        log.Fatal(err)
    }
}()
```

### Custom Data Store

Implement the `DataStore` interface for custom storage:

```go
type CustomDataStore struct {
    // Your storage implementation
    db *sql.DB
}

func (ds *CustomDataStore) ReadHoldingRegisters(address modbus.Address, quantity modbus.Quantity) ([]uint16, error) {
    // Read from database
    values := make([]uint16, quantity)
    rows, err := ds.db.Query("SELECT value FROM registers WHERE address >= ? LIMIT ?", address, quantity)
    // ... process rows
    return values, nil
}

func (ds *CustomDataStore) WriteHoldingRegisters(address modbus.Address, values []uint16) error {
    // Write to database
    tx, err := ds.db.Begin()
    // ... insert/update values
    return tx.Commit()
}

// Implement other required methods...
```

### Server Configuration

```go
// Custom device identification
handler := modbus.NewServerRequestHandler(dataStore)
handler.SetDeviceIdentification(&modbus.DeviceIdentification{
    VendorName:         "Your Company",
    ProductCode:        "PROD-001",
    MajorMinorRevision: "1.0.0",
    VendorURL:          "https://example.com",
    ProductName:        "Industrial Controller",
    ModelName:          "IC-2024",
    UserApplicationName: "Process Control",
    ConformityLevel:    modbus.ConformityLevelBasicStream,
})
```

### Server Event Handling

```go
// Monitor server events
type ServerMonitor struct {
    server *modbus.Server
    stats  struct {
        requests  uint64
        errors    uint64
        lastError error
    }
}

func (m *ServerMonitor) Start() {
    ticker := time.NewTicker(5 * time.Second)
    for range ticker.C {
        // Log statistics
        log.Printf("Requests: %d, Errors: %d", m.stats.requests, m.stats.errors)
        
        // Update diagnostic counters
        dataStore.IncrementDiagnosticCounter("ServerMessage")
    }
}
```

## Advanced Features

### File Records

File records provide access to extended memory beyond the 65536 address limit:

```go
// Server-side: Initialize file records
records := []modbus.FileRecord{
    {
        ReferenceType: modbus.FileRecordTypeExtended,
        FileNumber:    1,        // File 1
        RecordNumber:  0,        // Record 0
        RecordLength:  100,      // 100 registers
        RecordData:    data,     // Your data
    },
}
dataStore.WriteFileRecords(records)

// Client-side: Read file records
readReq := []modbus.FileRecord{
    {
        ReferenceType: modbus.FileRecordTypeExtended,
        FileNumber:    1,
        RecordNumber:  0,
        RecordLength:  100,
    },
}
result, err := client.ReadFileRecords(readReq)
```

### FIFO Queues

FIFO queues are useful for buffering time-series data:

```go
// Server-side: Manage FIFO queue
type FIFOManager struct {
    dataStore *modbus.DefaultDataStore
    address   modbus.Address
    maxSize   int
}

func (f *FIFOManager) Push(value uint16) error {
    current, _ := f.dataStore.ReadFIFOQueue(f.address)
    if len(current) >= f.maxSize {
        current = current[1:] // Remove oldest
    }
    current = append(current, value)
    return f.dataStore.WriteFIFOQueue(f.address, current)
}

// Client-side: Read FIFO
values, err := client.ReadFIFOQueue(1000)
for i, v := range values {
    fmt.Printf("Entry %d: %d\n", i, v)
}
```

### Diagnostics

Implement comprehensive diagnostics for troubleshooting:

```go
// Server-side: Track diagnostics
func trackDiagnostics(ds *modbus.DefaultDataStore, req *pdu.Request, resp *pdu.Response) {
    ds.IncrementDiagnosticCounter("BusMessage")
    
    if resp.IsException() {
        ds.IncrementDiagnosticCounter("BusException")
        ds.SetExceptionStatus(0xFF) // Set exception flag
    }
    
    if resp == nil {
        ds.IncrementDiagnosticCounter("ServerNoResp")
    }
}

// Client-side: Monitor link quality
func monitorLink(client modbus.Client) {
    // Echo test
    testData := []byte{0xAA, 0x55}
    result, err := client.Diagnostic(modbus.DiagSubReturnQueryData, testData)
    if err != nil || !bytes.Equal(result, testData) {
        log.Println("Link quality issue detected")
    }
    
    // Get error counts
    status, count, err := client.GetCommEventCounter()
    if count > lastCount {
        successRate := float64(count-errorCount) / float64(count) * 100
        log.Printf("Success rate: %.2f%%\n", successRate)
    }
}
```

## Transport Layers

### TCP/IP Transport

TCP transport uses the MODBUS TCP protocol with MBAP header:

```go
// MBAP Header Structure
type MBAPHeader struct {
    TransactionID uint16  // Transaction identifier
    ProtocolID    uint16  // Always 0 for MODBUS
    Length        uint16  // Number of following bytes
    UnitID        uint8   // Unit identifier (slave ID)
}

// Custom TCP transport options
transport := transport.NewTCPTransport(transport.TCPConfig{
    Address:         "192.168.1.100:502",
    ConnectTimeout:  5 * time.Second,
    ReadTimeout:     3 * time.Second,
    WriteTimeout:    3 * time.Second,
    KeepAlive:       30 * time.Second,
    MaxConnections:  10,
})
```

### Serial RTU Transport

RTU uses binary encoding with CRC error checking:

```go
// RTU Frame Structure
type RTUFrame struct {
    SlaveID      uint8    // Device address
    FunctionCode uint8    // Function code
    Data         []byte   // Data bytes
    CRC          uint16   // CRC-16 checksum
}

// Calculate CRC for RTU
func calculateCRC(data []byte) uint16 {
    crc := uint16(0xFFFF)
    for _, b := range data {
        crc ^= uint16(b)
        for i := 0; i < 8; i++ {
            if crc&0x0001 != 0 {
                crc = (crc >> 1) ^ 0xA001
            } else {
                crc >>= 1
            }
        }
    }
    return crc
}
```

### Serial ASCII Transport

ASCII uses text encoding with LRC error checking:

```go
// ASCII Frame Structure
type ASCIIFrame struct {
    Start        byte     // ':' character
    SlaveID      [2]byte  // Two ASCII hex chars
    FunctionCode [2]byte  // Two ASCII hex chars
    Data         []byte   // ASCII hex pairs
    LRC          [2]byte  // LRC checksum
    End          [2]byte  // CR LF
}

// Calculate LRC for ASCII
func calculateLRC(data []byte) uint8 {
    var lrc uint8
    for _, b := range data {
        lrc += b
    }
    return uint8(-int8(lrc))
}
```

## Error Handling

### Exception Codes

The library defines standard MODBUS exception codes:

| Code | Name | Description |
|------|------|-------------|
| 0x01 | Illegal Function | Function code not supported |
| 0x02 | Illegal Data Address | Invalid address or address range |
| 0x03 | Illegal Data Value | Invalid value in data field |
| 0x04 | Server Device Failure | Unrecoverable error occurred |
| 0x05 | Acknowledge | Long duration command acknowledged |
| 0x06 | Server Device Busy | Device busy, retry later |
| 0x08 | Memory Parity Error | Memory parity error detected |
| 0x0A | Gateway Path Unavailable | Gateway misconfigured |
| 0x0B | Gateway Target Failed | Target device failed to respond |

### Error Types

```go
// MODBUS protocol error
type ModbusError struct {
    FunctionCode  FunctionCode
    ExceptionCode ExceptionCode
    Message       string
}

// Transport error
type TransportError struct {
    Operation string
    Cause     error
}

// Timeout error
type TimeoutError struct {
    Operation string
    Duration  time.Duration
}
```

### Error Handling Patterns

```go
// Comprehensive error handling
func robustRead(client modbus.Client, address modbus.Address, count modbus.Quantity) ([]uint16, error) {
    maxRetries := 3
    backoff := 100 * time.Millisecond
    
    for retry := 0; retry < maxRetries; retry++ {
        values, err := client.ReadHoldingRegisters(address, count)
        if err == nil {
            return values, nil
        }
        
        // Check error type
        switch e := err.(type) {
        case *modbus.ModbusError:
            if e.ExceptionCode == modbus.ExceptionCodeServerDeviceBusy {
                // Device busy, wait and retry
                time.Sleep(backoff)
                backoff *= 2
                continue
            }
            // Other MODBUS errors are fatal
            return nil, err
            
        case *modbus.TimeoutError:
            // Network timeout, retry
            log.Printf("Timeout on attempt %d: %v", retry+1, e)
            continue
            
        case *modbus.TransportError:
            // Transport failure, might need reconnection
            if retry < maxRetries-1 {
                client.Reconnect()
                continue
            }
        }
        
        return nil, err
    }
    
    return nil, fmt.Errorf("failed after %d retries", maxRetries)
}
```

## Testing

### Unit Testing

```go
func TestReadHoldingRegisters(t *testing.T) {
    // Create mock data store
    ds := modbus.NewDefaultDataStore(100, 100, 100, 100)
    
    // Set test data
    testData := []uint16{100, 200, 300, 400, 500}
    for i, v := range testData {
        ds.SetHoldingRegister(modbus.Address(i), v)
    }
    
    // Create handler
    handler := modbus.NewServerRequestHandler(ds)
    
    // Create request
    req := pdu.NewRequest(modbus.FuncCodeReadHoldingRegisters, 
        append(pdu.EncodeUint16(0), pdu.EncodeUint16(5)...))
    
    // Process request
    resp := handler.HandleRequest(1, req)
    
    // Verify response
    if resp.FunctionCode != modbus.FuncCodeReadHoldingRegisters {
        t.Errorf("Wrong function code: %v", resp.FunctionCode)
    }
    
    // Decode response data
    if resp.Data[0] != 10 { // Byte count
        t.Errorf("Wrong byte count: %d", resp.Data[0])
    }
    
    values, _ := pdu.DecodeUint16Slice(resp.Data[1:])
    if !reflect.DeepEqual(values, testData) {
        t.Errorf("Data mismatch: got %v, want %v", values, testData)
    }
}
```

### Integration Testing

```go
func TestClientServerIntegration(t *testing.T) {
    // Setup server
    ds := modbus.NewDefaultDataStore(1000, 1000, 1000, 1000)
    server, _ := modbus.NewTCPServer(":15502", ds)
    go server.Start()
    defer server.Stop()
    
    // Wait for server
    time.Sleep(100 * time.Millisecond)
    
    // Create client
    client, err := modbus.NewTCPClient("localhost:15502", 1)
    if err != nil {
        t.Fatal(err)
    }
    defer client.Close()
    
    // Test write and read
    testValues := []uint16{111, 222, 333}
    err = client.WriteMultipleRegisters(100, testValues)
    if err != nil {
        t.Fatal(err)
    }
    
    readValues, err := client.ReadHoldingRegisters(100, 3)
    if err != nil {
        t.Fatal(err)
    }
    
    if !reflect.DeepEqual(readValues, testValues) {
        t.Errorf("Mismatch: wrote %v, read %v", testValues, readValues)
    }
}
```

### Load Testing

```go
func BenchmarkReadHoldingRegisters(b *testing.B) {
    ds := modbus.NewDefaultDataStore(10000, 10000, 10000, 10000)
    handler := modbus.NewServerRequestHandler(ds)
    
    req := pdu.NewRequest(modbus.FuncCodeReadHoldingRegisters,
        append(pdu.EncodeUint16(0), pdu.EncodeUint16(100)...))
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        resp := handler.HandleRequest(1, req)
        if resp.IsException() {
            b.Fatal("Unexpected exception")
        }
    }
}

func TestConcurrentAccess(t *testing.T) {
    ds := modbus.NewDefaultDataStore(1000, 1000, 1000, 1000)
    server, _ := modbus.NewTCPServer(":25502", ds)
    go server.Start()
    defer server.Stop()
    
    // Create multiple concurrent clients
    var wg sync.WaitGroup
    errors := make(chan error, 10)
    
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            
            client, err := modbus.NewTCPClient("localhost:25502", uint8(id))
            if err != nil {
                errors <- err
                return
            }
            defer client.Close()
            
            // Perform operations
            for j := 0; j < 100; j++ {
                addr := modbus.Address(id * 100)
                value := uint16(id*1000 + j)
                
                err := client.WriteSingleRegister(addr, value)
                if err != nil {
                    errors <- err
                    return
                }
                
                readValue, err := client.ReadHoldingRegisters(addr, 1)
                if err != nil {
                    errors <- err
                    return
                }
                
                if readValue[0] != value {
                    errors <- fmt.Errorf("Value mismatch")
                    return
                }
            }
        }(i)
    }
    
    wg.Wait()
    close(errors)
    
    for err := range errors {
        if err != nil {
            t.Error(err)
        }
    }
}
```

## Performance Considerations

### Optimization Tips

1. **Batch Operations**
   - Use multi-register reads/writes instead of single operations
   - Maximum efficiency: read up to 125 registers at once

2. **Connection Pooling**
```go
type ClientPool struct {
    clients chan modbus.Client
    factory func() (modbus.Client, error)
}

func NewClientPool(size int, factory func() (modbus.Client, error)) *ClientPool {
    pool := &ClientPool{
        clients: make(chan modbus.Client, size),
        factory: factory,
    }
    
    // Pre-populate pool
    for i := 0; i < size; i++ {
        client, _ := factory()
        pool.clients <- client
    }
    
    return pool
}

func (p *ClientPool) Get() modbus.Client {
    return <-p.clients
}

func (p *ClientPool) Put(client modbus.Client) {
    p.clients <- client
}
```

3. **Caching**
```go
type CachedDataStore struct {
    modbus.DataStore
    cache     map[string]cacheEntry
    cacheLock sync.RWMutex
    ttl       time.Duration
}

type cacheEntry struct {
    data      interface{}
    timestamp time.Time
}

func (c *CachedDataStore) ReadHoldingRegisters(address modbus.Address, quantity modbus.Quantity) ([]uint16, error) {
    key := fmt.Sprintf("hr:%d:%d", address, quantity)
    
    // Check cache
    c.cacheLock.RLock()
    entry, found := c.cache[key]
    c.cacheLock.RUnlock()
    
    if found && time.Since(entry.timestamp) < c.ttl {
        return entry.data.([]uint16), nil
    }
    
    // Read from underlying store
    data, err := c.DataStore.ReadHoldingRegisters(address, quantity)
    if err != nil {
        return nil, err
    }
    
    // Update cache
    c.cacheLock.Lock()
    c.cache[key] = cacheEntry{data: data, timestamp: time.Now()}
    c.cacheLock.Unlock()
    
    return data, nil
}
```

### Benchmarks

Typical performance metrics on modern hardware:

| Operation | Throughput | Latency |
|-----------|------------|---------|
| Read 100 registers (TCP) | ~10,000 req/s | ~0.1ms |
| Write 100 registers (TCP) | ~8,000 req/s | ~0.125ms |
| Read single coil (TCP) | ~15,000 req/s | ~0.067ms |
| Read 100 registers (RTU 115200) | ~50 req/s | ~20ms |
| Read 100 registers (RTU 9600) | ~5 req/s | ~200ms |

## Troubleshooting

### Common Issues and Solutions

#### Connection Refused
```go
// Issue: dial tcp 192.168.1.100:502: connect: connection refused

// Solutions:
// 1. Check if server is running
// 2. Verify IP address and port
// 3. Check firewall settings
// 4. For Linux, check if port needs sudo (ports < 1024)
```

#### Timeout Errors
```go
// Issue: operation timeout after 3s

// Solutions:
// 1. Increase timeout
client.SetTimeout(10 * time.Second)

// 2. Check network latency
// 3. Reduce request size
// 4. For serial, check baud rate and cable length
```

#### Illegal Data Address
```go
// Issue: MODBUS Error [ReadHoldingRegisters]: IllegalDataAddress

// Solutions:
// 1. Verify address is within valid range
// 2. Check if registers are configured on device
// 3. Some devices use 1-based addressing (add 1 to address)
// 4. Check device documentation for memory map
```

#### CRC/Checksum Errors
```go
// Issue: CRC error in response

// Solutions:
// 1. For serial: Check cable quality and connections
// 2. Reduce baud rate
// 3. Add termination resistors (RS-485)
// 4. Check for electrical interference
```

### Debugging Tools

#### Request/Response Logging
```go
type LoggingTransport struct {
    transport.Transport
    logger *log.Logger
}

func (lt *LoggingTransport) SendRequest(slaveID byte, req *pdu.Request) (*pdu.Response, error) {
    lt.logger.Printf("Request: SlaveID=%d, FC=%02X, Data=%X", 
        slaveID, req.FunctionCode, req.Data)
    
    resp, err := lt.Transport.SendRequest(slaveID, req)
    
    if err != nil {
        lt.logger.Printf("Error: %v", err)
    } else {
        lt.logger.Printf("Response: FC=%02X, Data=%X", 
            resp.FunctionCode, resp.Data)
    }
    
    return resp, err
}
```

#### Protocol Analyzer
```go
func analyzeProtocol(data []byte) {
    fmt.Println("Protocol Analysis:")
    fmt.Printf("Raw bytes: % X\n", data)
    
    if len(data) >= 7 {
        // Check for MBAP header (TCP)
        transID := binary.BigEndian.Uint16(data[0:2])
        protoID := binary.BigEndian.Uint16(data[2:4])
        length := binary.BigEndian.Uint16(data[4:6])
        unitID := data[6]
        
        if protoID == 0 {
            fmt.Println("Protocol: MODBUS TCP")
            fmt.Printf("Transaction ID: %d\n", transID)
            fmt.Printf("Length: %d bytes\n", length)
            fmt.Printf("Unit ID: %d\n", unitID)
            
            if len(data) > 7 {
                pduData := data[7:]
                fmt.Printf("Function Code: 0x%02X\n", pduData[0])
                fmt.Printf("PDU Data: % X\n", pduData[1:])
            }
        }
    }
    
    // Check for RTU frame
    if len(data) >= 4 {
        slaveID := data[0]
        funcCode := data[1]
        crc := binary.LittleEndian.Uint16(data[len(data)-2:])
        calcCRC := calculateCRC(data[:len(data)-2])
        
        fmt.Println("Possible RTU frame:")
        fmt.Printf("Slave ID: %d\n", slaveID)
        fmt.Printf("Function Code: 0x%02X\n", funcCode)
        fmt.Printf("CRC: 0x%04X (calculated: 0x%04X)\n", crc, calcCRC)
        
        if crc == calcCRC {
            fmt.Println("CRC: Valid")
        } else {
            fmt.Println("CRC: Invalid")
        }
    }
}
```

### Performance Profiling

```go
import (
    "net/http"
    _ "net/http/pprof"
    "runtime/pprof"
)

// Enable profiling endpoint
go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()

// CPU profiling
cpuFile, _ := os.Create("cpu.prof")
pprof.StartCPUProfile(cpuFile)
defer pprof.StopCPUProfile()

// Memory profiling
memFile, _ := os.Create("mem.prof")
defer pprof.WriteHeapProfile(memFile)

// Analyze with: go tool pprof cpu.prof
```

## Conclusion

ModbusGo provides a complete, production-ready MODBUS implementation with:

- Full protocol compliance with MODBUS specification V1.1b3
- Support for all standard function codes
- Multiple transport options (TCP, RTU, ASCII)
- Comprehensive error handling
- Thread-safe operations
- Extensive testing
- Clean, maintainable architecture

The library is suitable for:
- Industrial automation systems
- SCADA applications
- IoT gateways
- PLC communication
- Building automation
- Energy management systems
- Testing and simulation tools

For additional support or contributions, please visit the [GitHub repository](https://github.com/adibhanna/modbus-go).