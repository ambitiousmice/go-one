package pktconn

import (
	"encoding/binary"
	"fmt"
	"github.com/ambitiousmice/go-one/common/consts"
	"github.com/ambitiousmice/go-one/common/log"
	"math"
	"sync"
	"sync/atomic"
	"unsafe"
)

const (
	minPayloadCap       = 128
	payloadCapGrowShift = uint(2)
)

var (
	packetEndian               = binary.LittleEndian
	predefinePayloadCapacities []uint32

	packetBufferPools = map[uint32]*sync.Pool{}
	packetPool        = &sync.Pool{
		New: func() interface{} {
			p := &Packet{}
			p.bytes = p.initialBytes[:]
			return p
		},
	}
)

func init() {
	payloadCap := uint32(minPayloadCap) << payloadCapGrowShift
	for payloadCap < MaxPayloadLength {
		predefinePayloadCapacities = append(predefinePayloadCapacities, payloadCap)
		payloadCap <<= payloadCapGrowShift
	}
	predefinePayloadCapacities = append(predefinePayloadCapacities, MaxPayloadLength)

	for _, payloadCap := range predefinePayloadCapacities {
		payloadCap := payloadCap
		packetBufferPools[payloadCap] = &sync.Pool{
			New: func() interface{} {
				return make([]byte, prePayloadSize+payloadCap)
			},
		}
	}
}

func getPayloadCapOfPayloadLen(payloadLen uint32) uint32 {
	for _, payloadCap := range predefinePayloadCapacities {
		if payloadCap >= payloadLen {
			return payloadCap
		}
	}
	return MaxPayloadLength
}

// Packet is a packet for sending data
type Packet struct {
	Src *PacketConn

	readCursor   uint32
	refcount     int64
	bytes        []byte
	initialBytes [prePayloadSize + minPayloadCap]byte
}

func allocPacket() *Packet {
	pkt := packetPool.Get().(*Packet)
	pkt.refcount = 1

	if pkt.GetPayloadLen() != 0 {
		log.Panic(fmt.Errorf("allocPacket: payload should be 0, but is %d", pkt.GetPayloadLen()))
	}

	return pkt
}

// NewPacket allocates a new packet
func NewPacket() *Packet {
	return allocPacket()
}

func (p *Packet) payloadSlice(i, j uint32) []byte {
	return p.bytes[i+prePayloadSize : j+prePayloadSize]
}

// GetPayloadLen returns the payload length
func (p *Packet) GetPayloadLen() uint32 {
	packetEndian.Uint32(p.bytes[0:4])
	return *(*uint32)(unsafe.Pointer(&p.bytes[0]))
}

func (p *Packet) SetPayloadLen(plen uint32) {
	pplen := (*uint32)(unsafe.Pointer(&p.bytes[0]))
	*pplen = plen
}

// Payload returns the total payload of packet
func (p *Packet) Payload() []byte {
	return p.bytes[prePayloadSize : prePayloadSize+p.GetPayloadLen()]
}

// UnreadPayload returns the unread payload
func (p *Packet) UnreadPayload() []byte {
	pos := p.readCursor + prePayloadSize
	payloadEnd := prePayloadSize + p.GetPayloadLen()
	return p.bytes[pos:payloadEnd]
}

// HasUnreadPayload returns if all payload is read
func (p *Packet) HasUnreadPayload() bool {
	pos := p.readCursor
	plen := p.GetPayloadLen()
	return pos < plen
}

func (p *Packet) data() []byte {
	return p.bytes[0 : prePayloadSize+p.GetPayloadLen()]
}

// PayloadCap returns the current payload capacity
func (p *Packet) PayloadCap() uint32 {
	return uint32(len(p.bytes) - prePayloadSize)
}

