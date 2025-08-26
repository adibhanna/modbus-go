package modbus

import (
	"encoding/binary"
	"fmt"
	"sync"

	"github.com/adibhanna/modbusgo/modbus"
	"github.com/adibhanna/modbusgo/pdu"
	"github.com/adibhanna/modbusgo/transport"
)

// Server represents a MODBUS server
type Server struct {
	transport  transport.RequestHandler
	dataStore  modbus.DataStore
	slaveID    modbus.SlaveID
	deviceInfo *modbus.DeviceIdentification
	mutex      sync.RWMutex
}

// DefaultDataStore provides a simple in-memory data store
type DefaultDataStore struct {
	coils            []bool
	discreteInputs   []bool
	holdingRegisters []uint16
	inputRegisters   []uint16
	fileRecords      map[uint16]map[uint16][]uint16 // fileNumber -> recordNumber -> data
	fifoQueues       map[uint16][]uint16            // address -> queue data
	exceptionStatus  uint8
	diagnosticData   modbus.DiagnosticData
	commEventLog     []byte
	mutex            sync.RWMutex
}

// NewDefaultDataStore creates a new default data store with the given sizes
func NewDefaultDataStore(coilCount, discreteInputCount, holdingRegCount, inputRegCount int) *DefaultDataStore {
	return &DefaultDataStore{
		coils:            make([]bool, coilCount),
		discreteInputs:   make([]bool, discreteInputCount),
		holdingRegisters: make([]uint16, holdingRegCount),
		inputRegisters:   make([]uint16, inputRegCount),
		fileRecords:      make(map[uint16]map[uint16][]uint16),
		fifoQueues:       make(map[uint16][]uint16),
		exceptionStatus:  0,
		diagnosticData:   modbus.DiagnosticData{},
		commEventLog:     make([]byte, 0, 64),
	}
}

// ReadCoils implements modbus.DataStore
func (ds *DefaultDataStore) ReadCoils(address modbus.Address, quantity modbus.Quantity) ([]bool, error) {
	ds.mutex.RLock()
	defer ds.mutex.RUnlock()

	start := int(address)
	end := start + int(quantity)

	if start < 0 || end > len(ds.coils) {
		return nil, modbus.NewModbusError(modbus.FuncCodeReadCoils, modbus.ExceptionCodeIllegalDataAddress,
			fmt.Sprintf("address range %d-%d out of bounds (0-%d)", start, end-1, len(ds.coils)-1))
	}

	result := make([]bool, quantity)
	copy(result, ds.coils[start:end])
	return result, nil
}

// WriteCoils implements modbus.DataStore
func (ds *DefaultDataStore) WriteCoils(address modbus.Address, values []bool) error {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	start := int(address)
	end := start + len(values)

	if start < 0 || end > len(ds.coils) {
		return modbus.NewModbusError(modbus.FuncCodeWriteMultipleCoils, modbus.ExceptionCodeIllegalDataAddress,
			fmt.Sprintf("address range %d-%d out of bounds (0-%d)", start, end-1, len(ds.coils)-1))
	}

	copy(ds.coils[start:end], values)
	return nil
}

// ReadDiscreteInputs implements modbus.DataStore
func (ds *DefaultDataStore) ReadDiscreteInputs(address modbus.Address, quantity modbus.Quantity) ([]bool, error) {
	ds.mutex.RLock()
	defer ds.mutex.RUnlock()

	start := int(address)
	end := start + int(quantity)

	if start < 0 || end > len(ds.discreteInputs) {
		return nil, modbus.NewModbusError(modbus.FuncCodeReadDiscreteInputs, modbus.ExceptionCodeIllegalDataAddress,
			fmt.Sprintf("address range %d-%d out of bounds (0-%d)", start, end-1, len(ds.discreteInputs)-1))
	}

	result := make([]bool, quantity)
	copy(result, ds.discreteInputs[start:end])
	return result, nil
}

// ReadHoldingRegisters implements modbus.DataStore
func (ds *DefaultDataStore) ReadHoldingRegisters(address modbus.Address, quantity modbus.Quantity) ([]uint16, error) {
	ds.mutex.RLock()
	defer ds.mutex.RUnlock()

	start := int(address)
	end := start + int(quantity)

	if start < 0 || end > len(ds.holdingRegisters) {
		return nil, modbus.NewModbusError(modbus.FuncCodeReadHoldingRegisters, modbus.ExceptionCodeIllegalDataAddress,
			fmt.Sprintf("address range %d-%d out of bounds (0-%d)", start, end-1, len(ds.holdingRegisters)-1))
	}

	result := make([]uint16, quantity)
	copy(result, ds.holdingRegisters[start:end])
	return result, nil
}

// WriteHoldingRegisters implements modbus.DataStore
func (ds *DefaultDataStore) WriteHoldingRegisters(address modbus.Address, values []uint16) error {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	start := int(address)
	end := start + len(values)

	if start < 0 || end > len(ds.holdingRegisters) {
		return modbus.NewModbusError(modbus.FuncCodeWriteMultipleRegisters, modbus.ExceptionCodeIllegalDataAddress,
			fmt.Sprintf("address range %d-%d out of bounds (0-%d)", start, end-1, len(ds.holdingRegisters)-1))
	}

	copy(ds.holdingRegisters[start:end], values)
	return nil
}

// ReadInputRegisters implements modbus.DataStore
func (ds *DefaultDataStore) ReadInputRegisters(address modbus.Address, quantity modbus.Quantity) ([]uint16, error) {
	ds.mutex.RLock()
	defer ds.mutex.RUnlock()

	start := int(address)
	end := start + int(quantity)

	if start < 0 || end > len(ds.inputRegisters) {
		return nil, modbus.NewModbusError(modbus.FuncCodeReadInputRegisters, modbus.ExceptionCodeIllegalDataAddress,
			fmt.Sprintf("address range %d-%d out of bounds (0-%d)", start, end-1, len(ds.inputRegisters)-1))
	}

	result := make([]uint16, quantity)
	copy(result, ds.inputRegisters[start:end])
	return result, nil
}

