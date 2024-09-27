package gateway

import (
	"bytes"
	"net"
	"tp1/pkg/ioutils"
	"tp1/pkg/messages"
)

var MsgIdSize = 1
var LenFieldSize = 8

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
	serverAddr, clientAddr := []byte(g.Listener.Addr().String()), []byte(conn.RemoteAddr().String())
	bufferSize := g.Config.Int("gateway.buffer_size", 1024)
	chunkSize := g.Config.Int("gateway.chunk_size", 100)
	read := 0 //not considering the message id and lengthSize (uint8 and uint64)
	data := make([]byte, 0, bufferSize)
	receivedEof := false
	msgId := uint8(0)
	payloadLength := uint64(0)
	chunk := make([]byte, 0)

	for !receivedEof {
		n, err := conn.Read(data[read:])
		if err != nil {
			return //TODO: handle error
		}

		read += n
		if read >= MsgIdSize && msgId == 0 {
			msgId = ioutils.ReadU8FromSlice(data)
			data = data[MsgIdSize:]
			read -= MsgIdSize
			if read >= LenFieldSize {
				payloadLength = ioutils.ReadU64FromSlice(data)
				read -= LenFieldSize
			}
		}

		if read >= int(payloadLength) {
			receivedEof, err = parsePayload(msgId, data, payloadLength, chunk, serverAddr, clientAddr)
			if err != nil {
				return //TODO: handle error
			}
			msgId = 0
			read = len(data)
		}

		sendChunk(g, chunk, chunkSize, !receivedEof)
	}
}

func sendChunk(g *Gateway, chunk []byte, chunkSize int, endOfFile bool) {
	if len(chunk) == chunkSize || endOfFile {
		return //TODO
	}
}

// parsePayload parses the data received from the client and appends it to the chunk
// Returns true if the end of the file was reached
// Moves the buffer payloadLen positions
func parsePayload(msgId uint8, payload []byte, payloadLen uint64, chunk []byte, serverAddr []byte, clientAddr []byte) (bool, error) {
	if isEndOfFile(msgId) {
		return true, nil
	}

	fields := []interface{}{
		msgId,
		payloadLen,
		payload[:payloadLen],
		uint64(len(serverAddr)), serverAddr,
		uint64(len(clientAddr)), clientAddr,
	}

	newMsg := new(bytes.Buffer)
	err := ioutils.WriteBytesToBuff(fields, newMsg)
	if err != nil {
		return false, err
	}

	chunk = append(chunk, newMsg.Bytes()...)
	payload = payload[:payloadLen]
	return false, nil
}

func isEndOfFile(msgId uint8) bool {
	return msgId == uint8(messages.EOF_MSG)
}
