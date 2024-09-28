package gateway

import (
	"bytes"
	"net"
	"tp1/pkg/ioutils"
	"tp1/pkg/message"
	"tp1/pkg/message/game"
	"tp1/pkg/message/review"
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
	reviewsCount, gamesCount := uint8(0), uint8(0)

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
			receivedEof, err = processPayload(msgId, data, payloadSize, chunkReviews, chunkGames, serverAddr, clientAddr, &reviewsCount, &gamesCount)
			if err != nil {
				return //TODO: handle error
			}
			sendCorrespondingChunk(g, msgId, chunkReviews, chunkSize, receivedEof, chunkGames, &reviewsCount, &gamesCount)
			msgId, read = 0, len(data)
		}
	}
}

func sendCorrespondingChunk(g *Gateway, msgId uint8, chunkReviews []byte, chunkSize uint8, receivedEof bool, chunkGames []byte, reviewsCount *uint8, gamesCount *uint8) {
	if msgId == uint8(message.ReviewIdMsg) {
		sendChunk(g, msgId, chunkReviews, chunkSize, receivedEof, reviewsCount)
	} else {
		sendChunk(g, msgId, chunkGames, chunkSize, receivedEof, gamesCount)
	}
}

func readPayloadSize(payloadSize *uint64, data []byte, read *int) {
	*payloadSize = ioutils.ReadU64FromSlice(data)
	*read -= LenFieldSize
}

func readId(msgId *uint8, data []byte, read *int) {
	*msgId = ioutils.ReadU8FromSlice(data)
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

func sendChunk(g *Gateway, key uint8, chunk []byte, maxChunkLen uint8, endOfFile bool, count *uint8) {
	if (*count == maxChunkLen || endOfFile) && *count > 0 {
		auxChunk := make([]byte, *count)
		auxChunk = append(auxChunk, key)
		auxChunk = append(auxChunk, *count)
		auxChunk = append(auxChunk, chunk...)

		err := g.broker.Publish(g.exchange, string(key), auxChunk)
		*count = 0
		if err != nil { //TODO: handle error
			return
		}
	}
}

// processPayload parses the data received from the client and appends it to the corresponding chunk
// Returns true if the end of the file was reached
// Moves the buffer payloadLen positions
func processPayload(msgId uint8, payload []byte,
	payloadLen uint64, chunkR []byte,
	chunkG []byte, serverAddr []byte, clientAddr []byte,
	reviewsCount *uint8, gamesCount *uint8) (bool, error) {

	if isEndOfFile(msgId) {
		return true, nil
	}

	newMsg, b, err := parsePayload(msgId, payload, payloadLen, serverAddr, clientAddr)
	if err != nil {
		return b, err
	}

	addToChunk(msgId, chunkR, newMsg, chunkG, reviewsCount, gamesCount)
	payload = payload[:payloadLen]

	return false, nil
}

func parsePayload(msgId uint8, payload []byte, payloadLen uint64, serverAddr []byte, clientAddr []byte) (*bytes.Buffer, bool, error) {
	payload = payload[:payloadLen]
	var msg []byte
	if msgId == uint8(message.ReviewIdMsg) {
		msg = parseReviewPayload(payload)
	} else {
		msg = parseGamePayload(payload)
	}

	fields := []interface{}{
		msgId,
		msg,
		uint64(len(serverAddr)), serverAddr,
		uint64(len(clientAddr)), clientAddr,
	}

	newMsg := new(bytes.Buffer)
	err := ioutils.WriteBytesToBuff(fields, newMsg)
	if err != nil {
		return nil, false, err
	}
	return newMsg, false, nil
}

func parseReviewPayload(payload []byte) []byte {
	return review.FromClientBytes(payload)
}

func parseGamePayload(payload []byte) []byte {
	return game.FromClientBytes(payload)
}

func addToChunk(msgId uint8, chunkR []byte, newMsg *bytes.Buffer, chunkG []byte, reviewsCount *uint8, gamesCount *uint8) {
	switch msgId {
	case uint8(message.ReviewIdMsg):
		chunkR = append(chunkR, newMsg.Bytes()...)
		*reviewsCount += 1
	case uint8(message.GameIdMsg):
		chunkG = append(chunkG, newMsg.Bytes()...)
		*gamesCount += 1
	}
}

func isEndOfFile(msgId uint8) bool {
	return msgId == uint8(message.EofMsg)
}
