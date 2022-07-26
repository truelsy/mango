// Packet.go
package mango

import (
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"time"
	"unsafe"
)

const (
	Stx = 0xff
)

type Packet struct {
	_Buf         []byte
	_ReadOffset  uint16
	_WriteOffset uint16
	_RefCount    int16
	_Tick        time.Time // CHK 패킷활성화시간
}

func GetHeaderLen() uint16 {
	// identify(1byte) + size(2byte) + command(2byte) + Identifier(8byte)
	return 1 + 2 + 2 + 8
}

///////////////////////////////////////////////////////////
// szdr7t6Utils.SyncPool interface 구현
func (p *Packet) Init() {
	p._ReadOffset = GetHeaderLen()
	p._WriteOffset = GetHeaderLen()
	p._RefCount = 0
	p._Tick = time.Time{}
}

func (p *Packet) Capture() {
	p._RefCount++
	p.NewTick()
}

func (p *Packet) Release() {
	p._RefCount--
	if p._RefCount < 0 {
		p._RefCount = 0
	}
}

func (p *Packet) GetRefCount() int16 {
	return p._RefCount
}

///////////////////////////////////////////////////////////

func (p *Packet) FinishedBytes() []byte {
	return p._Buf[:p._WriteOffset]
}

func (p *Packet) Copy(b []byte) {
	bLength := len(b)
	copy(p._Buf, b[:bLength])
	p._WriteOffset = uint16(bLength)
	//p._Buf = b
	//p._WriteOffset = uint16(len(p._Buf))
}

func (p *Packet) Append(b []byte) {
	copy(p._Buf[p._WriteOffset:], b)
	p._WriteOffset += uint16(len(b))
}

func (p *Packet) Len() int {
	//return len(p._Buf)
	return int(p._WriteOffset)
}

func (p *Packet) SetOffset(offset int) {
	p._WriteOffset = uint16(offset) + GetHeaderLen()
}

//// TODO 데이터만 얻어오는 걸 하나 만들까?
//// TODO append를 위해서 만든건데 append방식의 수정이 필요할듯
//func (p *Packet) GetData() []byte {
//	return p._Buf[GetHeaderLen():p._WriteOffset]
//}
//
//func (p *Packet) GetAll() []byte {
//	return p._Buf[:p._WriteOffset]
//}

func (p *Packet) GetBufAll() []byte {
	return p._Buf[GetHeaderLen():]
}

func (p *Packet) GetBuf() []byte {
	//return p._Buf[GetHeaderLen():]
	return p._Buf[GetHeaderLen():p._WriteOffset]
}

func (p *Packet) GetCommand() uint16 {
	return binary.LittleEndian.Uint16(p._Buf[3:])
}

func (p *Packet) GetIdentifier() uint64 {
	return binary.LittleEndian.Uint64(p._Buf[5:])
}

func (p *Packet) SetIdentifier(identifier uint64) {
	binary.LittleEndian.PutUint64(p._Buf[5:], identifier)
}

func (p *Packet) Read(args ...interface{}) error {
	for i := 0; i < len(args); i++ {

		t := reflect.ValueOf(args[i])

		switch t.Interface().(type) {
		case *int8:
			t.Elem().SetInt(int64(p.ReadInt8()))
		case *uint8:
			t.Elem().SetUint(uint64(p.ReadUInt8()))
		case *int16:
			t.Elem().SetInt(int64(p.ReadInt16()))
		case *uint16:
			t.Elem().SetUint(uint64(p.ReadUInt16()))
		case *int, *int32:
			t.Elem().SetInt(int64(p.ReadInt32()))
		case *uint, *uint32:
			t.Elem().SetUint(uint64(p.ReadUInt32()))
		case *int64:
			t.Elem().SetInt(p.ReadInt64())
		case *uint64:
			t.Elem().SetUint(p.ReadUInt64())
		case *string:
			t.Elem().SetString(p.ReadString())
		default:
			switch t.Kind() {
			case reflect.Ptr:
				return fmt.Errorf("%v pointer is not supported", t.Elem().Kind())
			default:
				return fmt.Errorf("%v is not supported", t.Kind())
			}
		}
	}
	return nil
}

func (p *Packet) ReadString() string {
	strLen := p.ReadUInt16()
	v := string(p._Buf[p._ReadOffset : p._ReadOffset+strLen])
	p._ReadOffset += strLen
	return v
}

func (p *Packet) ReadInt8() int8 {
	return int8(p.ReadUInt8())
}

func (p *Packet) ReadUInt8() uint8 {
	v := p._Buf[p._ReadOffset] // uint8
	p._ReadOffset += uint16(unsafe.Sizeof(v))
	return v
}

func (p *Packet) ReadInt16() int16 {
	return int16(p.ReadUInt16())
}

func (p *Packet) ReadUInt16() uint16 {
	v := binary.LittleEndian.Uint16(p._Buf[p._ReadOffset:])
	p._ReadOffset += uint16(unsafe.Sizeof(v))
	return v
}

