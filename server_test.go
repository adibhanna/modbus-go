package modbus

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/adibhanna/modbusgo/modbus"
	"github.com/adibhanna/modbusgo/pdu"
)

func TestDefaultDataStore(t *testing.T) {
	ds := NewDefaultDataStore(100, 100, 100, 100)

	t.Run("ReadCoils", func(t *testing.T) {
		// Set some test coils
		ds.SetCoil(0, true)
		ds.SetCoil(1, false)
		ds.SetCoil(2, true)

		// Read coils
		values, err := ds.ReadCoils(0, 3)
		if err != nil {
			t.Fatalf("Failed to read coils: %v", err)
		}

		expected := []bool{true, false, true}
		if !reflect.DeepEqual(values, expected) {
			t.Errorf("Expected %v, got %v", expected, values)
		}

		// Test out of bounds
		_, err = ds.ReadCoils(99, 2)
		if err == nil {
			t.Error("Expected error for out of bounds read")
		}
	})

	t.Run("WriteCoils", func(t *testing.T) {
		values := []bool{false, true, false, true}
		err := ds.WriteCoils(10, values)
		if err != nil {
			t.Fatalf("Failed to write coils: %v", err)
		}

		// Read back
		readValues, err := ds.ReadCoils(10, 4)
		if err != nil {
			t.Fatalf("Failed to read coils: %v", err)
		}

		if !reflect.DeepEqual(values, readValues) {
			t.Errorf("Expected %v, got %v", values, readValues)
		}

		// Test out of bounds
		err = ds.WriteCoils(98, values)
		if err == nil {
			t.Error("Expected error for out of bounds write")
		}
	})

	t.Run("ReadDiscreteInputs", func(t *testing.T) {
		// Set some test discrete inputs
		ds.SetDiscreteInput(0, true)
		ds.SetDiscreteInput(1, true)
		ds.SetDiscreteInput(2, false)

		// Read discrete inputs
		values, err := ds.ReadDiscreteInputs(0, 3)
		if err != nil {
			t.Fatalf("Failed to read discrete inputs: %v", err)
		}

		expected := []bool{true, true, false}
		if !reflect.DeepEqual(values, expected) {
			t.Errorf("Expected %v, got %v", expected, values)
		}
	})

	t.Run("ReadHoldingRegisters", func(t *testing.T) {
		// Set some test registers
		ds.SetHoldingRegister(0, 1234)
		ds.SetHoldingRegister(1, 5678)
		ds.SetHoldingRegister(2, 9012)

		// Read registers
		values, err := ds.ReadHoldingRegisters(0, 3)
		if err != nil {
			t.Fatalf("Failed to read holding registers: %v", err)
		}

		expected := []uint16{1234, 5678, 9012}
		if !reflect.DeepEqual(values, expected) {
			t.Errorf("Expected %v, got %v", expected, values)
		}
	})

	t.Run("WriteHoldingRegisters", func(t *testing.T) {
		values := []uint16{111, 222, 333}
		err := ds.WriteHoldingRegisters(20, values)
		if err != nil {
			t.Fatalf("Failed to write holding registers: %v", err)
		}

		// Read back
		readValues, err := ds.ReadHoldingRegisters(20, 3)
		if err != nil {
			t.Fatalf("Failed to read holding registers: %v", err)
		}

		if !reflect.DeepEqual(values, readValues) {
			t.Errorf("Expected %v, got %v", values, readValues)
		}
	})

	t.Run("ReadInputRegisters", func(t *testing.T) {
		// Set some test input registers
		ds.SetInputRegister(0, 4321)
		ds.SetInputRegister(1, 8765)

		// Read registers
		values, err := ds.ReadInputRegisters(0, 2)
		if err != nil {
			t.Fatalf("Failed to read input registers: %v", err)
		}

		expected := []uint16{4321, 8765}
		if !reflect.DeepEqual(values, expected) {
			t.Errorf("Expected %v, got %v", expected, values)
		}
	})
}

