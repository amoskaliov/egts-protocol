package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strconv"
)

//RecordDataSet описывает массив с подзаписями протокола ЕГТС
type RecordDataSet []RecordData

//Encode кодирование структуры в байты
func (f *RecordDataSet) Encode() ([]byte, error) {
	var result []byte
	buf := new(bytes.Buffer)

	for _, rd := range *f {
		rec, err := rd.Encode()
		if err != nil {
			return result, err
		}
		buf.Write(rec)
	}

	result = buf.Bytes()

	return result, nil
}

//Length получает длину массива записей
func (f *RecordDataSet) Length() uint16 {
	var result uint16

	if recBytes, err := f.Encode(); err != nil {
		result = uint16(0)
	} else {
		result = uint16(len(recBytes))
	}

	return result
}

type ServiceDataSet []ServiceDataRecord

//Encode кодирование структуры в байты
func (f *ServiceDataSet) Encode() ([]byte, error) {
	var result []byte
	buf := new(bytes.Buffer)

	for _, rd := range *f {
		rec, err := rd.Encode()
		if err != nil {
			return result, err
		}
		buf.Write(rec)
	}

	result = buf.Bytes()

	return result, nil
}

//Length получает длину массива записей
func (f *ServiceDataSet) Length() uint16 {
	var result uint16

	if recBytes, err := f.Encode(); err != nil {
		result = uint16(0)
	} else {
		result = uint16(len(recBytes))
	}

	return result
}

type ServiceDataRecord struct {
	RecordLength             uint16
	RecordNumber             uint16
	SourceServiceOnDevice    string
	RecipientServiceOnDevice string
	Group                    string
	RecordProcessingPriority string
	TimeFieldExists          string
	EventIDFieldExists       string
	ObjectIDFieldExists      string
	ObjectIdentifier         uint32
	EventIdentifier          uint32
	Time                     uint32
	SourceServiceType        byte
	RecipientServiceType     byte
	RecordDataSet
}

//Decode разбор байтов в структуру
func (sdr *ServiceDataRecord) Decode(sdrBytes []byte) error {
	var (
		err   error
		flags byte
	)
	buf := bytes.NewReader(sdrBytes)

	tmpIntBuf := make([]byte, 2)
	if _, err = buf.Read(tmpIntBuf); err != nil {
		return fmt.Errorf("Не удалось получить длину записи sdr: %v", err)
	}
	sdr.RecordLength = binary.LittleEndian.Uint16(tmpIntBuf)

	if _, err = buf.Read(tmpIntBuf); err != nil {
		return fmt.Errorf("Не удалось получить номер записи sdr: %v", err)
	}
	sdr.RecordNumber = binary.LittleEndian.Uint16(tmpIntBuf)

	if flags, err = buf.ReadByte(); err != nil {
		return fmt.Errorf("Не удалось считать байт флагов srd: %v", err)
	}
	flagBits := fmt.Sprintf("%08b", flags)
	sdr.SourceServiceOnDevice = flagBits[:1]
	sdr.RecipientServiceOnDevice = flagBits[1:2]
	sdr.Group = flagBits[2:3]
	sdr.RecordProcessingPriority = flagBits[3:5]
	sdr.TimeFieldExists = flagBits[5:6]
	sdr.EventIDFieldExists = flagBits[6:7]
	sdr.ObjectIDFieldExists = flagBits[7:]

	if sdr.ObjectIDFieldExists == "1" {
		oid := make([]byte, 4)
		if _, err := buf.Read(oid); err != nil {
			return fmt.Errorf("Не удалось получить идентификатор объекта sdr: %v", err)
		}
		sdr.ObjectIdentifier = binary.LittleEndian.Uint32(oid)
	}

	if sdr.EventIDFieldExists == "1" {
		event := make([]byte, 4)
		if _, err := buf.Read(event); err != nil {
			return fmt.Errorf("Не удалось получить идентификатор события sdr: %v", err)
		}
		sdr.EventIdentifier = binary.LittleEndian.Uint32(event)
	}

	if sdr.TimeFieldExists == "1" {
		tm := make([]byte, 4)
		if _, err := buf.Read(tm); err != nil {
			return fmt.Errorf("Не удалось получить время формирования записи на стороне отправителя sdr: %v", err)
		}
		sdr.Time = binary.LittleEndian.Uint32(tm)
	}

	if sdr.SourceServiceType, err = buf.ReadByte(); err != nil {
		return fmt.Errorf("Не удалось считать идентификатор тип сервиса-отправителя sdr: %v", err)
	}

	if sdr.RecipientServiceType, err = buf.ReadByte(); err != nil {
		return fmt.Errorf("Не удалось считать идентификатор тип сервиса-получателя srd: %v", err)
	}
	return err
}

