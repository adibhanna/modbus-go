package pdu

import (
	"fmt"

	"github.com/adibhanna/modbusgo/modbus"
)

// ParseReadCoilsResponse parses a response PDU for read coils
func ParseReadCoilsResponse(resp *Response, expectedQuantity modbus.Quantity) ([]bool, error) {
	if resp.IsException() {
		ec, _ := resp.GetExceptionCode()
		return nil, modbus.NewModbusError(resp.FunctionCode.FromException(), ec, "")
	}

	if len(resp.Data) < 1 {
		return nil, fmt.Errorf("invalid read coils response: no byte count")
	}

	byteCount := int(resp.Data[0])
	if len(resp.Data) != 1+byteCount {
		return nil, fmt.Errorf("invalid read coils response: expected %d data bytes, got %d",
			byteCount, len(resp.Data)-1)
	}

	return DecodeBoolSlice(resp.Data[1:], int(expectedQuantity)), nil
}

// ParseReadDiscreteInputsResponse parses a response PDU for read discrete inputs
func ParseReadDiscreteInputsResponse(resp *Response, expectedQuantity modbus.Quantity) ([]bool, error) {
	if resp.IsException() {
		ec, _ := resp.GetExceptionCode()
		return nil, modbus.NewModbusError(resp.FunctionCode.FromException(), ec, "")
	}

	if len(resp.Data) < 1 {
		return nil, fmt.Errorf("invalid read discrete inputs response: no byte count")
	}

	byteCount := int(resp.Data[0])
	if len(resp.Data) != 1+byteCount {
		return nil, fmt.Errorf("invalid read discrete inputs response: expected %d data bytes, got %d",
			byteCount, len(resp.Data)-1)
	}

	return DecodeBoolSlice(resp.Data[1:], int(expectedQuantity)), nil
}

// ParseReadHoldingRegistersResponse parses a response PDU for read holding registers
func ParseReadHoldingRegistersResponse(resp *Response, expectedQuantity modbus.Quantity) ([]uint16, error) {
	if resp.IsException() {
		ec, _ := resp.GetExceptionCode()
		return nil, modbus.NewModbusError(resp.FunctionCode.FromException(), ec, "")
	}

	if len(resp.Data) < 1 {
		return nil, fmt.Errorf("invalid read holding registers response: no byte count")
	}

	byteCount := int(resp.Data[0])
	if len(resp.Data) != 1+byteCount {
		return nil, fmt.Errorf("invalid read holding registers response: expected %d data bytes, got %d",
			byteCount, len(resp.Data)-1)
	}

	if byteCount != int(expectedQuantity)*2 {
		return nil, fmt.Errorf("invalid read holding registers response: expected %d bytes for %d registers, got %d",
			expectedQuantity*2, expectedQuantity, byteCount)
	}

	return DecodeUint16Slice(resp.Data[1:])
}

// ParseReadInputRegistersResponse parses a response PDU for read input registers
func ParseReadInputRegistersResponse(resp *Response, expectedQuantity modbus.Quantity) ([]uint16, error) {
	if resp.IsException() {
		ec, _ := resp.GetExceptionCode()
		return nil, modbus.NewModbusError(resp.FunctionCode.FromException(), ec, "")
	}

	if len(resp.Data) < 1 {
		return nil, fmt.Errorf("invalid read input registers response: no byte count")
	}

	byteCount := int(resp.Data[0])
	if len(resp.Data) != 1+byteCount {
		return nil, fmt.Errorf("invalid read input registers response: expected %d data bytes, got %d",
			byteCount, len(resp.Data)-1)
	}

	if byteCount != int(expectedQuantity)*2 {
		return nil, fmt.Errorf("invalid read input registers response: expected %d bytes for %d registers, got %d",
			expectedQuantity*2, expectedQuantity, byteCount)
	}

	return DecodeUint16Slice(resp.Data[1:])
}