func (p *Packet) ReadInt32() int32 {
	return int32(p.ReadUInt32())
}

func (p *Packet) ReadUInt32() uint32 {
	v := binary.LittleEndian.Uint32(p._Buf[p._ReadOffset:])
	p._ReadOffset += uint16(unsafe.Sizeof(v))
	return v
}

func (p *Packet) ReadInt64() int64 {
	return int64(p.ReadUInt64())
}

func (p *Packet) ReadUInt64() uint64 {
	v := binary.LittleEndian.Uint64(p._Buf[p._ReadOffset:])
	p._ReadOffset += uint16(unsafe.Sizeof(v))
	return v
}

func (p *Packet) ReadFloat32() float32 {
	v := p.ReadUInt32()
	return math.Float32frombits(v)
}

func (p *Packet) ReadFloat64() float64 {
	v := p.ReadUInt64()
	return math.Float64frombits(v)
}

func (p *Packet) WriteInt8(v int8) {
	p.WriteUInt8(uint8(v))
}

func (p *Packet) WriteUInt8(v uint8) {
	p._Buf[p._WriteOffset] = v
	p._WriteOffset += uint16(unsafe.Sizeof(v))
}

func (p *Packet) WriteInt16(v int16) {
	p.WriteUInt16(uint16(v))
}

func (p *Packet) WriteUInt16(v uint16) {
	binary.LittleEndian.PutUint16(p._Buf[p._WriteOffset:], v)
	p._WriteOffset += uint16(unsafe.Sizeof(v))
}

func (p *Packet) WriteInt32(v int32) {
	p.WriteUInt32(uint32(v))
}

func (p *Packet) WriteUInt32(v uint32) {
	binary.LittleEndian.PutUint32(p._Buf[p._WriteOffset:], v)
	p._WriteOffset += uint16(unsafe.Sizeof(v))
}

func (p *Packet) WriteInt64(v int64) {
	p.WriteUInt64(uint64(v))
}

func (p *Packet) WriteUInt64(v uint64) {
	binary.LittleEndian.PutUint64(p._Buf[p._WriteOffset:], v)
	p._WriteOffset += uint16(unsafe.Sizeof(v))
}

func (p *Packet) WriteString(v string) {
	strLen := uint16(len(v))
	p.WriteUInt16(strLen)
	copy(p._Buf[p._WriteOffset:], v)
	p._WriteOffset += strLen
}

func (p *Packet) WriteFloat32(v float32) {
	bits := math.Float32bits(v)
	binary.LittleEndian.PutUint32(p._Buf[p._WriteOffset:], bits)
	p._WriteOffset += uint16(unsafe.Sizeof(v))
}

func (p *Packet) WriteFloat64(v float64) {
	bits := math.Float64bits(v)
	binary.LittleEndian.PutUint64(p._Buf[p._WriteOffset:], bits)
	p._WriteOffset += uint16(unsafe.Sizeof(v))
}

func (p *Packet) Write(arg interface{}) *Packet {
	switch arg.(type) {
	case int8:
		p.WriteInt8(arg.(int8))
	case uint8:
		p.WriteUInt8(arg.(uint8))
	case int16:
		p.WriteInt16(arg.(int16))
	case uint16:
		p.WriteUInt16(arg.(uint16))
	case int32:
		p.WriteInt32(arg.(int32))
	case uint32:
		p.WriteUInt32(arg.(uint32))
	case int64:
		p.WriteInt64(arg.(int64))
	case uint64:
		p.WriteUInt64(arg.(uint64))
	case string:
		p.WriteString(arg.(string))
	}
	return p
}

func (p *Packet) MakeHeader(cmd uint16, iden uint64) {
	p._Buf[0] = byte(Stx)                                                    // 패킷의 시작 (1byte)
	binary.LittleEndian.PutUint16(p._Buf[1:], p._WriteOffset-GetHeaderLen()) // header길이는 포함하지 않는다. (2byte)
	binary.LittleEndian.PutUint16(p._Buf[3:], cmd)                           // 명령어 (2byte)
	p.SetIdentifier(iden)                                                    // 패킷 식별자 (8byte)
}

func (p *Packet) NewTick() {
	p._Tick = time.Now()
}

func (p *Packet) GetTick() time.Duration {
	if p._Tick.IsZero() {
		return time.Duration(0)
	}
	return time.Since(p._Tick)
}

func (p *Packet) ElaspedTime() float64 {
	if p._Tick.IsZero() {
		return 0
	}

	dur := time.Since(p._Tick)
	elapsed := float64(dur) / float64(time.Millisecond)
	return elapsed
}

// Echo 테스트 코드
func (p *Packet) TestEchoHeader() {
	binary.LittleEndian.PutUint32(p._Buf[0:], uint32(p._WriteOffset))
}

func (p *Packet) TestEchoFinishedBuf() []byte {
	return p._Buf[4:p._WriteOffset]
}
