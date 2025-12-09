package transport

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/adibhanna/modbus-go/modbus"
	"github.com/adibhanna/modbus-go/pdu"
)

// Logger interface for custom logging
type Logger interface {
	Printf(format string, v ...interface{})
}

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
	conn           net.Conn
	transactionID  uint16
	timeout        time.Duration
	idleTimeout    time.Duration
	connectTimeout time.Duration
	mutex          sync.Mutex
	address        string
	connected      bool
	tlsConfig      *tls.Config
	logger         Logger
	lastActivity   time.Time
}

// TCPTransportConfig holds configuration for TCP transport
type TCPTransportConfig struct {
	Address        string
	Timeout        time.Duration
	IdleTimeout    time.Duration
	ConnectTimeout time.Duration
	TLSConfig      *tls.Config
	Logger         Logger
}

// NewTCPTransport creates a new TCP transport
func NewTCPTransport(address string) *TCPTransport {
	return &TCPTransport{
		address:        address,
		timeout:        time.Duration(modbus.DefaultResponseTimeout) * time.Millisecond,
		connectTimeout: time.Duration(modbus.DefaultConnectTimeout) * time.Millisecond,
		idleTimeout:    60 * time.Second,
		transactionID:  1,
	}
}

// NewTCPTransportWithConfig creates a new TCP transport with full configuration
func NewTCPTransportWithConfig(config TCPTransportConfig) *TCPTransport {
	t := &TCPTransport{
		address:        config.Address,
		timeout:        config.Timeout,
		idleTimeout:    config.IdleTimeout,
		connectTimeout: config.ConnectTimeout,
		tlsConfig:      config.TLSConfig,
		logger:         config.Logger,
		transactionID:  1,
	}

	if t.timeout == 0 {
		t.timeout = time.Duration(modbus.DefaultResponseTimeout) * time.Millisecond
	}
	if t.connectTimeout == 0 {
		t.connectTimeout = time.Duration(modbus.DefaultConnectTimeout) * time.Millisecond
	}
	if t.idleTimeout == 0 {
		t.idleTimeout = 60 * time.Second
	}

	return t
}

// NewTLSTransport creates a new TCP transport with TLS encryption
func NewTLSTransport(address string, tlsConfig *tls.Config) *TCPTransport {
	return &TCPTransport{
		address:        address,
		timeout:        time.Duration(modbus.DefaultResponseTimeout) * time.Millisecond,
		connectTimeout: time.Duration(modbus.DefaultConnectTimeout) * time.Millisecond,
		idleTimeout:    60 * time.Second,
		transactionID:  1,
		tlsConfig:      tlsConfig,
	}
}

// SetLogger sets a custom logger for the transport
func (t *TCPTransport) SetLogger(logger Logger) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.logger = logger
}

// SetIdleTimeout sets the idle timeout for the connection
func (t *TCPTransport) SetIdleTimeout(timeout time.Duration) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.idleTimeout = timeout
}

// GetIdleTimeout returns the current idle timeout
func (t *TCPTransport) GetIdleTimeout() time.Duration {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return t.idleTimeout
}

// SetConnectTimeout sets the connection timeout
func (t *TCPTransport) SetConnectTimeout(timeout time.Duration) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.connectTimeout = timeout
}

// GetConnectTimeout returns the current connection timeout
func (t *TCPTransport) GetConnectTimeout() time.Duration {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return t.connectTimeout
}

func (t *TCPTransport) logf(format string, v ...interface{}) {
	if t.logger != nil {
		t.logger.Printf(format, v...)
	}
}

// Connect establishes a TCP connection (with optional TLS)
func (t *TCPTransport) Connect() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.connected {
		return nil
	}

	var conn net.Conn
	var err error

	dialer := &net.Dialer{
		Timeout: t.connectTimeout,
	}

	if t.tlsConfig != nil {
		// TLS connection
		t.logf("Connecting to %s with TLS", t.address)
		tlsDialer := &tls.Dialer{
			NetDialer: dialer,
			Config:    t.tlsConfig,
		}
		conn, err = tlsDialer.Dial("tcp", t.address)
	} else {
		// Plain TCP connection
		t.logf("Connecting to %s", t.address)
		conn, err = dialer.Dial("tcp", t.address)
	}

	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", t.address, err)
	}

	t.conn = conn
	t.connected = true
	t.lastActivity = time.Now()
	t.logf("Connected to %s", t.address)
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
	if _, readErr := io.ReadFull(t.conn, pduBytes); readErr != nil {
		return nil, nil, fmt.Errorf("failed to read PDU: %w", readErr)
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
	if t.tlsConfig != nil {
		return fmt.Sprintf("TCP+TLS(%s)", t.address)
	}
	return fmt.Sprintf("TCP(%s)", t.address)
}

