package transport

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/adibhanna/modbusgo/modbus"
	"github.com/adibhanna/modbusgo/pdu"
)

// MBAP header structure for MODBUS TCP/IP
type MBAPHeader struct {
	TransactionID uint16
	ProtocolID    uint16 // Always 0x0000 for MODBUS
	Length        uint16 // Length of following bytes (unit ID + PDU)
	UnitID        uint8  // Slave/Unit ID
}

// EncodeMBAP encodes an MBAP header to bytes
func (h *MBAPHeader) EncodeMBAP() []byte {
	buf := make([]byte, modbus.MBAPHeaderSize)
	binary.BigEndian.PutUint16(buf[0:2], h.TransactionID)
	binary.BigEndian.PutUint16(buf[2:4], h.ProtocolID)
	binary.BigEndian.PutUint16(buf[4:6], h.Length)
	buf[6] = h.UnitID
	return buf
}

// DecodeMBAP decodes bytes to an MBAP header
func DecodeMBAP(data []byte) (*MBAPHeader, error) {
	if len(data) < modbus.MBAPHeaderSize {
		return nil, fmt.Errorf("insufficient data for MBAP header: need %d bytes, got %d",
			modbus.MBAPHeaderSize, len(data))
	}

	return &MBAPHeader{
		TransactionID: binary.BigEndian.Uint16(data[0:2]),
		ProtocolID:    binary.BigEndian.Uint16(data[2:4]),
		Length:        binary.BigEndian.Uint16(data[4:6]),
		UnitID:        data[6],
	}, nil
}

// TCPTransport implements MODBUS TCP/IP transport
type TCPTransport struct {
	conn          net.Conn
	transactionID uint16
	timeout       time.Duration
	mutex         sync.Mutex
	address       string
	connected     bool
}

// NewTCPTransport creates a new TCP transport
func NewTCPTransport(address string) *TCPTransport {
	return &TCPTransport{
		address:       address,
		timeout:       time.Duration(modbus.DefaultResponseTimeout) * time.Millisecond,
		transactionID: 1,
	}
}

// Connect establishes a TCP connection
func (t *TCPTransport) Connect() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.connected {
		return nil
	}

	dialer := &net.Dialer{
		Timeout: time.Duration(modbus.DefaultConnectTimeout) * time.Millisecond,
	}

	conn, err := dialer.Dial("tcp", t.address)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", t.address, err)
	}

	t.conn = conn
	t.connected = true
	return nil
}

// Close closes the TCP connection
func (t *TCPTransport) Close() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if !t.connected || t.conn == nil {
		return nil
	}

	err := t.conn.Close()
	t.conn = nil
	t.connected = false
	return err
}

// IsConnected returns true if the transport is connected
func (t *TCPTransport) IsConnected() bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return t.connected
}

// SetTimeout sets the response timeout
func (t *TCPTransport) SetTimeout(timeout time.Duration) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.timeout = timeout
}

// GetTimeout returns the current timeout
func (t *TCPTransport) GetTimeout() time.Duration {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return t.timeout
}

