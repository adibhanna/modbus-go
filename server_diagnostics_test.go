package modbus

import (
	"testing"

	"github.com/adibhanna/modbusgo/modbus"
	"github.com/adibhanna/modbusgo/pdu"
)

func TestDiagnosticsFunctions(t *testing.T) {
	t.Run("ReadExceptionStatus", func(t *testing.T) {
		ds := NewDefaultDataStore(100, 100, 100, 100)
		handler := NewServerRequestHandler(ds)

		// Set exception status
		ds.SetExceptionStatus(0xF5)

		// Create request
		req := pdu.NewRequest(modbus.FuncCodeReadExceptionStatus, []byte{})

		resp := handler.HandleRequest(1, req)

		if resp.FunctionCode != modbus.FuncCodeReadExceptionStatus {
			t.Errorf("Expected function code %d, got %d", modbus.FuncCodeReadExceptionStatus, resp.FunctionCode)
		}

		if len(resp.Data) != 1 || resp.Data[0] != 0xF5 {
			t.Errorf("Expected exception status 0xF5, got %v", resp.Data)
		}
	})

	t.Run("Diagnostic", func(t *testing.T) {
		ds := NewDefaultDataStore(100, 100, 100, 100)
		handler := NewServerRequestHandler(ds)

		// Test Return Query Data
		queryData := []byte{0x12, 0x34}
		req := pdu.NewRequest(modbus.FuncCodeDiagnostic,
			append(pdu.EncodeUint16(modbus.DiagSubReturnQueryData), queryData...))

		resp := handler.HandleRequest(1, req)

		if resp.FunctionCode != modbus.FuncCodeDiagnostic {
			t.Errorf("Expected function code %d, got %d", modbus.FuncCodeDiagnostic, resp.FunctionCode)
		}

		if len(resp.Data) != 4 {
			t.Errorf("Expected 4 bytes response, got %d", len(resp.Data))
		}

		subFunc, _ := pdu.DecodeUint16(resp.Data[0:2])
		if subFunc != modbus.DiagSubReturnQueryData {
			t.Errorf("Expected sub-function %d, got %d", modbus.DiagSubReturnQueryData, subFunc)
		}

		if resp.Data[2] != 0x12 || resp.Data[3] != 0x34 {
			t.Errorf("Expected query data echo, got %v", resp.Data[2:])
		}
	})

	t.Run("GetCommEventCounter", func(t *testing.T) {
		ds := NewDefaultDataStore(100, 100, 100, 100)
		handler := NewServerRequestHandler(ds)

		// Increment some counters
		ds.IncrementDiagnosticCounter("BusMessage")
		ds.IncrementDiagnosticCounter("BusMessage")

		req := pdu.NewRequest(modbus.FuncCodeGetCommEventCounter, []byte{})

		resp := handler.HandleRequest(1, req)

		if resp.FunctionCode != modbus.FuncCodeGetCommEventCounter {
			t.Errorf("Expected function code %d, got %d", modbus.FuncCodeGetCommEventCounter, resp.FunctionCode)
		}

		if len(resp.Data) != 4 {
			t.Errorf("Expected 4 bytes response, got %d", len(resp.Data))
		}

		status, _ := pdu.DecodeUint16(resp.Data[0:2])
		eventCount, _ := pdu.DecodeUint16(resp.Data[2:4])

		if status != 0xFFFF {
			t.Errorf("Expected status 0xFFFF, got 0x%04X", status)
		}

		if eventCount != 2 {
			t.Errorf("Expected event count 2, got %d", eventCount)
		}
	})

	t.Run("GetCommEventLog", func(t *testing.T) {
		ds := NewDefaultDataStore(100, 100, 100, 100)
		handler := NewServerRequestHandler(ds)

		// Increment some counters
		ds.IncrementDiagnosticCounter("BusMessage")
		ds.IncrementDiagnosticCounter("ServerMessage")

		req := pdu.NewRequest(modbus.FuncCodeGetCommEventLog, []byte{})

		resp := handler.HandleRequest(1, req)

		if resp.FunctionCode != modbus.FuncCodeGetCommEventLog {
			t.Errorf("Expected function code %d, got %d", modbus.FuncCodeGetCommEventLog, resp.FunctionCode)
		}

		if len(resp.Data) < 7 {
			t.Errorf("Expected at least 7 bytes response, got %d", len(resp.Data))
		}

		byteCount := resp.Data[0]
		if int(byteCount) != len(resp.Data)-1 {
			t.Errorf("Byte count mismatch: %d vs %d", byteCount, len(resp.Data)-1)
		}

		status, _ := pdu.DecodeUint16(resp.Data[1:3])
		eventCount, _ := pdu.DecodeUint16(resp.Data[3:5])
		messageCount, _ := pdu.DecodeUint16(resp.Data[5:7])

		if status != 0xFFFF {
			t.Errorf("Expected status 0xFFFF, got 0x%04X", status)
		}

		if eventCount != 1 {
			t.Errorf("Expected event count 1, got %d", eventCount)
		}

		if messageCount != 1 {
			t.Errorf("Expected message count 1, got %d", messageCount)
		}
	})

	t.Run("ReportServerID", func(t *testing.T) {
		ds := NewDefaultDataStore(100, 100, 100, 100)
		handler := NewServerRequestHandler(ds)

		req := pdu.NewRequest(modbus.FuncCodeReportServerID, []byte{})

		resp := handler.HandleRequest(1, req)

		if resp.FunctionCode != modbus.FuncCodeReportServerID {
			t.Errorf("Expected function code %d, got %d", modbus.FuncCodeReportServerID, resp.FunctionCode)
		}

		if len(resp.Data) < 2 {
			t.Errorf("Expected at least 2 bytes response, got %d", len(resp.Data))
		}

		byteCount := resp.Data[0]
		if int(byteCount) != len(resp.Data)-1 {
			t.Errorf("Byte count mismatch: %d vs %d", byteCount, len(resp.Data)-1)
		}

		runIndicator := resp.Data[1]
		if runIndicator != 0xFF {
			t.Errorf("Expected run indicator 0xFF, got 0x%02X", runIndicator)
		}

		serverID := string(resp.Data[2:])
		if serverID != "ModbusGo Server v1.0" {
			t.Errorf("Expected server ID 'ModbusGo Server v1.0', got '%s'", serverID)
		}
	})
}

