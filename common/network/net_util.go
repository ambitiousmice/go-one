package network

import "net"

// ConnectTCP connects to host:port in TCP
func ConnectTCP(addr string) (net.Conn, error) {
	conn, err := net.Dial("tcp", addr)
	return conn, err
}
