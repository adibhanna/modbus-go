package transport

import (
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/adibhanna/modbus-go/modbus"
	"github.com/adibhanna/modbus-go/pdu"
	"go.bug.st/serial"
)

// SerialConfig holds serial port configuration
type SerialConfig struct {
	Port     string
	BaudRate int
	DataBits int
	StopBits serial.StopBits
	Parity   serial.Parity
	Timeout  time.Duration
}

// NewSerialConfig creates a new serial configuration
func NewSerialConfig(port string, baudRate int, dataBits int, stopBits int, parity string) (*SerialConfig, error) {
	var sb serial.StopBits
	switch stopBits {
	case 1:
		sb = serial.OneStopBit
	case 2:
		sb = serial.TwoStopBits
	default:
		return nil, fmt.Errorf("invalid stop bits: %d (must be 1 or 2)", stopBits)
	}

	var p serial.Parity
	switch strings.ToUpper(parity) {
	case "N", "NONE":
		p = serial.NoParity
	case "E", "EVEN":
		p = serial.EvenParity
	case "O", "ODD":
		p = serial.OddParity
	default:
		return nil, fmt.Errorf("invalid parity: %s (must be N, E, or O)", parity)
	}

	return &SerialConfig{
		Port:     port,
		BaudRate: baudRate,
		DataBits: dataBits,
		StopBits: sb,
		Parity:   p,
		Timeout:  time.Duration(modbus.DefaultResponseTimeout) * time.Millisecond,
	}, nil
}

// RTUTransport implements MODBUS RTU over serial transport
type RTUTransport struct {
	config    *SerialConfig
	port      serial.Port
	connected bool
	mutex     sync.Mutex
}

// NewRTUTransport creates a new RTU transport
func NewRTUTransport(config *SerialConfig) *RTUTransport {
	return &RTUTransport{
		config: config,
	}
}

// Connect opens the serial port
func (t *RTUTransport) Connect() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.connected {
		return nil
	}

	mode := &serial.Mode{
		BaudRate: t.config.BaudRate,
		DataBits: t.config.DataBits,
		Parity:   t.config.Parity,
		StopBits: t.config.StopBits,
	}

	port, err := serial.Open(t.config.Port, mode)
	if err != nil {
		return fmt.Errorf("failed to open serial port %s: %w", t.config.Port, err)
	}

	// Set read timeout
	if err := port.SetReadTimeout(t.config.Timeout); err != nil {
		_ = port.Close()
		return fmt.Errorf("failed to set read timeout: %w", err)
	}

	t.port = port
	t.connected = true
	return nil
}

// Close closes the serial port
func (t *RTUTransport) Close() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if !t.connected || t.port == nil {
		return nil
	}

	err := t.port.Close()
	t.port = nil
	t.connected = false
	return err
}

// IsConnected returns true if the transport is connected
func (t *RTUTransport) IsConnected() bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return t.connected
}

// SetTimeout sets the response timeout
func (t *RTUTransport) SetTimeout(timeout time.Duration) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.config.Timeout = timeout
	if t.connected && t.port != nil {
		_ = t.port.SetReadTimeout(timeout)
	}
}

// GetTimeout returns the current timeout
func (t *RTUTransport) GetTimeout() time.Duration {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return t.config.Timeout
}

// SendRequest sends a request PDU and returns the response PDU
func (t *RTUTransport) SendRequest(slaveID modbus.SlaveID, request *pdu.Request) (*pdu.Response, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if !t.connected {
		return nil, fmt.Errorf("transport not connected")
	}

	// Create RTU ADU: SlaveID + PDU + CRC
	pduBytes := request.Bytes()
	adu := make([]byte, 1+len(pduBytes)+2)
	adu[0] = byte(slaveID)
	copy(adu[1:1+len(pduBytes)], pduBytes)

	// Calculate and append CRC
	crc := calculateCRC16(adu[:1+len(pduBytes)])
	adu[1+len(pduBytes)] = byte(crc)
	adu[1+len(pduBytes)+1] = byte(crc >> 8)

	// Send request
	if _, err := t.port.Write(adu); err != nil {
		return nil, fmt.Errorf("failed to write RTU request: %w", err)
	}

	// Calculate inter-character timeout for RTU
	// RTU requires 3.5 character times of silence between frames
	charTime := calculateCharacterTime(t.config.BaudRate, t.config.DataBits, int(t.config.StopBits), t.config.Parity)
	interCharTimeout := time.Duration(float64(charTime) * 1.5) // 1.5 character times for inter-character
	frameTimeout := time.Duration(float64(charTime) * 3.5)     // 3.5 character times for end-of-frame

	// Receive response
	var response []byte
	buf := make([]byte, 256)
	lastReceiveTime := time.Now()

	for {
		// Set short timeout for individual reads
		_ = t.port.SetReadTimeout(interCharTimeout)

		n, err := t.port.Read(buf)
		if err != nil {
			// Check if this is a timeout and we have some data
			if len(response) > 0 && time.Since(lastReceiveTime) >= frameTimeout {
				break // End of frame detected
			}
			return nil, fmt.Errorf("failed to read RTU response: %w", err)
		}

		if n > 0 {
			response = append(response, buf[:n]...)
			lastReceiveTime = time.Now()
		}

		// Check for minimum response length (SlaveID + FunctionCode + CRC)
		if len(response) >= 4 {
			// Check if we have a complete response
			if time.Since(lastReceiveTime) >= frameTimeout {
				break
			}
		}

		// Overall timeout check
		if time.Since(lastReceiveTime) > t.config.Timeout {
			return nil, fmt.Errorf("response timeout")
		}
	}

	return t.parseRTUResponse(response, slaveID)
}

