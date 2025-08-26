package pdu

import (
	"fmt"

	"github.com/adibhanna/modbusgo/modbus"
)

// ReadCoilsRequest creates a PDU for reading coils
func ReadCoilsRequest(address modbus.Address, quantity modbus.Quantity) (*Request, error) {
	if err := ValidateQuantity(modbus.FuncCodeReadCoils, quantity); err != nil {
		return nil, err
	}
	if err := ValidateAddress(address, quantity); err != nil {
		return nil, err
	}

	data := make([]byte, 4)
	copy(data[0:2], EncodeUint16(uint16(address)))
	copy(data[2:4], EncodeUint16(uint16(quantity)))

	return NewRequest(modbus.FuncCodeReadCoils, data), nil
}

// ReadDiscreteInputsRequest creates a PDU for reading discrete inputs
func ReadDiscreteInputsRequest(address modbus.Address, quantity modbus.Quantity) (*Request, error) {
	if err := ValidateQuantity(modbus.FuncCodeReadDiscreteInputs, quantity); err != nil {
		return nil, err
	}
	if err := ValidateAddress(address, quantity); err != nil {
		return nil, err
	}

	data := make([]byte, 4)
	copy(data[0:2], EncodeUint16(uint16(address)))
	copy(data[2:4], EncodeUint16(uint16(quantity)))

	return NewRequest(modbus.FuncCodeReadDiscreteInputs, data), nil
}

// ReadHoldingRegistersRequest creates a PDU for reading holding registers
func ReadHoldingRegistersRequest(address modbus.Address, quantity modbus.Quantity) (*Request, error) {
	if err := ValidateQuantity(modbus.FuncCodeReadHoldingRegisters, quantity); err != nil {
		return nil, err
	}
	if err := ValidateAddress(address, quantity); err != nil {
		return nil, err
	}

	data := make([]byte, 4)
	copy(data[0:2], EncodeUint16(uint16(address)))
	copy(data[2:4], EncodeUint16(uint16(quantity)))

	return NewRequest(modbus.FuncCodeReadHoldingRegisters, data), nil
}

// ReadInputRegistersRequest creates a PDU for reading input registers
func ReadInputRegistersRequest(address modbus.Address, quantity modbus.Quantity) (*Request, error) {
	if err := ValidateQuantity(modbus.FuncCodeReadInputRegisters, quantity); err != nil {
		return nil, err
	}
	if err := ValidateAddress(address, quantity); err != nil {
		return nil, err
	}

	data := make([]byte, 4)
	copy(data[0:2], EncodeUint16(uint16(address)))
	copy(data[2:4], EncodeUint16(uint16(quantity)))

	return NewRequest(modbus.FuncCodeReadInputRegisters, data), nil
}

// WriteSingleCoilRequest creates a PDU for writing a single coil
func WriteSingleCoilRequest(address modbus.Address, value bool) (*Request, error) {
	coilValue := uint16(modbus.CoilOff)
	if value {
		coilValue = modbus.CoilOn
	}

	data := make([]byte, 4)
	copy(data[0:2], EncodeUint16(uint16(address)))
	copy(data[2:4], EncodeUint16(coilValue))

	return NewRequest(modbus.FuncCodeWriteSingleCoil, data), nil
}

// WriteSingleRegisterRequest creates a PDU for writing a single register
func WriteSingleRegisterRequest(address modbus.Address, value uint16) (*Request, error) {
	data := make([]byte, 4)
	copy(data[0:2], EncodeUint16(uint16(address)))
	copy(data[2:4], EncodeUint16(value))

	return NewRequest(modbus.FuncCodeWriteSingleRegister, data), nil
}

