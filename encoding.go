package modbus

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/adibhanna/modbus-go/modbus"
)

// Endianness represents the byte order for multi-byte values
type Endianness int

const (
	// BigEndian is the default MODBUS byte order (MSB first)
	BigEndian Endianness = iota
	// LittleEndian is LSB first byte order
	LittleEndian
)

// WordOrder represents the word order for multi-register values (32/64 bit)
type WordOrder int

const (
	// HighWordFirst is the default word order (high word at lower address)
	HighWordFirst WordOrder = iota
	// LowWordFirst puts low word at lower address
	LowWordFirst
)

// EncodingConfig holds the encoding configuration for multi-byte/word values
type EncodingConfig struct {
	ByteOrder Endianness
	WordOrder WordOrder
}

// DefaultEncodingConfig returns the default MODBUS encoding (big endian, high word first)
func DefaultEncodingConfig() *EncodingConfig {
	return &EncodingConfig{
		ByteOrder: BigEndian,
		WordOrder: HighWordFirst,
	}
}

// SetEncoding configures the byte and word order for multi-byte values
func (c *Client) SetEncoding(byteOrder Endianness, wordOrder WordOrder) {
	c.encoding = &EncodingConfig{
		ByteOrder: byteOrder,
		WordOrder: wordOrder,
	}
}

// GetEncoding returns the current encoding configuration
func (c *Client) GetEncoding() *EncodingConfig {
	if c.encoding == nil {
		c.encoding = DefaultEncodingConfig()
	}
	return c.encoding
}

// --- Single Value Read Helpers ---

// ReadCoil reads a single coil and returns its boolean value
func (c *Client) ReadCoil(address modbus.Address) (bool, error) {
	values, err := c.ReadCoils(address, 1)
	if err != nil {
		return false, err
	}
	if len(values) == 0 {
		return false, fmt.Errorf("no coil value returned")
	}
	return values[0], nil
}

// ReadDiscreteInput reads a single discrete input and returns its boolean value
func (c *Client) ReadDiscreteInput(address modbus.Address) (bool, error) {
	values, err := c.ReadDiscreteInputs(address, 1)
	if err != nil {
		return false, err
	}
	if len(values) == 0 {
		return false, fmt.Errorf("no discrete input value returned")
	}
	return values[0], nil
}

// ReadHoldingRegister reads a single holding register
func (c *Client) ReadHoldingRegister(address modbus.Address) (uint16, error) {
	values, err := c.ReadHoldingRegisters(address, 1)
	if err != nil {
		return 0, err
	}
	if len(values) == 0 {
		return 0, fmt.Errorf("no register value returned")
	}
	return values[0], nil
}

// ReadInputRegister reads a single input register
func (c *Client) ReadInputRegister(address modbus.Address) (uint16, error) {
	values, err := c.ReadInputRegisters(address, 1)
	if err != nil {
		return 0, err
	}
	if len(values) == 0 {
		return 0, fmt.Errorf("no register value returned")
	}
	return values[0], nil
}

// --- 32-bit Integer Operations ---

// ReadUint32 reads a 32-bit unsigned integer from two consecutive holding registers
func (c *Client) ReadUint32(address modbus.Address) (uint32, error) {
	values, err := c.ReadHoldingRegisters(address, 2)
	if err != nil {
		return 0, err
	}
	return c.decodeUint32(values), nil
}

// ReadUint32s reads multiple 32-bit unsigned integers from holding registers
func (c *Client) ReadUint32s(address modbus.Address, quantity uint16) ([]uint32, error) {
	values, err := c.ReadHoldingRegisters(address, modbus.Quantity(quantity*2))
	if err != nil {
		return nil, err
	}
	result := make([]uint32, quantity)
	for i := uint16(0); i < quantity; i++ {
		result[i] = c.decodeUint32(values[i*2 : i*2+2])
	}
	return result, nil
}

// ReadInt32 reads a 32-bit signed integer from two consecutive holding registers
func (c *Client) ReadInt32(address modbus.Address) (int32, error) {
	val, err := c.ReadUint32(address)
	if err != nil {
		return 0, err
	}
	return int32(val), nil
}

