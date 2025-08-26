package modbus

import (
	"testing"
	"time"

	"github.com/adibhanna/modbusgo/modbus"
)

func TestTCPClient(t *testing.T) {
	// Start a test server
	dataStore := NewDefaultDataStore(1000, 1000, 1000, 1000)

	// Initialize test data
	for i := 0; i < 10; i++ {
		dataStore.SetCoil(modbus.Address(i), i%2 == 0)
		dataStore.SetHoldingRegister(modbus.Address(i), uint16(i*100))
	}

	server, err := NewTCPServer("localhost:15502", dataStore)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Create client
	client := NewTCPClient("localhost:15502")
	client.SetSlaveID(1)
	client.SetTimeout(2 * time.Second)

	t.Run("ConnectAndDisconnect", func(t *testing.T) {
		if err := client.Connect(); err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}

		if !client.IsConnected() {
			t.Error("Expected client to be connected")
		}

		client.Close()

		if client.IsConnected() {
			t.Error("Expected client to be disconnected")
		}

		// Reconnect for other tests
		if err := client.Connect(); err != nil {
			t.Fatalf("Failed to reconnect: %v", err)
		}
	})

	t.Run("ReadCoils", func(t *testing.T) {
		values, err := client.ReadCoils(0, 5)
		if err != nil {
			t.Fatalf("Failed to read coils: %v", err)
		}

		expected := []bool{true, false, true, false, true}
		for i, v := range values {
			if v != expected[i] {
				t.Errorf("Coil %d: expected %v, got %v", i, expected[i], v)
			}
		}
	})

	t.Run("WriteSingleCoil", func(t *testing.T) {
		// Write coil 10 to ON
		if err := client.WriteSingleCoil(10, true); err != nil {
			t.Fatalf("Failed to write coil: %v", err)
		}

		// Read back
		values, err := client.ReadCoils(10, 1)
		if err != nil {
			t.Fatalf("Failed to read coil: %v", err)
		}

		if !values[0] {
			t.Error("Expected coil to be ON")
		}
	})

	t.Run("ReadHoldingRegisters", func(t *testing.T) {
		values, err := client.ReadHoldingRegisters(0, 5)
		if err != nil {
			t.Fatalf("Failed to read holding registers: %v", err)
		}

		for i, v := range values {
			expected := uint16(i * 100)
			if v != expected {
				t.Errorf("Register %d: expected %d, got %d", i, expected, v)
			}
		}
	})

	t.Run("WriteSingleRegister", func(t *testing.T) {
		// Write register 20 to 12345
		if err := client.WriteSingleRegister(20, 12345); err != nil {
			t.Fatalf("Failed to write register: %v", err)
		}

		// Read back
		values, err := client.ReadHoldingRegisters(20, 1)
		if err != nil {
			t.Fatalf("Failed to read register: %v", err)
		}

		if values[0] != 12345 {
			t.Errorf("Expected 12345, got %d", values[0])
		}
	})

	t.Run("WriteMultipleRegisters", func(t *testing.T) {
		// Write multiple registers
		writeValues := []uint16{111, 222, 333, 444}
		if err := client.WriteMultipleRegisters(30, writeValues); err != nil {
			t.Fatalf("Failed to write multiple registers: %v", err)
		}

		// Read back
		readValues, err := client.ReadHoldingRegisters(30, modbus.Quantity(len(writeValues)))
		if err != nil {
			t.Fatalf("Failed to read registers: %v", err)
		}

		for i, v := range readValues {
			if v != writeValues[i] {
				t.Errorf("Register %d: expected %d, got %d", i+30, writeValues[i], v)
			}
		}
	})

	t.Run("MaskWriteRegister", func(t *testing.T) {
		// Set initial value
		if err := client.WriteSingleRegister(40, 0x12); err != nil {
			t.Fatalf("Failed to write initial value: %v", err)
		}

		// Apply mask
		if err := client.MaskWriteRegister(40, 0xF2, 0x25); err != nil {
			t.Fatalf("Failed to mask write register: %v", err)
		}

		// Read back
		values, err := client.ReadHoldingRegisters(40, 1)
		if err != nil {
			t.Fatalf("Failed to read register: %v", err)
		}

		// Expected: (0x12 & 0xF2) | (0x25 & ~0xF2) = 0x12 | 0x05 = 0x17
		if values[0] != 0x17 {
			t.Errorf("Expected 0x17, got 0x%02X", values[0])
		}
	})

	t.Run("ReadWriteMultipleRegisters", func(t *testing.T) {
		// Set initial values
		if err := client.WriteMultipleRegisters(50, []uint16{1111, 2222, 3333}); err != nil {
			t.Fatalf("Failed to set initial values: %v", err)
		}

		// Read from 50-52, write to 60-61
		writeValues := []uint16{9999, 8888}
		readValues, err := client.ReadWriteMultipleRegisters(50, 3, 60, writeValues)
		if err != nil {
			t.Fatalf("Failed to read/write multiple registers: %v", err)
		}

		// Check read values
		expectedRead := []uint16{1111, 2222, 3333}
		for i, v := range readValues {
			if v != expectedRead[i] {
				t.Errorf("Read register %d: expected %d, got %d", i+50, expectedRead[i], v)
			}
		}

		// Verify write
		writtenValues, err := client.ReadHoldingRegisters(60, modbus.Quantity(len(writeValues)))
		if err != nil {
			t.Fatalf("Failed to read written registers: %v", err)
		}

		for i, v := range writtenValues {
			if v != writeValues[i] {
				t.Errorf("Written register %d: expected %d, got %d", i+60, writeValues[i], v)
			}
		}
	})

	t.Run("ReadDeviceIdentification", func(t *testing.T) {
		deviceID, moreFollows, nextObjectID, err := client.ReadDeviceIdentification(
			modbus.DeviceIDReadBasic, 0)
		if err != nil {
			t.Fatalf("Failed to read device identification: %v", err)
		}

		if deviceID.VendorName == "" {
			t.Error("Expected vendor name")
		}

		if deviceID.ProductCode == "" {
			t.Error("Expected product code")
		}

		if moreFollows {
			t.Logf("More objects available, next object ID: %d", nextObjectID)
		}
	})

	// Close client
	client.Close()
}

