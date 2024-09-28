package gateway

import (
	"bytes"
	"net"
	"tp1/pkg/ioutils"
	"tp1/pkg/message"
)

const MsgIdSize = 1
const LenFieldSize = 8
const TransportProtocol = "tcp"

func CreateGatewaySocket(g *Gateway) error {
	conn, err := net.Listen(TransportProtocol, g.Config.String("gateway.address", ""))
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

// handleConnection reads the data from the client and sends it to the broker.
// It reads considering that Read can return less than the desired buffer size
func handleConnection(g *Gateway, conn net.Conn) {
	serverAddr, clientAddr := []byte(g.Listener.Addr().String()), []byte(conn.RemoteAddr().String())
	bufferSize, chunkSize := g.Config.Int("gateway.buffer_size", 1024),
		g.Config.Uint8("gateway.chunk_size", 100) //TODO chequear el casteo de Uint8
	read, msgId, payloadSize, receivedEof := 0, uint8(0), uint64(0), false
	data := make([]byte, 0, bufferSize)
	chunkReviews, chunkGames := make([]byte, 0), make([]byte, 0)

	for !receivedEof {
		n, err := conn.Read(data[read:])
		if err != nil {
			return //TODO: handle error
		}

		read += n
		if hasReadId(read, msgId) {
			readId(&msgId, data, &read)
		}

		if hasReadMsgSize(read, payloadSize) {
			readPayloadSize(&payloadSize, data, &read)
		}

		if hasReadCompletePayload(read, payloadSize) {
			receivedEof, err = parsePayload(msgId, data, payloadSize, chunkReviews, chunkGames, serverAddr, clientAddr)
			if err != nil {
				return //TODO: handle error
			}
			sendChunk(g, msgId, chunkReviews, chunkSize, receivedEof)
			sendChunk(g, msgId, chunkGames, chunkSize, receivedEof)
			msgId, read = 0, len(data)
		}
	}
}

func readPayloadSize(payloadSize *uint64, data []byte, read *int) {
	*payloadSize = ioutils.ReadU64FromSlice(data)
	*read -= LenFieldSize
}

func readId(msgId *uint8, data []byte, read *int) {
	*msgId = ioutils.ReadU8FromSlice(data)
	data = data[MsgIdSize:]
	*read -= MsgIdSize
}

func hasReadCompletePayload(read int, payloadSize uint64) bool {
	return read >= int(payloadSize)
}

func hasReadMsgSize(read int, payloadSize uint64) bool {
	return read >= LenFieldSize && payloadSize == 0
}

func hasReadId(read int, msgId uint8) bool {
	return read >= MsgIdSize && msgId == 0
}

func sendChunk(g *Gateway, key uint8, chunk []byte, maxChunkSize uint8, endOfFile bool) {
	currentChunkSize := uint8(len(chunk))
	if currentChunkSize == maxChunkSize || endOfFile {
		auxChunk := make([]byte, currentChunkSize)
		auxChunk = append(auxChunk, currentChunkSize)
		auxChunk = append(auxChunk, chunk...)

		err := g.broker.Publish(g.exchange, string(key), auxChunk)
		if err != nil { //TODO: handle error
			return
		}
	}
}

// parsePayload parses the data received from the client and appends it to the corresponding chunk
// Returns true if the end of the file was reached
// Moves the buffer payloadLen positions
func parsePayload(msgId uint8, payload []byte, payloadLen uint64, chunkR []byte, chunkG []byte, serverAddr []byte, clientAddr []byte) (bool, error) {
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

	addToChunk(msgId, chunkR, newMsg, chunkG)
	payload = payload[:payloadLen]

	return false, nil
}

func addToChunk(msgId uint8, chunkR []byte, newMsg *bytes.Buffer, chunkG []byte) {
	switch msgId {
	case uint8(message.ReviewIdMsg):
		chunkR = append(chunkR, newMsg.Bytes()...)
	case uint8(message.GameIdMsg):
		chunkG = append(chunkG, newMsg.Bytes()...)
	}
}

func isEndOfFile(msgId uint8) bool {
	return msgId == uint8(message.EofMsg)
}