// RTUOverTCPTransport implements RTU framing over TCP/IP
// This is used for serial-to-Ethernet converters and remote serial devices
type RTUOverTCPTransport struct {
	conn           net.Conn
	timeout        time.Duration
	idleTimeout    time.Duration
	connectTimeout time.Duration
	mutex          sync.Mutex
	address        string
	connected      bool
	logger         Logger
	lastActivity   time.Time
}

// NewRTUOverTCPTransport creates a new RTU over TCP transport
func NewRTUOverTCPTransport(address string) *RTUOverTCPTransport {
	return &RTUOverTCPTransport{
		address:        address,
		timeout:        time.Duration(modbus.DefaultResponseTimeout) * time.Millisecond,
		connectTimeout: time.Duration(modbus.DefaultConnectTimeout) * time.Millisecond,
		idleTimeout:    60 * time.Second,
	}
}

// SetLogger sets a custom logger for the transport
func (t *RTUOverTCPTransport) SetLogger(logger Logger) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.logger = logger
}

func (t *RTUOverTCPTransport) logf(format string, v ...interface{}) {
	if t.logger != nil {
		t.logger.Printf(format, v...)
	}
}

// Connect establishes a TCP connection for RTU framing
func (t *RTUOverTCPTransport) Connect() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.connected {
		return nil
	}

	dialer := &net.Dialer{
		Timeout: t.connectTimeout,
	}

	t.logf("Connecting RTU over TCP to %s", t.address)
	conn, err := dialer.Dial("tcp", t.address)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", t.address, err)
	}

	t.conn = conn
	t.connected = true
	t.lastActivity = time.Now()
	t.logf("Connected to %s", t.address)
	return nil
}

// Close closes the connection
func (t *RTUOverTCPTransport) Close() error {
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

// IsConnected returns true if connected
func (t *RTUOverTCPTransport) IsConnected() bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return t.connected
}

// SetTimeout sets the response timeout
func (t *RTUOverTCPTransport) SetTimeout(timeout time.Duration) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.timeout = timeout
}

// GetTimeout returns the current timeout
func (t *RTUOverTCPTransport) GetTimeout() time.Duration {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return t.timeout
}

// SendRequest sends an RTU framed request over TCP
func (t *RTUOverTCPTransport) SendRequest(slaveID modbus.SlaveID, request *pdu.Request) (*pdu.Response, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if !t.connected {
		return nil, fmt.Errorf("transport not connected")
	}

	// Build RTU frame: SlaveID + PDU + CRC
	pduBytes := request.Bytes()
	frame := make([]byte, 1+len(pduBytes)+2)
	frame[0] = uint8(slaveID)
	copy(frame[1:], pduBytes)

	// Calculate and append CRC
	crc := calculateCRC16(frame[:len(frame)-2])
	frame[len(frame)-2] = byte(crc)
	frame[len(frame)-1] = byte(crc >> 8)

	// Set deadline
	if err := t.conn.SetDeadline(time.Now().Add(t.timeout)); err != nil {
		return nil, fmt.Errorf("failed to set deadline: %w", err)
	}

	t.logf("TX: % X", frame)

	// Send frame
	if _, err := t.conn.Write(frame); err != nil {
		return nil, fmt.Errorf("failed to send RTU frame: %w", err)
	}

	t.lastActivity = time.Now()

	// Read response
	response := make([]byte, 256)
	n, err := t.conn.Read(response)
	if err != nil {
		return nil, fmt.Errorf("failed to read RTU response: %w", err)
	}

	if n < 4 {
		return nil, fmt.Errorf("RTU response too short: %d bytes", n)
	}

	t.logf("RX: % X", response[:n])

	// Verify CRC
	respCRC := uint16(response[n-2]) | uint16(response[n-1])<<8
	calcCRC := calculateCRC16(response[:n-2])
	if respCRC != calcCRC {
		return nil, fmt.Errorf("CRC mismatch: expected 0x%04X, got 0x%04X", calcCRC, respCRC)
	}

	// Verify slave ID
	if response[0] != uint8(slaveID) {
		return nil, fmt.Errorf("slave ID mismatch: expected %d, got %d", slaveID, response[0])
	}

	// Parse PDU (skip slave ID, exclude CRC)
	responsePDU, err := pdu.ParsePDU(response[1 : n-2])
	if err != nil {
		return nil, fmt.Errorf("failed to parse response PDU: %w", err)
	}

	return &pdu.Response{PDU: responsePDU}, nil
}