// SetCoil sets a single coil value
func (ds *DefaultDataStore) SetCoil(address modbus.Address, value bool) error {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	if int(address) >= len(ds.coils) {
		return fmt.Errorf("coil address %d out of bounds (0-%d)", address, len(ds.coils)-1)
	}

	ds.coils[address] = value
	return nil
}

// SetDiscreteInput sets a single discrete input value
func (ds *DefaultDataStore) SetDiscreteInput(address modbus.Address, value bool) error {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	if int(address) >= len(ds.discreteInputs) {
		return fmt.Errorf("discrete input address %d out of bounds (0-%d)", address, len(ds.discreteInputs)-1)
	}

	ds.discreteInputs[address] = value
	return nil
}

// SetHoldingRegister sets a single holding register value
func (ds *DefaultDataStore) SetHoldingRegister(address modbus.Address, value uint16) error {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	if int(address) >= len(ds.holdingRegisters) {
		return fmt.Errorf("holding register address %d out of bounds (0-%d)", address, len(ds.holdingRegisters)-1)
	}

	ds.holdingRegisters[address] = value
	return nil
}

// SetInputRegister sets a single input register value
func (ds *DefaultDataStore) SetInputRegister(address modbus.Address, value uint16) error {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	if int(address) >= len(ds.inputRegisters) {
		return fmt.Errorf("input register address %d out of bounds (0-%d)", address, len(ds.inputRegisters)-1)
	}

	ds.inputRegisters[address] = value
	return nil
}

// ReadFileRecords implements modbus.DataStore
func (ds *DefaultDataStore) ReadFileRecords(records []modbus.FileRecord) ([]modbus.FileRecord, error) {
	ds.mutex.RLock()
	defer ds.mutex.RUnlock()

	result := make([]modbus.FileRecord, 0, len(records))
	for _, record := range records {
		if record.ReferenceType != modbus.FileRecordTypeExtended {
			return nil, modbus.NewModbusError(modbus.FuncCodeReadFileRecord, modbus.ExceptionCodeIllegalDataValue,
				fmt.Sprintf("unsupported reference type %d", record.ReferenceType))
		}

		fileMap, exists := ds.fileRecords[record.FileNumber]
		if !exists {
			return nil, modbus.NewModbusError(modbus.FuncCodeReadFileRecord, modbus.ExceptionCodeIllegalDataAddress,
				fmt.Sprintf("file number %d not found", record.FileNumber))
		}

		recordData, exists := fileMap[record.RecordNumber]
		if !exists || uint16(len(recordData)) < record.RecordLength {
			return nil, modbus.NewModbusError(modbus.FuncCodeReadFileRecord, modbus.ExceptionCodeIllegalDataAddress,
				fmt.Sprintf("record %d in file %d not found or too short", record.RecordNumber, record.FileNumber))
		}

		resultRecord := modbus.FileRecord{
			ReferenceType: record.ReferenceType,
			FileNumber:    record.FileNumber,
			RecordNumber:  record.RecordNumber,
			RecordLength:  record.RecordLength,
			RecordData:    make([]uint16, record.RecordLength),
		}
		copy(resultRecord.RecordData, recordData[:record.RecordLength])
		result = append(result, resultRecord)
	}

	return result, nil
}

// WriteFileRecords implements modbus.DataStore
func (ds *DefaultDataStore) WriteFileRecords(records []modbus.FileRecord) error {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	for _, record := range records {
		if record.ReferenceType != modbus.FileRecordTypeExtended {
			return modbus.NewModbusError(modbus.FuncCodeWriteFileRecord, modbus.ExceptionCodeIllegalDataValue,
				fmt.Sprintf("unsupported reference type %d", record.ReferenceType))
		}

		fileMap, exists := ds.fileRecords[record.FileNumber]
		if !exists {
			fileMap = make(map[uint16][]uint16)
			ds.fileRecords[record.FileNumber] = fileMap
		}

		fileMap[record.RecordNumber] = make([]uint16, len(record.RecordData))
		copy(fileMap[record.RecordNumber], record.RecordData)
	}

	return nil
}

// ReadFIFOQueue implements modbus.DataStore
func (ds *DefaultDataStore) ReadFIFOQueue(address modbus.Address) ([]uint16, error) {
	ds.mutex.RLock()
	defer ds.mutex.RUnlock()

	queue, exists := ds.fifoQueues[uint16(address)]
	if !exists {
		// Return empty queue if not exists
		return []uint16{}, nil
	}

	// Return a copy of the queue
	result := make([]uint16, len(queue))
	copy(result, queue)
	return result, nil
}

// WriteFIFOQueue writes data to a FIFO queue (helper method)
func (ds *DefaultDataStore) WriteFIFOQueue(address modbus.Address, values []uint16) error {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	if len(values) > modbus.MaxFIFOCount {
		return modbus.NewModbusError(modbus.FuncCodeReadFIFOQueue, modbus.ExceptionCodeIllegalDataValue,
			fmt.Sprintf("FIFO queue size %d exceeds maximum %d", len(values), modbus.MaxFIFOCount))
	}

	ds.fifoQueues[uint16(address)] = make([]uint16, len(values))
	copy(ds.fifoQueues[uint16(address)], values)
	return nil
}

// ReadExceptionStatus implements modbus.DataStore
func (ds *DefaultDataStore) ReadExceptionStatus() (uint8, error) {
	ds.mutex.RLock()
	defer ds.mutex.RUnlock()
	return ds.exceptionStatus, nil
}

