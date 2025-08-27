package main

import (
	"fmt"
	"log"

	modbus "github.com/adibhanna/modbus-go"
	modbuslib "github.com/adibhanna/modbus-go/modbus"
)

func main() {
	// Example 1: Using core library with JSON string configuration
	fmt.Println("=== Example 1: Loading from JSON string ===")
	jsonConfig1 := `{
		"slave_id": 1,
		"timeout_ms": 10000,
		"retry_count": 3,
		"retry_delay_ms": 100,
		"connect_timeout_ms": 5000,
		"transport_type": "tcp"
	}`
	
	client1, err := modbus.NewTCPClientFromJSONString(jsonConfig1, "192.168.1.102:502")
	if err != nil {
		log.Printf("Failed to create client from JSON string: %v", err)
	} else {
		config := client1.GetConfig()
		fmt.Printf("Loaded config - SlaveID: %d, Timeout: %v, RetryCount: %d\n", 
			config.SlaveID, config.Timeout, config.RetryCount)
	}

	// Example 2: Using core library with different JSON configuration
	fmt.Println("\n=== Example 2: Different JSON configuration ===")
	jsonConfig := `{
		"slave_id": 2,
		"timeout_ms": 5000,
		"retry_count": 2,
		"retry_delay_ms": 200,
		"connect_timeout_ms": 3000,
		"transport_type": "tcp"
	}`
	
	client2, err := modbus.NewTCPClientFromJSONString(jsonConfig, "192.168.1.102:502")
	if err != nil {
		log.Printf("Failed to create client from JSON string: %v", err)
	} else {
		config := client2.GetConfig()
		fmt.Printf("Loaded config - SlaveID: %d, Timeout: %v, RetryDelay: %v\n", 
			config.SlaveID, config.Timeout, config.RetryDelay)
	}

	// Example 3: Using ClientConfig struct directly
	fmt.Println("\n=== Example 3: Using ClientConfig struct ===")
	config := modbuslib.DefaultClientConfig()
	config.SlaveID = 3
	config.RetryCount = 1
	
	client3 := modbus.NewTCPClientFromConfig(config, "192.168.1.102:502")
	fmt.Printf("Created client with SlaveID: %d, RetryCount: %d\n", 
		client3.GetSlaveID(), client3.GetRetryCount())

	// Example 4: Modifying configuration at runtime
	fmt.Println("\n=== Example 4: Runtime configuration changes ===")
	client4 := modbus.NewTCPClient("192.168.1.102:502")
	
	// Show initial config
	initialConfig := client4.GetConfig()
	fmt.Printf("Initial config - SlaveID: %d, Timeout: %v\n", 
		initialConfig.SlaveID, initialConfig.Timeout)
	
	// Modify individual settings
	client4.SetSlaveID(5)
	client4.SetRetryCount(5)
	client4.SetRetryDelay(500)
	
	// Show modified config
	modifiedConfig := client4.GetConfig()
	fmt.Printf("Modified config - SlaveID: %d, RetryCount: %d, RetryDelay: %v\n", 
		modifiedConfig.SlaveID, modifiedConfig.RetryCount, modifiedConfig.RetryDelay)

	// Example 5: Saving configuration to JSON
	fmt.Println("\n=== Example 5: Saving configuration to JSON ===")
	configToSave := client4.GetConfig()
	
	// Save to file
	if err := configToSave.SaveClientConfigToJSON("./saved-config.json"); err != nil {
		log.Printf("Failed to save config to file: %v", err)
	} else {
		fmt.Println("Configuration saved to saved-config.json")
	}
	
	// Convert to JSON string
	if jsonStr, err := configToSave.ToJSONString(); err != nil {
		log.Printf("Failed to convert config to JSON string: %v", err)
	} else {
		fmt.Printf("Configuration as JSON:\n%s\n", jsonStr)
	}

	fmt.Println("\n=== All examples completed ===")
}
