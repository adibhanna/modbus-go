package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/adibhanna/modbus-go/modbus"
)

// ConnectionConfig holds connection-related settings
type ConnectionConfig struct {
	Address          string `json:"address"`
	Port             int    `json:"port"`
	TimeoutMs        int    `json:"timeout_ms"`
	ConnectTimeoutMs int    `json:"connect_timeout_ms"`
	RetryCount       int    `json:"retry_count"`
	TransportType    string `json:"transport_type"`
}

// GetFullAddress returns the full address string (host:port)
func (c *ConnectionConfig) GetFullAddress() string {
	return fmt.Sprintf("%s:%d", c.Address, c.Port)
}

// GetTimeout returns the timeout as a time.Duration
func (c *ConnectionConfig) GetTimeout() time.Duration {
	return time.Duration(c.TimeoutMs) * time.Millisecond
}

// GetConnectTimeout returns the connection timeout as a time.Duration
func (c *ConnectionConfig) GetConnectTimeout() time.Duration {
	return time.Duration(c.ConnectTimeoutMs) * time.Millisecond
}

// ModbusConfig holds MODBUS-specific settings
type ModbusConfig struct {
	SlaveID    int `json:"slave_id"`
	UnitID     int `json:"unit_id"`
	ProtocolID int `json:"protocol_id"`
}

// GetSlaveID returns the slave ID as a modbus.SlaveID type
func (m *ModbusConfig) GetSlaveID() modbus.SlaveID {
	return modbus.SlaveID(m.SlaveID)
}

// AddressRange represents a range of addresses for testing
type AddressRange struct {
	StartAddress int `json:"start_address"`
	Quantity     int `json:"quantity"`
}

// GetAddress returns the start address as modbus.Address
func (a *AddressRange) GetAddress() modbus.Address {
	return modbus.Address(a.StartAddress)
}

// GetQuantity returns the quantity as modbus.Quantity
func (a *AddressRange) GetQuantity() modbus.Quantity {
	return modbus.Quantity(a.Quantity)
}

// TestConfig holds testing-related settings
type TestConfig struct {
	EnabledTests  []string                   `json:"enabled_tests"`
	TestAddresses map[string]AddressRange    `json:"test_addresses"`
	TestValues    map[string]json.RawMessage `json:"test_values"`
}

// IsTestEnabled checks if a specific test is enabled
func (t *TestConfig) IsTestEnabled(testName string) bool {
	for _, enabled := range t.EnabledTests {
		if enabled == testName {
			return true
		}
	}
	return false
}

// GetSingleRegisterValue returns the test value for single register writes
func (t *TestConfig) GetSingleRegisterValue() (uint16, error) {
	var value uint16
	if raw, exists := t.TestValues["single_register_value"]; exists {
		err := json.Unmarshal(raw, &value)
		return value, err
	}
	return 0, fmt.Errorf("single_register_value not found in config")
}

// GetMultipleRegisterValues returns the test values for multiple register writes
func (t *TestConfig) GetMultipleRegisterValues() ([]uint16, error) {
	var values []uint16
	if raw, exists := t.TestValues["multiple_register_values"]; exists {
		err := json.Unmarshal(raw, &values)
		return values, err
	}
	return nil, fmt.Errorf("multiple_register_values not found in config")
}

// GetCoilPattern returns the test pattern for coil operations
func (t *TestConfig) GetCoilPattern() ([]bool, error) {
	var pattern []bool
	if raw, exists := t.TestValues["coil_pattern"]; exists {
		err := json.Unmarshal(raw, &pattern)
		return pattern, err
	}
	return nil, fmt.Errorf("coil_pattern not found in config")
}

// AdvancedConfig holds advanced operation settings
type AdvancedConfig struct {
	MaskWrite struct {
		Address int `json:"address"`
		AndMask int `json:"and_mask"`
		OrMask  int `json:"or_mask"`
	} `json:"mask_write"`
	ReadWriteMultiple struct {
		ReadAddress  int      `json:"read_address"`
		ReadQuantity int      `json:"read_quantity"`
		WriteAddress int      `json:"write_address"`
		WriteValues  []uint16 `json:"write_values"`
	} `json:"read_write_multiple"`
	FIFOQueue struct {
		Address int `json:"address"`
	} `json:"fifo_queue"`
}

// LoggingConfig holds logging settings
type LoggingConfig struct {
	Level       string `json:"level"`
	Verbose     bool   `json:"verbose"`
	ShowTiming  bool   `json:"show_timing"`
	ShowRawData bool   `json:"show_raw_data"`
}

