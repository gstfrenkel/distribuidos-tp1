package gateway

import (
	"net"

	"tp1/pkg/ioutils"
)

var MSG_ID_SIZE = 1
var LEN_SIZE = 8

func CreateGatewaySocket(g *Gateway) error {
	conn, err := net.Listen("tcp", g.Config.String("gateway.address", ""))
	if err != nil {
		return err
	}

	g.Listener = conn
	return nil
}

func ListenForNewClients(g *Gateway) error {
	for {
		c, err := g.Listener.Accept()
		if err != nil {
			return err
		}
		go handleConnection(g, c)
	}
}

func handleConnection(g *Gateway, conn net.Conn) {
	//server_addr, client_addr := []byte(g.Listener.Addr().String()), []byte(conn.RemoteAddr().String())
	bufferSize := g.Config.Int("gateway.buffer_size", 1024)
	read := 0
	data := make([]byte, bufferSize)
	notEof := true
	msgId := uint8(0)
	payloadLength := uint64(0)

	for notEof {
		n, err := conn.Read(data[read:])
		if err != nil {
			return //TODO: handle errors
		}

		read += n
		if read >= MSG_ID_SIZE && msgId == 0 {
			msgId = ioutils.ReadU8FromSlice(data)
			data = data[MSG_ID_SIZE:]
			read -= MSG_ID_SIZE
			if read >= LEN_SIZE {
				payloadLength = ioutils.ReadU64FromSlice(data)
				read -= LEN_SIZE
			}
		}

		if read >= int(payloadLength) {
			read = 0
			notEof, err = ParseData(data)
			if err != nil {
				return //TODO: handle errors
			}
			data = make([]byte, bufferSize)
		}
	}
}

func ParseData(data []byte) (bool, error) {
	//TODO: implement
	return false, nil
}
