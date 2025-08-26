package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	modbus "github.com/adibhanna/modbusgo"
)

func main() {
	fmt.Println("Starting MODBUS TCP Server...")

	// Create a data store with some initial data
	dataStore := modbus.NewDefaultDataStore(1000, 1000, 1000, 1000)

	// Initialize some test data
	fmt.Println("Initializing test data...")

	// Set some coils
	for i := 0; i < 10; i++ {
		if err := dataStore.SetCoil(modbus.Address(i), i%2 == 0); err != nil {
			log.Printf("Warning: failed to set coil %d: %v", i, err)
		}
	}

	// Set some discrete inputs
	for i := 0; i < 10; i++ {
		if err := dataStore.SetDiscreteInput(modbus.Address(i), i%3 == 0); err != nil {
			log.Printf("Warning: failed to set discrete input %d: %v", i, err)
		}
	}

	// Set some holding registers with test pattern
	for i := 0; i < 20; i++ {
		if err := dataStore.SetHoldingRegister(modbus.Address(i), uint16(i*100)); err != nil {
			log.Printf("Warning: failed to set holding register %d: %v", i, err)
		}
	}

	// Set some input registers
	for i := 0; i < 20; i++ {
		if err := dataStore.SetInputRegister(modbus.Address(i), uint16(i*10+5)); err != nil {
			log.Printf("Warning: failed to set input register %d: %v", i, err)
		}
	}

	// Create and start the server
	server, err := modbus.NewTCPServer(":5502", dataStore)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	fmt.Println("MODBUS TCP Server started on port 5502")
	fmt.Println("Test data initialized:")
	fmt.Println("  - Coils 0-9: alternating pattern (ON/OFF)")
	fmt.Println("  - Discrete inputs 0-9: every 3rd is ON")
	fmt.Println("  - Holding registers 0-19: values 0, 100, 200, ..., 1900")
	fmt.Println("  - Input registers 0-19: values 5, 15, 25, ..., 195")
	fmt.Println()
	fmt.Println("You can now connect with a MODBUS client to test the server.")
	fmt.Println("Example client connection: go run examples/tcp_client/main.go")
	fmt.Println()
	fmt.Println("Press Ctrl+C to stop the server")

	// Create a goroutine to periodically update some values to show dynamic behavior
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		counter := uint16(0)
		for range ticker.C {
			counter++
			// Update register 999 with an incrementing counter
			if err := dataStore.SetHoldingRegister(999, counter); err != nil {
				log.Printf("Failed to update holding register: %v", err)
			} else {
				log.Printf("Updated holding register 999 to %d", counter)
			}

			if err := dataStore.SetCoil(999, counter%2 == 0); err != nil {
				log.Printf("Failed to update coil: %v", err)
			}

			fmt.Printf("Updated test values: register[999]=%d, coil[999]=%t\n",
				counter, counter%2 == 0)
		}
	}()

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("\nShutting down server...")
	if err := server.Stop(); err != nil {
		log.Printf("Error stopping server: %v", err)
	}

	fmt.Println("Server stopped")
}