// SendRequest sends a request PDU and returns the response PDU
func (t *TCPTransport) SendRequest(slaveID modbus.SlaveID, request *pdu.Request) (*pdu.Response, error) {
	if !t.IsConnected() {
		return nil, fmt.Errorf("transport not connected")
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	// Get next transaction ID
	txID := t.transactionID
	t.transactionID++
	if t.transactionID == 0 {
		t.transactionID = 1
	}

	// Create MBAP header
	pduBytes := request.Bytes()
	header := &MBAPHeader{
		TransactionID: txID,
		ProtocolID:    modbus.MBAPProtocolID,
		Length:        uint16(1 + len(pduBytes)), // UnitID + PDU
		UnitID:        uint8(slaveID),
	}

	// Send request
	if err := t.sendADU(header, pduBytes); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Receive response
	responseHeader, responsePDU, err := t.receiveADU()
	if err != nil {
		return nil, fmt.Errorf("failed to receive response: %w", err)
	}

	// Validate response
	if responseHeader.TransactionID != txID {
		return nil, fmt.Errorf("transaction ID mismatch: expected %d, got %d",
			txID, responseHeader.TransactionID)
	}

	if responseHeader.ProtocolID != modbus.MBAPProtocolID {
		return nil, fmt.Errorf("protocol ID mismatch: expected %d, got %d",
			modbus.MBAPProtocolID, responseHeader.ProtocolID)
	}

	if responseHeader.UnitID != uint8(slaveID) {
		return nil, fmt.Errorf("unit ID mismatch: expected %d, got %d",
			slaveID, responseHeader.UnitID)
	}

	return &pdu.Response{PDU: responsePDU}, nil
}

// sendADU sends an Application Data Unit (MBAP + PDU)
func (t *TCPTransport) sendADU(header *MBAPHeader, pduBytes []byte) error {
	// Set write timeout
	if err := t.conn.SetWriteDeadline(time.Now().Add(t.timeout)); err != nil {
		return fmt.Errorf("failed to set write deadline: %w", err)
	}

	// Send MBAP header
	mbapBytes := header.EncodeMBAP()
	if _, err := t.conn.Write(mbapBytes); err != nil {
		return fmt.Errorf("failed to write MBAP header: %w", err)
	}

	// Send PDU
	if _, err := t.conn.Write(pduBytes); err != nil {
		return fmt.Errorf("failed to write PDU: %w", err)
	}

	return nil
}

// receiveADU receives an Application Data Unit (MBAP + PDU)
func (t *TCPTransport) receiveADU() (*MBAPHeader, *pdu.PDU, error) {
	// Set read timeout
	if err := t.conn.SetReadDeadline(time.Now().Add(t.timeout)); err != nil {
		return nil, nil, fmt.Errorf("failed to set read deadline: %w", err)
	}

	// Read MBAP header
	headerBytes := make([]byte, modbus.MBAPHeaderSize)
	if _, err := io.ReadFull(t.conn, headerBytes); err != nil {
		return nil, nil, fmt.Errorf("failed to read MBAP header: %w", err)
	}

	header, err := DecodeMBAP(headerBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode MBAP header: %w", err)
	}

	// Validate length
	if header.Length < 2 { // At least UnitID + function code
		return nil, nil, fmt.Errorf("invalid MBAP length: %d", header.Length)
	}

	if header.Length > modbus.MaxPDUSize+1 { // UnitID + max PDU size
		return nil, nil, fmt.Errorf("MBAP length too large: %d", header.Length)
	}

	// Read PDU (length includes UnitID which we already have in header)
	pduBytes := make([]byte, header.Length-1)
	if _, err := io.ReadFull(t.conn, pduBytes); err != nil {
		return nil, nil, fmt.Errorf("failed to read PDU: %w", err)
	}

	responsePDU, err := pdu.ParsePDU(pduBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse PDU: %w", err)
	}

	return header, responsePDU, nil
}

// GetTransportType returns the transport type
func (t *TCPTransport) GetTransportType() modbus.TransportType {
	return modbus.TransportTCP
}

// String returns a string representation of the transport
func (t *TCPTransport) String() string {
	return fmt.Sprintf("TCP(%s)", t.address)
}

// TCPServer implements a MODBUS TCP server
type TCPServer struct {
	listener    net.Listener
	address     string
	handler     RequestHandler
	connections map[net.Conn]bool
	mutex       sync.RWMutex
	running     bool
	stopChan    chan struct{}
}

// RequestHandler defines the interface for handling MODBUS requests
type RequestHandler interface {
	HandleRequest(slaveID modbus.SlaveID, req *pdu.Request) *pdu.Response
}

// NewTCPServer creates a new TCP server
func NewTCPServer(address string, handler RequestHandler) *TCPServer {
	return &TCPServer{
		address:     address,
		handler:     handler,
		connections: make(map[net.Conn]bool),
		stopChan:    make(chan struct{}),
	}
}

// Start starts the TCP server
func (s *TCPServer) Start() error {
	// Start listening
	lc := net.ListenConfig{}
	listener, err := lc.Listen(context.Background(), "tcp", s.address)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.address, err)
	}

	s.listener = listener
	s.running = true

	go s.acceptLoop()

	return nil
}

// Stop stops the TCP server
func (s *TCPServer) Stop() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.running {
		return nil
	}

	close(s.stopChan)
	s.running = false

	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			// Log error but don't fail stop
			fmt.Printf("Warning: error closing listener: %v\n", err)
		}
	}

	// Close all active connections
	for conn := range s.connections {
		_ = conn.Close() // Best effort close, ignore errors
	}
	s.connections = make(map[net.Conn]bool)

	return nil
}

// IsRunning returns true if the server is running
func (s *TCPServer) IsRunning() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.running
}

// acceptLoop accepts incoming connections
func (s *TCPServer) acceptLoop() {
	for {
		select {
		case <-s.stopChan:
			return
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				if s.IsRunning() {
					// Log error if server is still supposed to be running
					fmt.Printf("TCP server accept error: %v\n", err)
				}
				continue
			}

			s.mutex.Lock()
			s.connections[conn] = true
			s.mutex.Unlock()

			go s.handleConnection(conn)
		}
	}
}

// handleConnection handles a single connection
func (s *TCPServer) handleConnection(conn net.Conn) {
	defer func() {
		_ = conn.Close() // Best effort close, ignore errors
		s.mutex.Lock()
		delete(s.connections, conn)
		s.mutex.Unlock()
	}()

	transport := &TCPTransport{
		conn:      conn,
		connected: true,
		timeout:   time.Duration(modbus.DefaultResponseTimeout) * time.Millisecond,
	}

	for {
		select {
		case <-s.stopChan:
			return
		default:
			// Receive request
			header, requestPDU, err := transport.receiveADU()
			if err != nil {
				if s.IsRunning() {
					// Log error if server is still running
					fmt.Printf("TCP server receive error: %v\n", err)
				}
				return
			}

			// Handle request
			request := &pdu.Request{PDU: requestPDU}
			response := s.handler.HandleRequest(modbus.SlaveID(header.UnitID), request)

			// Send response
			responseHeader := &MBAPHeader{
				TransactionID: header.TransactionID,
				ProtocolID:    modbus.MBAPProtocolID,
				Length:        uint16(1 + response.Size()), // UnitID + PDU
				UnitID:        header.UnitID,
			}

			if err := transport.sendADU(responseHeader, response.Bytes()); err != nil {
				if s.IsRunning() {
					fmt.Printf("TCP server send error: %v\n", err)
				}
				return
			}
		}
	}
}