// SetExceptionStatus sets the exception status (helper method)
func (ds *DefaultDataStore) SetExceptionStatus(status uint8) {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()
	ds.exceptionStatus = status
}

// GetDiagnosticData implements modbus.DataStore
func (ds *DefaultDataStore) GetDiagnosticData(subFunction uint16, data []byte) ([]byte, error) {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	switch subFunction {
	case modbus.DiagSubReturnQueryData:
		// Echo back the query data
		return data, nil

	case modbus.DiagSubRestartCommOption:
		// Clear event log
		ds.commEventLog = ds.commEventLog[:0]
		ds.diagnosticData = modbus.DiagnosticData{}
		return data, nil

	case modbus.DiagSubReturnDiagRegister:
		// Return diagnostic register (16-bit value)
		result := make([]byte, 2)
		result[0] = 0x00 // Diagnostic register high byte
		result[1] = 0x00 // Diagnostic register low byte
		return result, nil

	case modbus.DiagSubClearCounters:
		// Clear all counters and diagnostic register
		ds.diagnosticData = modbus.DiagnosticData{}
		return data, nil

	case modbus.DiagSubReturnBusMessageCount:
		return pdu.EncodeUint16(ds.diagnosticData.BusMessageCount), nil

	case modbus.DiagSubReturnBusCommErrorCount:
		return pdu.EncodeUint16(ds.diagnosticData.BusCommErrorCount), nil

	case modbus.DiagSubReturnBusExceptionCount:
		return pdu.EncodeUint16(ds.diagnosticData.BusExceptionCount), nil

	case modbus.DiagSubReturnServerMessageCount:
		return pdu.EncodeUint16(ds.diagnosticData.ServerMessageCount), nil

	case modbus.DiagSubReturnServerNoRespCount:
		return pdu.EncodeUint16(ds.diagnosticData.ServerNoRespCount), nil

	case modbus.DiagSubReturnServerNAKCount:
		return pdu.EncodeUint16(ds.diagnosticData.ServerNAKCount), nil

	case modbus.DiagSubReturnServerBusyCount:
		return pdu.EncodeUint16(ds.diagnosticData.ServerBusyCount), nil

	case modbus.DiagSubReturnBusCharOverrunCount:
		return pdu.EncodeUint16(ds.diagnosticData.BusCharOverrunCount), nil

	default:
		return nil, modbus.NewModbusError(modbus.FuncCodeDiagnostic, modbus.ExceptionCodeIllegalFunction,
			fmt.Sprintf("unsupported diagnostic sub-function %d", subFunction))
	}
}

// GetCommEventCounter implements modbus.DataStore
func (ds *DefaultDataStore) GetCommEventCounter() (uint16, uint16, error) {
	ds.mutex.RLock()
	defer ds.mutex.RUnlock()

	// Status: 0xFFFF = Ready, 0x0000 = Not Ready
	status := uint16(0xFFFF)
	eventCount := ds.diagnosticData.BusMessageCount

	return status, eventCount, nil
}

// GetCommEventLog implements modbus.DataStore
func (ds *DefaultDataStore) GetCommEventLog() (uint16, uint16, uint16, []byte, error) {
	ds.mutex.RLock()
	defer ds.mutex.RUnlock()

	// Status: 0xFFFF = Ready, 0x0000 = Not Ready
	status := uint16(0xFFFF)
	eventCount := ds.diagnosticData.BusMessageCount
	messageCount := ds.diagnosticData.ServerMessageCount

	// Copy event log
	events := make([]byte, len(ds.commEventLog))
	copy(events, ds.commEventLog)

	return status, eventCount, messageCount, events, nil
}

// IncrementDiagnosticCounter increments a diagnostic counter (helper method)
func (ds *DefaultDataStore) IncrementDiagnosticCounter(counter string) {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	switch counter {
	case "BusMessage":
		ds.diagnosticData.BusMessageCount++
	case "BusCommError":
		ds.diagnosticData.BusCommErrorCount++
	case "BusException":
		ds.diagnosticData.BusExceptionCount++
	case "ServerMessage":
		ds.diagnosticData.ServerMessageCount++
	case "ServerNoResp":
		ds.diagnosticData.ServerNoRespCount++
	case "ServerNAK":
		ds.diagnosticData.ServerNAKCount++
	case "ServerBusy":
		ds.diagnosticData.ServerBusyCount++
	case "BusCharOverrun":
		ds.diagnosticData.BusCharOverrunCount++
	}
}

// ServerRequestHandler implements the RequestHandler interface
type ServerRequestHandler struct {
	dataStore  modbus.DataStore
	deviceInfo *modbus.DeviceIdentification
}

// NewServerRequestHandler creates a new server request handler
func NewServerRequestHandler(dataStore modbus.DataStore) *ServerRequestHandler {
	return &ServerRequestHandler{
		dataStore: dataStore,
		deviceInfo: &modbus.DeviceIdentification{
			VendorName:         "ModbusGo",
			ProductCode:        "MG001",
			MajorMinorRevision: "1.0.0",
			ConformityLevel:    modbus.ConformityLevelBasicStream,
		},
	}
}

// SetDeviceIdentification sets the device identification information
func (h *ServerRequestHandler) SetDeviceIdentification(deviceInfo *modbus.DeviceIdentification) {
	h.deviceInfo = deviceInfo
}