func TestServerRequestHandler(t *testing.T) {
	ds := NewDefaultDataStore(100, 100, 100, 100)
	handler := NewServerRequestHandler(ds)

	t.Run("HandleReadCoils", func(t *testing.T) {
		// Set test data
		ds.SetCoil(0, true)
		ds.SetCoil(1, false)
		ds.SetCoil(2, true)

		// Create request
		reqData := make([]byte, 4)
		copy(reqData[0:2], pdu.EncodeUint16(0)) // Starting address
		copy(reqData[2:4], pdu.EncodeUint16(3)) // Quantity

		req := pdu.NewRequest(modbus.FuncCodeReadCoils, reqData)

		// Handle request
		resp := handler.HandleRequest(1, req)

		// Check response
		if resp.FunctionCode != modbus.FuncCodeReadCoils {
			t.Errorf("Expected function code %d, got %d", modbus.FuncCodeReadCoils, resp.FunctionCode)
		}

		if resp.IsException() {
			ec, _ := resp.GetExceptionCode()
			t.Errorf("Expected no exception, got %d", ec)
		}

		// Check data - first byte is byte count
		if resp.Data[0] != 1 {
			t.Errorf("Expected byte count 1, got %d", resp.Data[0])
		}

		// Check coil values - packed as bits
		// Expected: true, false, true = 0b00000101 = 0x05
		if resp.Data[1] != 0x05 {
			t.Errorf("Expected coil byte 0x05, got 0x%02X", resp.Data[1])
		}
	})

	t.Run("HandleWriteSingleCoil", func(t *testing.T) {
		// Create request to write coil 5 to ON
		reqData := make([]byte, 4)
		copy(reqData[0:2], pdu.EncodeUint16(5))      // Address
		copy(reqData[2:4], pdu.EncodeUint16(0xFF00)) // Value (ON)

		req := pdu.NewRequest(modbus.FuncCodeWriteSingleCoil, reqData)

		// Handle request
		resp := handler.HandleRequest(1, req)

		// Check response
		if resp.FunctionCode != modbus.FuncCodeWriteSingleCoil {
			t.Errorf("Expected function code %d, got %d", modbus.FuncCodeWriteSingleCoil, resp.FunctionCode)
		}

		// Response should echo the request
		if !bytes.Equal(resp.Data, reqData) {
			t.Error("Response data should echo request data")
		}

		// Verify coil was written
		values, _ := ds.ReadCoils(5, 1)
		if !values[0] {
			t.Error("Expected coil 5 to be ON")
		}
	})

	t.Run("HandleReadHoldingRegisters", func(t *testing.T) {
		// Set test data
		ds.SetHoldingRegister(10, 0x1234)
		ds.SetHoldingRegister(11, 0x5678)

		// Create request
		reqData := make([]byte, 4)
		copy(reqData[0:2], pdu.EncodeUint16(10)) // Starting address
		copy(reqData[2:4], pdu.EncodeUint16(2))  // Quantity

		req := pdu.NewRequest(modbus.FuncCodeReadHoldingRegisters, reqData)

		// Handle request
		resp := handler.HandleRequest(1, req)

		// Check response
		if resp.FunctionCode != modbus.FuncCodeReadHoldingRegisters {
			t.Errorf("Expected function code %d, got %d", modbus.FuncCodeReadHoldingRegisters, resp.FunctionCode)
		}

		// Check data - first byte is byte count
		if resp.Data[0] != 4 {
			t.Errorf("Expected byte count 4, got %d", resp.Data[0])
		}

		// Check register values
		reg1, _ := pdu.DecodeUint16(resp.Data[1:3])
		reg2, _ := pdu.DecodeUint16(resp.Data[3:5])

		if reg1 != 0x1234 {
			t.Errorf("Expected register 1 = 0x1234, got 0x%04X", reg1)
		}
		if reg2 != 0x5678 {
			t.Errorf("Expected register 2 = 0x5678, got 0x%04X", reg2)
		}
	})

	t.Run("HandleWriteMultipleRegisters", func(t *testing.T) {
		// Create request to write 3 registers starting at address 20
		values := []uint16{0xAAAA, 0xBBBB, 0xCCCC}

		reqData := make([]byte, 5+len(values)*2)
		copy(reqData[0:2], pdu.EncodeUint16(20))                  // Starting address
		copy(reqData[2:4], pdu.EncodeUint16(uint16(len(values)))) // Quantity
		reqData[4] = byte(len(values) * 2)                        // Byte count

		// Copy register values
		for i, v := range values {
			copy(reqData[5+i*2:7+i*2], pdu.EncodeUint16(v))
		}

		req := pdu.NewRequest(modbus.FuncCodeWriteMultipleRegisters, reqData)

		// Handle request
		resp := handler.HandleRequest(1, req)

		// Check response
		if resp.FunctionCode != modbus.FuncCodeWriteMultipleRegisters {
			t.Errorf("Expected function code %d, got %d", modbus.FuncCodeWriteMultipleRegisters, resp.FunctionCode)
		}

		// Response should contain address and quantity
		respAddr, _ := pdu.DecodeUint16(resp.Data[0:2])
		respQty, _ := pdu.DecodeUint16(resp.Data[2:4])

		if respAddr != 20 {
			t.Errorf("Expected response address 20, got %d", respAddr)
		}
		if respQty != uint16(len(values)) {
			t.Errorf("Expected response quantity %d, got %d", len(values), respQty)
		}

		// Verify registers were written
		readValues, _ := ds.ReadHoldingRegisters(20, modbus.Quantity(len(values)))
		if !reflect.DeepEqual(values, readValues) {
			t.Errorf("Expected registers %v, got %v", values, readValues)
		}
	})

	t.Run("HandleMaskWriteRegister", func(t *testing.T) {
		// Set initial value
		ds.SetHoldingRegister(30, 0x12)

		// Create request - AND mask 0xF2, OR mask 0x25
		// Result should be (0x12 & 0xF2) | (0x25 & ~0xF2) = 0x12 | 0x05 = 0x17
		reqData := make([]byte, 6)
		copy(reqData[0:2], pdu.EncodeUint16(30))     // Address
		copy(reqData[2:4], pdu.EncodeUint16(0x00F2)) // AND mask
		copy(reqData[4:6], pdu.EncodeUint16(0x0025)) // OR mask

		req := pdu.NewRequest(modbus.FuncCodeMaskWriteRegister, reqData)

		// Handle request
		resp := handler.HandleRequest(1, req)

		// Check response
		if resp.FunctionCode != modbus.FuncCodeMaskWriteRegister {
			t.Errorf("Expected function code %d, got %d", modbus.FuncCodeMaskWriteRegister, resp.FunctionCode)
		}

		// Response should echo the request
		if !bytes.Equal(resp.Data, reqData) {
			t.Error("Response data should echo request data")
		}

		// Verify register was masked correctly
		values, _ := ds.ReadHoldingRegisters(30, 1)
		if values[0] != 0x17 {
			t.Errorf("Expected register value 0x17, got 0x%02X", values[0])
		}
	})

	t.Run("HandleReadWriteMultipleRegisters", func(t *testing.T) {
		// Set initial values for read
		ds.SetHoldingRegister(40, 0x1111)
		ds.SetHoldingRegister(41, 0x2222)

		// Create request - read 2 registers from 40, write 2 registers to 50
		writeValues := []uint16{0x3333, 0x4444}

		reqData := make([]byte, 9+len(writeValues)*2)
		copy(reqData[0:2], pdu.EncodeUint16(40))                       // Read address
		copy(reqData[2:4], pdu.EncodeUint16(2))                        // Read quantity
		copy(reqData[4:6], pdu.EncodeUint16(50))                       // Write address
		copy(reqData[6:8], pdu.EncodeUint16(uint16(len(writeValues)))) // Write quantity
		reqData[8] = byte(len(writeValues) * 2)                        // Write byte count

		// Copy write values
		for i, v := range writeValues {
			copy(reqData[9+i*2:11+i*2], pdu.EncodeUint16(v))
		}

		req := pdu.NewRequest(modbus.FuncCodeReadWriteMultipleRegs, reqData)

		// Handle request
		resp := handler.HandleRequest(1, req)

		// Check response
		if resp.FunctionCode != modbus.FuncCodeReadWriteMultipleRegs {
			t.Errorf("Expected function code %d, got %d", modbus.FuncCodeReadWriteMultipleRegs, resp.FunctionCode)
		}

		// Check read values in response
		if resp.Data[0] != 4 { // Byte count
			t.Errorf("Expected byte count 4, got %d", resp.Data[0])
		}

		reg1, _ := pdu.DecodeUint16(resp.Data[1:3])
		reg2, _ := pdu.DecodeUint16(resp.Data[3:5])

		if reg1 != 0x1111 {
			t.Errorf("Expected read register 1 = 0x1111, got 0x%04X", reg1)
		}
		if reg2 != 0x2222 {
			t.Errorf("Expected read register 2 = 0x2222, got 0x%04X", reg2)
		}

		// Verify write was successful
		writtenValues, _ := ds.ReadHoldingRegisters(50, modbus.Quantity(len(writeValues)))
		if !reflect.DeepEqual(writeValues, writtenValues) {
			t.Errorf("Expected written registers %v, got %v", writeValues, writtenValues)
		}
	})

	t.Run("HandleIllegalFunction", func(t *testing.T) {
		req := pdu.NewRequest(0x99, []byte{})

		resp := handler.HandleRequest(1, req)

		// Should return exception
		expectedFC := modbus.FunctionCode(0x99).ToException()
		if resp.FunctionCode != expectedFC {
			t.Errorf("Expected exception function code %d, got %d", expectedFC, resp.FunctionCode)
		}

		if !resp.IsException() {
			t.Error("Expected exception response")
		}

		ec, _ := resp.GetExceptionCode()
		if ec != modbus.ExceptionCodeIllegalFunction {
			t.Errorf("Expected exception code %d, got %d", modbus.ExceptionCodeIllegalFunction, ec)
		}
	})

	t.Run("HandleIllegalDataAddress", func(t *testing.T) {
		// Request to read coils beyond the data store size
		reqData := make([]byte, 4)
		copy(reqData[0:2], pdu.EncodeUint16(99)) // Starting address
		copy(reqData[2:4], pdu.EncodeUint16(5))  // Quantity - will exceed bounds

		req := pdu.NewRequest(modbus.FuncCodeReadCoils, reqData)

		resp := handler.HandleRequest(1, req)

		// Should return exception
		if !resp.IsException() {
			t.Error("Expected exception response")
		}

		ec, _ := resp.GetExceptionCode()
		if ec != modbus.ExceptionCodeIllegalDataAddress {
			t.Errorf("Expected exception code %d, got %d", modbus.ExceptionCodeIllegalDataAddress, ec)
		}
	})

	t.Run("HandleIllegalDataValue", func(t *testing.T) {
		// Request with invalid data length
		req := pdu.NewRequest(modbus.FuncCodeReadCoils, []byte{0x00}) // Too short - should be 4 bytes

		resp := handler.HandleRequest(1, req)

		// Should return exception
		if !resp.IsException() {
			t.Error("Expected exception response")
		}

		ec, _ := resp.GetExceptionCode()
		if ec != modbus.ExceptionCodeIllegalDataValue {
			t.Errorf("Expected exception code %d, got %d", modbus.ExceptionCodeIllegalDataValue, ec)
		}
	})
}

