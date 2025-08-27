package main

import (
	"flag"
	"fmt"
	"log"
	"net"
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

	fmt.Printf("=== MODBUS TCP Diagnostic Tool ===\n")
	fmt.Printf("Target: %s\n", cfg.Connection.GetFullAddress())
	fmt.Printf("Slave ID: %d\n", cfg.Modbus.SlaveID)
	fmt.Printf("Timeout: %v\n", cfg.Connection.GetTimeout())
	fmt.Printf("Device Profile: %s\n\n", cfg.CurrentProfile)

	// Step 1: Test basic TCP connectivity
	fmt.Println("--- Step 1: Testing Basic TCP Connectivity ---")
	if err := testTCPConnection(cfg.Connection.GetFullAddress(), cfg.Connection.GetTimeout()); err != nil {
		log.Fatalf("TCP connection failed: %v", err)
	}
	fmt.Println("✓ Basic TCP connection successful")

	// Step 2: Test MODBUS client creation and connection
	fmt.Println("--- Step 2: Testing MODBUS Client Connection ---")
	client := modbus.NewTCPClient(cfg.Connection.GetFullAddress())
	client.SetSlaveID(cfg.Modbus.GetSlaveID())
	client.SetTimeout(cfg.Connection.GetTimeout())
	client.SetRetryCount(1) // Reduce retries for faster diagnosis

	if err := client.Connect(); err != nil {
		log.Fatalf("MODBUS client connection failed: %v", err)
	}
	defer client.Close()
	fmt.Println("✓ MODBUS client connected successfully")

	// Step 3: Test simple MODBUS operations with detailed error reporting
	fmt.Println("--- Step 3: Testing MODBUS Operations ---")

	// Test only enabled functions
	if cfg.Testing.IsTestEnabled("read_holding_registers") {
		fmt.Println("Testing Read Holding Registers (Function Code 0x03)...")
		testReadHoldingRegisters(client, cfg)
	}

	if cfg.Testing.IsTestEnabled("read_coils") {
		fmt.Println("\nTesting Read Coils (Function Code 0x01)...")
		testReadCoils(client, cfg)
	}

	if cfg.Testing.IsTestEnabled("read_input_registers") {
		fmt.Println("\nTesting Read Input Registers (Function Code 0x04)...")
		testReadInputRegisters(client, cfg)
	}

	if cfg.Testing.IsTestEnabled("read_discrete_inputs") {
		fmt.Println("\nTesting Read Discrete Inputs (Function Code 0x02)...")
		testReadDiscreteInputs(client, cfg)
	}

	if cfg.Testing.IsTestEnabled("device_identification") {
		fmt.Println("\nTesting Device Identification (Function Code 0x2B/0x0E)...")
		testDeviceIdentification(client, cfg)
	}

	fmt.Println("\n--- Diagnostic Complete ---")
	fmt.Println("If all tests failed with timeouts, the device may not support MODBUS TCP")
	fmt.Println("or may require different configuration (slave ID, addresses, etc.)")
}

func testTCPConnection(address string, timeout time.Duration) error {
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Try to set a short read timeout and see if we can detect any response
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)

	if err != nil {
		// This is expected - most MODBUS devices won't send data without a request
		fmt.Printf("  No immediate data from server (expected): %v\n", err)
	} else {
		fmt.Printf("  Received %d bytes immediately: %x\n", n, buffer[:n])
	}

	return nil
}

func testReadHoldingRegisters(client *modbus.Client, cfg *config.Config) {
	startTime := time.Now()
	addr := cfg.Testing.TestAddresses["holding_registers"]
	registers, err := client.ReadHoldingRegisters(addr.GetAddress(), addr.GetQuantity())
	duration := time.Since(startTime)

	if err != nil {
		fmt.Printf("  ✗ Failed after %v: %v\n", duration, err)
	} else {
		fmt.Printf("  ✓ Success in %v: %v (addresses %d-%d)\n", duration, registers, addr.StartAddress, addr.StartAddress+addr.Quantity-1)
	}
}

func testReadCoils(client *modbus.Client, cfg *config.Config) {
	startTime := time.Now()
	addr := cfg.Testing.TestAddresses["coils"]
	coils, err := client.ReadCoils(addr.GetAddress(), addr.GetQuantity())
	duration := time.Since(startTime)

	if err != nil {
		fmt.Printf("  ✗ Failed after %v: %v\n", duration, err)
	} else {
		fmt.Printf("  ✓ Success in %v: %v (addresses %d-%d)\n", duration, coils, addr.StartAddress, addr.StartAddress+addr.Quantity-1)
	}
}

func testReadInputRegisters(client *modbus.Client, cfg *config.Config) {
	startTime := time.Now()
	addr := cfg.Testing.TestAddresses["input_registers"]
	registers, err := client.ReadInputRegisters(addr.GetAddress(), addr.GetQuantity())
	duration := time.Since(startTime)

	if err != nil {
		fmt.Printf("  ✗ Failed after %v: %v\n", duration, err)
	} else {
		fmt.Printf("  ✓ Success in %v: %v (addresses %d-%d)\n", duration, registers, addr.StartAddress, addr.StartAddress+addr.Quantity-1)
	}
}

func testReadDiscreteInputs(client *modbus.Client, cfg *config.Config) {
	startTime := time.Now()
	addr := cfg.Testing.TestAddresses["discrete_inputs"]
	inputs, err := client.ReadDiscreteInputs(addr.GetAddress(), addr.GetQuantity())
	duration := time.Since(startTime)

	if err != nil {
		fmt.Printf("  ✗ Failed after %v: %v\n", duration, err)
	} else {
		fmt.Printf("  ✓ Success in %v: %v (addresses %d-%d)\n", duration, inputs, addr.StartAddress, addr.StartAddress+addr.Quantity-1)
	}
}

func testDeviceIdentification(client *modbus.Client, cfg *config.Config) {
	startTime := time.Now()
	deviceID, moreFollows, nextObjectID, err := client.ReadDeviceIdentification(modbuslib.DeviceIDReadBasic, 0)
	duration := time.Since(startTime)

	if err != nil {
		fmt.Printf("  ✗ Failed after %v: %v\n", duration, err)
	} else {
		fmt.Printf("  ✓ Success in %v:\n", duration)
		fmt.Printf("    Vendor Name: %s\n", deviceID.VendorName)
		fmt.Printf("    Product Code: %s\n", deviceID.ProductCode)
		fmt.Printf("    Major/Minor Revision: %s\n", deviceID.MajorMinorRevision)
		fmt.Printf("    More Follows: %t, Next Object ID: %d\n", moreFollows, nextObjectID)
	}
}
