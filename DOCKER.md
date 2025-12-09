# ModbusGo Docker Image

A production-ready MODBUS TCP server and client implementation in Go.

## Quick Start

```bash
# Run the MODBUS server
docker run -d -p 5502:5502 adibhanna/modbus-go:latest

# Or with standard MODBUS port
docker run -d -p 502:502 -p 5502:5502 adibhanna/modbus-go:latest
```

## Tags

| Tag | Description |
|-----|-------------|
| `latest` | Latest stable release |
| `v1.1.0` | Version 1.1.0 - Go 1.25, Docker support |
| `v1.0.0` | Version 1.0.0 - Initial release |

## Features

- **Complete MODBUS Protocol** - All 19 standard function codes
- **Multiple Transports** - TCP, TLS, RTU over TCP, UDP
- **High-Level Data Types** - Float32, Float64, Uint32, Uint64, strings
- **Configurable Endianness** - Big/Little endian byte and word order
- **Production Ready** - Auto-reconnect, graceful shutdown, thread-safe

## Exposed Ports

| Port | Description |
|------|-------------|
| 502 | Standard MODBUS TCP port |
| 5502 | Alternative MODBUS TCP port |

## Available Binaries

The image includes these pre-built binaries:

| Binary | Description |
|--------|-------------|
| `/app/bin/advanced_server` | Full-featured MODBUS server (default) |
| `/app/bin/tcp_server` | Simple TCP server |
| `/app/bin/tcp_client` | TCP client example |
| `/app/bin/integration_test` | Integration test suite |
| `/app/bin/config_showcase` | Configuration examples |

## Usage Examples

### Run Advanced Server (Default)

```bash
docker run -d \
  --name modbus-server \
  -p 5502:5502 \
  adibhanna/modbus-go:latest
```

### Run Simple Server

```bash
docker run -d \
  --name modbus-simple \
  -p 5502:5502 \
  adibhanna/modbus-go:latest \
  /app/bin/tcp_server
```

### Run with Custom Configuration

```bash
docker run -d \
  --name modbus-server \
  -p 5502:5502 \
  -v /path/to/config.json:/app/config.json \
  adibhanna/modbus-go:latest
```

### Docker Compose

```yaml
version: '3.8'
services:
  modbus-server:
    image: adibhanna/modbus-go:latest
    ports:
      - "5502:5502"
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "nc", "-z", "localhost", "5502"]
      interval: 30s
      timeout: 10s
      retries: 3
```

## Environment Variables

The container runs as non-root user `modbus` (UID 1000) for security.

## Health Check

The image includes a built-in health check:

```bash
nc -z localhost 5502
```

## Source Code

- **GitHub**: https://github.com/adibhanna/modbus-go
- **Documentation**: https://github.com/adibhanna/modbus-go/blob/main/docs/DOCUMENTATION.md

## License

MIT License - see [LICENSE](https://github.com/adibhanna/modbus-go/blob/main/LICENSE)