//Encode кодирование структуры в байты
func (sdr *ServiceDataRecord) Encode() ([]byte, error) {
	var (
		result []byte
		err    error
		flags  uint64
	)

	buf := new(bytes.Buffer)

	tmpIntBuf := make([]byte, 2)
	binary.LittleEndian.PutUint16(tmpIntBuf, sdr.RecordLength)
	if _, err = buf.Write(tmpIntBuf); err != nil {
		return result, fmt.Errorf("Не удалось записать длину записи sdr: %v", err)
	}

	binary.LittleEndian.PutUint16(tmpIntBuf, sdr.RecordNumber)
	if _, err = buf.Write(tmpIntBuf); err != nil {
		return result, fmt.Errorf("Не удалось записать номер записи sdr: %v", err)
	}

	//составной байт
	flagsBits := sdr.SourceServiceOnDevice + sdr.RecipientServiceOnDevice + sdr.Group + sdr.RecordProcessingPriority +
		sdr.TimeFieldExists + sdr.EventIDFieldExists + sdr.ObjectIDFieldExists
	if flags, err = strconv.ParseUint(flagsBits, 2, 8); err != nil {
		return result, fmt.Errorf("Не удалось сгенерировать байт флагов sdr: %v", err)
	}
	if err = buf.WriteByte(uint8(flags)); err != nil {
		return result, fmt.Errorf("Не удалось записать флаги sdr: %v", err)
	}

	tmpInt32Buf := make([]byte, 4)
	if sdr.ObjectIDFieldExists == "1" {
		binary.LittleEndian.PutUint32(tmpInt32Buf, sdr.ObjectIdentifier)
		if _, err = buf.Write(tmpInt32Buf); err != nil {
			return result, fmt.Errorf("Не удалось записать идентификатор объекта sdr: %v", err)
		}
	}

	if sdr.EventIDFieldExists == "1" {
		binary.LittleEndian.PutUint32(tmpInt32Buf, sdr.EventIdentifier)
		if _, err = buf.Write(tmpInt32Buf); err != nil {
			return result, fmt.Errorf("Не удалось записать идентификатор события sdr: %v", err)
		}
	}

	if sdr.TimeFieldExists == "1" {
		binary.LittleEndian.PutUint32(tmpInt32Buf, sdr.Time)
		if _, err = buf.Write(tmpInt32Buf); err != nil {
			return result, fmt.Errorf("Не удалось время формирования записи на стороне отправителя sdr: %v", err)
		}
	}

	if err := buf.WriteByte(sdr.SourceServiceType); err != nil {
		return result, fmt.Errorf("Не удалось записать идентификатор тип сервиса-отправителя sdr: %v", err)
	}

	if err := buf.WriteByte(sdr.RecipientServiceType); err != nil {
		return result, fmt.Errorf("Не удалось записать идентификатор тип сервиса-получателя sdr: %v", err)
	}

	rd, err := sdr.RecordDataSet.Encode()
	if err != nil {
		return result, err
	}
	buf.Write(rd)

	result = buf.Bytes()
	return result, err
}

//EgtsPtAppdata тип пакета ЕГТС
type EgtsPtAppdata struct {
	ServiceDataSet
}

//Encode кодирование структуры в байты
func (ad *EgtsPtAppdata) Encode() ([]byte, error) {
	return ad.ServiceDataSet.Encode()
}

//Length получает длину массива записей
func (ad *EgtsPtAppdata) Length() uint16 {
	var result uint16

	if recBytes, err := ad.Encode(); err != nil {
		result = uint16(0)
	} else {
		result = uint16(len(recBytes))
	}
	return result
}