// ReadInt32s reads multiple 32-bit signed integers from holding registers
func (c *Client) ReadInt32s(address modbus.Address, quantity uint16) ([]int32, error) {
	values, err := c.ReadUint32s(address, quantity)
	if err != nil {
		return nil, err
	}
	result := make([]int32, len(values))
	for i, v := range values {
		result[i] = int32(v)
	}
	return result, nil
}

// WriteUint32 writes a 32-bit unsigned integer to two consecutive holding registers
func (c *Client) WriteUint32(address modbus.Address, value uint32) error {
	regs := c.encodeUint32(value)
	return c.WriteMultipleRegisters(address, regs)
}

// WriteUint32s writes multiple 32-bit unsigned integers to holding registers
func (c *Client) WriteUint32s(address modbus.Address, values []uint32) error {
	regs := make([]uint16, len(values)*2)
	for i, v := range values {
		encoded := c.encodeUint32(v)
		regs[i*2] = encoded[0]
		regs[i*2+1] = encoded[1]
	}
	return c.WriteMultipleRegisters(address, regs)
}

// WriteInt32 writes a 32-bit signed integer to two consecutive holding registers
func (c *Client) WriteInt32(address modbus.Address, value int32) error {
	return c.WriteUint32(address, uint32(value))
}

// WriteInt32s writes multiple 32-bit signed integers to holding registers
func (c *Client) WriteInt32s(address modbus.Address, values []int32) error {
	uvals := make([]uint32, len(values))
	for i, v := range values {
		uvals[i] = uint32(v)
	}
	return c.WriteUint32s(address, uvals)
}

// ReadInputUint32 reads a 32-bit unsigned integer from two consecutive input registers
func (c *Client) ReadInputUint32(address modbus.Address) (uint32, error) {
	values, err := c.ReadInputRegisters(address, 2)
	if err != nil {
		return 0, err
	}
	return c.decodeUint32(values), nil
}

// ReadInputUint32s reads multiple 32-bit unsigned integers from input registers
func (c *Client) ReadInputUint32s(address modbus.Address, quantity uint16) ([]uint32, error) {
	values, err := c.ReadInputRegisters(address, modbus.Quantity(quantity*2))
	if err != nil {
		return nil, err
	}
	result := make([]uint32, quantity)
	for i := uint16(0); i < quantity; i++ {
		result[i] = c.decodeUint32(values[i*2 : i*2+2])
	}
	return result, nil
}

// --- 64-bit Integer Operations ---

// ReadUint64 reads a 64-bit unsigned integer from four consecutive holding registers
func (c *Client) ReadUint64(address modbus.Address) (uint64, error) {
	values, err := c.ReadHoldingRegisters(address, 4)
	if err != nil {
		return 0, err
	}
	return c.decodeUint64(values), nil
}

// ReadUint64s reads multiple 64-bit unsigned integers from holding registers
func (c *Client) ReadUint64s(address modbus.Address, quantity uint16) ([]uint64, error) {
	values, err := c.ReadHoldingRegisters(address, modbus.Quantity(quantity*4))
	if err != nil {
		return nil, err
	}
	result := make([]uint64, quantity)
	for i := uint16(0); i < quantity; i++ {
		result[i] = c.decodeUint64(values[i*4 : i*4+4])
	}
	return result, nil
}

// ReadInt64 reads a 64-bit signed integer from four consecutive holding registers
func (c *Client) ReadInt64(address modbus.Address) (int64, error) {
	val, err := c.ReadUint64(address)
	if err != nil {
		return 0, err
	}
	return int64(val), nil
}

// ReadInt64s reads multiple 64-bit signed integers from holding registers
func (c *Client) ReadInt64s(address modbus.Address, quantity uint16) ([]int64, error) {
	values, err := c.ReadUint64s(address, quantity)
	if err != nil {
		return nil, err
	}
	result := make([]int64, len(values))
	for i, v := range values {
		result[i] = int64(v)
	}
	return result, nil
}