// DeviceProfile holds device-specific configuration
type DeviceProfile struct {
	SlaveID               int    `json:"slave_id"`
	HoldingRegistersStart int    `json:"holding_registers_start"`
	InputRegistersStart   int    `json:"input_registers_start"`
	CoilsStart            int    `json:"coils_start"`
	DiscreteInputsStart   int    `json:"discrete_inputs_start"`
	SupportedFunctions    []int  `json:"supported_functions"`
	Notes                 string `json:"notes,omitempty"`
}

// Config holds the complete configuration
type Config struct {
	Connection     ConnectionConfig         `json:"connection"`
	Modbus         ModbusConfig             `json:"modbus"`
	Testing        TestConfig               `json:"testing"`
	Advanced       AdvancedConfig           `json:"advanced"`
	Logging        LoggingConfig            `json:"logging"`
	DeviceProfiles map[string]DeviceProfile `json:"device_profiles"`
	CurrentProfile string                   `json:"current_profile"`
}

// GetCurrentProfile returns the currently selected device profile
func (c *Config) GetCurrentProfile() (*DeviceProfile, error) {
	if profile, exists := c.DeviceProfiles[c.CurrentProfile]; exists {
		return &profile, nil
	}
	return nil, fmt.Errorf("profile '%s' not found", c.CurrentProfile)
}

// ApplyProfile applies the current device profile settings to the config
func (c *Config) ApplyProfile() error {
	profile, err := c.GetCurrentProfile()
	if err != nil {
		return err
	}

	// Override modbus settings with profile settings
	c.Modbus.SlaveID = profile.SlaveID

	// Override test address settings with profile settings
	if c.Testing.TestAddresses == nil {
		c.Testing.TestAddresses = make(map[string]AddressRange)
	}

	c.Testing.TestAddresses["holding_registers"] = AddressRange{
		StartAddress: profile.HoldingRegistersStart,
		Quantity:     c.Testing.TestAddresses["holding_registers"].Quantity,
	}

	c.Testing.TestAddresses["input_registers"] = AddressRange{
		StartAddress: profile.InputRegistersStart,
		Quantity:     c.Testing.TestAddresses["input_registers"].Quantity,
	}

	c.Testing.TestAddresses["coils"] = AddressRange{
		StartAddress: profile.CoilsStart,
		Quantity:     c.Testing.TestAddresses["coils"].Quantity,
	}

	c.Testing.TestAddresses["discrete_inputs"] = AddressRange{
		StartAddress: profile.DiscreteInputsStart,
		Quantity:     c.Testing.TestAddresses["discrete_inputs"].Quantity,
	}

	return nil
}

// LoadConfig loads configuration from a JSON file
func LoadConfig(configPath string) (*Config, error) {
	// If no path provided, look for config.json in current directory and parent directories
	if configPath == "" {
		configPath = findConfigFile()
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", configPath, err)
	}

	// Apply the current device profile
	if err := config.ApplyProfile(); err != nil {
		return nil, fmt.Errorf("failed to apply device profile: %w", err)
	}

	return &config, nil
}

// findConfigFile searches for config.json in current and parent directories
func findConfigFile() string {
	// Try current directory first
	if _, err := os.Stat("config.json"); err == nil {
		return "config.json"
	}

	// Try parent directories
	dir, _ := os.Getwd()
	for {
		configPath := filepath.Join(dir, "config.json")
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}

		parent := filepath.Dir(dir)
		if parent == dir { // Reached root
			break
		}
		dir = parent
	}

	return "config.json" // Default, will cause error if not found
}

// SaveConfig saves the configuration to a JSON file
func (c *Config) SaveConfig(configPath string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	err = os.WriteFile(configPath, data, 0600)
	if err != nil {
		return fmt.Errorf("failed to write config file %s: %w", configPath, err)
	}

	return nil
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Connection: ConnectionConfig{
			Address:          "localhost",
			Port:             502,
			TimeoutMs:        5000,
			ConnectTimeoutMs: 5000,
			RetryCount:       3,
			TransportType:    "tcp",
		},
		Modbus: ModbusConfig{
			SlaveID:    1,
			UnitID:     1,
			ProtocolID: 0,
		},
		Testing: TestConfig{
			EnabledTests: []string{"read_holding_registers"},
			TestAddresses: map[string]AddressRange{
				"holding_registers": {StartAddress: 0, Quantity: 1},
			},
		},
		Logging: LoggingConfig{
			Level:      "info",
			Verbose:    false,
			ShowTiming: true,
		},
		CurrentProfile: "generic",
	}
}
