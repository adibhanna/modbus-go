package main

import (
	"fmt"
	"log"
	"time"

	modbus "github.com/adibhanna/modbusgo"
	modbuslib "github.com/adibhanna/modbusgo/modbus"
)

func main() {
	// Create a new TCP client
	client := modbus.NewTCPClient("localhost:5502")
	client.SetSlaveID(1)
	client.SetTimeout(5 * time.Second)

	// Connect to the server
	fmt.Println("Connecting to MODBUS TCP server...")
	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("Error closing client: %v", err)
		}
	}()

	fmt.Println("Connected successfully!")

	// Example 1: Read holding registers
	fmt.Println("\n--- Reading Holding Registers ---")
	registers, err := client.ReadHoldingRegisters(0, 5)
	if err != nil {
		log.Printf("Failed to read holding registers: %v", err)
	} else {
		fmt.Printf("Holding registers [0-4]: %v\n", registers)
	}

	// Example 2: Write single register
	fmt.Println("\n--- Writing Single Register ---")
	if err := client.WriteSingleRegister(0, 12345); err != nil {
		log.Printf("Failed to write single register: %v", err)
	} else {
		fmt.Println("Successfully wrote value 12345 to register 0")
	}

	// Example 3: Read back the register to verify
	fmt.Println("\n--- Reading Back Register 0 ---")
	registers, err = client.ReadHoldingRegisters(0, 1)
	if err != nil {
		log.Printf("Failed to read register 0: %v", err)
	} else {
		fmt.Printf("Register 0 value: %d\n", registers[0])
	}

	// Example 4: Write multiple registers
	fmt.Println("\n--- Writing Multiple Registers ---")
	values := []uint16{100, 200, 300, 400, 500}
	if err := client.WriteMultipleRegisters(10, values); err != nil {
		log.Printf("Failed to write multiple registers: %v", err)
	} else {
		fmt.Printf("Successfully wrote values %v to registers 10-14\n", values)
	}

	// Example 5: Read multiple registers
	fmt.Println("\n--- Reading Multiple Registers ---")
	registers, err = client.ReadHoldingRegisters(10, 5)
	if err != nil {
		log.Printf("Failed to read multiple registers: %v", err)
	} else {
		fmt.Printf("Registers [10-14]: %v\n", registers)
	}

	// Example 6: Work with coils
	fmt.Println("\n--- Working with Coils ---")

	// Write single coil
	if err := client.WriteSingleCoil(0, true); err != nil {
		log.Printf("Failed to write coil 0: %v", err)
	} else {
		fmt.Println("Set coil 0 to ON")
	}

	// Read coils
	coils, err := client.ReadCoils(0, 8)
	if err != nil {
		log.Printf("Failed to read coils: %v", err)
	} else {
		fmt.Printf("Coils [0-7]: %v\n", coils)
	}

	// Write multiple coils
	coilValues := []bool{true, false, true, false, true, false, true, false}
	if err := client.WriteMultipleCoils(0, coilValues); err != nil {
		log.Printf("Failed to write multiple coils: %v", err)
	} else {
		fmt.Printf("Successfully wrote coil pattern: %v\n", coilValues)
	}

	// Read coils again
	coils, err = client.ReadCoils(0, 8)
	if err != nil {
		log.Printf("Failed to read coils: %v", err)
	} else {
		fmt.Printf("Coils after write [0-7]: %v\n", coils)
	}

	// Example 7: Read/Write Multiple Registers in one transaction
	fmt.Println("\n--- Read/Write Multiple Registers ---")
	writeValues := []uint16{999, 888, 777}
	readValues, err := client.ReadWriteMultipleRegisters(0, 3, 20, writeValues)
	if err != nil {
		log.Printf("Failed to read/write multiple registers: %v", err)
	} else {
		fmt.Printf("Read values [0-2]: %v\n", readValues)
		fmt.Printf("Wrote values %v to registers 20-22\n", writeValues)
	}

	// Example 8: Mask Write Register
	fmt.Println("\n--- Mask Write Register ---")
	// Set bits 0, 2, 4 (AND mask = 0xFFFF to preserve all bits, OR mask = 0x0015 to set specific bits)
	if err := client.MaskWriteRegister(5, 0xFFFF, 0x0015); err != nil {
		log.Printf("Failed to mask write register: %v", err)
	} else {
		fmt.Println("Applied mask to register 5 (set bits 0, 2, 4)")

		// Read back to see the result
		if registers, err := client.ReadHoldingRegisters(5, 1); err == nil {
			fmt.Printf("Register 5 after mask write: %d (0x%04X)\n", registers[0], registers[0])
		}
	}

	// Example 9: Read Device Identification (if supported)
	fmt.Println("\n--- Reading Device Identification ---")
	deviceID, moreFollows, nextObjectID, err := client.ReadDeviceIdentification(modbuslib.DeviceIDReadBasic, 0)
	if err != nil {
		log.Printf("Failed to read device identification: %v", err)
	} else {
		fmt.Printf("Device Identification:\n")
		fmt.Printf("  Vendor Name: %s\n", deviceID.VendorName)
		fmt.Printf("  Product Code: %s\n", deviceID.ProductCode)
		fmt.Printf("  Major/Minor Revision: %s\n", deviceID.MajorMinorRevision)
		fmt.Printf("  Conformity Level: 0x%02X\n", deviceID.ConformityLevel)
		fmt.Printf("  More Follows: %t, Next Object ID: %d\n", moreFollows, nextObjectID)
	}

	fmt.Println("\n--- All examples completed ---")
}
