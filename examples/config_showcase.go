package main

import (
	"fmt"
	"log"
	"time"

	modbus "github.com/adibhanna/modbus-go"
	modbuslib "github.com/adibhanna/modbus-go/modbus"
)

func main() {
	fmt.Println("=== MODBUS Configuration Showcase ===\n")

	// Example 1: Default client (original behavior)
	fmt.Println("1. Default Client Configuration:")
	defaultClient := modbus.NewTCPClient("192.168.1.102:502")
	config := defaultClient.GetConfig()
	fmt.Printf("   SlaveID: %d\n", config.SlaveID)
	fmt.Printf("   Timeout: %v\n", config.Timeout)
	fmt.Printf("   RetryCount: %d\n", config.RetryCount)
	fmt.Printf("   RetryDelay: %v\n", config.RetryDelay)
	fmt.Printf("   ConnectTimeout: %v\n\n", config.ConnectTimeout)

	// Example 2: JSON string configuration  
	fmt.Println("2. JSON String Configuration:")
	jsonConfigStr1 := `{
		"slave_id": 1,
		"timeout_ms": 10000,
		"retry_count": 3,
		"retry_delay_ms": 100,
		"connect_timeout_ms": 5000,
		"transport_type": "tcp"
	}`
	
	jsonClient, err := modbus.NewTCPClientFromJSONString(jsonConfigStr1, "192.168.1.102:502")
	if err != nil {
		log.Printf("   Error loading from JSON string: %v\n", err)
	} else {
		config := jsonClient.GetConfig()
		fmt.Printf("   Successfully loaded from JSON string\n")
		fmt.Printf("   SlaveID: %d\n", config.SlaveID)
		fmt.Printf("   Timeout: %v\n", config.Timeout)
		fmt.Printf("   RetryCount: %d\n\n", config.RetryCount)
	}

	// Example 3: Advanced JSON configuration
	fmt.Println("3. Advanced JSON Configuration:")
	jsonConfigStr := `{
		"slave_id": 2,
		"timeout_ms": 15000,
		"retry_count": 5,
		"retry_delay_ms": 250,
		"connect_timeout_ms": 8000,
		"transport_type": "tcp"
	}`

	jsonStrClient, err := modbus.NewTCPClientFromJSONString(jsonConfigStr, "192.168.1.102:502")
	if err != nil {
		log.Printf("   Error loading from JSON string: %v\n", err)
	} else {
		config := jsonStrClient.GetConfig()
		fmt.Printf("   Successfully loaded from JSON string\n")
		fmt.Printf("   SlaveID: %d\n", config.SlaveID)
		fmt.Printf("   Timeout: %v\n", config.Timeout)
		fmt.Printf("   RetryDelay: %v\n\n", config.RetryDelay)
	}

	// Example 4: ClientConfig struct configuration
	fmt.Println("4. ClientConfig Struct Configuration:")
	configStruct := modbuslib.DefaultClientConfig()
	configStruct.SlaveID = 3
	configStruct.Timeout = 20 * time.Second
	configStruct.RetryCount = 1
	configStruct.RetryDelay = 500 * time.Millisecond

	configClient := modbus.NewTCPClientFromConfig(configStruct, "192.168.1.102:502")
	fmt.Printf("   Created with custom ClientConfig\n")
	fmt.Printf("   SlaveID: %d\n", configClient.GetSlaveID())
	fmt.Printf("   Timeout: %v\n", configClient.GetTimeout())
	fmt.Printf("   RetryDelay: %v\n\n", configClient.GetRetryDelay())

	// Example 5: Runtime configuration changes
	fmt.Println("5. Runtime Configuration Changes:")
	runtimeClient := modbus.NewTCPClient("192.168.1.102:502")
	
	fmt.Printf("   Before changes:\n")
	fmt.Printf("     SlaveID: %d\n", runtimeClient.GetSlaveID())
	fmt.Printf("     RetryCount: %d\n", runtimeClient.GetRetryCount())
	
	runtimeClient.SetSlaveID(10)
	runtimeClient.SetRetryCount(7)
	runtimeClient.SetRetryDelay(1 * time.Second)
	runtimeClient.SetConnectTimeout(30 * time.Second)
	
	fmt.Printf("   After changes:\n")
	fmt.Printf("     SlaveID: %d\n", runtimeClient.GetSlaveID())
	fmt.Printf("     RetryCount: %d\n", runtimeClient.GetRetryCount())
	fmt.Printf("     RetryDelay: %v\n", runtimeClient.GetRetryDelay())
	fmt.Printf("     ConnectTimeout: %v\n\n", runtimeClient.GetConnectTimeout())

	// Example 6: Configuration persistence
	fmt.Println("6. Configuration Persistence:")
	persistConfig := runtimeClient.GetConfig()
	
	// Save to JSON file
	if err := persistConfig.SaveClientConfigToJSON("./runtime-config.json"); err != nil {
		log.Printf("   Error saving to file: %v\n", err)
	} else {
		fmt.Printf("   Configuration saved to runtime-config.json\n")
	}
	
	// Convert to JSON string
	if jsonStr, err := persistConfig.ToJSONString(); err != nil {
		log.Printf("   Error converting to JSON: %v\n", err)
	} else {
		fmt.Printf("   Configuration as JSON:\n%s\n\n", jsonStr)
	}

	// Example 7: Configuration comparison
	fmt.Println("7. Configuration Comparison:")
	defaultConfig := modbuslib.DefaultClientConfig()
	customConfig := &modbuslib.ClientConfig{
		SlaveID:        5,
		Timeout:        10 * time.Second,
		RetryCount:     2,
		RetryDelay:     300 * time.Millisecond,
		ConnectTimeout: 5 * time.Second,
		TransportType:  modbuslib.TransportTCP,
	}

	fmt.Printf("   Default vs Custom Configuration:\n")
	fmt.Printf("   SlaveID: %d vs %d\n", defaultConfig.SlaveID, customConfig.SlaveID)
	fmt.Printf("   Timeout: %v vs %v\n", defaultConfig.Timeout, customConfig.Timeout)
	fmt.Printf("   RetryCount: %d vs %d\n", defaultConfig.RetryCount, customConfig.RetryCount)
	fmt.Printf("   RetryDelay: %v vs %v\n\n", defaultConfig.RetryDelay, customConfig.RetryDelay)

	fmt.Println("=== Configuration Showcase Complete ===")
	fmt.Println("\nAll configuration methods are now available in the core library!")
	fmt.Println("You can use any of these approaches based on your needs:")
	fmt.Println("- Simple clients with modbus.NewTCPClient()")
	fmt.Println("- JSON-based configuration with NewTCPClientFromJSONFile()")
	fmt.Println("- Runtime configuration changes with Set/Get methods") 
	fmt.Println("- Configuration persistence with Save/Load methods")
}