// WriteMultipleCoilsRequest creates a PDU for writing multiple coils
func WriteMultipleCoilsRequest(address modbus.Address, values []bool) (*Request, error) {
	quantity := modbus.Quantity(len(values))
	if err := ValidateQuantity(modbus.FuncCodeWriteMultipleCoils, quantity); err != nil {
		return nil, err
	}
	if err := ValidateAddress(address, quantity); err != nil {
		return nil, err
	}

	coilBytes := EncodeBoolSlice(values)
	byteCount := len(coilBytes)

	data := make([]byte, 5+byteCount)
	copy(data[0:2], EncodeUint16(uint16(address)))
	copy(data[2:4], EncodeUint16(uint16(quantity)))
	data[4] = byte(byteCount)
	copy(data[5:], coilBytes)

	return NewRequest(modbus.FuncCodeWriteMultipleCoils, data), nil
}

// WriteMultipleRegistersRequest creates a PDU for writing multiple registers
func WriteMultipleRegistersRequest(address modbus.Address, values []uint16) (*Request, error) {
	quantity := modbus.Quantity(len(values))
	if err := ValidateQuantity(modbus.FuncCodeWriteMultipleRegisters, quantity); err != nil {
		return nil, err
	}
	if err := ValidateAddress(address, quantity); err != nil {
		return nil, err
	}

	registerBytes := EncodeUint16Slice(values)
	byteCount := len(registerBytes)

	data := make([]byte, 5+byteCount)
	copy(data[0:2], EncodeUint16(uint16(address)))
	copy(data[2:4], EncodeUint16(uint16(quantity)))
	data[4] = byte(byteCount)
	copy(data[5:], registerBytes)

	return NewRequest(modbus.FuncCodeWriteMultipleRegisters, data), nil
}

// MaskWriteRegisterRequest creates a PDU for mask write register
func MaskWriteRegisterRequest(address modbus.Address, andMask, orMask uint16) (*Request, error) {
	data := make([]byte, 6)
	copy(data[0:2], EncodeUint16(uint16(address)))
	copy(data[2:4], EncodeUint16(andMask))
	copy(data[4:6], EncodeUint16(orMask))

	return NewRequest(modbus.FuncCodeMaskWriteRegister, data), nil
}

// ReadWriteMultipleRegistersRequest creates a PDU for read/write multiple registers
func ReadWriteMultipleRegistersRequest(readAddress modbus.Address, readQuantity modbus.Quantity,
	writeAddress modbus.Address, writeValues []uint16) (*Request, error) {
	if readQuantity < 1 || readQuantity > modbus.MaxReadWriteRegs {
		return nil, fmt.Errorf("invalid read quantity %d: must be 1-%d", readQuantity, modbus.MaxReadWriteRegs)
	}

	writeQuantity := modbus.Quantity(len(writeValues))
	if writeQuantity < 1 || writeQuantity > modbus.MaxWriteReadWriteRegs {
		return nil, fmt.Errorf("invalid write quantity %d: must be 1-%d", writeQuantity, modbus.MaxWriteReadWriteRegs)
	}

	if err := ValidateAddress(readAddress, readQuantity); err != nil {
		return nil, fmt.Errorf("read address validation failed: %w", err)
	}
	if err := ValidateAddress(writeAddress, writeQuantity); err != nil {
		return nil, fmt.Errorf("write address validation failed: %w", err)
	}

	writeBytes := EncodeUint16Slice(writeValues)
	writeByteCount := len(writeBytes)

	data := make([]byte, 9+writeByteCount)
	copy(data[0:2], EncodeUint16(uint16(readAddress)))
	copy(data[2:4], EncodeUint16(uint16(readQuantity)))
	copy(data[4:6], EncodeUint16(uint16(writeAddress)))
	copy(data[6:8], EncodeUint16(uint16(writeQuantity)))
	data[8] = byte(writeByteCount)
	copy(data[9:], writeBytes)

	return NewRequest(modbus.FuncCodeReadWriteMultipleRegs, data), nil
}

// ReadFIFOQueueRequest creates a PDU for reading FIFO queue
func ReadFIFOQueueRequest(address modbus.Address) (*Request, error) {
	data := EncodeUint16(uint16(address))
	return NewRequest(modbus.FuncCodeReadFIFOQueue, data), nil
}

