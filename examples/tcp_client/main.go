package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	modbus "github.com/adibhanna/modbus-go"
	"github.com/adibhanna/modbus-go/config"
	modbuslib "github.com/adibhanna/modbus-go/modbus"
)

func main() {
	// Parse command-line flags
	configPath := flag.String("config", "", "Path to configuration file (defaults to config.json)")
	flag.Parse()

	// Load configuration
	fmt.Println("Loading configuration...")
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	fmt.Printf("Configuration loaded successfully!\n")
	fmt.Printf("Target: %s\n", cfg.Connection.GetFullAddress())
	fmt.Printf("Slave ID: %d\n", cfg.Modbus.SlaveID)
	fmt.Printf("Timeout: %v\n", cfg.Connection.GetTimeout())
	fmt.Printf("Device Profile: %s\n\n", cfg.CurrentProfile)

	// Create a new TCP client using configuration
	client := modbus.NewTCPClient(cfg.Connection.GetFullAddress())
	client.SetSlaveID(cfg.Modbus.GetSlaveID())
	client.SetTimeout(cfg.Connection.GetTimeout())
	client.SetRetryCount(cfg.Connection.RetryCount)

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
	if cfg.Testing.IsTestEnabled("read_holding_registers") {
		fmt.Println("\n--- Reading Holding Registers ---")
		addr := cfg.Testing.TestAddresses["holding_registers"]
		registers, err := client.ReadHoldingRegisters(addr.GetAddress(), addr.GetQuantity())
		if err != nil {
			log.Printf("Failed to read holding registers: %v", err)
		} else {
			fmt.Printf("Holding registers [%d-%d]: %v\n", addr.StartAddress, addr.StartAddress+addr.Quantity-1, registers)
		}
	}

	// Example 2: Write single register
	if cfg.Testing.IsTestEnabled("write_single_register") {
		fmt.Println("\n--- Writing Single Register ---")
		addr := cfg.Testing.TestAddresses["holding_registers"]
		value, _ := cfg.Testing.GetSingleRegisterValue()
		if err := client.WriteSingleRegister(addr.GetAddress(), value); err != nil {
			log.Printf("Failed to write single register: %v", err)
		} else {
			fmt.Printf("Successfully wrote value %d to register %d\n", value, addr.StartAddress)
		}
	}

	// Example 3: Read back the register to verify
	if cfg.Testing.IsTestEnabled("write_single_register") {
		fmt.Println("\n--- Reading Back Register to Verify ---")
		addr := cfg.Testing.TestAddresses["holding_registers"]
		registers, err := client.ReadHoldingRegisters(addr.GetAddress(), 1)
		if err != nil {
			log.Printf("Failed to read register %d: %v", addr.StartAddress, err)
		} else {
			fmt.Printf("Register %d value: %d\n", addr.StartAddress, registers[0])
		}
	}

	// Example 4: Write multiple registers
	if cfg.Testing.IsTestEnabled("write_multiple_registers") {
		fmt.Println("\n--- Writing Multiple Registers ---")
		values, _ := cfg.Testing.GetMultipleRegisterValues()
		writeAddr := cfg.Advanced.ReadWriteMultiple.WriteAddress
		if err := client.WriteMultipleRegisters(modbus.Address(writeAddr), values); err != nil {
			log.Printf("Failed to write multiple registers: %v", err)
		} else {
			fmt.Printf("Successfully wrote values %v to registers %d-%d\n", values, writeAddr, writeAddr+len(values)-1)
		}

		// Read back to verify
		fmt.Println("\n--- Reading Multiple Registers ---")
		registers, err := client.ReadHoldingRegisters(modbus.Address(writeAddr), modbus.Quantity(len(values)))
		if err != nil {
			log.Printf("Failed to read multiple registers: %v", err)
		} else {
			fmt.Printf("Registers [%d-%d]: %v\n", writeAddr, writeAddr+len(values)-1, registers)
		}
	}

	// Example 6: Work with coils
	if cfg.Testing.IsTestEnabled("read_coils") || cfg.Testing.IsTestEnabled("write_single_coil") || cfg.Testing.IsTestEnabled("write_multiple_coils") {
		fmt.Println("\n--- Working with Coils ---")
		coilAddr := cfg.Testing.TestAddresses["coils"]

		// Write single coil
		if cfg.Testing.IsTestEnabled("write_single_coil") {
			if err := client.WriteSingleCoil(coilAddr.GetAddress(), true); err != nil {
				log.Printf("Failed to write coil %d: %v", coilAddr.StartAddress, err)
			} else {
				fmt.Printf("Set coil %d to ON\n", coilAddr.StartAddress)
			}
		}

		// Read coils
		if cfg.Testing.IsTestEnabled("read_coils") {
			coils, err := client.ReadCoils(coilAddr.GetAddress(), coilAddr.GetQuantity())
			if err != nil {
				log.Printf("Failed to read coils: %v", err)
			} else {
				fmt.Printf("Coils [%d-%d]: %v\n", coilAddr.StartAddress, coilAddr.StartAddress+coilAddr.Quantity-1, coils)
			}
		}

		// Write multiple coils
		if cfg.Testing.IsTestEnabled("write_multiple_coils") {
			coilValues, _ := cfg.Testing.GetCoilPattern()
			if len(coilValues) > coilAddr.Quantity {
				coilValues = coilValues[:coilAddr.Quantity] // Trim to fit configured quantity
			}
			if err := client.WriteMultipleCoils(coilAddr.GetAddress(), coilValues); err != nil {
				log.Printf("Failed to write multiple coils: %v", err)
			} else {
				fmt.Printf("Successfully wrote coil pattern: %v\n", coilValues)
			}

			// Read coils again to verify
			coils, err := client.ReadCoils(coilAddr.GetAddress(), modbus.Quantity(len(coilValues)))
			if err != nil {
				log.Printf("Failed to read coils after write: %v", err)
			} else {
				fmt.Printf("Coils after write [%d-%d]: %v\n", coilAddr.StartAddress, coilAddr.StartAddress+len(coilValues)-1, coils)
			}
		}
	}

	// Example 7: Read/Write Multiple Registers in one transaction
	if cfg.Testing.IsTestEnabled("read_write_multiple_registers") {
		fmt.Println("\n--- Read/Write Multiple Registers ---")
		rwConfig := cfg.Advanced.ReadWriteMultiple
		readValues, err := client.ReadWriteMultipleRegisters(
			modbus.Address(rwConfig.ReadAddress),
			modbus.Quantity(rwConfig.ReadQuantity),
			modbus.Address(rwConfig.WriteAddress),
			rwConfig.WriteValues)
		if err != nil {
			log.Printf("Failed to read/write multiple registers: %v", err)
		} else {
			fmt.Printf("Read values [%d-%d]: %v\n", rwConfig.ReadAddress, rwConfig.ReadAddress+rwConfig.ReadQuantity-1, readValues)
			fmt.Printf("Wrote values %v to registers %d-%d\n", rwConfig.WriteValues, rwConfig.WriteAddress, rwConfig.WriteAddress+len(rwConfig.WriteValues)-1)
		}
	}

	// Example 8: Mask Write Register
	if cfg.Testing.IsTestEnabled("mask_write_register") {
		fmt.Println("\n--- Mask Write Register ---")
		maskConfig := cfg.Advanced.MaskWrite
		if err := client.MaskWriteRegister(modbus.Address(maskConfig.Address), uint16(maskConfig.AndMask), uint16(maskConfig.OrMask)); err != nil {
			log.Printf("Failed to mask write register: %v", err)
		} else {
			fmt.Printf("Applied mask to register %d (AND: 0x%04X, OR: 0x%04X)\n", maskConfig.Address, maskConfig.AndMask, maskConfig.OrMask)

			// Read back to see the result
			if registers, err := client.ReadHoldingRegisters(modbus.Address(maskConfig.Address), 1); err == nil {
				fmt.Printf("Register %d after mask write: %d (0x%04X)\n", maskConfig.Address, registers[0], registers[0])
			}
		}
	}

	// Example 9: Read Device Identification (if supported)
	if cfg.Testing.IsTestEnabled("device_identification") {
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
	}

	fmt.Println("\n--- All examples completed ---")
	if cfg.Logging.Verbose {
		fmt.Printf("Configuration used: %s\n", cfg.CurrentProfile)
		fmt.Printf("Total tests enabled: %d\n", len(cfg.Testing.EnabledTests))
	}
}
