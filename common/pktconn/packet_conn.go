package pktconn

import (
	"context"
	"errors"
	"fmt"
	"go-one/common/log"
	"io"
	"net"
	"sync/atomic"
	"time"
)

const (
	MaxPayloadLength       = 32 * 1024 * 1024
	DefaultReceiveChanSize = 64
	sendChanSize           = 64

	payloadLengthSize = 4 // payloadLengthSize is the packet size field (uint32) size
	prePayloadSize    = payloadLengthSize
	sendTimeout       = 500 * time.Millisecond
)

// PacketConn is a connection that send and receive data packets upon a network stream connection
type PacketConn struct {
	Proxy    interface{}
	ctx      context.Context
	conn     net.Conn
	sendChan chan *Packet
	cancel   context.CancelFunc
	err      error
	once     uint32
}

// NewPacketConn creates a packet connection based on network connection
func NewPacketConn(ctx context.Context, conn net.Conn, proxy interface{}) *PacketConn {
	if conn == nil {
		panic("conn is nil")
	}

	pcCtx, pcCancel := context.WithCancel(ctx)

	pc := &PacketConn{
		Proxy:    proxy,
		conn:     conn,
		ctx:      pcCtx,
		cancel:   pcCancel,
		sendChan: make(chan *Packet, sendChanSize),
	}

	go pc.flushMessage()
	return pc
}

func (pc *PacketConn) flushMessage() {
	ctxDone := pc.ctx.Done()
loop:
	for {
		select {
		case packet := <-pc.sendChan:
			err := pc.flush(packet)
			if err != nil {
				pc.closeWithError(err)
				break loop
			}
		case <-ctxDone:
			pc.closeWithError(pc.ctx.Err())
			break loop
		}
	}
}

func (pc *PacketConn) Receive() <-chan *Packet {
	return pc.ReceiveChanSize(DefaultReceiveChanSize)
}

func (pc *PacketConn) ReceiveChanSize(chanSize uint) <-chan *Packet {
	receiveChan := make(chan *Packet, chanSize)

	go func() {
		defer close(receiveChan)
		_ = pc.ReceiveChan(receiveChan)
	}()

	return receiveChan
}

func (pc *PacketConn) ReceiveChan(receiveChan chan *Packet) (err error) {
	for {
		packet, err := pc.receive()
		if err != nil {
			_ = pc.closeWithError(err)
			break
		}

		receiveChan <- packet
	}

	return
}

// Send send packets to remote
func (pc *PacketConn) Send(packet *Packet) error {
	if atomic.LoadInt64(&packet.refcount) <= 0 {
		panic(fmt.Errorf("sending packet with refcount=%d", packet.refcount))
	}

	packet.addRefCount(1)
	select {
	case pc.sendChan <- packet:
		return nil
	case <-time.After(sendTimeout):
		//packet.addRefCount(-1) // Decrement the refcount since the send failed
		return errors.New("send operation timed out")
	}
}

// SendAndRelease send a packet to remote and then release the packet
func (pc *PacketConn) SendAndRelease(packet *Packet) error {
	err := pc.Send(packet)
	packet.Release()

	return err
}

// Flush connection writes
func (pc *PacketConn) flushAll(packets []*Packet) (err error) {
	if len(packets) == 1 {
		// only 1 packet to send, just send it directly, no need to use send buffer
		packet := packets[0]

		err = pc.writePacket(packet)
		packet.Release()
		if err == nil {
			err = tryFlush(pc.conn)
		}
		return
	}

	for _, packet := range packets {
		err = pc.writePacket(packet)
		packet.Release()

		if err != nil {
			break
		}
	}

	// now we send all data in the send buffer
	if err == nil {
		err = tryFlush(pc.conn)
	}
	return
}

func (pc *PacketConn) flush(packet *Packet) (err error) {
	err = pc.writePacket(packet)
	packet.Release()
	if err == nil {
		err = tryFlush(pc.conn)
	}
	return
}

func (pc *PacketConn) writePacket(packet *Packet) error {
	data := packet.data()

	err := writeFull(pc.conn, data)

	if err != nil {
		return err
	}

	return nil
}

// receive receives the next packet
func (pc *PacketConn) receive() (*Packet, error) {
	var uint32Buffer [4]byte
	//var crcChecksumBuffer [4]byte
	var err error

	// receive payload length (uint32)
	err = readFull(pc.conn, uint32Buffer[:])
	if err != nil {
		return nil, err
	}

	payloadSize := packetEndian.Uint32(uint32Buffer[:])
	if payloadSize > MaxPayloadLength {
		return nil, errPayloadTooLarge
	}

	// allocate a packet to receive payload
	packet := NewPacket()
	packet.Src = pc
	payload := packet.extendPayload(int(payloadSize))
	err = readFull(pc.conn, payload)
	if err != nil {
		return nil, err
	}

	packet.SetPayloadLen(payloadSize)

	return packet, nil
}

// Close the connection
func (pc *PacketConn) Close() error {
	return pc.closeWithError(io.EOF)
}

func (pc *PacketConn) closeWithError(err error) error {
	defer func() {
		err := recover()
		if err != nil {
			log.Errorf("closeWithError failed, error: %v", err)
		}
	}()
	if atomic.CompareAndSwapUint32(&pc.once, 0, 1) {
		// close exactly once
		pc.err = err
		err := pc.conn.Close()
		pc.cancel()
		return err
	} else {
		return nil
	}
}

func (pc *PacketConn) Done() <-chan struct{} {
	return pc.ctx.Done()
}

func (pc *PacketConn) Err() error {
	return pc.err
}

// RemoteAddr return the remote address
func (pc *PacketConn) RemoteAddr() net.Addr {
	return pc.conn.RemoteAddr()
}

// LocalAddr returns the local address
func (pc *PacketConn) LocalAddr() net.Addr {
	return pc.conn.LocalAddr()
}

func (pc *PacketConn) String() string {
	return fmt.Sprintf("PacketConn<%s-%s>", pc.LocalAddr(), pc.RemoteAddr())
}