func (p *Packet) extendPayload(size int) []byte {
	if size > MaxPayloadLength {
		log.Panic(ErrPayloadTooLarge)
	}

	payloadLen := p.GetPayloadLen()
	newPayloadLen := payloadLen + uint32(size)
	oldCap := p.PayloadCap()

	if newPayloadLen <= oldCap { // most case
		p.SetPayloadLen(newPayloadLen)
		return p.payloadSlice(payloadLen, newPayloadLen)
	}

	if newPayloadLen > MaxPayloadLength {
		log.Panic(ErrPayloadTooLarge)
	}

	// try to find the proper capacity for the size bytes
	resizeToCap := getPayloadCapOfPayloadLen(newPayloadLen)

	buffer := packetBufferPools[resizeToCap].Get().([]byte)
	if len(buffer) != int(resizeToCap+prePayloadSize) {
		log.Panic(fmt.Errorf("buffer size should be %d, but is %d", resizeToCap, len(buffer)))
	}
	copy(buffer, p.data())
	oldBytes := p.bytes
	p.bytes = buffer

	if oldCap > minPayloadCap {
		// release old bytes
		packetBufferPools[oldCap].Put(oldBytes)
	}

	p.SetPayloadLen(newPayloadLen)
	return p.payloadSlice(payloadLen, newPayloadLen)
}

func (p *Packet) ClearLastPayload(size int) {
	payloadLen := p.GetPayloadLen()

	if size >= int(payloadLen) {
		// If size is greater than or equal to current payload length,
		// set payload length to 0 and clear the payload data.
		p.SetPayloadLen(0)
		// Determine the start index to clear
		startIndex := len(p.bytes) - size

		// Clear the payload data by setting the last few elements to zero value
		for i := startIndex; i < len(p.bytes); i++ {
			p.bytes[i] = 0
		}
		return
	}

	// Determine the start index to clear
	startIndex := len(p.bytes) - size

	// Clear the last few elements of the payload data by setting them to zero value
	for i := startIndex; i < len(p.bytes); i++ {
		p.bytes[i] = 0
	}

	// Update the payload length to reflect the truncation without changing the slice length
	newPayloadLen := payloadLen - uint32(size)
	p.SetPayloadLen(newPayloadLen)
}

// addRefCount adds reference count of packet
func (p *Packet) addRefCount(add int64) {
	atomic.AddInt64(&p.refcount, add)
}

func (p *Packet) Retain() {
	p.addRefCount(1)
}

// Release releases the packet to packet pool
func (p *Packet) Release() {
	refcount := atomic.AddInt64(&p.refcount, -1)

	if refcount == 0 {
		p.Src = nil

		payloadCap := p.PayloadCap()
		if payloadCap > minPayloadCap {
			buffer := p.bytes
			p.bytes = p.initialBytes[:]
			resizeToCap := getPayloadCapOfPayloadLen(payloadCap)
			packetBufferPools[resizeToCap].Put(buffer)
		}

		p.readCursor = 0
		p.SetPayloadLen(0)
		packetPool.Put(p)
	} else if refcount < 0 {
		log.Panic(fmt.Errorf("releasing packet with refcount=%d", p.refcount))
	}
}

// ClearPayload clears packet payload
func (p *Packet) ClearPayload() {
	p.readCursor = 0
	p.SetPayloadLen(0)
}

func (p *Packet) SetReadPos(pos uint32) {
	plen := p.GetPayloadLen()
	if pos > plen {
		pos = plen
	}

	p.readCursor = pos
}

func (p *Packet) GetReadPos() uint32 {
	return p.readCursor
}

// WriteOneByte appends one byte to the end of payload
func (p *Packet) WriteOneByte(b byte) {
	pl := p.extendPayload(1)
	pl[0] = b
}

// WriteBool appends one byte 1/0 to the end of payload
func (p *Packet) WriteBool(b bool) {
	if b {
		p.WriteOneByte(1)
	} else {
		p.WriteOneByte(0)
	}
}

// WriteUint16 appends one uint16 to the end of payload
func (p *Packet) WriteUint16(v uint16) {
	pl := p.extendPayload(2)
	packetEndian.PutUint16(pl, v)
}

