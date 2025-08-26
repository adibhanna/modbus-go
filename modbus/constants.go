package modbus

// MODBUS Function Codes
const (
	// Bit Access
	FuncCodeReadCoils          = 0x01
	FuncCodeReadDiscreteInputs = 0x02
	FuncCodeWriteSingleCoil    = 0x05
	FuncCodeWriteMultipleCoils = 0x0F

	// 16-bit Register Access
	FuncCodeReadHoldingRegisters   = 0x03
	FuncCodeReadInputRegisters     = 0x04
	FuncCodeWriteSingleRegister    = 0x06
	FuncCodeWriteMultipleRegisters = 0x10
	FuncCodeMaskWriteRegister      = 0x16
	FuncCodeReadWriteMultipleRegs  = 0x17

	// File Record Access
	FuncCodeReadFileRecord  = 0x14
	FuncCodeWriteFileRecord = 0x15

	// Diagnostics (Serial Line only)
	FuncCodeReadExceptionStatus = 0x07
	FuncCodeDiagnostic          = 0x08
	FuncCodeGetCommEventCounter = 0x0B
	FuncCodeGetCommEventLog     = 0x0C
	FuncCodeReportServerID      = 0x11

	// FIFO Queue
	FuncCodeReadFIFOQueue = 0x18

	// Encapsulated Interface Transport
	FuncCodeEncapsulatedInterface = 0x2B
)

// MODBUS Exception Codes
const (
	ExceptionCodeIllegalFunction     = 0x01
	ExceptionCodeIllegalDataAddress  = 0x02
	ExceptionCodeIllegalDataValue    = 0x03
	ExceptionCodeServerDeviceFailure = 0x04
	ExceptionCodeAcknowledge         = 0x05
	ExceptionCodeServerDeviceBusy    = 0x06
	ExceptionCodeMemoryParityError   = 0x08
	ExceptionCodeGatewayPathUnavail  = 0x0A
	ExceptionCodeGatewayTargetFail   = 0x0B
)

// MEI Types for Encapsulated Interface Transport
const (
	MEITypeCANopenGeneralReference = 0x0D
	MEITypeDeviceIdentification    = 0x0E
)

// Device Identification Read Device ID Codes
const (
	DeviceIDReadBasic    = 0x01
	DeviceIDReadRegular  = 0x02
	DeviceIDReadExtended = 0x03
	DeviceIDReadSpecific = 0x04
)

// Device Identification Object IDs
const (
	DeviceIDVendorName         = 0x00
	DeviceIDProductCode        = 0x01
	DeviceIDMajorMinorRevision = 0x02
	DeviceIDVendorURL          = 0x03
	DeviceIDProductName        = 0x04
	DeviceIDModelName          = 0x05
	DeviceIDUserAppName        = 0x06
)

// Conformity Levels
const (
	ConformityLevelBasicStream        = 0x01
	ConformityLevelRegularStream      = 0x02
	ConformityLevelExtendedStream     = 0x03
	ConformityLevelBasicIndividual    = 0x81
	ConformityLevelRegularIndividual  = 0x82
	ConformityLevelExtendedIndividual = 0x83
)

// MODBUS Protocol Limits
const (
	MaxPDUSize       = 253 // Maximum PDU size in bytes
	MaxTCPADUSize    = 260 // Maximum TCP ADU size (PDU + MBAP)
	MaxSerialADUSize = 256 // Maximum Serial ADU size (PDU + Address + CRC)

	// Quantity limits per function code
	MaxReadCoils            = 2000
	MaxReadDiscreteInputs   = 2000
	MaxReadHoldingRegs      = 125
	MaxReadInputRegs        = 125
	MaxWriteMultipleCoils   = 1968 // 0x7B0
	MaxWriteMultipleRegs    = 123  // 0x7B
	MaxReadWriteRegs        = 125  // Read quantity
	MaxWriteReadWriteRegs   = 121  // Write quantity
	MaxReadFileRecordBytes  = 245  // 0xF5
	MaxWriteFileRecordBytes = 251  // 0xFB
	MaxFIFOCount            = 31
)

// MODBUS TCP/IP specific constants
const (
	MBAPHeaderSize = 7
	MBAPProtocolID = 0x0000
	TCPDefaultPort = 502
)

// Diagnostic Sub-function codes
const (
	DiagSubReturnQueryData           = 0x0000
	DiagSubRestartCommOption         = 0x0001
	DiagSubReturnDiagRegister        = 0x0002
	DiagSubChangeASCIIDelimiter      = 0x0003
	DiagSubForceListenOnlyMode       = 0x0004
	DiagSubClearCounters             = 0x000A
	DiagSubReturnBusMessageCount     = 0x000B
	DiagSubReturnBusCommErrorCount   = 0x000C
	DiagSubReturnBusExceptionCount   = 0x000D
	DiagSubReturnServerMessageCount  = 0x000E
	DiagSubReturnServerNoRespCount   = 0x000F
	DiagSubReturnServerNAKCount      = 0x0010
	DiagSubReturnServerBusyCount     = 0x0011
	DiagSubReturnBusCharOverrunCount = 0x0012
	DiagSubClearOverrunCounter       = 0x0014
)

// Common coil values
const (
	CoilOn  = 0xFF00
	CoilOff = 0x0000
)

// File Record Reference Types
const (
	FileRecordTypeExtended = 0x06
)

// Address ranges
const (
	MinAddress = 0x0000
	MaxAddress = 0xFFFF
)

// Timeout defaults (in milliseconds)
const (
	DefaultResponseTimeout = 1000
	DefaultConnectTimeout  = 5000
)
