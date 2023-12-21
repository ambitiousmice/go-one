package network

import (
	"errors"
	"github.com/ambitiousmice/go-one/common/log"
	"net"
	"time"
)

const (
	restartTcpServerInterval = 3 * time.Second
)

// TCPServerDelegate is the implementations that a TCP server should provide
type TCPServerDelegate interface {
	ServeTCPConnection(net.Conn)
}

// ServeTCPForever serves on specified address as TCP server, forever ...
func ServeTCPForever(listenAddr string, delegate TCPServerDelegate) {
	for {
		err := ServeTCPForeverOnce(listenAddr, delegate)
		log.Errorf("server@%s failed with error: %v, will restart after %s", listenAddr, err, restartTcpServerInterval)
		time.Sleep(restartTcpServerInterval)
	}
}

func ServeTCPForeverOnce(listenAddr string, delegate TCPServerDelegate) error {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("serveTCPImpl: panic with error %s", err)
		}
	}()

	return ServeTCP(listenAddr, delegate)

}

// ServeTCP serves on specified address as TCP server
func ServeTCP(listenAddr string, delegate TCPServerDelegate) error {
	if len(listenAddr) == 0 {
		return errors.New("tcp listenAddr is empty")
	}

	listener, err := net.Listen("tcp", listenAddr)
	log.Infof("Listening on TCP: %s ...", listenAddr)

	if err != nil {
		return err
	}

	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		log.Infof("new tcp connection from: %s", conn.RemoteAddr())

		go delegate.ServeTCPConnection(conn)
	}
}