// HandleRequest implements transport.RequestHandler
func (h *ServerRequestHandler) HandleRequest(slaveID modbus.SlaveID, req *pdu.Request) *pdu.Response {
	switch req.FunctionCode {
	case modbus.FuncCodeReadCoils:
		return h.handleReadCoils(req)
	case modbus.FuncCodeReadDiscreteInputs:
		return h.handleReadDiscreteInputs(req)
	case modbus.FuncCodeReadHoldingRegisters:
		return h.handleReadHoldingRegisters(req)
	case modbus.FuncCodeReadInputRegisters:
		return h.handleReadInputRegisters(req)
	case modbus.FuncCodeWriteSingleCoil:
		return h.handleWriteSingleCoil(req)
	case modbus.FuncCodeWriteSingleRegister:
		return h.handleWriteSingleRegister(req)
	case modbus.FuncCodeWriteMultipleCoils:
		return h.handleWriteMultipleCoils(req)
	case modbus.FuncCodeWriteMultipleRegisters:
		return h.handleWriteMultipleRegisters(req)
	case modbus.FuncCodeMaskWriteRegister:
		return h.handleMaskWriteRegister(req)
	case modbus.FuncCodeReadWriteMultipleRegs:
		return h.handleReadWriteMultipleRegisters(req)
	case modbus.FuncCodeReadExceptionStatus:
		return h.handleReadExceptionStatus(req)
	case modbus.FuncCodeDiagnostic:
		return h.handleDiagnostic(req)
	case modbus.FuncCodeGetCommEventCounter:
		return h.handleGetCommEventCounter(req)
	case modbus.FuncCodeGetCommEventLog:
		return h.handleGetCommEventLog(req)
	case modbus.FuncCodeReportServerID:
		return h.handleReportServerID(req)
	case modbus.FuncCodeReadFileRecord:
		return h.handleReadFileRecord(req)
	case modbus.FuncCodeWriteFileRecord:
		return h.handleWriteFileRecord(req)
	case modbus.FuncCodeReadFIFOQueue:
		return h.handleReadFIFOQueue(req)
	case modbus.FuncCodeEncapsulatedInterface:
		return h.handleEncapsulatedInterface(req)
	default:
		return pdu.NewExceptionResponse(req.FunctionCode, modbus.ExceptionCodeIllegalFunction)
	}
}

// handleReadCoils handles read coils request
func (h *ServerRequestHandler) handleReadCoils(req *pdu.Request) *pdu.Response {
	if len(req.Data) != 4 {
		return pdu.NewExceptionResponse(req.FunctionCode, modbus.ExceptionCodeIllegalDataValue)
	}

	address, _ := pdu.DecodeUint16(req.Data[0:2])
	quantity, _ := pdu.DecodeUint16(req.Data[2:4])

	values, err := h.dataStore.ReadCoils(modbus.Address(address), modbus.Quantity(quantity))
	if err != nil {
		if modbusErr, ok := err.(*modbus.ModbusError); ok {
			return pdu.NewExceptionResponse(req.FunctionCode, modbusErr.ExceptionCode)
		}
		return pdu.NewExceptionResponse(req.FunctionCode, modbus.ExceptionCodeServerDeviceFailure)
	}

	coilBytes := pdu.EncodeBoolSlice(values)
	responseData := make([]byte, 1+len(coilBytes))
	responseData[0] = byte(len(coilBytes))
	copy(responseData[1:], coilBytes)

	return pdu.NewResponse(req.FunctionCode, responseData)
}

// handleReadDiscreteInputs handles read discrete inputs request
func (h *ServerRequestHandler) handleReadDiscreteInputs(req *pdu.Request) *pdu.Response {
	if len(req.Data) != 4 {
		return pdu.NewExceptionResponse(req.FunctionCode, modbus.ExceptionCodeIllegalDataValue)
	}

	address, _ := pdu.DecodeUint16(req.Data[0:2])
	quantity, _ := pdu.DecodeUint16(req.Data[2:4])

	values, err := h.dataStore.ReadDiscreteInputs(modbus.Address(address), modbus.Quantity(quantity))
	if err != nil {
		if modbusErr, ok := err.(*modbus.ModbusError); ok {
			return pdu.NewExceptionResponse(req.FunctionCode, modbusErr.ExceptionCode)
		}
		return pdu.NewExceptionResponse(req.FunctionCode, modbus.ExceptionCodeServerDeviceFailure)
	}

	inputBytes := pdu.EncodeBoolSlice(values)
	responseData := make([]byte, 1+len(inputBytes))
	responseData[0] = byte(len(inputBytes))
	copy(responseData[1:], inputBytes)

	return pdu.NewResponse(req.FunctionCode, responseData)
}

// handleReadHoldingRegisters handles read holding registers request
func (h *ServerRequestHandler) handleReadHoldingRegisters(req *pdu.Request) *pdu.Response {
	if len(req.Data) != 4 {
		return pdu.NewExceptionResponse(req.FunctionCode, modbus.ExceptionCodeIllegalDataValue)
	}

	address, _ := pdu.DecodeUint16(req.Data[0:2])
	quantity, _ := pdu.DecodeUint16(req.Data[2:4])

	values, err := h.dataStore.ReadHoldingRegisters(modbus.Address(address), modbus.Quantity(quantity))
	if err != nil {
		if modbusErr, ok := err.(*modbus.ModbusError); ok {
			return pdu.NewExceptionResponse(req.FunctionCode, modbusErr.ExceptionCode)
		}
		return pdu.NewExceptionResponse(req.FunctionCode, modbus.ExceptionCodeServerDeviceFailure)
	}

	registerBytes := pdu.EncodeUint16Slice(values)
	responseData := make([]byte, 1+len(registerBytes))
	responseData[0] = byte(len(registerBytes))
	copy(responseData[1:], registerBytes)

	return pdu.NewResponse(req.FunctionCode, responseData)
}

