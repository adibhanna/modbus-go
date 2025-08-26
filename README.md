# ModbusGo

A comprehensive, production-ready MODBUS implementation in Go supporting the complete MODBUS Application Protocol Specification V1.1b3.

## 📚 Documentation

- **[Complete Documentation](DOCUMENTATION.md)** - Comprehensive guide covering all features with examples
- **[API Reference](API_REFERENCE.md)** - Detailed API documentation for all packages
- **[Examples](examples/)** - Ready-to-run example implementations

## ✨ Features

- **Complete Protocol Implementation** - All 19 standard MODBUS function codes
- **Multiple Transport Protocols** - TCP/IP, RTU (serial), and ASCII
- **Client and Server Support** - Full-featured client and server implementations  
- **Advanced Features** - File records, FIFO queues, diagnostics, device identification
- **Thread-Safe** - Concurrent-safe operations with proper synchronization
- **Production Ready** - Comprehensive error handling and recovery mechanisms
- **Well Tested** - Extensive test coverage for all components
- **Zero Dependencies** - Uses only Go standard library

## 📦 Installation

```bash
go get github.com/adibhanna/modbusgo
```

## 🚀 Quick Start

### TCP Client

```go
package main

import (
    "fmt"
    "log"
    modbus "github.com/adibhanna/modbusgo"
)

func main() {
    // Connect to MODBUS TCP server
    client, err := modbus.NewTCPClient("192.168.1.100:502", 1)
    if err != nil {
        log.Fatal(err)
    }
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
    modbus "github.com/adibhanna/modbusgo"
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

- **Documentation**: See [DOCUMENTATION.md](DOCUMENTATION.md)
- **API Reference**: See [API_REFERENCE.md](API_REFERENCE.md)
- **Issues**: [GitHub Issues](https://github.com/adibhanna/modbusgo/issues)
- **Discussions**: [GitHub Discussions](https://github.com/adibhanna/modbusgo/discussions)