// ReadExceptionStatusRequest creates a PDU for reading exception status (Serial line only)
func ReadExceptionStatusRequest() (*Request, error) {
	return NewRequest(modbus.FuncCodeReadExceptionStatus, []byte{}), nil
}

// DiagnosticRequest creates a PDU for diagnostic function (Serial line only)
func DiagnosticRequest(subFunction uint16, data []byte) (*Request, error) {
	reqData := make([]byte, 2+len(data))
	copy(reqData[0:2], EncodeUint16(subFunction))
	copy(reqData[2:], data)
	return NewRequest(modbus.FuncCodeDiagnostic, reqData), nil
}

// GetCommEventCounterRequest creates a PDU for getting comm event counter (Serial line only)
func GetCommEventCounterRequest() (*Request, error) {
	return NewRequest(modbus.FuncCodeGetCommEventCounter, []byte{}), nil
}

// GetCommEventLogRequest creates a PDU for getting comm event log (Serial line only)
func GetCommEventLogRequest() (*Request, error) {
	return NewRequest(modbus.FuncCodeGetCommEventLog, []byte{}), nil
}

// ReportServerIDRequest creates a PDU for reporting server ID (Serial line only)
func ReportServerIDRequest() (*Request, error) {
	return NewRequest(modbus.FuncCodeReportServerID, []byte{}), nil
}

// ReadFileRecordRequest creates a PDU for reading file record
func ReadFileRecordRequest(records []modbus.FileRecord) (*Request, error) {
	if len(records) == 0 {
		return nil, fmt.Errorf("at least one file record must be specified")
	}

	var data []byte
	for _, record := range records {
		subReq := make([]byte, 7)
		subReq[0] = record.ReferenceType
		copy(subReq[1:3], EncodeUint16(record.FileNumber))
		copy(subReq[3:5], EncodeUint16(record.RecordNumber))
		copy(subReq[5:7], EncodeUint16(record.RecordLength))
		data = append(data, subReq...)
	}

	if len(data) > modbus.MaxReadFileRecordBytes {
		return nil, fmt.Errorf("file record request too large: %d bytes, max %d",
			len(data), modbus.MaxReadFileRecordBytes)
	}

	fullData := make([]byte, 1+len(data))
	fullData[0] = byte(len(data))
	copy(fullData[1:], data)

	return NewRequest(modbus.FuncCodeReadFileRecord, fullData), nil
}

// WriteFileRecordRequest creates a PDU for writing file record
func WriteFileRecordRequest(records []modbus.FileRecord) (*Request, error) {
	if len(records) == 0 {
		return nil, fmt.Errorf("at least one file record must be specified")
	}

	var data []byte
	for _, record := range records {
		recordDataBytes := EncodeUint16Slice(record.RecordData)
		subReq := make([]byte, 7+len(recordDataBytes))
		subReq[0] = record.ReferenceType
		copy(subReq[1:3], EncodeUint16(record.FileNumber))
		copy(subReq[3:5], EncodeUint16(record.RecordNumber))
		copy(subReq[5:7], EncodeUint16(record.RecordLength))
		copy(subReq[7:], recordDataBytes)
		data = append(data, subReq...)
	}

	if len(data) > modbus.MaxWriteFileRecordBytes {
		return nil, fmt.Errorf("file record request too large: %d bytes, max %d",
			len(data), modbus.MaxWriteFileRecordBytes)
	}

	fullData := make([]byte, 1+len(data))
	fullData[0] = byte(len(data))
	copy(fullData[1:], data)

	return NewRequest(modbus.FuncCodeWriteFileRecord, fullData), nil
}

// ReadDeviceIdentificationRequest creates a PDU for reading device identification
func ReadDeviceIdentificationRequest(readCode uint8, objectID uint8) (*Request, error) {
	data := []byte{
		modbus.MEITypeDeviceIdentification,
		readCode,
		objectID,
	}
	return NewRequest(modbus.FuncCodeEncapsulatedInterface, data), nil
}