// handleReadInputRegisters handles read input registers request
func (h *ServerRequestHandler) handleReadInputRegisters(req *pdu.Request) *pdu.Response {
	if len(req.Data) != 4 {
		return pdu.NewExceptionResponse(req.FunctionCode, modbus.ExceptionCodeIllegalDataValue)
	}

	address, _ := pdu.DecodeUint16(req.Data[0:2])
	quantity, _ := pdu.DecodeUint16(req.Data[2:4])

	values, err := h.dataStore.ReadInputRegisters(modbus.Address(address), modbus.Quantity(quantity))
	if err != nil {
		if modbusErr, ok := err.(*modbus.ModbusError); ok {
			return pdu.NewExceptionResponse(req.FunctionCode, modbusErr.ExceptionCode)
		}
		return pdu.NewExceptionResponse(req.FunctionCode, modbus.ExceptionCodeServerDeviceFailure)
	}

	registerBytes := pdu.EncodeUint16Slice(values)
	responseData := make([]byte, 1+len(registerBytes))
	responseData[0] = byte(len(registerBytes))
	copy(responseData[1:], registerBytes)

	return pdu.NewResponse(req.FunctionCode, responseData)
}

// handleWriteSingleCoil handles write single coil request
func (h *ServerRequestHandler) handleWriteSingleCoil(req *pdu.Request) *pdu.Response {
	if len(req.Data) != 4 {
		return pdu.NewExceptionResponse(req.FunctionCode, modbus.ExceptionCodeIllegalDataValue)
	}

	address, _ := pdu.DecodeUint16(req.Data[0:2])
	value, _ := pdu.DecodeUint16(req.Data[2:4])

	// Validate coil value
	if value != modbus.CoilOff && value != modbus.CoilOn {
		return pdu.NewExceptionResponse(req.FunctionCode, modbus.ExceptionCodeIllegalDataValue)
	}

	coilValue := value == modbus.CoilOn
	err := h.dataStore.WriteCoils(modbus.Address(address), []bool{coilValue})
	if err != nil {
		if modbusErr, ok := err.(*modbus.ModbusError); ok {
			return pdu.NewExceptionResponse(req.FunctionCode, modbusErr.ExceptionCode)
		}
		return pdu.NewExceptionResponse(req.FunctionCode, modbus.ExceptionCodeServerDeviceFailure)
	}

	// Echo back the request
	return pdu.NewResponse(req.FunctionCode, req.Data)
}

// handleWriteSingleRegister handles write single register request
func (h *ServerRequestHandler) handleWriteSingleRegister(req *pdu.Request) *pdu.Response {
	if len(req.Data) != 4 {
		return pdu.NewExceptionResponse(req.FunctionCode, modbus.ExceptionCodeIllegalDataValue)
	}

	address, _ := pdu.DecodeUint16(req.Data[0:2])
	value, _ := pdu.DecodeUint16(req.Data[2:4])

	err := h.dataStore.WriteHoldingRegisters(modbus.Address(address), []uint16{value})
	if err != nil {
		if modbusErr, ok := err.(*modbus.ModbusError); ok {
			return pdu.NewExceptionResponse(req.FunctionCode, modbusErr.ExceptionCode)
		}
		return pdu.NewExceptionResponse(req.FunctionCode, modbus.ExceptionCodeServerDeviceFailure)
	}

	// Echo back the request
	return pdu.NewResponse(req.FunctionCode, req.Data)
}

// handleWriteMultipleCoils handles write multiple coils request
func (h *ServerRequestHandler) handleWriteMultipleCoils(req *pdu.Request) *pdu.Response {
	if len(req.Data) < 5 {
		return pdu.NewExceptionResponse(req.FunctionCode, modbus.ExceptionCodeIllegalDataValue)
	}

	address, _ := pdu.DecodeUint16(req.Data[0:2])
	quantity, _ := pdu.DecodeUint16(req.Data[2:4])
	byteCount := req.Data[4]

	if len(req.Data) != 5+int(byteCount) {
		return pdu.NewExceptionResponse(req.FunctionCode, modbus.ExceptionCodeIllegalDataValue)
	}

	values := pdu.DecodeBoolSlice(req.Data[5:], int(quantity))
	err := h.dataStore.WriteCoils(modbus.Address(address), values)
	if err != nil {
		if modbusErr, ok := err.(*modbus.ModbusError); ok {
			return pdu.NewExceptionResponse(req.FunctionCode, modbusErr.ExceptionCode)
		}
		return pdu.NewExceptionResponse(req.FunctionCode, modbus.ExceptionCodeServerDeviceFailure)
	}

	// Return address and quantity
	responseData := make([]byte, 4)
	copy(responseData[0:2], pdu.EncodeUint16(uint16(address)))
	copy(responseData[2:4], pdu.EncodeUint16(uint16(quantity)))

	return pdu.NewResponse(req.FunctionCode, responseData)
}

// handleWriteMultipleRegisters handles write multiple registers request
func (h *ServerRequestHandler) handleWriteMultipleRegisters(req *pdu.Request) *pdu.Response {
	if len(req.Data) < 5 {
		return pdu.NewExceptionResponse(req.FunctionCode, modbus.ExceptionCodeIllegalDataValue)
	}

	address, _ := pdu.DecodeUint16(req.Data[0:2])
	quantity, _ := pdu.DecodeUint16(req.Data[2:4])
	byteCount := req.Data[4]

	if len(req.Data) != 5+int(byteCount) || int(byteCount) != int(quantity)*2 {
		return pdu.NewExceptionResponse(req.FunctionCode, modbus.ExceptionCodeIllegalDataValue)
	}

	values, err := pdu.DecodeUint16Slice(req.Data[5:])
	if err != nil {
		return pdu.NewExceptionResponse(req.FunctionCode, modbus.ExceptionCodeIllegalDataValue)
	}

	err = h.dataStore.WriteHoldingRegisters(modbus.Address(address), values)
	if err != nil {
		if modbusErr, ok := err.(*modbus.ModbusError); ok {
			return pdu.NewExceptionResponse(req.FunctionCode, modbusErr.ExceptionCode)
		}
		return pdu.NewExceptionResponse(req.FunctionCode, modbus.ExceptionCodeServerDeviceFailure)
	}

	// Return address and quantity
	responseData := make([]byte, 4)
	copy(responseData[0:2], pdu.EncodeUint16(uint16(address)))
	copy(responseData[2:4], pdu.EncodeUint16(uint16(quantity)))

	return pdu.NewResponse(req.FunctionCode, responseData)
}

