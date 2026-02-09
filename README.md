# ModbusGo

A comprehensive, production-ready MODBUS implementation in Go supporting the complete MODBUS Application Protocol Specification V1.1b3.

## 📚 Documentation

- **[Complete Documentation](docs/DOCUMENTATION.md)** - Comprehensive guide covering all features with examples
- **[API Reference](docs/API_REFERENCE.md)** - Detailed API documentation for all packages
- **[Configuration Guide](docs/DOCUMENTATION.md#configuration)** - Complete configuration system documentation
- **[Transport Guide](docs/TRANSPORTS.md)** - TCP, RTU, and ASCII transport details
- **[Examples](examples/)** - Ready-to-run example implementations
- **[Configuration Examples](config-examples/)** - Device-specific configuration templates

## ✨ Features

- **Complete Protocol Implementation** - All 19 standard MODBUS function codes
- **Multiple Transport Protocols** - TCP/IP, TLS, RTU over TCP, UDP, RTU (serial), and ASCII
- **Client and Server Support** - Full-featured client and server implementations
- **Flexible Configuration** - JSON-based configuration with device profiles and runtime management
- **High-Level Data Types** - Read/write uint32, uint64, float32, float64, strings, bytes
- **Configurable Endianness** - Support for different byte/word orderings
- **TLS Support** - Secure MODBUS TCP with certificate authentication
- **RTU over TCP** - Support for serial-to-Ethernet converters
- **UDP Transport** - Low-latency connectionless MODBUS communication
- **Advanced Features** - File records, FIFO queues, diagnostics, device identification
- **Auto-Reconnect** - Automatic connection recovery on failure
- **Broadcast Support** - Send commands to all devices (slave ID 0)
- **Graceful Shutdown** - Server shutdown with timeout and proper cleanup
- **Custom Logging** - Pluggable logger interface for debugging
- **Idle Timeout** - Automatic cleanup of idle connections
- **Thread-Safe** - Concurrent-safe operations with proper synchronization
- **Production Ready** - Comprehensive error handling and recovery mechanisms
- **Well Tested** - Extensive test coverage for all components
- **Minimal Dependencies** - Only Go standard library plus serial port support

## 📦 Installation

```bash
go get github.com/adibhanna/modbus-go@latest
```

Or use Docker:

```bash
docker pull adibhanna/modbus-go:latest
```

## 🚀 Quick Start

### TCP Client

```go
package main

import (
    "fmt"
    "log"
    modbus "github.com/adibhanna/modbus-go"
)

func main() {
    // Connect to MODBUS TCP server
    client := modbus.NewTCPClient("192.168.1.100:502")
    client.SetSlaveID(1)
    defer client.Close()

    // Read 10 holding registers starting from address 100
    values, err := client.ReadHoldingRegisters(100, 10)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Register values: %v\n", values)
    
    // Write single register
    err = client.WriteSingleRegister(200, 1234)
    if err != nil {
        log.Fatal(err)
    }
}
```

### TCP Server

```go
package main

import (
    "log"
    modbus "github.com/adibhanna/modbus-go"
)

func main() {
    // Create data store (10000 addresses for each data type)
    dataStore := modbus.NewDefaultDataStore(10000, 10000, 10000, 10000)
    
    // Initialize some data
    dataStore.SetHoldingRegister(100, 42)
    dataStore.SetCoil(0, true)
    
    // Create and start TCP server
    server, err := modbus.NewTCPServer(":502", dataStore)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Println("MODBUS TCP Server starting on :502")
    if err := server.Start(); err != nil {
        log.Fatal(err)
    }
}
```

## ⚙️ Configuration

The library supports both programmatic and JSON-based configuration for flexible client setup:

### JSON Configuration

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

### Configuration Examples

```go
// Load client from JSON file (create your own config file)
client, err := modbus.NewTCPClientFromJSONFile("my-config.json", "192.168.1.102:502")

// Load client from JSON string  
jsonConfig := `{"slave_id": 1, "timeout_ms": 5000, "retry_count": 2}`
client, err := modbus.NewTCPClientFromJSONString(jsonConfig, "192.168.1.102:502")

// Use configuration struct
config := modbus.DefaultClientConfig()
config.SlaveID = 2
config.RetryCount = 5
client := modbus.NewTCPClientFromConfig(config, "192.168.1.102:502")

// Runtime configuration changes
client.SetSlaveID(3)
client.SetRetryDelay(200 * time.Millisecond)
client.SetRetryCount(1)

// Save current configuration
config := client.GetConfig()
config.SaveClientConfigToJSON("saved-config.json")
```

### Advanced Configuration

For comprehensive configuration with testing parameters, device profiles, and logging options, use the extended configuration system:

```go
import "github.com/adibhanna/modbus-go/config"

// Load extended configuration
cfg, err := config.LoadConfig("config.json")
client := modbus.NewTCPClient(cfg.Connection.GetFullAddress())
client.SetSlaveID(cfg.Modbus.GetSlaveID())
client.SetTimeout(cfg.Connection.GetTimeout())
```

### Device-Specific Configurations

Pre-configured templates for common MODBUS devices are available in [`config-examples/`](config-examples/):

- **Schneider Electric**: `config-examples/schneider-electric.json`
- **Siemens**: `config-examples/siemens.json`  
- **Diagnostic/Troubleshooting**: `config-examples/diagnostic.json`

```bash
# Use device-specific configuration
go run examples/tcp_client/main.go -config=config-examples/schneider-electric.json

# Run diagnostics with specific configuration
go run examples/tcp_client_diagnostic.go -config=config-examples/diagnostic.json
```

## 📋 Supported Function Codes

| Code     | Function                         | Description                           |
| -------- | -------------------------------- | ------------------------------------- |
| **0x01** | Read Coils                       | Read multiple coil status             |
| **0x02** | Read Discrete Inputs             | Read multiple discrete input status   |
| **0x03** | Read Holding Registers           | Read multiple holding registers       |
| **0x04** | Read Input Registers             | Read multiple input registers         |
| **0x05** | Write Single Coil                | Write single coil                     |
| **0x06** | Write Single Register            | Write single holding register         |
| **0x07** | Read Exception Status            | Read exception status (serial only)   |
| **0x08** | Diagnostics                      | Various diagnostic functions          |
| **0x0B** | Get Comm Event Counter           | Get communication event counter       |
| **0x0C** | Get Comm Event Log               | Get communication event log           |
| **0x0F** | Write Multiple Coils             | Write multiple coils                  |
| **0x10** | Write Multiple Registers         | Write multiple holding registers      |
| **0x11** | Report Server ID                 | Report server identification          |
| **0x14** | Read File Record                 | Read file record from extended memory |
| **0x15** | Write File Record                | Write file record to extended memory  |
| **0x16** | Mask Write Register              | Modify register using AND/OR masks    |
| **0x17** | Read/Write Multiple Registers    | Atomic read and write operation       |
| **0x18** | Read FIFO Queue                  | Read FIFO queue contents              |
| **0x2B** | Encapsulated Interface Transport | Device identification and other MEI   |

## 🏗️ Architecture

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
├── types.go         # Package-level type exports
└── examples/        # Example implementations
    ├── tcp_client/
    ├── tcp_server/
    └── advanced_server/
```

## 🔧 Advanced Features

### File Records (Extended Memory)

```go
// Read file records for accessing memory beyond 65536 addresses
records := []modbus.FileRecord{
    {
        ReferenceType: modbus.FileRecordTypeExtended,
        FileNumber:    4,
        RecordNumber:  1,
        RecordLength:  3,
    },
}
result, err := client.ReadFileRecords(records)
```

### FIFO Queues

```go
// Read FIFO queue for buffered/time-series data
values, err := client.ReadFIFOQueue(1000)
fmt.Printf("FIFO contains %d values: %v\n", len(values), values)
```

### Diagnostics

```go
// Perform diagnostic echo test
testData := []byte{0xAA, 0x55}
response, err := client.Diagnostic(modbus.DiagSubReturnQueryData, testData)

// Get communication statistics
status, eventCount, err := client.GetCommEventCounter()
```

### Device Identification

```go
// Read device identification information
objects, err := client.ReadDeviceIdentification(
    modbus.DeviceIDReadBasic, 
    0x00,
)
```

## 🐳 Docker

Run the MODBUS server and development environment using Docker:

```bash
# Build and run the MODBUS server
docker compose up modbus-server

# Run in background
docker compose up -d modbus-server

# Development with hot reload
docker compose --profile dev up

# Run tests in Docker
docker compose --profile test up

# Run CI checks in Docker
docker compose --profile ci up

# Test client against server
docker compose --profile client up
```

### Available Services

| Service | Description | Port |
|---------|-------------|------|
| `modbus-server` | Advanced MODBUS TCP server | 502, 5502 |
| `simple-server` | Simple TCP server | 5503 |
| `dev` | Development environment | 5502 |
| `test` | Run unit tests | - |
| `ci` | Run full CI pipeline | - |
| `client` | Test client | - |

### Makefile Docker Targets

```bash
make docker-build   # Build Docker image
make docker-run     # Run MODBUS server container
make docker-up      # Start all services with docker-compose
make docker-down    # Stop all services
make docker-test    # Run tests in Docker
make docker-ci      # Run CI checks in Docker
make docker-dev     # Start development environment
make docker-clean   # Remove Docker images and containers
make docker-shell   # Open shell in running container
```

## 🧪 Testing

Run the complete test suite:

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific test
go test -run TestTCPClient ./...

# Run benchmarks
go test -bench=. ./...

# Integration tests
./test_integration.sh
```

## 📊 Performance

Typical performance on modern hardware:

| Operation           | Transport  | Throughput    | Latency  |
| ------------------- | ---------- | ------------- | -------- |
| Read 100 registers  | TCP        | ~10,000 req/s | ~0.1ms   |
| Write 100 registers | TCP        | ~8,000 req/s  | ~0.125ms |
| Read single coil    | TCP        | ~15,000 req/s | ~0.067ms |
| Read 100 registers  | RTU 115200 | ~50 req/s     | ~20ms    |
| Read 100 registers  | RTU 9600   | ~5 req/s      | ~200ms   |

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 📚 References

- [MODBUS Application Protocol Specification V1.1b3](https://modbus.org/docs/Modbus_Application_Protocol_V1_1b3.pdf)
- [MODBUS over Serial Line Specification V1.02](https://modbus.org/docs/Modbus_over_serial_line_V1_02.pdf)
- [MODBUS Messaging on TCP/IP Implementation Guide V1.0b](https://modbus.org/docs/Modbus_Messaging_Implementation_Guide_V1_0b.pdf)

## 🙏 Acknowledgments

- The MODBUS Organization for the protocol specifications
- The Go community for excellent tools and libraries
- All contributors who help improve this library

## 📞 Support

- **Documentation**: See [docs/DOCUMENTATION.md](docs/DOCUMENTATION.md)
- **API Reference**: See [docs/API_REFERENCE.md](docs/API_REFERENCE.md)
- **Docker Hub**: [adibhanna/modbus-go](https://hub.docker.com/r/adibhanna/modbus-go)
- **Issues**: [GitHub Issues](https://github.com/adibhanna/modbus-go/issues)
- **Discussions**: [GitHub Discussions](https://github.com/adibhanna/modbus-go/discussions)