func (p *Packet) WriteInt16(v int16) {
	p.WriteUint16(uint16(v))
}

// WriteUint32 appends one uint32 to the end of payload
func (p *Packet) WriteUint32(v uint32) {
	pl := p.extendPayload(4)
	packetEndian.PutUint32(pl, v)
}

func (p *Packet) WriteInt32(v int32) {
	p.WriteUint32(uint32(v))
}

// WriteUint64 appends one uint64 to the end of payload
func (p *Packet) WriteUint64(v uint64) {
	pl := p.extendPayload(8)
	packetEndian.PutUint64(pl, v)
}

func (p *Packet) WriteInt64(v int64) {
	p.WriteUint64(uint64(v))
}

// WriteFloat32 appends one float32 to the end of payload
func (p *Packet) WriteFloat32(f float32) {
	p.WriteUint32(math.Float32bits(f))
}

// ReadFloat32 reads one float32 from the beginning of unread payload
func (p *Packet) ReadFloat32() float32 {
	return math.Float32frombits(p.ReadUint32())
}

// WriteFloat64 appends one float64 to the end of payload
func (p *Packet) WriteFloat64(f float64) {
	p.WriteUint64(math.Float64bits(f))
}

// ReadFloat64 reads one float64 from the beginning of unread payload
func (p *Packet) ReadFloat64() float64 {
	return math.Float64frombits(p.ReadUint64())
}

// WriteBytes appends slice of bytes to the end of payload
func (p *Packet) WriteBytes(b []byte) {
	pl := p.extendPayload(len(b))
	copy(pl, b)
}

// WriteVarBytesI appends varsize bytes to the end of payload
func (p *Packet) WriteVarBytesI(b []byte) {
	p.WriteUint32(uint32(len(b)))
	p.WriteBytes(b)
}

// WriteVarBytesH appends varsize bytes to the end of payload
func (p *Packet) WriteVarBytesH(b []byte) {
	if len(b) > 0xFFFF {
		log.Panic(ErrPayloadTooLarge)
	}

	p.WriteUint16(uint16(len(b)))
	p.WriteBytes(b)
}

func (p *Packet) WriteVarStrI(s string) {
	p.WriteVarBytesI([]byte(s))
}

func (p *Packet) WriteVarStrH(s string) {
	p.WriteVarBytesH([]byte(s))
}

// ReadOneByte reads one byte from the beginning
func (p *Packet) ReadOneByte() (v byte) {
	pos := p.readCursor + prePayloadSize
	v = p.bytes[pos]
	p.readCursor += 1
	return
}

// ReadBool reads one byte 1/0 from the beginning of unread payload
func (p *Packet) ReadBool() (v bool) {
	return p.ReadOneByte() != 0
}

// ReadBytes reads bytes from the beginning of unread payload
func (p *Packet) ReadBytes(size int) []byte {
	readPos := p.readCursor
	readEnd := readPos + uint32(size)

	if size > MaxPayloadLength || readEnd > p.GetPayloadLen() {
		log.Panic(ErrPayloadTooSmall)
	}

	p.readCursor = readEnd
	return p.payloadSlice(readPos, readEnd)
}

// ReadUint16 reads one uint16 from the beginning of unread payload
func (p *Packet) ReadUint16() uint16 {
	return packetEndian.Uint16(p.ReadBytes(2))
}

func (p *Packet) ReadInt16() int16 {
	return int16(p.ReadUint16())
}

// ReadUint32 reads one uint32 from the beginning of unread payload
func (p *Packet) ReadUint32() uint32 {
	return packetEndian.Uint32(p.ReadBytes(4))
}

func (p *Packet) ReadInt32() int32 {
	return int32(p.ReadUint32())
}

// ReadUint64 reads one uint64 from the beginning of unread payload
func (p *Packet) ReadUint64() (v uint64) {
	return packetEndian.Uint64(p.ReadBytes(8))
}

func (p *Packet) ReadInt64() int64 {
	return int64(p.ReadUint64())
}