// handleMaskWriteRegister handles mask write register request
func (h *ServerRequestHandler) handleMaskWriteRegister(req *pdu.Request) *pdu.Response {
	if len(req.Data) != 6 {
		return pdu.NewExceptionResponse(req.FunctionCode, modbus.ExceptionCodeIllegalDataValue)
	}

	address, _ := pdu.DecodeUint16(req.Data[0:2])
	andMask, _ := pdu.DecodeUint16(req.Data[2:4])
	orMask, _ := pdu.DecodeUint16(req.Data[4:6])

	// Read current value
	currentValues, err := h.dataStore.ReadHoldingRegisters(modbus.Address(address), 1)
	if err != nil {
		if modbusErr, ok := err.(*modbus.ModbusError); ok {
			return pdu.NewExceptionResponse(req.FunctionCode, modbusErr.ExceptionCode)
		}
		return pdu.NewExceptionResponse(req.FunctionCode, modbus.ExceptionCodeServerDeviceFailure)
	}

	// Apply mask: Result = (Current AND And_Mask) OR (Or_Mask AND (NOT And_Mask))
	current := currentValues[0]
	result := (current & andMask) | (orMask & (^andMask))

	// Write back
	err = h.dataStore.WriteHoldingRegisters(modbus.Address(address), []uint16{result})
	if err != nil {
		if modbusErr, ok := err.(*modbus.ModbusError); ok {
			return pdu.NewExceptionResponse(req.FunctionCode, modbusErr.ExceptionCode)
		}
		return pdu.NewExceptionResponse(req.FunctionCode, modbus.ExceptionCodeServerDeviceFailure)
	}

	// Echo back the request
	return pdu.NewResponse(req.FunctionCode, req.Data)
}

// handleReadWriteMultipleRegisters handles read/write multiple registers request
func (h *ServerRequestHandler) handleReadWriteMultipleRegisters(req *pdu.Request) *pdu.Response {
	if len(req.Data) < 9 {
		return pdu.NewExceptionResponse(req.FunctionCode, modbus.ExceptionCodeIllegalDataValue)
	}

	readAddress, _ := pdu.DecodeUint16(req.Data[0:2])
	readQuantity, _ := pdu.DecodeUint16(req.Data[2:4])
	writeAddress, _ := pdu.DecodeUint16(req.Data[4:6])
	writeQuantity, _ := pdu.DecodeUint16(req.Data[6:8])
	writeByteCount := req.Data[8]

	if len(req.Data) != 9+int(writeByteCount) || int(writeByteCount) != int(writeQuantity)*2 {
		return pdu.NewExceptionResponse(req.FunctionCode, modbus.ExceptionCodeIllegalDataValue)
	}

	// Write first
	writeValues, err := pdu.DecodeUint16Slice(req.Data[9:])
	if err != nil {
		return pdu.NewExceptionResponse(req.FunctionCode, modbus.ExceptionCodeIllegalDataValue)
	}

	err = h.dataStore.WriteHoldingRegisters(modbus.Address(writeAddress), writeValues)
	if err != nil {
		if modbusErr, ok := err.(*modbus.ModbusError); ok {
			return pdu.NewExceptionResponse(req.FunctionCode, modbusErr.ExceptionCode)
		}
		return pdu.NewExceptionResponse(req.FunctionCode, modbus.ExceptionCodeServerDeviceFailure)
	}

	// Then read
	readValues, err := h.dataStore.ReadHoldingRegisters(modbus.Address(readAddress), modbus.Quantity(readQuantity))
	if err != nil {
		if modbusErr, ok := err.(*modbus.ModbusError); ok {
			return pdu.NewExceptionResponse(req.FunctionCode, modbusErr.ExceptionCode)
		}
		return pdu.NewExceptionResponse(req.FunctionCode, modbus.ExceptionCodeServerDeviceFailure)
	}

	registerBytes := pdu.EncodeUint16Slice(readValues)
	responseData := make([]byte, 1+len(registerBytes))
	responseData[0] = byte(len(registerBytes))
	copy(responseData[1:], registerBytes)

	return pdu.NewResponse(req.FunctionCode, responseData)
}

// handleEncapsulatedInterface handles encapsulated interface transport
func (h *ServerRequestHandler) handleEncapsulatedInterface(req *pdu.Request) *pdu.Response {
	if len(req.Data) < 1 {
		return pdu.NewExceptionResponse(req.FunctionCode, modbus.ExceptionCodeIllegalDataValue)
	}

	meiType := req.Data[0]
	switch meiType {
	case modbus.MEITypeDeviceIdentification:
		return h.handleReadDeviceIdentification(req)
	default:
		return pdu.NewExceptionResponse(req.FunctionCode, modbus.ExceptionCodeIllegalFunction)
	}
}