// WriteUint64 writes a 64-bit unsigned integer to four consecutive holding registers
func (c *Client) WriteUint64(address modbus.Address, value uint64) error {
	regs := c.encodeUint64(value)
	return c.WriteMultipleRegisters(address, regs)
}

// WriteUint64s writes multiple 64-bit unsigned integers to holding registers
func (c *Client) WriteUint64s(address modbus.Address, values []uint64) error {
	regs := make([]uint16, len(values)*4)
	for i, v := range values {
		encoded := c.encodeUint64(v)
		for j := 0; j < 4; j++ {
			regs[i*4+j] = encoded[j]
		}
	}
	return c.WriteMultipleRegisters(address, regs)
}

// WriteInt64 writes a 64-bit signed integer to four consecutive holding registers
func (c *Client) WriteInt64(address modbus.Address, value int64) error {
	return c.WriteUint64(address, uint64(value))
}

// WriteInt64s writes multiple 64-bit signed integers to holding registers
func (c *Client) WriteInt64s(address modbus.Address, values []int64) error {
	uvals := make([]uint64, len(values))
	for i, v := range values {
		uvals[i] = uint64(v)
	}
	return c.WriteUint64s(address, uvals)
}

// --- Float32 Operations ---

// ReadFloat32 reads a 32-bit float from two consecutive holding registers
func (c *Client) ReadFloat32(address modbus.Address) (float32, error) {
	val, err := c.ReadUint32(address)
	if err != nil {
		return 0, err
	}
	return math.Float32frombits(val), nil
}

// ReadFloat32s reads multiple 32-bit floats from holding registers
func (c *Client) ReadFloat32s(address modbus.Address, quantity uint16) ([]float32, error) {
	values, err := c.ReadUint32s(address, quantity)
	if err != nil {
		return nil, err
	}
	result := make([]float32, len(values))
	for i, v := range values {
		result[i] = math.Float32frombits(v)
	}
	return result, nil
}

// WriteFloat32 writes a 32-bit float to two consecutive holding registers
func (c *Client) WriteFloat32(address modbus.Address, value float32) error {
	return c.WriteUint32(address, math.Float32bits(value))
}

// WriteFloat32s writes multiple 32-bit floats to holding registers
func (c *Client) WriteFloat32s(address modbus.Address, values []float32) error {
	uvals := make([]uint32, len(values))
	for i, v := range values {
		uvals[i] = math.Float32bits(v)
	}
	return c.WriteUint32s(address, uvals)
}

// ReadInputFloat32 reads a 32-bit float from two consecutive input registers
func (c *Client) ReadInputFloat32(address modbus.Address) (float32, error) {
	val, err := c.ReadInputUint32(address)
	if err != nil {
		return 0, err
	}
	return math.Float32frombits(val), nil
}

// ReadInputFloat32s reads multiple 32-bit floats from input registers
func (c *Client) ReadInputFloat32s(address modbus.Address, quantity uint16) ([]float32, error) {
	values, err := c.ReadInputUint32s(address, quantity)
	if err != nil {
		return nil, err
	}
	result := make([]float32, len(values))
	for i, v := range values {
		result[i] = math.Float32frombits(v)
	}
	return result, nil
}

// --- Float64 Operations ---

// ReadFloat64 reads a 64-bit float from four consecutive holding registers
func (c *Client) ReadFloat64(address modbus.Address) (float64, error) {
	val, err := c.ReadUint64(address)
	if err != nil {
		return 0, err
	}
	return math.Float64frombits(val), nil
}

// ReadFloat64s reads multiple 64-bit floats from holding registers
func (c *Client) ReadFloat64s(address modbus.Address, quantity uint16) ([]float64, error) {
	values, err := c.ReadUint64s(address, quantity)
	if err != nil {
		return nil, err
	}
	result := make([]float64, len(values))
	for i, v := range values {
		result[i] = math.Float64frombits(v)
	}
	return result, nil
}

