package main

import (
	"github.com/ambitiousmice/go-one/common/pktconn"
)

func main() {
	packet := pktconn.NewPacket()
	packet.WriteUint16(1)
	packet.WriteUint32(1)

	cmd := packet.ReadUint16()
	b := packet.ReadVarBytesI()

	println(cmd)
	println(len(b))

}