// ParseWriteSingleCoilResponse parses a response PDU for write single coil
func ParseWriteSingleCoilResponse(resp *Response, expectedAddress modbus.Address, expectedValue bool) error {
	if resp.IsException() {
		ec, _ := resp.GetExceptionCode()
		return modbus.NewModbusError(resp.FunctionCode.FromException(), ec, "")
	}

	if len(resp.Data) != 4 {
		return fmt.Errorf("invalid write single coil response: expected 4 bytes, got %d", len(resp.Data))
	}

	address, err := DecodeUint16(resp.Data[0:2])
	if err != nil {
		return fmt.Errorf("invalid write single coil response: %w", err)
	}

	value, err := DecodeUint16(resp.Data[2:4])
	if err != nil {
		return fmt.Errorf("invalid write single coil response: %w", err)
	}

	if address != uint16(expectedAddress) {
		return fmt.Errorf("write single coil response address mismatch: expected %d, got %d",
			expectedAddress, address)
	}

	expectedCoilValue := uint16(modbus.CoilOff)
	if expectedValue {
		expectedCoilValue = modbus.CoilOn
	}

	if value != expectedCoilValue {
		return fmt.Errorf("write single coil response value mismatch: expected %04X, got %04X",
			expectedCoilValue, value)
	}

	return nil
}

// ParseWriteSingleRegisterResponse parses a response PDU for write single register
func ParseWriteSingleRegisterResponse(resp *Response, expectedAddress modbus.Address, expectedValue uint16) error {
	if resp.IsException() {
		ec, _ := resp.GetExceptionCode()
		return modbus.NewModbusError(resp.FunctionCode.FromException(), ec, "")
	}

	if len(resp.Data) != 4 {
		return fmt.Errorf("invalid write single register response: expected 4 bytes, got %d", len(resp.Data))
	}

	address, err := DecodeUint16(resp.Data[0:2])
	if err != nil {
		return fmt.Errorf("invalid write single register response: %w", err)
	}

	value, err := DecodeUint16(resp.Data[2:4])
	if err != nil {
		return fmt.Errorf("invalid write single register response: %w", err)
	}

	if address != uint16(expectedAddress) {
		return fmt.Errorf("write single register response address mismatch: expected %d, got %d",
			expectedAddress, address)
	}

	if value != expectedValue {
		return fmt.Errorf("write single register response value mismatch: expected %d, got %d",
			expectedValue, value)
	}

	return nil
}

// ParseWriteMultipleCoilsResponse parses a response PDU for write multiple coils
func ParseWriteMultipleCoilsResponse(resp *Response, expectedAddress modbus.Address, expectedQuantity modbus.Quantity) error {
	if resp.IsException() {
		ec, _ := resp.GetExceptionCode()
		return modbus.NewModbusError(resp.FunctionCode.FromException(), ec, "")
	}

	if len(resp.Data) != 4 {
		return fmt.Errorf("invalid write multiple coils response: expected 4 bytes, got %d", len(resp.Data))
	}

	address, err := DecodeUint16(resp.Data[0:2])
	if err != nil {
		return fmt.Errorf("invalid write multiple coils response: %w", err)
	}

	quantity, err := DecodeUint16(resp.Data[2:4])
	if err != nil {
		return fmt.Errorf("invalid write multiple coils response: %w", err)
	}

	if address != uint16(expectedAddress) {
		return fmt.Errorf("write multiple coils response address mismatch: expected %d, got %d",
			expectedAddress, address)
	}

	if quantity != uint16(expectedQuantity) {
		return fmt.Errorf("write multiple coils response quantity mismatch: expected %d, got %d",
			expectedQuantity, quantity)
	}

	return nil
}

// ParseWriteMultipleRegistersResponse parses a response PDU for write multiple registers
func ParseWriteMultipleRegistersResponse(resp *Response, expectedAddress modbus.Address, expectedQuantity modbus.Quantity) error {
	if resp.IsException() {
		ec, _ := resp.GetExceptionCode()
		return modbus.NewModbusError(resp.FunctionCode.FromException(), ec, "")
	}

	if len(resp.Data) != 4 {
		return fmt.Errorf("invalid write multiple registers response: expected 4 bytes, got %d", len(resp.Data))
	}

	address, err := DecodeUint16(resp.Data[0:2])
	if err != nil {
		return fmt.Errorf("invalid write multiple registers response: %w", err)
	}

	quantity, err := DecodeUint16(resp.Data[2:4])
	if err != nil {
		return fmt.Errorf("invalid write multiple registers response: %w", err)
	}

	if address != uint16(expectedAddress) {
		return fmt.Errorf("write multiple registers response address mismatch: expected %d, got %d",
			expectedAddress, address)
	}

	if quantity != uint16(expectedQuantity) {
		return fmt.Errorf("write multiple registers response quantity mismatch: expected %d, got %d",
			expectedQuantity, quantity)
	}

	return nil
}

