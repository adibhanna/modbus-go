package pdu

import (
	"encoding/binary"
	"fmt"

	"github.com/adibhanna/modbusgo/modbus"
)

// PDU represents a MODBUS Protocol Data Unit
type PDU struct {
	FunctionCode modbus.FunctionCode
	Data         []byte
}

// NewPDU creates a new PDU with the given function code and data
func NewPDU(functionCode modbus.FunctionCode, data []byte) *PDU {
	return &PDU{
		FunctionCode: functionCode,
		Data:         data,
	}
}

// Bytes returns the PDU as a byte slice
func (p *PDU) Bytes() []byte {
	result := make([]byte, 1+len(p.Data))
	result[0] = byte(p.FunctionCode)
	copy(result[1:], p.Data)
	return result
}

// Size returns the total size of the PDU in bytes
func (p *PDU) Size() int {
	return 1 + len(p.Data)
}

// IsException returns true if this is an exception response PDU
func (p *PDU) IsException() bool {
	return p.FunctionCode.IsException()
}

// GetExceptionCode returns the exception code if this is an exception PDU
func (p *PDU) GetExceptionCode() (modbus.ExceptionCode, error) {
	if !p.IsException() {
		return 0, fmt.Errorf("PDU is not an exception response")
	}
	if len(p.Data) < 1 {
		return 0, fmt.Errorf("invalid exception PDU: no exception code")
	}
	return modbus.ExceptionCode(p.Data[0]), nil
}

// ParsePDU parses a byte slice into a PDU
func ParsePDU(data []byte) (*PDU, error) {
	if len(data) < 1 {
		return nil, fmt.Errorf("PDU too short: need at least 1 byte")
	}

	functionCode := modbus.FunctionCode(data[0])
	pduData := make([]byte, len(data)-1)
	copy(pduData, data[1:])

	return &PDU{
		FunctionCode: functionCode,
		Data:         pduData,
	}, nil
}

// CreateExceptionPDU creates an exception response PDU
func CreateExceptionPDU(functionCode modbus.FunctionCode, exceptionCode modbus.ExceptionCode) *PDU {
	return &PDU{
		FunctionCode: functionCode.ToException(),
		Data:         []byte{byte(exceptionCode)},
	}
}

// Request represents a MODBUS request PDU
type Request struct {
	*PDU
}

// NewRequest creates a new request PDU
func NewRequest(functionCode modbus.FunctionCode, data []byte) *Request {
	return &Request{
		PDU: NewPDU(functionCode, data),
	}
}

// Response represents a MODBUS response PDU
type Response struct {
	*PDU
}

// NewResponse creates a new response PDU
func NewResponse(functionCode modbus.FunctionCode, data []byte) *Response {
	return &Response{
		PDU: NewPDU(functionCode, data),
	}
}

// NewExceptionResponse creates a new exception response PDU
func NewExceptionResponse(functionCode modbus.FunctionCode, exceptionCode modbus.ExceptionCode) *Response {
	return &Response{
		PDU: CreateExceptionPDU(functionCode, exceptionCode),
	}
}

// Helper functions for encoding/decoding common data types

// EncodeUint16 encodes a uint16 value in big-endian format
func EncodeUint16(value uint16) []byte {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, value)
	return buf
}

// DecodeUint16 decodes a big-endian uint16 value
func DecodeUint16(data []byte) (uint16, error) {
	if len(data) < 2 {
		return 0, fmt.Errorf("insufficient data for uint16: need 2 bytes, got %d", len(data))
	}
	return binary.BigEndian.Uint16(data), nil
}

// EncodeUint16Slice encodes a slice of uint16 values in big-endian format
func EncodeUint16Slice(values []uint16) []byte {
	buf := make([]byte, len(values)*2)
	for i, value := range values {
		binary.BigEndian.PutUint16(buf[i*2:], value)
	}
	return buf
}

// DecodeUint16Slice decodes a slice of big-endian uint16 values
func DecodeUint16Slice(data []byte) ([]uint16, error) {
	if len(data)%2 != 0 {
		return nil, fmt.Errorf("invalid data length for uint16 slice: must be even, got %d", len(data))
	}

	count := len(data) / 2
	values := make([]uint16, count)

	for i := 0; i < count; i++ {
		values[i] = binary.BigEndian.Uint16(data[i*2:])
	}

	return values, nil
}

// EncodeBoolSlice encodes a slice of bool values as a bit-packed byte slice
func EncodeBoolSlice(values []bool) []byte {
	if len(values) == 0 {
		return []byte{}
	}

	// Calculate number of bytes needed
	byteCount := (len(values) + 7) / 8
	result := make([]byte, byteCount)

	for i, value := range values {
		if value {
			byteIndex := i / 8
			bitIndex := i % 8
			result[byteIndex] |= 1 << bitIndex
		}
	}

	return result
}

// DecodeBoolSlice decodes a bit-packed byte slice to a slice of bool values
func DecodeBoolSlice(data []byte, count int) []bool {
	result := make([]bool, count)

	for i := 0; i < count && i < len(data)*8; i++ {
		byteIndex := i / 8
		bitIndex := i % 8
		result[i] = (data[byteIndex] & (1 << bitIndex)) != 0
	}

	return result
}

// ValidateQuantity validates that a quantity is within acceptable limits for a function code
func ValidateQuantity(functionCode modbus.FunctionCode, quantity modbus.Quantity) error {
	switch functionCode {
	case modbus.FuncCodeReadCoils, modbus.FuncCodeReadDiscreteInputs:
		if quantity < 1 || quantity > modbus.MaxReadCoils {
			return fmt.Errorf("invalid quantity %d for %s: must be 1-%d",
				quantity, functionCode.String(), modbus.MaxReadCoils)
		}
	case modbus.FuncCodeReadHoldingRegisters, modbus.FuncCodeReadInputRegisters:
		if quantity < 1 || quantity > modbus.MaxReadHoldingRegs {
			return fmt.Errorf("invalid quantity %d for %s: must be 1-%d",
				quantity, functionCode.String(), modbus.MaxReadHoldingRegs)
		}
	case modbus.FuncCodeWriteMultipleCoils:
		if quantity < 1 || quantity > modbus.MaxWriteMultipleCoils {
			return fmt.Errorf("invalid quantity %d for %s: must be 1-%d",
				quantity, functionCode.String(), modbus.MaxWriteMultipleCoils)
		}
	case modbus.FuncCodeWriteMultipleRegisters:
		if quantity < 1 || quantity > modbus.MaxWriteMultipleRegs {
			return fmt.Errorf("invalid quantity %d for %s: must be 1-%d",
				quantity, functionCode.String(), modbus.MaxWriteMultipleRegs)
		}
	}
	return nil
}

// ValidateAddress validates that an address and quantity combination is valid
func ValidateAddress(address modbus.Address, quantity modbus.Quantity) error {
	if uint32(address)+uint32(quantity) > 0x10000 {
		return fmt.Errorf("address range overflow: address %d + quantity %d exceeds 65535",
			address, quantity)
	}
	return nil
}