// parseRTUResponse parses an RTU response
func (t *RTUTransport) parseRTUResponse(data []byte, expectedSlaveID modbus.SlaveID) (*pdu.Response, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("RTU response too short: need at least 4 bytes, got %d", len(data))
	}

	// Extract components
	receivedSlaveID := modbus.SlaveID(data[0])
	pduData := data[1 : len(data)-2]
	receivedCRC := uint16(data[len(data)-2]) | (uint16(data[len(data)-1]) << 8)

	// Validate slave ID
	if receivedSlaveID != expectedSlaveID {
		return nil, fmt.Errorf("slave ID mismatch: expected %d, got %d", expectedSlaveID, receivedSlaveID)
	}

	// Validate CRC
	calculatedCRC := calculateCRC16(data[:len(data)-2])
	if receivedCRC != calculatedCRC {
		return nil, fmt.Errorf("CRC mismatch: expected %04X, got %04X", calculatedCRC, receivedCRC)
	}

	// Parse PDU
	responsePDU, err := pdu.ParsePDU(pduData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RTU response PDU: %w", err)
	}

	return &pdu.Response{PDU: responsePDU}, nil
}

// GetTransportType returns the transport type
func (t *RTUTransport) GetTransportType() modbus.TransportType {
	return modbus.TransportRTU
}

// String returns a string representation of the transport
func (t *RTUTransport) String() string {
	return fmt.Sprintf("RTU(%s@%d)", t.config.Port, t.config.BaudRate)
}

// ASCIITransport implements MODBUS ASCII over serial transport
type ASCIITransport struct {
	config    *SerialConfig
	port      serial.Port
	connected bool
	mutex     sync.Mutex
}

// NewASCIITransport creates a new ASCII transport
func NewASCIITransport(config *SerialConfig) *ASCIITransport {
	return &ASCIITransport{
		config: config,
	}
}

// Connect opens the serial port
func (t *ASCIITransport) Connect() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.connected {
		return nil
	}

	mode := &serial.Mode{
		BaudRate: t.config.BaudRate,
		DataBits: t.config.DataBits,
		Parity:   t.config.Parity,
		StopBits: t.config.StopBits,
	}

	port, err := serial.Open(t.config.Port, mode)
	if err != nil {
		return fmt.Errorf("failed to open serial port %s: %w", t.config.Port, err)
	}

	if err := port.SetReadTimeout(t.config.Timeout); err != nil {
		_ = port.Close()
		return fmt.Errorf("failed to set read timeout: %w", err)
	}

	t.port = port
	t.connected = true
	return nil
}

// Close closes the serial port
func (t *ASCIITransport) Close() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if !t.connected || t.port == nil {
		return nil
	}

	err := t.port.Close()
	t.port = nil
	t.connected = false
	return err
}

// IsConnected returns true if the transport is connected
func (t *ASCIITransport) IsConnected() bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return t.connected
}

// SetTimeout sets the response timeout
func (t *ASCIITransport) SetTimeout(timeout time.Duration) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.config.Timeout = timeout
	if t.connected && t.port != nil {
		_ = t.port.SetReadTimeout(timeout)
	}
}

// GetTimeout returns the current timeout
func (t *ASCIITransport) GetTimeout() time.Duration {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return t.config.Timeout
}