// ParseReadWriteMultipleRegistersResponse parses a response PDU for read/write multiple registers
func ParseReadWriteMultipleRegistersResponse(resp *Response, expectedReadQuantity modbus.Quantity) ([]uint16, error) {
	if resp.IsException() {
		ec, _ := resp.GetExceptionCode()
		return nil, modbus.NewModbusError(resp.FunctionCode.FromException(), ec, "")
	}

	if len(resp.Data) < 1 {
		return nil, fmt.Errorf("invalid read/write multiple registers response: no byte count")
	}

	byteCount := int(resp.Data[0])
	if len(resp.Data) != 1+byteCount {
		return nil, fmt.Errorf("invalid read/write multiple registers response: expected %d data bytes, got %d",
			byteCount, len(resp.Data)-1)
	}

	if byteCount != int(expectedReadQuantity)*2 {
		return nil, fmt.Errorf("invalid read/write multiple registers response: expected %d bytes for %d registers, got %d",
			expectedReadQuantity*2, expectedReadQuantity, byteCount)
	}

	return DecodeUint16Slice(resp.Data[1:])
}

// ParseMaskWriteRegisterResponse parses a response PDU for mask write register
func ParseMaskWriteRegisterResponse(resp *Response, expectedAddress modbus.Address, expectedAndMask, expectedOrMask uint16) error {
	if resp.IsException() {
		ec, _ := resp.GetExceptionCode()
		return modbus.NewModbusError(resp.FunctionCode.FromException(), ec, "")
	}

	if len(resp.Data) != 6 {
		return fmt.Errorf("invalid mask write register response: expected 6 bytes, got %d", len(resp.Data))
	}

	address, err := DecodeUint16(resp.Data[0:2])
	if err != nil {
		return fmt.Errorf("invalid mask write register response: %w", err)
	}

	andMask, err := DecodeUint16(resp.Data[2:4])
	if err != nil {
		return fmt.Errorf("invalid mask write register response: %w", err)
	}

	orMask, err := DecodeUint16(resp.Data[4:6])
	if err != nil {
		return fmt.Errorf("invalid mask write register response: %w", err)
	}

	if address != uint16(expectedAddress) {
		return fmt.Errorf("mask write register response address mismatch: expected %d, got %d",
			expectedAddress, address)
	}

	if andMask != expectedAndMask {
		return fmt.Errorf("mask write register response AND mask mismatch: expected %04X, got %04X",
			expectedAndMask, andMask)
	}

	if orMask != expectedOrMask {
		return fmt.Errorf("mask write register response OR mask mismatch: expected %04X, got %04X",
			expectedOrMask, orMask)
	}

	return nil
}

// ParseReadFIFOQueueResponse parses a response PDU for read FIFO queue
func ParseReadFIFOQueueResponse(resp *Response) ([]uint16, error) {
	if resp.IsException() {
		ec, _ := resp.GetExceptionCode()
		return nil, modbus.NewModbusError(resp.FunctionCode.FromException(), ec, "")
	}

	if len(resp.Data) < 4 {
		return nil, fmt.Errorf("invalid read FIFO queue response: need at least 4 bytes")
	}

	byteCount, err := DecodeUint16(resp.Data[0:2])
	if err != nil {
		return nil, fmt.Errorf("invalid read FIFO queue response: %w", err)
	}

	if len(resp.Data) != int(byteCount)+2 {
		return nil, fmt.Errorf("invalid read FIFO queue response: expected %d total bytes, got %d",
			byteCount+2, len(resp.Data))
	}

	fifoCount, err := DecodeUint16(resp.Data[2:4])
	if err != nil {
		return nil, fmt.Errorf("invalid read FIFO queue response: %w", err)
	}

	if fifoCount > modbus.MaxFIFOCount {
		return nil, fmt.Errorf("invalid FIFO count: %d, max allowed: %d", fifoCount, modbus.MaxFIFOCount)
	}

	if fifoCount == 0 {
		return []uint16{}, nil
	}

	expectedDataBytes := int(fifoCount) * 2
	if len(resp.Data[4:]) != expectedDataBytes {
		return nil, fmt.Errorf("invalid read FIFO queue response: expected %d data bytes, got %d",
			expectedDataBytes, len(resp.Data[4:]))
	}

	return DecodeUint16Slice(resp.Data[4:])
}