// handleReadDeviceIdentification handles read device identification
func (h *ServerRequestHandler) handleReadDeviceIdentification(req *pdu.Request) *pdu.Response {
	if len(req.Data) < 3 {
		return pdu.NewExceptionResponse(req.FunctionCode, modbus.ExceptionCodeIllegalDataValue)
	}

	readCode := req.Data[1]
	objectID := req.Data[2]

	// Basic implementation - return basic device info
	responseData := []byte{
		modbus.MEITypeDeviceIdentification,
		readCode,
		h.deviceInfo.ConformityLevel,
		0x00, // More follows = false
		0x00, // Next object ID
		0x03, // Number of objects (VendorName, ProductCode, MajorMinorRevision)
	}

	// Add VendorName
	responseData = append(responseData, modbus.DeviceIDVendorName)
	responseData = append(responseData, byte(len(h.deviceInfo.VendorName)))
	responseData = append(responseData, []byte(h.deviceInfo.VendorName)...)

	// Add ProductCode
	responseData = append(responseData, modbus.DeviceIDProductCode)
	responseData = append(responseData, byte(len(h.deviceInfo.ProductCode)))
	responseData = append(responseData, []byte(h.deviceInfo.ProductCode)...)

	// Add MajorMinorRevision
	responseData = append(responseData, modbus.DeviceIDMajorMinorRevision)
	responseData = append(responseData, byte(len(h.deviceInfo.MajorMinorRevision)))
	responseData = append(responseData, []byte(h.deviceInfo.MajorMinorRevision)...)

	_ = objectID // For future use with individual access

	return pdu.NewResponse(req.FunctionCode, responseData)
}

// handleReadExceptionStatus handles read exception status request
func (h *ServerRequestHandler) handleReadExceptionStatus(req *pdu.Request) *pdu.Response {
	status, err := h.dataStore.ReadExceptionStatus()
	if err != nil {
		if modbusErr, ok := err.(*modbus.ModbusError); ok {
			return pdu.NewExceptionResponse(req.FunctionCode, modbusErr.ExceptionCode)
		}
		return pdu.NewExceptionResponse(req.FunctionCode, modbus.ExceptionCodeServerDeviceFailure)
	}

	return pdu.NewResponse(req.FunctionCode, []byte{status})
}

// handleDiagnostic handles diagnostic request
func (h *ServerRequestHandler) handleDiagnostic(req *pdu.Request) *pdu.Response {
	if len(req.Data) < 2 {
		return pdu.NewExceptionResponse(req.FunctionCode, modbus.ExceptionCodeIllegalDataValue)
	}

	subFunction, _ := pdu.DecodeUint16(req.Data[0:2])
	var data []byte
	if len(req.Data) > 2 {
		data = req.Data[2:]
	}

	result, err := h.dataStore.GetDiagnosticData(subFunction, data)
	if err != nil {
		if modbusErr, ok := err.(*modbus.ModbusError); ok {
			return pdu.NewExceptionResponse(req.FunctionCode, modbusErr.ExceptionCode)
		}
		return pdu.NewExceptionResponse(req.FunctionCode, modbus.ExceptionCodeServerDeviceFailure)
	}

	responseData := make([]byte, 2+len(result))
	copy(responseData[0:2], pdu.EncodeUint16(subFunction))
	copy(responseData[2:], result)

	return pdu.NewResponse(req.FunctionCode, responseData)
}

// handleGetCommEventCounter handles get communication event counter request
func (h *ServerRequestHandler) handleGetCommEventCounter(req *pdu.Request) *pdu.Response {
	status, eventCount, err := h.dataStore.GetCommEventCounter()
	if err != nil {
		if modbusErr, ok := err.(*modbus.ModbusError); ok {
			return pdu.NewExceptionResponse(req.FunctionCode, modbusErr.ExceptionCode)
		}
		return pdu.NewExceptionResponse(req.FunctionCode, modbus.ExceptionCodeServerDeviceFailure)
	}

	responseData := make([]byte, 4)
	copy(responseData[0:2], pdu.EncodeUint16(status))
	copy(responseData[2:4], pdu.EncodeUint16(eventCount))

	return pdu.NewResponse(req.FunctionCode, responseData)
}

// handleGetCommEventLog handles get communication event log request
func (h *ServerRequestHandler) handleGetCommEventLog(req *pdu.Request) *pdu.Response {
	status, eventCount, messageCount, events, err := h.dataStore.GetCommEventLog()
	if err != nil {
		if modbusErr, ok := err.(*modbus.ModbusError); ok {
			return pdu.NewExceptionResponse(req.FunctionCode, modbusErr.ExceptionCode)
		}
		return pdu.NewExceptionResponse(req.FunctionCode, modbus.ExceptionCodeServerDeviceFailure)
	}

	responseData := make([]byte, 7+len(events))
	responseData[0] = byte(6 + len(events)) // Byte count
	copy(responseData[1:3], pdu.EncodeUint16(status))
	copy(responseData[3:5], pdu.EncodeUint16(eventCount))
	copy(responseData[5:7], pdu.EncodeUint16(messageCount))
	copy(responseData[7:], events)

	return pdu.NewResponse(req.FunctionCode, responseData)
}

// handleReportServerID handles report server ID request
func (h *ServerRequestHandler) handleReportServerID(req *pdu.Request) *pdu.Response {
	// Basic implementation - return server ID and run indicator status
	serverID := []byte("ModbusGo Server v1.0")
	runIndicator := byte(0xFF) // 0xFF = ON

	responseData := make([]byte, 2+len(serverID))
	responseData[0] = byte(1 + len(serverID)) // Byte count
	responseData[1] = runIndicator
	copy(responseData[2:], serverID)

	return pdu.NewResponse(req.FunctionCode, responseData)
}

