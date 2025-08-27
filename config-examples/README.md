# Configuration Examples

This directory contains example configuration files for different MODBUS devices and use cases. These configurations can be used with both the core library's JSON configuration system and the extended configuration system.

## Core Library Configuration Examples

### Basic JSON Configuration

Simple JSON configuration for the core library:

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

**Usage:**
```go
// Save above JSON to a file, then load it
client, err := modbus.NewTCPClientFromJSONFile("my-config.json", "192.168.1.102:502")

// Or use inline JSON string
jsonConfig := `{"slave_id": 1, "timeout_ms": 10000, ...}`
client, err := modbus.NewTCPClientFromJSONString(jsonConfig, "192.168.1.102:502")
```

## Device-Specific Configurations

### Schneider Electric (`schneider-electric.json`)

Configuration optimized for Schneider Electric MODBUS devices:

- Uses 1-based addressing (typical for Schneider devices)
- Longer timeout for industrial environments
- Comprehensive test suite enabled
- Device profile with supported function codes

**Key Features:**
- Slave ID: 1
- Timeout: 10 seconds
- 1-based register addressing
- Extended test coverage

**Usage:**
```go
import "github.com/adibhanna/modbus-go/config"

cfg, err := config.LoadConfig("config-examples/schneider-electric.json")
client := modbus.NewTCPClient(cfg.Connection.GetFullAddress())
client.SetSlaveID(cfg.Modbus.GetSlaveID())
```

### Siemens (`siemens.json`)

Configuration tailored for Siemens MODBUS devices:

- Uses 0-based addressing (typical for Siemens devices)
- Optimized timeout settings
- Focus on essential MODBUS functions

**Key Features:**
- Slave ID: 1
- Timeout: 10 seconds
- 0-based register addressing
- Efficient retry settings

### Diagnostic Configuration (`diagnostic.json`)

Minimal configuration for device diagnosis and troubleshooting:

- Short timeouts for quick feedback
- Single retry for fast failure detection
- Basic test set for connectivity verification
- Verbose logging enabled

**Key Features:**
- Timeout: 5 seconds
- Retry count: 1
- Minimal test set
- Debug logging enabled

**Usage:**
```bash
# Use with diagnostic tool
go run examples/tcp_client_diagnostic.go -config=config-examples/diagnostic.json
```

## Extended Configuration Features

All extended configuration files include:

### Connection Settings
```json
{
  "connection": {
    "address": "192.168.1.102",
    "port": 502,
    "timeout_ms": 10000,
    "connect_timeout_ms": 5000,
    "retry_count": 3,
    "transport_type": "tcp"
  }
}
```

### MODBUS Protocol Settings
```json
{
  "modbus": {
    "slave_id": 1,
    "unit_id": 1,
    "protocol_id": 0
  }
}
```

### Testing Configuration
```json
{
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
  }
}
```

### Device Profiles
```json
{
  "device_profiles": {
    "generic": {
      "slave_id": 1,
      "holding_registers_start": 0,
      "supported_functions": [1, 2, 3, 4, 5, 6, 15, 16]
    }
  },
  "current_profile": "generic"
}
```

## How to Use These Configurations

### 1. Core Library (Simple JSON)

Extract just the basic parameters for core library use:

```go
// For core library, create simplified config
jsonConfig := `{
    "slave_id": 1,
    "timeout_ms": 10000,
    "retry_count": 3,
    "retry_delay_ms": 100,
    "connect_timeout_ms": 5000,
    "transport_type": "tcp"
}`

client, err := modbus.NewTCPClientFromJSONString(jsonConfig, "192.168.1.102:502")
```

### 2. Extended Configuration System

Use the full configuration files directly:

```go
import "github.com/adibhanna/modbus-go/config"

// Load full configuration
cfg, err := config.LoadConfig("config-examples/schneider-electric.json")
if err != nil {
    log.Fatal(err)
}

// Create client with full configuration
client := modbus.NewTCPClient(cfg.Connection.GetFullAddress())
client.SetSlaveID(cfg.Modbus.GetSlaveID())
client.SetTimeout(cfg.Connection.GetTimeout())
```

### 3. Configuration Customization

Copy and modify these configurations for your specific devices:

```bash
# Copy a template
cp config-examples/diagnostic.json my-device-config.json

# Edit for your device
# - Change IP address and port
# - Adjust slave ID
# - Modify timeout values
# - Enable/disable specific tests
```

## Configuration Selection Guide

| Use Case                    | Recommended Configuration  | Rationale                               |
| --------------------------- | -------------------------- | --------------------------------------- |
| **Schneider Electric PLCs** | `schneider-electric.json`  | 1-based addressing, longer timeouts     |
| **Siemens PLCs**            | `siemens.json`             | 0-based addressing, efficient settings  |
| **Troubleshooting**         | `diagnostic.json`          | Fast failure detection, verbose logging |
| **General Industrial**      | Inline JSON or custom file | Balanced settings for most devices      |
| **Development/Testing**     | `diagnostic.json`          | Quick feedback, detailed information    |

## Creating Your Own Configuration

### Step 1: Identify Your Device Requirements

- What MODBUS functions does it support?
- Does it use 0-based or 1-based addressing?
- What are typical response times?
- Any device-specific quirks?

### Step 2: Start with a Template

Copy the closest matching configuration:
```bash
cp config-examples/diagnostic.json my-config.json
```

### Step 3: Customize Settings

Adjust these key parameters:
- `address` and `port` - Your device network location
- `slave_id` - Your device's MODBUS address
- `timeout_ms` - Based on device response times
- `retry_count` - Based on network reliability
- `test_addresses` - Valid addresses for your device

### Step 4: Test and Validate

```bash
# Test with diagnostic tool
go run examples/tcp_client_diagnostic.go -config=my-config.json

# Test with example client
go run examples/tcp_client/main.go -config=my-config.json
```

### Step 5: Document Device-Specific Notes

Add a device profile with notes about your configuration:

```json
{
  "device_profiles": {
    "my_device": {
      "slave_id": 1,
      "holding_registers_start": 0,
      "supported_functions": [1, 3, 6, 16],
      "notes": "Custom device configuration - registers 0-99 for data, 100-199 for settings"
    }
  }
}
```

## Contributing New Configurations

If you've created a configuration for a specific device that others might find useful:

1. Create a descriptive filename (e.g., `allen-bradley-micrologix.json`)
2. Include comprehensive device profile information
3. Add documentation comments
4. Test thoroughly with the actual device
5. Submit a pull request with your configuration

See [CONTRIBUTING.md](../CONTRIBUTING.md) for more details on contributing device profiles and configuration enhancements.