func TestClientRetry(t *testing.T) {
	// Test with non-existent server
	client := NewTCPClient("localhost:19999")
	client.SetSlaveID(1)
	client.SetTimeout(100 * time.Millisecond)
	client.SetRetryCount(2)

	// Try to connect - should fail
	err := client.Connect()
	if err == nil {
		t.Error("Expected connection error")
		client.Close()
	}
}

func TestClientTimeout(t *testing.T) {
	// Start a test server
	dataStore := NewDefaultDataStore(100, 100, 100, 100)
	server, err := NewTCPServer("localhost:15503", dataStore)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Create client with very short timeout
	client := NewTCPClient("localhost:15503")
	client.SetSlaveID(1)
	client.SetTimeout(1 * time.Nanosecond) // Extremely short timeout

	if err := client.Connect(); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	// This should timeout
	_, err = client.ReadCoils(0, 10)
	if err == nil {
		t.Error("Expected timeout error")
	}
}

// Benchmark client operations
func BenchmarkClientReadHoldingRegisters(b *testing.B) {
	// Start server
	dataStore := NewDefaultDataStore(1000, 1000, 1000, 1000)
	server, _ := NewTCPServer("localhost:15504", dataStore)
	server.Start()
	defer server.Stop()

	time.Sleep(100 * time.Millisecond)

	// Create client
	client := NewTCPClient("localhost:15504")
	client.SetSlaveID(1)
	client.Connect()
	defer client.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.ReadHoldingRegisters(0, 100)
	}
}

func BenchmarkClientWriteMultipleRegisters(b *testing.B) {
	// Start server
	dataStore := NewDefaultDataStore(1000, 1000, 1000, 1000)
	server, _ := NewTCPServer("localhost:15505", dataStore)
	server.Start()
	defer server.Stop()

	time.Sleep(100 * time.Millisecond)

	// Create client
	client := NewTCPClient("localhost:15505")
	client.SetSlaveID(1)
	client.Connect()
	defer client.Close()

	values := make([]uint16, 100)
	for i := range values {
		values[i] = uint16(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.WriteMultipleRegisters(0, values)
	}
}
