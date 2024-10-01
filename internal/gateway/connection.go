package gateway

import (
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
	bufferSize := g.Config.Int("gateway.buffer_size", 1024)
	read, msgId, payloadSize, receivedEof := 0, uint8(0), uint64(0), false
	data := make([]byte, 0, bufferSize)

	for !receivedEof {
		n, err := conn.Read(data[read:])
		if err != nil {
			return //TODO: handle errors
		}

		read += n
		if hasReadId(read, msgId) {
			readId(&msgId, data, &read)
		}

		if hasReadMsgSize(read, payloadSize) {
			readPayloadSize(&payloadSize, data, &read)
		}

		if hasReadCompletePayload(read, payloadSize) {
			receivedEof, err = processPayload(g, message.ID(msgId), data, payloadSize)
			if err != nil {
				return //TODO: handle errors
			}
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

// processPayload parses the data received from the client and appends it to the corresponding chunk
// Returns true if the end of the file was reached
// Moves the buffer payloadLen positions
func processPayload(g *Gateway, msgId message.ID, payload []byte, payloadLen uint64) (bool, error) {

	if isEndOfFile(payloadLen) {
		sendMsgToChunkSender(g, msgId, nil)
		return true, nil
	}

	newMsg, err := parseClientPayload(msgId, payload, payloadLen)
	if err != nil {
		return false, err
	}

	sendMsgToChunkSender(g, msgId, newMsg)

	return false, nil
}

// parseClientPayload parses the payload received from the client and returns a DataCsvReviews or DataCsvGames
func parseClientPayload(msgId message.ID, payload []byte, payloadLen uint64) (any, error) {
	payload = payload[:payloadLen]

	var (
		data any
		err  error
	)

	switch msgId {
	case message.ReviewIdMsg:
		data, err = message.DataCSVReviewsFromBytes(payload)
	case message.GameIdMsg:
		data, err = message.DataCSVGamesFromBytes(payload)
	default:
		//TODO: handle error
	}

	if err != nil {
		return nil, err
	}

	return data, nil
}

func sendMsgToChunkSender(g *Gateway, msgId message.ID, newMsg any) {
	g.ChunkChan <- ChunkItem{msgId, newMsg}
}

func isEndOfFile(payloadLen uint64) bool {
	return payloadLen == 0
}
