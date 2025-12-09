# ModbusGo Transport Layer Guide

This guide covers the transport layer implementations in ModbusGo, including TCP/IP, RTU (serial), ASCII, TLS, RTU-over-TCP, and UDP transports.

## Table of Contents

1. [Overview](#overview)
2. [TCP/IP Transport](#tcpip-transport)
3. [TLS Transport](#tls-transport)
4. [RTU over TCP Transport](#rtu-over-tcp-transport)
5. [UDP Transport](#udp-transport)
6. [Serial RTU Transport](#serial-rtu-transport)
7. [Serial ASCII Transport](#serial-ascii-transport)
8. [Transport Interface](#transport-interface)
9. [Error Handling](#error-handling)
10. [Performance Tuning](#performance-tuning)

## Overview

ModbusGo supports multiple transport protocols for different use cases:

| Transport | Use Case | Frame Format | Error Detection |
|-----------|----------|--------------|-----------------|
| **TCP/IP** | Ethernet networks | MBAP header + PDU | TCP checksums |
| **TLS** | Secure TCP communications | MBAP header + PDU | TLS + TCP checksums |
| **RTU over TCP** | Serial-to-Ethernet converters | Address + PDU + CRC | CRC-16 |
| **UDP** | Low-latency, connectionless | MBAP header + PDU | Application layer |
| **RTU** | Serial lines (RS-232/485) | Address + PDU + CRC | CRC-16 |
| **ASCII** | Serial lines (human-readable) | ':' + hex chars + LRC + CRLF | LRC |

## TCP/IP Transport

### MBAP Header Structure

The MODBUS Application Protocol header (MBAP) is used for TCP/IP transport:

```
+-------------------+-------------------+-------------------+----------+
| Transaction ID    | Protocol ID       | Length            | Unit ID  |
| (2 bytes)         | (2 bytes)         | (2 bytes)         | (1 byte) |
+-------------------+-------------------+-------------------+----------+
```

- **Transaction ID**: Unique identifier for request/response pairing
- **Protocol ID**: Always 0x0000 for MODBUS
- **Length**: Number of following bytes (Unit ID + PDU)
- **Unit ID**: MODBUS slave ID (for gateway routing)

### TCP Client Usage

```go
import (
    modbus "github.com/adibhanna/modbus-go"
    "github.com/adibhanna/modbus-go/transport"
)

// Simple TCP client
client, err := modbus.NewTCPClient("192.168.1.100:502", 1)
if err != nil {
    log.Fatal(err)
}
defer client.Close()

// Configure timeouts
client.SetTimeout(5 * time.Second)
client.SetConnectTimeout(10 * time.Second)

// Enable auto-reconnect for resilient connections
client.SetAutoReconnect(true)

// Perform operations
values, err := client.ReadHoldingRegisters(0, 10)
```

### TCP Server Usage

```go
// Create data store
dataStore := modbus.NewDefaultDataStore(10000, 10000, 10000, 10000)

// Create server
server := transport.NewTCPServer(":502", handler)

// Start server
if err := server.Start(); err != nil {
    log.Fatal(err)
}

// Graceful shutdown with timeout
err = server.StopWithTimeout(5 * time.Second)
```

### TCP Configuration Options

| Option | Default | Description |
|--------|---------|-------------|
| Address | Required | Server address (host:port) |
| Timeout | 1000ms | Response timeout |
| ConnectTimeout | 5000ms | Connection timeout |
| RetryCount | 3 | Number of retry attempts |
| RetryDelay | 100ms | Delay between retries |
| IdleTimeout | 0 (disabled) | Auto-close idle connections |

### Idle Timeout

Configure automatic connection cleanup after periods of inactivity:

```go
tcpTransport := transport.NewTCPTransport("192.168.1.100:502")

// Set idle timeout - connection will be closed after 5 minutes of inactivity
tcpTransport.SetIdleTimeout(5 * time.Minute)

// Check current idle timeout
idleTimeout := tcpTransport.GetIdleTimeout()

// Set connect timeout separately
tcpTransport.SetConnectTimeout(10 * time.Second)
```

### Custom Logger

Add custom logging for debugging and monitoring:

```go
// Logger interface - implement Printf method
type MyLogger struct{}

func (l *MyLogger) Printf(format string, v ...interface{}) {
    log.Printf("[MODBUS] "+format, v...)
}

// Set logger on transport
tcpTransport := transport.NewTCPTransport("192.168.1.100:502")
tcpTransport.SetLogger(&MyLogger{})
```

## TLS Transport

Secure your MODBUS TCP communications with TLS encryption.

### Basic TLS Usage

```go
import (
    "crypto/tls"
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

// Create client
client := modbus.NewClient(tcpTransport)
client.SetSlaveID(1)

values, err := client.ReadHoldingRegisters(0, 10)
```

### Mutual TLS (mTLS)

For environments requiring client certificate authentication:

```go
import (
    "crypto/tls"
    "crypto/x509"
    "io/ioutil"
)

// Load client certificate
cert, err := tls.LoadX509KeyPair("client-cert.pem", "client-key.pem")
if err != nil {
    log.Fatal(err)
}

// Load CA certificate
caCert, err := ioutil.ReadFile("ca-cert.pem")
if err != nil {
    log.Fatal(err)
}
caCertPool := x509.NewCertPool()
caCertPool.AppendCertsFromPEM(caCert)

// Configure TLS
tlsConfig := &tls.Config{
    Certificates: []tls.Certificate{cert},
    RootCAs:      caCertPool,
    MinVersion:   tls.VersionTLS12,
}

secureTransport := transport.NewTLSTransport("192.168.1.100:802", tlsConfig)
```

### TLS Configuration Options

| Option | Description |
|--------|-------------|
| MinVersion | Minimum TLS version (recommended: TLS 1.2+) |
| Certificates | Client certificates for mTLS |
| RootCAs | CA certificates for server verification |
| InsecureSkipVerify | Skip certificate verification (NOT for production) |

## RTU over TCP Transport

For serial-to-Ethernet converters that use RTU framing over TCP connections.

### When to Use RTU over TCP

- Serial-to-Ethernet converters in "raw" or "transparent" mode
- Devices that expect RTU framing regardless of transport
- Legacy systems upgraded with Ethernet adapters
- Gateways that don't translate to standard MODBUS TCP

### RTU over TCP Frame Structure

Unlike standard MODBUS TCP which uses MBAP headers, RTU over TCP sends RTU frames directly:

```
+----------+---------------+------------+----------+
| Address  | Function Code | Data       | CRC      |
| (1 byte) | (1 byte)      | (0-252)    | (2 bytes)|
+----------+---------------+------------+----------+
```

### RTU over TCP Usage

```go
import "github.com/adibhanna/modbus-go/transport"

// Create RTU over TCP transport
rtuOverTCP := transport.NewRTUOverTCPTransport("192.168.1.100:4001")

// Connect
if err := rtuOverTCP.Connect(); err != nil {
    log.Fatal(err)
}
defer rtuOverTCP.Close()

// Create client
client := modbus.NewClient(rtuOverTCP)
client.SetSlaveID(1)

// Use normally - transport handles RTU framing over TCP
values, err := client.ReadHoldingRegisters(0, 10)
```

### RTU over TCP with Logger

```go
rtuOverTCP := transport.NewRTUOverTCPTransport("192.168.1.100:4001")
rtuOverTCP.SetLogger(&MyLogger{})
```

## UDP Transport

MODBUS over UDP for low-latency, connectionless communication.

### When to Use UDP

- Non-critical, high-frequency polling
- Environments where latency is more important than reliability
- Broadcast/multicast scenarios
- Networks with low packet loss

### UDP Considerations

| Aspect | Description |
|--------|-------------|
| Reliability | No automatic retransmission of lost packets |
| Ordering | No guarantee of packet order |
| Latency | Lower than TCP (no connection overhead) |
| Error Handling | Application must handle lost packets |

### UDP Usage

```go
import "github.com/adibhanna/modbus-go/transport"

// Create UDP transport
udpTransport := transport.NewUDPTransport("192.168.1.100:502")

// Connect (resolves address and creates socket)
if err := udpTransport.Connect(); err != nil {
    log.Fatal(err)
}
defer udpTransport.Close()

// Create client
client := modbus.NewClient(udpTransport)
client.SetSlaveID(1)

// Configure shorter timeout for UDP
client.SetTimeout(500 * time.Millisecond)

// Use normally
values, err := client.ReadHoldingRegisters(0, 10)
```

### UDP Best Practices

```go
// 1. Use shorter timeouts
udpTransport.SetTimeout(500 * time.Millisecond)

// 2. Implement application-level retries
func readWithRetry(client *modbus.Client, addr, qty uint16) ([]uint16, error) {
    var lastErr error
    for i := 0; i < 3; i++ {
        values, err := client.ReadHoldingRegisters(modbus.Address(addr), modbus.Quantity(qty))
        if err == nil {
            return values, nil
        }
        lastErr = err
        time.Sleep(50 * time.Millisecond)
    }
    return nil, lastErr
}

// 3. Monitor packet loss
// UDP doesn't report lost packets - track expected vs received
```

## Serial RTU Transport

RTU (Remote Terminal Unit) mode uses binary encoding for efficient communication over serial lines.

### RTU Frame Structure

```
+----------+---------------+------------+----------+
| Address  | Function Code | Data       | CRC      |
| (1 byte) | (1 byte)      | (0-252)    | (2 bytes)|
+----------+---------------+------------+----------+
```

### CRC-16 Calculation

RTU uses CRC-16 with polynomial 0xA001:

```go
func calculateCRC16(data []byte) uint16 {
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

### RTU Client Usage

```go
import "github.com/adibhanna/modbus-go/transport"

// Create serial configuration
config := transport.NewSerialConfig(
    "/dev/ttyUSB0",  // Device path
    9600,            // Baud rate
    8,               // Data bits
    1,               // Stop bits
    serial.NoParity, // Parity
)

// Create RTU transport
rtuTransport := transport.NewRTUTransport(config)

// Connect
if err := rtuTransport.Connect(); err != nil {
    log.Fatal(err)
}
defer rtuTransport.Close()

// Create client with transport
client := modbus.NewClient(rtuTransport, 1)
```

### RTU Timing Requirements

Per MODBUS specification, RTU mode has strict timing requirements:

| Parameter | Value |
|-----------|-------|
| Inter-character timeout | 1.5 character times |
| Inter-frame delay | 3.5 character times |

Character time at different baud rates:

| Baud Rate | Character Time | Inter-frame Delay |
|-----------|---------------|-------------------|
| 9600 | 1.04ms | 3.65ms |
| 19200 | 0.52ms | 1.82ms |
| 38400 | 0.26ms | 0.91ms |
| 57600 | 0.17ms | 0.61ms |
| 115200 | 0.09ms | 0.30ms |

### Serial Port Configuration

```go
// Common configurations
config9600 := transport.NewSerialConfig("/dev/ttyUSB0", 9600, 8, 1, serial.NoParity)
config19200 := transport.NewSerialConfig("/dev/ttyUSB0", 19200, 8, 1, serial.EvenParity)

// For RS-485 half-duplex
// The driver/hardware handles RTS/CTS automatically
```

## Serial ASCII Transport

ASCII mode uses human-readable hexadecimal encoding, making it easier to debug but less efficient.

### ASCII Frame Structure

```
+-------+----------+---------------+------+-----+------+------+
| Start | Address  | Function Code | Data | LRC | CR   | LF   |
| ':'   | (2 hex)  | (2 hex)       | hex  |(2)  | 0x0D | 0x0A |
+-------+----------+---------------+------+-----+------+------+
```

### LRC Calculation

ASCII mode uses Longitudinal Redundancy Check:

```go
func calculateLRC(data []byte) uint8 {
    var lrc uint8
    for _, b := range data {
        lrc += b
    }
    return uint8(-int8(lrc)) // Two's complement
}
```

### ASCII Client Usage

```go
// Create serial configuration
config := transport.NewSerialConfig(
    "/dev/ttyUSB0",
    9600,
    7,               // 7 data bits for ASCII
    1,
    serial.EvenParity, // Even parity common for ASCII
)

// Create ASCII transport
asciiTransport := transport.NewASCIITransport(config)

// Connect
if err := asciiTransport.Connect(); err != nil {
    log.Fatal(err)
}
defer asciiTransport.Close()

// Create client
client := modbus.NewClient(asciiTransport, 1)
```

## Transport Interface

All transports implement the common `Transport` interface:

```go
type Transport interface {
    // Connect establishes the transport connection
    Connect() error

    // Close closes the transport connection
    Close() error

    // IsConnected returns true if connected
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
```

### Custom Transport Implementation

You can create custom transports by implementing the interface:

```go
type CustomTransport struct {
    // Your fields
}

func (t *CustomTransport) Connect() error {
    // Custom connection logic
    return nil
}

func (t *CustomTransport) SendRequest(slaveID modbus.SlaveID, req *pdu.Request) (*pdu.Response, error) {
    // Custom request/response handling
    return nil, nil
}

// Implement other methods...
```

## Error Handling

### Transport Errors

Common transport errors and handling:

```go
resp, err := client.ReadHoldingRegisters(0, 10)
if err != nil {
    switch {
    case errors.Is(err, os.ErrDeadlineExceeded):
        // Timeout - device not responding
        log.Println("Timeout: Device not responding")

    case errors.Is(err, io.EOF):
        // Connection closed
        log.Println("Connection closed by remote")

    default:
        // Check for MODBUS errors
        var modbusErr *modbus.ModbusError
        if errors.As(err, &modbusErr) {
            log.Printf("MODBUS Exception: %s", modbusErr.ExceptionCode)
        }
    }
}
```

### CRC Errors (RTU)

```go
// CRC errors indicate data corruption
// - Check cable connections
// - Reduce baud rate
// - Add termination resistors (RS-485)
// - Check for electrical interference
```

### LRC Errors (ASCII)

```go
// LRC errors are less common due to ASCII encoding
// Similar causes to CRC errors
```

## Performance Tuning

### TCP Performance

```go
// Optimize for high-throughput scenarios
client.SetTimeout(500 * time.Millisecond) // Reduce timeout for fast networks
client.SetRetryCount(1)                   // Reduce retries for low latency

// Use bulk reads instead of multiple small reads
values, _ := client.ReadHoldingRegisters(0, 125) // Max 125 registers
```

### Serial Performance

```go
// Higher baud rates = faster communication
config := transport.NewSerialConfig(device, 115200, 8, 1, serial.NoParity)

// Reduce inter-request delay
// Note: Must respect MODBUS timing requirements
```

### Batch Operations

```go
// Efficient: Single request for multiple registers
values, _ := client.ReadHoldingRegisters(0, 100)

// Inefficient: Multiple requests
for i := 0; i < 100; i++ {
    val, _ := client.ReadHoldingRegisters(uint16(i), 1) // Avoid!
}

// Use Read/Write Multiple Registers for atomic operations
readVals, _ := client.ReadWriteMultipleRegisters(0, 10, 100, writeVals)
```

### Connection Pooling (TCP)

For high-concurrency applications:

```go
type ClientPool struct {
    clients chan *modbus.Client
    address string
    size    int
}

func NewClientPool(address string, size int) *ClientPool {
    pool := &ClientPool{
        clients: make(chan *modbus.Client, size),
        address: address,
        size:    size,
    }

    // Pre-create clients
    for i := 0; i < size; i++ {
        client, _ := modbus.NewTCPClient(address, 1)
        client.Connect()
        pool.clients <- client
    }

    return pool
}

func (p *ClientPool) Get() *modbus.Client {
    return <-p.clients
}

func (p *ClientPool) Put(client *modbus.Client) {
    p.clients <- client
}
```

## Choosing the Right Transport

| Requirement | Recommended Transport |
|-------------|----------------------|
| Ethernet/WiFi network | TCP/IP |
| Secure communications | TLS |
| Long distance (>15m) | TCP/IP or RS-485 RTU |
| High speed (>100 req/s) | TCP/IP or UDP |
| Low latency critical | UDP |
| Serial-to-Ethernet converters | RTU over TCP |
| Noisy environment | RTU with RS-485 |
| Debugging/monitoring | ASCII |
| Legacy equipment | RTU (most common) |
| Simple wiring | TCP/IP |
| Industrial environment | RS-485 RTU |
| Non-critical high-frequency polling | UDP |

## Troubleshooting

### TCP Issues

| Problem | Solution |
|---------|----------|
| Connection refused | Check server is running, firewall rules |
| Connection timeout | Increase timeout, check network |
| Transaction ID mismatch | Check for network proxies/load balancers |

### Serial Issues

| Problem | Solution |
|---------|----------|
| No response | Check wiring, baud rate, slave ID |
| CRC errors | Reduce baud rate, check cable quality |
| Intermittent failures | Add termination resistors (RS-485) |
| Garbled data | Verify data bits, parity, stop bits |

### Debug Logging

```go
// Enable debug output
import "log"

// Custom transport wrapper for logging
type LoggingTransport struct {
    transport.Transport
}

func (t *LoggingTransport) SendRequest(slaveID modbus.SlaveID, req *pdu.Request) (*pdu.Response, error) {
    log.Printf("TX: SlaveID=%d FC=0x%02X Data=%X", slaveID, req.FunctionCode, req.Data)

    resp, err := t.Transport.SendRequest(slaveID, req)

    if err != nil {
        log.Printf("Error: %v", err)
    } else {
        log.Printf("RX: FC=0x%02X Data=%X", resp.FunctionCode, resp.Data)
    }

    return resp, err
}
```