func (p *Packet) ReadVarBytesI() []byte {
	bl := p.ReadUint32()
	return p.ReadBytes(int(bl))
}

func (p *Packet) ReadVarBytesH() []byte {
	bl := p.ReadUint16()
	return p.ReadBytes(int(bl))
}

func (p *Packet) ReadVarStrI() string {
	return string(p.ReadVarBytesI())
}

func (p *Packet) ReadVarStrH() string {
	return string(p.ReadVarBytesH())
}

// AppendClientID appends one Client ID to the end of payload
func (p *Packet) AppendClientID(id string) {
	p.WriteBytes([]byte(id))
}

// ReadClientID reads one ClientID from the beginning of unread  payload
func (p *Packet) ReadClientID() string {
	return string(p.ReadBytes(consts.ClientIDLength))
}

// ReadVarStr reads a varsize string from the beginning of unread  payload
func (p *Packet) ReadVarStr() string {
	b := p.ReadVarBytes()
	return string(b)
}

// ReadVarBytes reads a varsize slice of bytes from the beginning of unread  payload
func (p *Packet) ReadVarBytes() []byte {
	blen := p.ReadUint32()
	return p.ReadBytes(int(blen))
}

func (p *Packet) AppendMapStringString(m map[string]string) {
	p.WriteUint32(uint32(len(m)))
	for k, v := range m {
		p.WriteVarStrI(k)
		p.WriteVarStrI(v)
	}
}

func (p *Packet) ReadMapStringString() map[string]string {
	size := p.ReadUint32()
	m := make(map[string]string, size)
	for i := uint32(0); i < size; i++ {
		k := p.ReadVarStr()
		v := p.ReadVarStr()
		m[k] = v
	}
	return m
}

// AppendData appends data of any type to the end of payload
func (p *Packet) AppendData(msg interface{}) {
	dataBytes, err := MSG_PACKER.PackMsg(msg, nil)
	if err != nil {
		log.Error(err)
	}

	p.WriteVarBytesI(dataBytes)
}

// ReadData reads one data of any type from the beginning of unread payload
func (p *Packet) ReadData(msg interface{}) {
	b := p.ReadVarBytes()
	//gwlog.Infof("ReadData: %s", string(b))
	err := MSG_PACKER.UnpackMsg(b, msg)
	if err != nil {
		log.Error(err)
	}
}

// AppendArgs appends arguments to the end of payload one by one
func (p *Packet) AppendArgs(args []interface{}) {
	argCount := uint16(len(args))
	p.WriteUint16(argCount)

	for _, arg := range args {
		p.AppendData(arg)
	}
}

// ReadArgs reads a number of arguments from the beginning of unread payload
func (p *Packet) ReadArgs() [][]byte {
	argCount := p.ReadUint16()
	args := make([][]byte, argCount)
	var i uint16
	for i = 0; i < argCount; i++ {
		args[i] = p.ReadVarBytes() // just read bytes, but not parse it
	}
	return args
}

// AppendStringList appends a list of strings to the end of payload
func (p *Packet) AppendStringList(list []string) {
	p.WriteUint16(uint16(len(list)))
	for _, s := range list {
		p.WriteVarStrI(s)
	}
}

// ReadStringList reads a list of strings from the beginning of unread payload
func (p *Packet) ReadStringList() []string {
	listen := int(p.ReadUint16())
	list := make([]string, listen)
	for i := 0; i < listen; i++ {
		list[i] = p.ReadVarStr()
	}
	return list
}

/*func (p *Packet) AddInt64AtPosition(position int, value int64) {
	// Check if the position is within the bounds of the slice
	if position < 0 || position > len(p.bytes) {
		log.Panic("position is out of bounds")
	}

	p.extendPayload(8)
	// Shift the elements to the right to make space for the int64
	p.bytes = append(p.bytes[:position+8], p.bytes[position:]...)

	// Write the int64 at the specified position
	binary.LittleEndian.PutUint64(p.bytes[position:], uint64(value))
}*/
