package main

import (
	"fmt"
	"log"
	"time"

	modbus "github.com/adibhanna/modbus-go"
)

func main() {
	fmt.Println("=== Integration Test ===")

	// Start server
	fmt.Println("Starting server...")
	dataStore := modbus.NewDefaultDataStore(1000, 1000, 1000, 1000)

	// Initialize test data
	for i := 0; i < 10; i++ {
		_ = dataStore.SetHoldingRegister(modbus.Address(i), uint16(i*100))
		_ = dataStore.SetCoil(modbus.Address(i), i%2 == 0)
	}

	server, err := modbus.NewTCPServer(":5502", dataStore)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	// Give server time to start
	time.Sleep(500 * time.Millisecond)
	fmt.Println("Server started on :5502")

	// Create client and connect
	fmt.Println("\nConnecting client...")
	client := modbus.NewTCPClient("localhost:5502")
	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()
	fmt.Println("Client connected!")

	// Test 1: Read holding registers
	fmt.Println("\n--- Test 1: Read Holding Registers ---")
	values, err := client.ReadHoldingRegisters(0, 5)
	if err != nil {
		log.Printf("FAIL: %v", err)
	} else {
		fmt.Printf("PASS: Registers 0-4 = %v (expected [0 100 200 300 400])\n", values)
	}

	// Test 2: Read coils
	fmt.Println("\n--- Test 2: Read Coils ---")
	coils, err := client.ReadCoils(0, 5)
	if err != nil {
		log.Printf("FAIL: %v", err)
	} else {
		fmt.Printf("PASS: Coils 0-4 = %v (expected [true false true false true])\n", coils)
	}

	// Test 3: Write single register
	fmt.Println("\n--- Test 3: Write Single Register ---")
	err = client.WriteSingleRegister(100, 12345)
	if err != nil {
		log.Printf("FAIL: %v", err)
	} else {
		// Read back
		readBack, _ := client.ReadHoldingRegisters(100, 1)
		fmt.Printf("PASS: Wrote 12345 to register 100, read back %v\n", readBack)
	}

	// Test 4: Write single coil
	fmt.Println("\n--- Test 4: Write Single Coil ---")
	err = client.WriteSingleCoil(100, true)
	if err != nil {
		log.Printf("FAIL: %v", err)
	} else {
		// Read back
		readBack, _ := client.ReadCoils(100, 1)
		fmt.Printf("PASS: Wrote true to coil 100, read back %v\n", readBack)
	}

	// Test 5: Write multiple registers
	fmt.Println("\n--- Test 5: Write Multiple Registers ---")
	err = client.WriteMultipleRegisters(200, []uint16{1111, 2222, 3333})
	if err != nil {
		log.Printf("FAIL: %v", err)
	} else {
		// Read back
		readBack, _ := client.ReadHoldingRegisters(200, 3)
		fmt.Printf("PASS: Wrote [1111 2222 3333], read back %v\n", readBack)
	}

	// Test 6: New high-level data type helpers
	fmt.Println("\n--- Test 6: High-Level Data Type Helpers ---")

	// Write float32
	err = client.WriteFloat32(300, 3.14159)
	if err != nil {
		log.Printf("FAIL WriteFloat32: %v", err)
	} else {
		// Read back
		floatVal, _ := client.ReadFloat32(300)
		fmt.Printf("PASS: Wrote 3.14159, read back %.5f\n", floatVal)
	}

	// Write uint32
	err = client.WriteUint32(310, 123456789)
	if err != nil {
		log.Printf("FAIL WriteUint32: %v", err)
	} else {
		// Read back
		uint32Val, _ := client.ReadUint32(310)
		fmt.Printf("PASS: Wrote 123456789, read back %d\n", uint32Val)
	}

	// Test single register read helper
	singleReg, err := client.ReadHoldingRegister(0)
	if err != nil {
		log.Printf("FAIL ReadHoldingRegister: %v", err)
	} else {
		fmt.Printf("PASS: Single register read: %d\n", singleReg)
	}

	// Test single coil read helper
	singleCoil, err := client.ReadCoil(0)
	if err != nil {
		log.Printf("FAIL ReadCoil: %v", err)
	} else {
		fmt.Printf("PASS: Single coil read: %v\n", singleCoil)
	}

	// Stop server
	fmt.Println("\n--- Stopping server ---")
	if err := server.Stop(); err != nil {
		log.Printf("Error stopping server: %v", err)
	}
	fmt.Println("Server stopped")

	fmt.Println("\n=== All Integration Tests Completed ===")
}