func TestDeviceIdentification(t *testing.T) {
	ds := NewDefaultDataStore(100, 100, 100, 100)
	handler := NewServerRequestHandler(ds)

	// Set custom device identification
	deviceInfo := &modbus.DeviceIdentification{
		VendorName:          "TestVendor",
		ProductCode:         "TEST-001",
		MajorMinorRevision:  "1.2.3",
		VendorURL:           "https://example.com",
		ProductName:         "Test Product",
		ModelName:           "Model X",
		UserApplicationName: "Test App",
		ConformityLevel:     modbus.ConformityLevelBasicStream,
	}
	handler.SetDeviceIdentification(deviceInfo)

	t.Run("ReadDeviceIdentification", func(t *testing.T) {
		// Create request for basic device identification
		reqData := []byte{
			modbus.MEITypeDeviceIdentification,
			modbus.DeviceIDReadBasic,
			0x00, // Object ID
		}

		req := pdu.NewRequest(modbus.FuncCodeEncapsulatedInterface, reqData)

		resp := handler.HandleRequest(1, req)

		// Check response
		if resp.FunctionCode != modbus.FuncCodeEncapsulatedInterface {
			t.Errorf("Expected function code %d, got %d", modbus.FuncCodeEncapsulatedInterface, resp.FunctionCode)
		}

		// Check MEI type in response
		if resp.Data[0] != modbus.MEITypeDeviceIdentification {
			t.Errorf("Expected MEI type %d, got %d", modbus.MEITypeDeviceIdentification, resp.Data[0])
		}

		// Check conformity level
		if resp.Data[2] != modbus.ConformityLevelBasicStream {
			t.Errorf("Expected conformity level %d, got %d", modbus.ConformityLevelBasicStream, resp.Data[2])
		}

		// Check number of objects (should be 3 for basic)
		if resp.Data[5] != 3 {
			t.Errorf("Expected 3 objects, got %d", resp.Data[5])
		}
	})
}

// Benchmark tests
func BenchmarkDataStoreReadCoils(b *testing.B) {
	ds := NewDefaultDataStore(1000, 1000, 1000, 1000)

	// Set some test data
	for i := 0; i < 100; i++ {
		ds.SetCoil(modbus.Address(i), i%2 == 0)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ds.ReadCoils(0, 100)
	}
}

func BenchmarkDataStoreWriteRegisters(b *testing.B) {
	ds := NewDefaultDataStore(1000, 1000, 1000, 1000)
	values := make([]uint16, 100)
	for i := range values {
		values[i] = uint16(i * 100)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ds.WriteHoldingRegisters(0, values)
	}
}

func BenchmarkServerHandleRequest(b *testing.B) {
	ds := NewDefaultDataStore(1000, 1000, 1000, 1000)
	handler := NewServerRequestHandler(ds)

	// Create a read holding registers request
	reqData := make([]byte, 4)
	copy(reqData[0:2], pdu.EncodeUint16(0))   // Starting address
	copy(reqData[2:4], pdu.EncodeUint16(100)) // Quantity

	req := pdu.NewRequest(modbus.FuncCodeReadHoldingRegisters, reqData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.HandleRequest(1, req)
	}
}