// SendRequest sends a request PDU and returns the response PDU
func (t *ASCIITransport) SendRequest(slaveID modbus.SlaveID, request *pdu.Request) (*pdu.Response, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if !t.connected {
		return nil, fmt.Errorf("transport not connected")
	}

	// Create ASCII frame: : + SlaveID + PDU + LRC + CRLF
	pduBytes := request.Bytes()
	dataBytes := make([]byte, 1+len(pduBytes))
	dataBytes[0] = byte(slaveID)
	copy(dataBytes[1:], pduBytes)

	// Calculate LRC
	lrc := calculateLRC(dataBytes)
	dataBytes = append(dataBytes, lrc)

	// Convert to ASCII hex
	asciiData := strings.ToUpper(hex.EncodeToString(dataBytes))
	frame := ":" + asciiData + "\r\n"

	// Send request
	if _, err := t.port.Write([]byte(frame)); err != nil {
		return nil, fmt.Errorf("failed to write ASCII request: %w", err)
	}

	// Receive response
	response, err := t.readASCIIFrame()
	if err != nil {
		return nil, fmt.Errorf("failed to read ASCII response: %w", err)
	}

	return t.parseASCIIResponse(response, slaveID)
}

// readASCIIFrame reads a complete ASCII frame
func (t *ASCIITransport) readASCIIFrame() ([]byte, error) {
	var frame []byte
	buf := make([]byte, 1)

	// Look for start character ':'
	for {
		n, err := t.port.Read(buf)
		if err != nil {
			return nil, fmt.Errorf("failed to read start character: %w", err)
		}
		if n > 0 && buf[0] == ':' {
			break
		}
	}

	// Read until CRLF
	for {
		n, err := t.port.Read(buf)
		if err != nil {
			return nil, fmt.Errorf("failed to read frame data: %w", err)
		}
		if n > 0 {
			frame = append(frame, buf[0])
			if len(frame) >= 2 && frame[len(frame)-2] == '\r' && frame[len(frame)-1] == '\n' {
				break
			}
		}
	}

	// Remove CRLF
	return frame[:len(frame)-2], nil
}

// parseASCIIResponse parses an ASCII response
func (t *ASCIITransport) parseASCIIResponse(asciiData []byte, expectedSlaveID modbus.SlaveID) (*pdu.Response, error) {
	// Convert from ASCII hex to binary
	if len(asciiData)%2 != 0 {
		return nil, fmt.Errorf("invalid ASCII frame length: %d", len(asciiData))
	}

	data, err := hex.DecodeString(string(asciiData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode ASCII hex: %w", err)
	}

	if len(data) < 3 { // SlaveID + FunctionCode + LRC minimum
		return nil, fmt.Errorf("ASCII response too short: need at least 3 bytes, got %d", len(data))
	}

	// Extract components
	receivedSlaveID := modbus.SlaveID(data[0])
	pduData := data[1 : len(data)-1]
	receivedLRC := data[len(data)-1]

	// Validate slave ID
	if receivedSlaveID != expectedSlaveID {
		return nil, fmt.Errorf("slave ID mismatch: expected %d, got %d", expectedSlaveID, receivedSlaveID)
	}

	// Validate LRC
	calculatedLRC := calculateLRC(data[:len(data)-1])
	if receivedLRC != calculatedLRC {
		return nil, fmt.Errorf("LRC mismatch: expected %02X, got %02X", calculatedLRC, receivedLRC)
	}

	// Parse PDU
	responsePDU, err := pdu.ParsePDU(pduData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ASCII response PDU: %w", err)
	}

	return &pdu.Response{PDU: responsePDU}, nil
}

// GetTransportType returns the transport type
func (t *ASCIITransport) GetTransportType() modbus.TransportType {
	return modbus.TransportASCII
}

// String returns a string representation of the transport
func (t *ASCIITransport) String() string {
	return fmt.Sprintf("ASCII(%s@%d)", t.config.Port, t.config.BaudRate)
}

// Helper functions

// calculateCRC16 calculates MODBUS CRC-16
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

// calculateLRC calculates MODBUS LRC (Longitudinal Redundancy Check)
func calculateLRC(data []byte) uint8 {
	lrc := uint8(0)
	for _, b := range data {
		lrc += b
	}
	return uint8(-int8(lrc))
}

// calculateCharacterTime calculates the time for one character transmission
func calculateCharacterTime(baudRate int, dataBits int, stopBits int, parity serial.Parity) time.Duration {
	// Start bit (1) + data bits + parity bit (if any) + stop bits
	bitsPerChar := 1 + dataBits + stopBits
	if parity != serial.NoParity {
		bitsPerChar++
	}

	// Time per bit in nanoseconds
	nsPerBit := int64(1_000_000_000) / int64(baudRate)

	// Total time per character
	return time.Duration(int64(bitsPerChar) * nsPerBit)
}