// GetTransportType returns the transport type
func (t *RTUOverTCPTransport) GetTransportType() modbus.TransportType {
	return modbus.TransportRTU
}

// String returns a string representation
func (t *RTUOverTCPTransport) String() string {
	return fmt.Sprintf("RTU-over-TCP(%s)", t.address)
}

// UDPTransport implements MODBUS over UDP
type UDPTransport struct {
	conn          *net.UDPConn
	remoteAddr    *net.UDPAddr
	transactionID uint16
	timeout       time.Duration
	mutex         sync.Mutex
	address       string
	connected     bool
	logger        Logger
}

// NewUDPTransport creates a new UDP transport
func NewUDPTransport(address string) *UDPTransport {
	return &UDPTransport{
		address:       address,
		timeout:       time.Duration(modbus.DefaultResponseTimeout) * time.Millisecond,
		transactionID: 1,
	}
}

// SetLogger sets a custom logger
func (t *UDPTransport) SetLogger(logger Logger) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.logger = logger
}

func (t *UDPTransport) logf(format string, v ...interface{}) {
	if t.logger != nil {
		t.logger.Printf(format, v...)
	}
}

// Connect resolves the remote address and creates a UDP connection
func (t *UDPTransport) Connect() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.connected {
		return nil
	}

	remoteAddr, err := net.ResolveUDPAddr("udp", t.address)
	if err != nil {
		return fmt.Errorf("failed to resolve UDP address %s: %w", t.address, err)
	}

	conn, err := net.DialUDP("udp", nil, remoteAddr)
	if err != nil {
		return fmt.Errorf("failed to create UDP connection: %w", err)
	}

	t.conn = conn
	t.remoteAddr = remoteAddr
	t.connected = true
	t.logf("UDP connected to %s", t.address)
	return nil
}

// Close closes the UDP connection
func (t *UDPTransport) Close() error {
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

// IsConnected returns true if connected
func (t *UDPTransport) IsConnected() bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return t.connected
}

// SetTimeout sets the response timeout
func (t *UDPTransport) SetTimeout(timeout time.Duration) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.timeout = timeout
}

// GetTimeout returns the current timeout
func (t *UDPTransport) GetTimeout() time.Duration {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return t.timeout
}

// SendRequest sends a MODBUS request over UDP using MBAP framing
func (t *UDPTransport) SendRequest(slaveID modbus.SlaveID, request *pdu.Request) (*pdu.Response, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if !t.connected {
		return nil, fmt.Errorf("transport not connected")
	}

	// Increment transaction ID
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
		Length:        uint16(1 + len(pduBytes)),
		UnitID:        uint8(slaveID),
	}

	// Build complete ADU
	headerBytes := header.EncodeMBAP()
	adu := append(headerBytes, pduBytes...)

	// Set deadline
	if err := t.conn.SetDeadline(time.Now().Add(t.timeout)); err != nil {
		return nil, fmt.Errorf("failed to set deadline: %w", err)
	}

	t.logf("TX UDP: % X", adu)

	// Send request
	if _, err := t.conn.Write(adu); err != nil {
		return nil, fmt.Errorf("failed to send UDP request: %w", err)
	}

	// Receive response
	response := make([]byte, modbus.MaxTCPADUSize)
	n, err := t.conn.Read(response)
	if err != nil {
		return nil, fmt.Errorf("failed to receive UDP response: %w", err)
	}

	if n < modbus.MBAPHeaderSize+1 {
		return nil, fmt.Errorf("UDP response too short: %d bytes", n)
	}

	t.logf("RX UDP: % X", response[:n])

	// Parse MBAP header
	respHeader, err := DecodeMBAP(response[:modbus.MBAPHeaderSize])
	if err != nil {
		return nil, fmt.Errorf("failed to decode MBAP header: %w", err)
	}

	// Validate response
	if respHeader.TransactionID != txID {
		return nil, fmt.Errorf("transaction ID mismatch: expected %d, got %d",
			txID, respHeader.TransactionID)
	}

	// Parse PDU
	responsePDU, err := pdu.ParsePDU(response[modbus.MBAPHeaderSize:n])
	if err != nil {
		return nil, fmt.Errorf("failed to parse response PDU: %w", err)
	}

	return &pdu.Response{PDU: responsePDU}, nil
}