// ParseReadExceptionStatusResponse parses a response PDU for read exception status
func ParseReadExceptionStatusResponse(resp *Response) (uint8, error) {
	if resp.IsException() {
		ec, _ := resp.GetExceptionCode()
		return 0, modbus.NewModbusError(resp.FunctionCode.FromException(), ec, "")
	}

	if len(resp.Data) != 1 {
		return 0, fmt.Errorf("invalid read exception status response: expected 1 byte, got %d", len(resp.Data))
	}

	return resp.Data[0], nil
}

// ParseDiagnosticResponse parses a response PDU for diagnostic function
func ParseDiagnosticResponse(resp *Response) (uint16, []byte, error) {
	if resp.IsException() {
		ec, _ := resp.GetExceptionCode()
		return 0, nil, modbus.NewModbusError(resp.FunctionCode.FromException(), ec, "")
	}

	if len(resp.Data) < 2 {
		return 0, nil, fmt.Errorf("invalid diagnostic response: need at least 2 bytes")
	}

	subFunction, err := DecodeUint16(resp.Data[0:2])
	if err != nil {
		return 0, nil, fmt.Errorf("invalid diagnostic response: %w", err)
	}

	data := make([]byte, len(resp.Data)-2)
	copy(data, resp.Data[2:])

	return subFunction, data, nil
}

// ParseReadDeviceIdentificationResponse parses a response PDU for read device identification
func ParseReadDeviceIdentificationResponse(resp *Response) (*modbus.DeviceIdentification, bool, uint8, error) {
	if resp.IsException() {
		ec, _ := resp.GetExceptionCode()
		return nil, false, 0, modbus.NewModbusError(resp.FunctionCode.FromException(), ec, "")
	}

	if len(resp.Data) < 6 {
		return nil, false, 0, fmt.Errorf("invalid read device identification response: need at least 6 bytes")
	}

	meiType := resp.Data[0]
	if meiType != modbus.MEITypeDeviceIdentification {
		return nil, false, 0, fmt.Errorf("invalid MEI type: expected %02X, got %02X",
			modbus.MEITypeDeviceIdentification, meiType)
	}

	_ = resp.Data[1] // readDevIDCode - not used in response parsing
	conformityLevel := resp.Data[2]
	moreFollows := resp.Data[3] != 0x00
	nextObjectID := resp.Data[4]
	numberOfObjects := resp.Data[5]

	deviceID := &modbus.DeviceIdentification{
		ConformityLevel: conformityLevel,
	}

	offset := 6
	for i := uint8(0); i < numberOfObjects && offset < len(resp.Data); i++ {
		if offset+2 >= len(resp.Data) {
			break
		}

		objectID := resp.Data[offset]
		objectLength := resp.Data[offset+1]
		offset += 2

		if offset+int(objectLength) > len(resp.Data) {
			break
		}

		objectValue := string(resp.Data[offset : offset+int(objectLength)])
		offset += int(objectLength)

		switch objectID {
		case modbus.DeviceIDVendorName:
			deviceID.VendorName = objectValue
		case modbus.DeviceIDProductCode:
			deviceID.ProductCode = objectValue
		case modbus.DeviceIDMajorMinorRevision:
			deviceID.MajorMinorRevision = objectValue
		case modbus.DeviceIDVendorURL:
			deviceID.VendorURL = objectValue
		case modbus.DeviceIDProductName:
			deviceID.ProductName = objectValue
		case modbus.DeviceIDModelName:
			deviceID.ModelName = objectValue
		case modbus.DeviceIDUserAppName:
			deviceID.UserApplicationName = objectValue
		}
	}

	return deviceID, moreFollows, nextObjectID, nil
}