// WriteFloat64 writes a 64-bit float to four consecutive holding registers
func (c *Client) WriteFloat64(address modbus.Address, value float64) error {
	return c.WriteUint64(address, math.Float64bits(value))
}

// WriteFloat64s writes multiple 64-bit floats to holding registers
func (c *Client) WriteFloat64s(address modbus.Address, values []float64) error {
	uvals := make([]uint64, len(values))
	for i, v := range values {
		uvals[i] = math.Float64bits(v)
	}
	return c.WriteUint64s(address, uvals)
}

// --- Byte Operations ---

// ReadBytes reads raw bytes from holding registers
// The byte order depends on the encoding configuration
func (c *Client) ReadBytes(address modbus.Address, byteCount uint16) ([]byte, error) {
	regCount := (byteCount + 1) / 2 // Round up to get enough registers
	values, err := c.ReadHoldingRegisters(address, modbus.Quantity(regCount))
	if err != nil {
		return nil, err
	}

	result := make([]byte, byteCount)
	enc := c.GetEncoding()

	for i := 0; i < len(values) && i*2 < int(byteCount); i++ {
		if enc.ByteOrder == BigEndian {
			if i*2 < int(byteCount) {
				result[i*2] = byte(values[i] >> 8)
			}
			if i*2+1 < int(byteCount) {
				result[i*2+1] = byte(values[i])
			}
		} else {
			if i*2 < int(byteCount) {
				result[i*2] = byte(values[i])
			}
			if i*2+1 < int(byteCount) {
				result[i*2+1] = byte(values[i] >> 8)
			}
		}
	}

	return result, nil
}

// WriteBytes writes raw bytes to holding registers
func (c *Client) WriteBytes(address modbus.Address, data []byte) error {
	regCount := (len(data) + 1) / 2
	regs := make([]uint16, regCount)
	enc := c.GetEncoding()

	for i := 0; i < regCount; i++ {
		var high, low byte
		if i*2 < len(data) {
			if enc.ByteOrder == BigEndian {
				high = data[i*2]
			} else {
				low = data[i*2]
			}
		}
		if i*2+1 < len(data) {
			if enc.ByteOrder == BigEndian {
				low = data[i*2+1]
			} else {
				high = data[i*2+1]
			}
		}
		regs[i] = uint16(high)<<8 | uint16(low)
	}

	return c.WriteMultipleRegisters(address, regs)
}

// ReadInputBytes reads raw bytes from input registers
func (c *Client) ReadInputBytes(address modbus.Address, byteCount uint16) ([]byte, error) {
	regCount := (byteCount + 1) / 2
	values, err := c.ReadInputRegisters(address, modbus.Quantity(regCount))
	if err != nil {
		return nil, err
	}

	result := make([]byte, byteCount)
	enc := c.GetEncoding()

	for i := 0; i < len(values) && i*2 < int(byteCount); i++ {
		if enc.ByteOrder == BigEndian {
			if i*2 < int(byteCount) {
				result[i*2] = byte(values[i] >> 8)
			}
			if i*2+1 < int(byteCount) {
				result[i*2+1] = byte(values[i])
			}
		} else {
			if i*2 < int(byteCount) {
				result[i*2] = byte(values[i])
			}
			if i*2+1 < int(byteCount) {
				result[i*2+1] = byte(values[i] >> 8)
			}
		}
	}

	return result, nil
}

// --- String Operations ---

// ReadString reads a string from holding registers
// The string is read as bytes and trimmed of null characters
func (c *Client) ReadString(address modbus.Address, maxLength uint16) (string, error) {
	data, err := c.ReadBytes(address, maxLength)
	if err != nil {
		return "", err
	}

	// Find null terminator or use full length
	end := len(data)
	for i, b := range data {
		if b == 0 {
			end = i
			break
		}
	}

	return string(data[:end]), nil
}

// WriteString writes a string to holding registers
func (c *Client) WriteString(address modbus.Address, value string, maxLength uint16) error {
	data := make([]byte, maxLength)
	copy(data, value)
	return c.WriteBytes(address, data)
}