// GetTransportType returns the transport type
func (t *UDPTransport) GetTransportType() modbus.TransportType {
	return modbus.TransportTCP // Uses same protocol, just UDP transport
}

// String returns a string representation
func (t *UDPTransport) String() string {
	return fmt.Sprintf("UDP(%s)", t.address)
}

// TCPServer implements a MODBUS TCP server
type TCPServer struct {
	listener       net.Listener
	address        string
	handler        RequestHandler
	connections    map[net.Conn]bool
	mutex          sync.RWMutex
	running        bool
	stopChan       chan struct{}
	wg             sync.WaitGroup
	shutdownCtx    context.Context
	shutdownCancel context.CancelFunc
}

// RequestHandler defines the interface for handling MODBUS requests
type RequestHandler interface {
	HandleRequest(slaveID modbus.SlaveID, req *pdu.Request) *pdu.Response
}

// NewTCPServer creates a new TCP server
func NewTCPServer(address string, handler RequestHandler) *TCPServer {
	ctx, cancel := context.WithCancel(context.Background())
	return &TCPServer{
		address:        address,
		handler:        handler,
		connections:    make(map[net.Conn]bool),
		stopChan:       make(chan struct{}),
		shutdownCtx:    ctx,
		shutdownCancel: cancel,
	}
}

// Start starts the TCP server
func (s *TCPServer) Start() error {
	s.mutex.Lock()
	if s.running {
		s.mutex.Unlock()
		return fmt.Errorf("server already running")
	}

	// Reset shutdown context if restarting
	s.shutdownCtx, s.shutdownCancel = context.WithCancel(context.Background())
	s.stopChan = make(chan struct{})
	s.mutex.Unlock()

	// Start listening
	lc := net.ListenConfig{}
	listener, err := lc.Listen(context.Background(), "tcp", s.address)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.address, err)
	}

	s.mutex.Lock()
	s.listener = listener
	s.running = true
	s.mutex.Unlock()

	s.wg.Add(1)
	go s.acceptLoop()

	return nil
}

// Stop stops the TCP server gracefully
func (s *TCPServer) Stop() error {
	s.mutex.Lock()
	if !s.running {
		s.mutex.Unlock()
		return nil
	}

	// Signal shutdown
	s.shutdownCancel()
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
	s.mutex.Unlock()

	// Wait for all goroutines to finish
	s.wg.Wait()

	return nil
}

// StopWithTimeout stops the server with a timeout for graceful shutdown
func (s *TCPServer) StopWithTimeout(timeout time.Duration) error {
	done := make(chan error, 1)
	go func() {
		done <- s.Stop()
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		return fmt.Errorf("server shutdown timed out after %v", timeout)
	}
}

// IsRunning returns true if the server is running
func (s *TCPServer) IsRunning() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.running
}

// acceptLoop accepts incoming connections
func (s *TCPServer) acceptLoop() {
	defer s.wg.Done()

	for {
		select {
		case <-s.stopChan:
			return
		case <-s.shutdownCtx.Done():
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

			s.wg.Add(1)
			go s.handleConnection(conn)
		}
	}
}

// handleConnection handles a single connection
func (s *TCPServer) handleConnection(conn net.Conn) {
	defer func() {
		s.wg.Done()
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
		case <-s.shutdownCtx.Done():
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