// handleReadFileRecord handles read file record request
func (h *ServerRequestHandler) handleReadFileRecord(req *pdu.Request) *pdu.Response {
	if len(req.Data) < 1 {
		return pdu.NewExceptionResponse(req.FunctionCode, modbus.ExceptionCodeIllegalDataValue)
	}

	byteCount := req.Data[0]
	if len(req.Data) != 1+int(byteCount) {
		return pdu.NewExceptionResponse(req.FunctionCode, modbus.ExceptionCodeIllegalDataValue)
	}

	// Parse file record sub-requests
	records := make([]modbus.FileRecord, 0)
	offset := 1
	for offset < len(req.Data) {
		if offset+7 > len(req.Data) {
			return pdu.NewExceptionResponse(req.FunctionCode, modbus.ExceptionCodeIllegalDataValue)
		}

		record := modbus.FileRecord{
			ReferenceType: req.Data[offset],
			FileNumber:    binary.BigEndian.Uint16(req.Data[offset+1 : offset+3]),
			RecordNumber:  binary.BigEndian.Uint16(req.Data[offset+3 : offset+5]),
			RecordLength:  binary.BigEndian.Uint16(req.Data[offset+5 : offset+7]),
		}
		records = append(records, record)
		offset += 7
	}

	// Read the file records
	resultRecords, err := h.dataStore.ReadFileRecords(records)
	if err != nil {
		if modbusErr, ok := err.(*modbus.ModbusError); ok {
			return pdu.NewExceptionResponse(req.FunctionCode, modbusErr.ExceptionCode)
		}
		return pdu.NewExceptionResponse(req.FunctionCode, modbus.ExceptionCodeServerDeviceFailure)
	}

	// Build response
	var responseData []byte
	for _, record := range resultRecords {
		subResp := make([]byte, 1+1+len(record.RecordData)*2)
		subResp[0] = 1 + byte(len(record.RecordData)*2) // Sub-req length
		subResp[1] = record.ReferenceType
		recordBytes := pdu.EncodeUint16Slice(record.RecordData)
		copy(subResp[2:], recordBytes)
		responseData = append(responseData, subResp...)
	}

	fullResponse := make([]byte, 1+len(responseData))
	fullResponse[0] = byte(len(responseData))
	copy(fullResponse[1:], responseData)

	return pdu.NewResponse(req.FunctionCode, fullResponse)
}

// handleWriteFileRecord handles write file record request
func (h *ServerRequestHandler) handleWriteFileRecord(req *pdu.Request) *pdu.Response {
	if len(req.Data) < 1 {
		return pdu.NewExceptionResponse(req.FunctionCode, modbus.ExceptionCodeIllegalDataValue)
	}

	byteCount := req.Data[0]
	if len(req.Data) != 1+int(byteCount) {
		return pdu.NewExceptionResponse(req.FunctionCode, modbus.ExceptionCodeIllegalDataValue)
	}

	// Parse file record sub-requests
	records := make([]modbus.FileRecord, 0)
	offset := 1
	for offset < len(req.Data) {
		if offset+7 > len(req.Data) {
			return pdu.NewExceptionResponse(req.FunctionCode, modbus.ExceptionCodeIllegalDataValue)
		}

		record := modbus.FileRecord{
			ReferenceType: req.Data[offset],
			FileNumber:    binary.BigEndian.Uint16(req.Data[offset+1 : offset+3]),
			RecordNumber:  binary.BigEndian.Uint16(req.Data[offset+3 : offset+5]),
			RecordLength:  binary.BigEndian.Uint16(req.Data[offset+5 : offset+7]),
		}

		// Read the record data
		dataByteCount := int(record.RecordLength) * 2
		if offset+7+dataByteCount > len(req.Data) {
			return pdu.NewExceptionResponse(req.FunctionCode, modbus.ExceptionCodeIllegalDataValue)
		}

		recordData, err := pdu.DecodeUint16Slice(req.Data[offset+7 : offset+7+dataByteCount])
		if err != nil {
			return pdu.NewExceptionResponse(req.FunctionCode, modbus.ExceptionCodeIllegalDataValue)
		}
		record.RecordData = recordData

		records = append(records, record)
		offset += 7 + dataByteCount
	}

	// Write the file records
	err := h.dataStore.WriteFileRecords(records)
	if err != nil {
		if modbusErr, ok := err.(*modbus.ModbusError); ok {
			return pdu.NewExceptionResponse(req.FunctionCode, modbusErr.ExceptionCode)
		}
		return pdu.NewExceptionResponse(req.FunctionCode, modbus.ExceptionCodeServerDeviceFailure)
	}

	// Echo back the request as response
	return pdu.NewResponse(req.FunctionCode, req.Data)
}

// handleReadFIFOQueue handles read FIFO queue request
func (h *ServerRequestHandler) handleReadFIFOQueue(req *pdu.Request) *pdu.Response {
	if len(req.Data) != 2 {
		return pdu.NewExceptionResponse(req.FunctionCode, modbus.ExceptionCodeIllegalDataValue)
	}

	address, _ := pdu.DecodeUint16(req.Data[0:2])

	values, err := h.dataStore.ReadFIFOQueue(modbus.Address(address))
	if err != nil {
		if modbusErr, ok := err.(*modbus.ModbusError); ok {
			return pdu.NewExceptionResponse(req.FunctionCode, modbusErr.ExceptionCode)
		}
		return pdu.NewExceptionResponse(req.FunctionCode, modbus.ExceptionCodeServerDeviceFailure)
	}

	if len(values) > modbus.MaxFIFOCount {
		return pdu.NewExceptionResponse(req.FunctionCode, modbus.ExceptionCodeIllegalDataValue)
	}

	fifoCount := uint16(len(values))
	fifoBytes := pdu.EncodeUint16Slice(values)

	responseData := make([]byte, 4+len(fifoBytes))
	copy(responseData[0:2], pdu.EncodeUint16(uint16(2+len(fifoBytes)))) // Byte count
	copy(responseData[2:4], pdu.EncodeUint16(fifoCount))                // FIFO count
	copy(responseData[4:], fifoBytes)                                   // FIFO value register

	return pdu.NewResponse(req.FunctionCode, responseData)
}

// NewTCPServer creates a new MODBUS TCP server
func NewTCPServer(address string, dataStore modbus.DataStore) (*transport.TCPServer, error) {
	handler := NewServerRequestHandler(dataStore)
	return transport.NewTCPServer(address, handler), nil
}