// --- Internal Encoding/Decoding Helpers ---

func (c *Client) decodeUint32(regs []uint16) uint32 {
	if len(regs) < 2 {
		return 0
	}

	enc := c.GetEncoding()
	var high, low uint16

	if enc.WordOrder == HighWordFirst {
		high, low = regs[0], regs[1]
	} else {
		high, low = regs[1], regs[0]
	}

	if enc.ByteOrder == BigEndian {
		return uint32(high)<<16 | uint32(low)
	}
	// Little endian: swap bytes within each word
	high = (high >> 8) | (high << 8)
	low = (low >> 8) | (low << 8)
	return uint32(high)<<16 | uint32(low)
}

func (c *Client) encodeUint32(value uint32) []uint16 {
	enc := c.GetEncoding()
	var high, low uint16

	if enc.ByteOrder == BigEndian {
		high = uint16(value >> 16)
		low = uint16(value)
	} else {
		// Little endian: swap bytes within each word
		high = uint16(value >> 16)
		low = uint16(value)
		high = (high >> 8) | (high << 8)
		low = (low >> 8) | (low << 8)
	}

	if enc.WordOrder == HighWordFirst {
		return []uint16{high, low}
	}
	return []uint16{low, high}
}

func (c *Client) decodeUint64(regs []uint16) uint64 {
	if len(regs) < 4 {
		return 0
	}

	enc := c.GetEncoding()
	var words [4]uint16

	if enc.WordOrder == HighWordFirst {
		words = [4]uint16{regs[0], regs[1], regs[2], regs[3]}
	} else {
		words = [4]uint16{regs[3], regs[2], regs[1], regs[0]}
	}

	var result uint64
	if enc.ByteOrder == BigEndian {
		result = uint64(words[0])<<48 | uint64(words[1])<<32 | uint64(words[2])<<16 | uint64(words[3])
	} else {
		for i := range words {
			words[i] = (words[i] >> 8) | (words[i] << 8)
		}
		result = uint64(words[0])<<48 | uint64(words[1])<<32 | uint64(words[2])<<16 | uint64(words[3])
	}

	return result
}

func (c *Client) encodeUint64(value uint64) []uint16 {
	enc := c.GetEncoding()
	var words [4]uint16

	if enc.ByteOrder == BigEndian {
		words[0] = uint16(value >> 48)
		words[1] = uint16(value >> 32)
		words[2] = uint16(value >> 16)
		words[3] = uint16(value)
	} else {
		words[0] = uint16(value >> 48)
		words[1] = uint16(value >> 32)
		words[2] = uint16(value >> 16)
		words[3] = uint16(value)
		for i := range words {
			words[i] = (words[i] >> 8) | (words[i] << 8)
		}
	}

	if enc.WordOrder == HighWordFirst {
		return words[:]
	}
	return []uint16{words[3], words[2], words[1], words[0]}
}

// RegistersToBytes converts register values to bytes using the client's encoding
func (c *Client) RegistersToBytes(regs []uint16) []byte {
	result := make([]byte, len(regs)*2)
	enc := c.GetEncoding()

	for i, reg := range regs {
		if enc.ByteOrder == BigEndian {
			binary.BigEndian.PutUint16(result[i*2:], reg)
		} else {
			binary.LittleEndian.PutUint16(result[i*2:], reg)
		}
	}

	return result
}

// BytesToRegisters converts bytes to register values using the client's encoding
func (c *Client) BytesToRegisters(data []byte) []uint16 {
	regCount := (len(data) + 1) / 2
	result := make([]uint16, regCount)
	enc := c.GetEncoding()

	for i := 0; i < regCount; i++ {
		start := i * 2
		end := start + 2
		if end > len(data) {
			end = len(data)
		}

		buf := make([]byte, 2)
		copy(buf, data[start:end])

		if enc.ByteOrder == BigEndian {
			result[i] = binary.BigEndian.Uint16(buf)
		} else {
			result[i] = binary.LittleEndian.Uint16(buf)
		}
	}

	return result
}
