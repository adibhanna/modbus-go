package main

import (
	"fmt"
	"log"
	"time"

	modbus "github.com/adibhanna/modbus-go"
	modbustypes "github.com/adibhanna/modbus-go/modbus"
	"github.com/adibhanna/modbus-go/transport"
)

func main() {
	// Create a comprehensive data store
	dataStore := modbus.NewDefaultDataStore(10000, 10000, 10000, 10000)

	// Initialize some example data
	initializeExampleData(dataStore)

	// Create TCP server
	address := ":5502"
	handler := modbus.NewServerRequestHandler(dataStore)

	// Set custom device identification
	handler.SetDeviceIdentification(&modbustypes.DeviceIdentification{
		VendorName:          "ModbusGo Advanced",
		ProductCode:         "MGA-001",
		MajorMinorRevision:  "2.0.0",
		VendorURL:           "https://github.com/adibhanna/modbus-go",
		ProductName:         "Advanced MODBUS Server",
		ModelName:           "AGS-2024",
		UserApplicationName: "Industrial Control System",
		ConformityLevel:     modbustypes.ConformityLevelBasicStream,
	})

	server := transport.NewTCPServer(address, handler)

	fmt.Printf("Starting advanced MODBUS TCP server on %s...\n", address)
	fmt.Println("\nSupported advanced features:")
	fmt.Println("- Diagnostics (Function Code 0x08)")
	fmt.Println("- Exception Status (Function Code 0x07)")
	fmt.Println("- Communication Event Counters (Function Code 0x0B)")
	fmt.Println("- Communication Event Log (Function Code 0x0C)")
	fmt.Println("- Server ID Reporting (Function Code 0x11)")
	fmt.Println("- File Record Operations (Function Codes 0x14, 0x15)")
	fmt.Println("- FIFO Queue Operations (Function Code 0x18)")
	fmt.Println("- Device Identification (Function Code 0x2B)")
	fmt.Println("\nPress Ctrl+C to stop the server...")

	// Start periodic data updates
	go periodicDataUpdates(dataStore)

	// Start server
	if err := server.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func initializeExampleData(ds *modbus.DefaultDataStore) {
	// Set some initial holding registers
	for i := 0; i < 10; i++ {
		_ = ds.SetHoldingRegister(modbustypes.Address(i), uint16(1000+i))
	}

	// Set some initial input registers
	for i := 0; i < 10; i++ {
		_ = ds.SetInputRegister(modbustypes.Address(i), uint16(2000+i))
	}

	// Initialize some coils
	_ = ds.SetCoil(0, true)
	_ = ds.SetCoil(1, false)
	_ = ds.SetCoil(2, true)

	// Set exception status
	ds.SetExceptionStatus(0x00) // All OK

	// Initialize FIFO queues
	fifoData1 := []uint16{100, 200, 300, 400, 500}
	_ = ds.WriteFIFOQueue(1000, fifoData1)

	fifoData2 := []uint16{0xABCD, 0x1234, 0x5678}
	_ = ds.WriteFIFOQueue(2000, fifoData2)

	// Initialize file records
	fileRecords := []modbustypes.FileRecord{
		{
			ReferenceType: modbustypes.FileRecordTypeExtended,
			FileNumber:    1,
			RecordNumber:  0,
			RecordLength:  5,
			RecordData:    []uint16{0x1111, 0x2222, 0x3333, 0x4444, 0x5555},
		},
		{
			ReferenceType: modbustypes.FileRecordTypeExtended,
			FileNumber:    1,
			RecordNumber:  1,
			RecordLength:  3,
			RecordData:    []uint16{0xAAAA, 0xBBBB, 0xCCCC},
		},
		{
			ReferenceType: modbustypes.FileRecordTypeExtended,
			FileNumber:    2,
			RecordNumber:  0,
			RecordLength:  4,
			RecordData:    []uint16{0x0001, 0x0002, 0x0003, 0x0004},
		},
	}

	if err := ds.WriteFileRecords(fileRecords); err != nil {
		log.Printf("Error initializing file records: %v", err)
	}

	fmt.Println("Example data initialized:")
	fmt.Println("- Holding registers 0-9: 1000-1009")
	fmt.Println("- Input registers 0-9: 2000-2009")
	fmt.Println("- Coils 0-2: ON, OFF, ON")
	fmt.Println("- FIFO Queue at address 1000: [100, 200, 300, 400, 500]")
	fmt.Println("- FIFO Queue at address 2000: [0xABCD, 0x1234, 0x5678]")
	fmt.Println("- File 1, Record 0: [0x1111, 0x2222, 0x3333, 0x4444, 0x5555]")
	fmt.Println("- File 1, Record 1: [0xAAAA, 0xBBBB, 0xCCCC]")
	fmt.Println("- File 2, Record 0: [0x0001, 0x0002, 0x0003, 0x0004]")
}

func periodicDataUpdates(ds *modbus.DefaultDataStore) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	counter := uint16(0)
	for range ticker.C {
		// Update some input registers with changing values
		counter++

		// Simulate temperature reading
		temperature := uint16(2000 + (counter % 100))
		_ = ds.SetInputRegister(100, temperature)

		// Simulate pressure reading
		pressure := uint16(1000 + (counter % 50))
		_ = ds.SetInputRegister(101, pressure)

		// Simulate flow rate
		flowRate := uint16(500 + (counter % 200))
		_ = ds.SetInputRegister(102, flowRate)

		// Increment diagnostic counters
		ds.IncrementDiagnosticCounter("BusMessage")
		ds.IncrementDiagnosticCounter("ServerMessage")

		// Occasionally update exception status
		if counter%10 == 0 {
			if counter%20 == 0 {
				ds.SetExceptionStatus(0xFF) // Some exceptions
			} else {
				ds.SetExceptionStatus(0x00) // Clear exceptions
			}
		}

		// Update FIFO queue with new data
		if counter%5 == 0 {
			newFIFO := []uint16{counter, counter + 1, counter + 2}
			_ = ds.WriteFIFOQueue(3000, newFIFO)
		}

		log.Printf("Updated: Temperature=%d, Pressure=%d, FlowRate=%d, Counter=%d",
			temperature, pressure, flowRate, counter)
	}
}