func TestFileRecordFunctions(t *testing.T) {
	t.Run("ReadWriteFileRecord", func(t *testing.T) {
		ds := NewDefaultDataStore(100, 100, 100, 100)

		// First write some file records
		writeRecords := []modbus.FileRecord{
			{
				ReferenceType: modbus.FileRecordTypeExtended,
				FileNumber:    4,
				RecordNumber:  1,
				RecordLength:  3,
				RecordData:    []uint16{0x1234, 0x5678, 0x9ABC},
			},
		}

		err := ds.WriteFileRecords(writeRecords)
		if err != nil {
			t.Fatalf("Failed to write file records: %v", err)
		}

		// Now read them back
		readRecords := []modbus.FileRecord{
			{
				ReferenceType: modbus.FileRecordTypeExtended,
				FileNumber:    4,
				RecordNumber:  1,
				RecordLength:  3,
			},
		}

		result, err := ds.ReadFileRecords(readRecords)
		if err != nil {
			t.Fatalf("Failed to read file records: %v", err)
		}

		if len(result) != 1 {
			t.Errorf("Expected 1 record, got %d", len(result))
		}

		if len(result[0].RecordData) != 3 {
			t.Errorf("Expected 3 data values, got %d", len(result[0].RecordData))
		}

		if result[0].RecordData[0] != 0x1234 || result[0].RecordData[1] != 0x5678 || result[0].RecordData[2] != 0x9ABC {
			t.Errorf("Data mismatch: got %v", result[0].RecordData)
		}
	})
}

func TestFIFOQueue(t *testing.T) {
	t.Run("ReadWriteFIFOQueue", func(t *testing.T) {
		ds := NewDefaultDataStore(100, 100, 100, 100)
		handler := NewServerRequestHandler(ds)

		// Write some data to FIFO queue
		fifoData := []uint16{0x1111, 0x2222, 0x3333, 0x4444}
		err := ds.WriteFIFOQueue(0x1234, fifoData)
		if err != nil {
			t.Fatalf("Failed to write FIFO queue: %v", err)
		}

		// Create read FIFO queue request
		req := pdu.NewRequest(modbus.FuncCodeReadFIFOQueue, pdu.EncodeUint16(0x1234))

		resp := handler.HandleRequest(1, req)

		if resp.FunctionCode != modbus.FuncCodeReadFIFOQueue {
			t.Errorf("Expected function code %d, got %d", modbus.FuncCodeReadFIFOQueue, resp.FunctionCode)
		}

		if len(resp.Data) < 4 {
			t.Errorf("Expected at least 4 bytes response, got %d", len(resp.Data))
		}

		byteCount, _ := pdu.DecodeUint16(resp.Data[0:2])
		fifoCount, _ := pdu.DecodeUint16(resp.Data[2:4])

		if fifoCount != 4 {
			t.Errorf("Expected FIFO count 4, got %d", fifoCount)
		}

		if int(byteCount) != 2+int(fifoCount)*2 {
			t.Errorf("Byte count mismatch: %d vs expected %d", byteCount, 2+fifoCount*2)
		}

		// Decode FIFO values
		values, _ := pdu.DecodeUint16Slice(resp.Data[4:])
		if len(values) != 4 {
			t.Errorf("Expected 4 values, got %d", len(values))
		}

		for i, expected := range fifoData {
			if values[i] != expected {
				t.Errorf("Value mismatch at index %d: expected 0x%04X, got 0x%04X", i, expected, values[i])
			}
		}
	})

	t.Run("EmptyFIFOQueue", func(t *testing.T) {
		ds := NewDefaultDataStore(100, 100, 100, 100)
		handler := NewServerRequestHandler(ds)

		// Read non-existent FIFO queue (should return empty)
		req := pdu.NewRequest(modbus.FuncCodeReadFIFOQueue, pdu.EncodeUint16(0x9999))

		resp := handler.HandleRequest(1, req)

		if resp.FunctionCode != modbus.FuncCodeReadFIFOQueue {
			t.Errorf("Expected function code %d, got %d", modbus.FuncCodeReadFIFOQueue, resp.FunctionCode)
		}

		byteCount, _ := pdu.DecodeUint16(resp.Data[0:2])
		fifoCount, _ := pdu.DecodeUint16(resp.Data[2:4])

		if fifoCount != 0 {
			t.Errorf("Expected FIFO count 0 for empty queue, got %d", fifoCount)
		}

		if byteCount != 2 {
			t.Errorf("Expected byte count 2 for empty queue, got %d", byteCount)
		}
	})
}
