package modbus

import (
	"github.com/adibhanna/modbusgo/modbus"
)

// Re-export types from modbus package
type (
	SlaveID              = modbus.SlaveID
	Address              = modbus.Address
	Quantity             = modbus.Quantity
	FunctionCode         = modbus.FunctionCode
	ExceptionCode        = modbus.ExceptionCode
	ModbusError          = modbus.ModbusError
	TransportType        = modbus.TransportType
	ClientConfig         = modbus.ClientConfig
	ServerConfig         = modbus.ServerConfig
	DataStore            = modbus.DataStore
	DeviceIdentification = modbus.DeviceIdentification
	FileRecord           = modbus.FileRecord
	DiagnosticData       = modbus.DiagnosticData
)

// Re-export constants from modbus package
const (
	// Function codes
	FuncCodeReadCoils              = modbus.FuncCodeReadCoils
	FuncCodeReadDiscreteInputs     = modbus.FuncCodeReadDiscreteInputs
	FuncCodeReadHoldingRegisters   = modbus.FuncCodeReadHoldingRegisters
	FuncCodeReadInputRegisters     = modbus.FuncCodeReadInputRegisters
	FuncCodeWriteSingleCoil        = modbus.FuncCodeWriteSingleCoil
	FuncCodeWriteSingleRegister    = modbus.FuncCodeWriteSingleRegister
	FuncCodeReadExceptionStatus    = modbus.FuncCodeReadExceptionStatus
	FuncCodeDiagnostic             = modbus.FuncCodeDiagnostic
	FuncCodeGetCommEventCounter    = modbus.FuncCodeGetCommEventCounter
	FuncCodeGetCommEventLog        = modbus.FuncCodeGetCommEventLog
	FuncCodeWriteMultipleCoils     = modbus.FuncCodeWriteMultipleCoils
	FuncCodeWriteMultipleRegisters = modbus.FuncCodeWriteMultipleRegisters
	FuncCodeReportServerID         = modbus.FuncCodeReportServerID
	FuncCodeReadFileRecord         = modbus.FuncCodeReadFileRecord
	FuncCodeWriteFileRecord        = modbus.FuncCodeWriteFileRecord
	FuncCodeMaskWriteRegister      = modbus.FuncCodeMaskWriteRegister
	FuncCodeReadWriteMultipleRegs  = modbus.FuncCodeReadWriteMultipleRegs
	FuncCodeReadFIFOQueue          = modbus.FuncCodeReadFIFOQueue
	FuncCodeEncapsulatedInterface  = modbus.FuncCodeEncapsulatedInterface

	// Exception codes
	ExceptionCodeIllegalFunction     = modbus.ExceptionCodeIllegalFunction
	ExceptionCodeIllegalDataAddress  = modbus.ExceptionCodeIllegalDataAddress
	ExceptionCodeIllegalDataValue    = modbus.ExceptionCodeIllegalDataValue
	ExceptionCodeServerDeviceFailure = modbus.ExceptionCodeServerDeviceFailure
	ExceptionCodeAcknowledge         = modbus.ExceptionCodeAcknowledge
	ExceptionCodeServerDeviceBusy    = modbus.ExceptionCodeServerDeviceBusy
	ExceptionCodeMemoryParityError   = modbus.ExceptionCodeMemoryParityError
	ExceptionCodeGatewayPathUnavail  = modbus.ExceptionCodeGatewayPathUnavail
	ExceptionCodeGatewayTargetFail   = modbus.ExceptionCodeGatewayTargetFail

	// Coil values
	CoilOff = modbus.CoilOff
	CoilOn  = modbus.CoilOn

	// Transport types
	TransportTCP   = modbus.TransportTCP
	TransportRTU   = modbus.TransportRTU
	TransportASCII = modbus.TransportASCII

	// Other constants
	DefaultResponseTimeout      = modbus.DefaultResponseTimeout
	ConformityLevelBasicStream  = modbus.ConformityLevelBasicStream
	MEITypeDeviceIdentification = modbus.MEITypeDeviceIdentification
	DeviceIDVendorName          = modbus.DeviceIDVendorName
	DeviceIDProductCode         = modbus.DeviceIDProductCode
	DeviceIDMajorMinorRevision  = modbus.DeviceIDMajorMinorRevision

	// Device ID Read Codes
	DeviceIDReadBasic    = modbus.DeviceIDReadBasic
	DeviceIDReadRegular  = modbus.DeviceIDReadRegular
	DeviceIDReadExtended = modbus.DeviceIDReadExtended
	DeviceIDReadSpecific = modbus.DeviceIDReadSpecific
)

// Re-export functions from modbus package
var (
	NewModbusError      = modbus.NewModbusError
	DefaultClientConfig = modbus.DefaultClientConfig
	DefaultServerConfig = modbus.DefaultServerConfig
)
