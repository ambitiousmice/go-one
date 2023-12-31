package pktconn

import (
	"bufio"
	"net"
	"sync"
	"time"
)

// NewBufferedConn creates a new connection with buffered write based on underlying connection
func NewBufferedConn(conn net.Conn, readBufferSize, writeBufferSize int) net.Conn {
	bc := &bufferedConn{
		Conn:      conn,
		bufReader: bufio.NewReaderSize(conn, readBufferSize),
		bufWriter: bufio.NewWriterSize(conn, writeBufferSize),
	}
	return bc
}

// bufferedConn provides buffered write to connections
type bufferedConn struct {
	net.Conn
	bufReader *bufio.Reader
	bufWriter *bufio.Writer
	flushLock sync.Mutex
}

// Read
func (bc *bufferedConn) Read(p []byte) (int, error) {
	return bc.bufReader.Read(p)
}

func (bc *bufferedConn) Write(p []byte) (int, error) {
	return bc.bufWriter.Write(p)
}

func (bc *bufferedConn) Close() error {
	bc.Flush()
	return bc.Conn.Close()
}

func (bc *bufferedConn) Flush() error {
	bc.flushLock.Lock()
	err := bc.bufWriter.Flush()
	bc.flushLock.Unlock()
	if err != nil {
		return err
	}

	return nil
}

func (bc *bufferedConn) SetReadTimeout(timeout time.Duration) error {
	return bc.SetDeadline(time.Now().Add(timeout))
}

func (bc *bufferedConn) SetWriteTimeout(timeout time.Duration) error {
	return bc.SetWriteDeadline(time.Now().Add(timeout))
}
