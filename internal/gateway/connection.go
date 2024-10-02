package gateway

import (
	"net"
	"tp1/pkg/ioutils"
	"tp1/pkg/logs"
	"tp1/pkg/message"
)

const MsgIdSize = 1
const LenFieldSize = 8
const TransportProtocol = "tcp"
const maxEofs = 2

var logger, _ = logs.GetLogger("gateway")

func CreateGatewaySocket(g *Gateway) error {
	addr := g.Config.String("gateway.address", "")
	conn, err := net.Listen(TransportProtocol, addr)
	if err != nil {
		return err
	}
	logger.Infof("Gateway listening on %s", conn.Addr().String())
	g.Listener = conn
	return nil
}

func ListenForNewClients(g *Gateway) error {
	for {
		logger.Infof("Waiting for new client connection...")
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
	logger.Infof("New client connected: %s", conn.RemoteAddr().String())
	bufferSize := g.Config.Int("gateway.buffer_size", 1024)
	read, msgId, payloadSize, eofs := 0, uint8(0), uint64(0), uint8(0)
	data := make([]byte, bufferSize)

	for eofs < maxEofs {
		n, err := conn.Read(data[read:])
		if err != nil {
			return //TODO: handle errors
		}

		read += n
		if hasReadId(read, msgId) {
			data = readId(&msgId, data, &read)
			logger.Infof("Received message ID: %d", msgId)
		}

		if hasReadPayloadSize(read, payloadSize) {
			data = readPayloadSize(&payloadSize, data, &read)
			logger.Infof("Read payload size: %d", payloadSize)
		}

		if hasReadCompletePayload(read, payloadSize) {
			logger.Infof("Received payload")
			err, data = processPayload(g, message.ID(msgId), data, payloadSize, &eofs)
			if err != nil {
				return //TODO: handle errors
			}
			msgId, read = 0, len(data)
		}
	}
}

func readPayloadSize(payloadSize *uint64, data []byte, read *int) []byte {
	var buf []byte
	*payloadSize, buf = ioutils.ReadU64FromSlice(data)
	*read -= LenFieldSize
	return buf
}

func readId(msgId *uint8, data []byte, read *int) []byte {
	var buf []byte
	*msgId, buf = ioutils.ReadU8FromSlice(data)
	*read -= MsgIdSize
	return buf
}

func hasReadCompletePayload(read int, payloadSize uint64) bool {
	return read >= int(payloadSize)
}

// hasReadPayloadSize returns true if the payload size field has been read (8 bytes)
func hasReadPayloadSize(read int, payloadSize uint64) bool {
	return read >= LenFieldSize && payloadSize == 0
}

// hasReadId returns true if the message ID field has been read (1 byte)
func hasReadId(read int, msgId uint8) bool {
	return read >= MsgIdSize && msgId == 0
}

// processPayload parses the data received from the client and appends it to the corresponding chunk
// Returns true if the end of the file was reached
// Moves the buffer payloadLen positions
func processPayload(g *Gateway, msgId message.ID, payload []byte, payloadLen uint64, eofs *uint8) (error, []byte) {

	if isEndOfFile(payloadLen) {
		logger.Infof("End of file received for message ID: %d", msgId)
		sendMsgToChunkSender(g, msgId, nil)
		*eofs++ //TODO podria validar q sea de diferente tipo de ID el eof... hashmap?
		return nil, payload
	}

	newMsg, err, payload := parseClientPayload(msgId, payload, payloadLen)
	if err != nil {
		return err, payload
	}

	sendMsgToChunkSender(g, msgId, newMsg)

	return nil, payload
}

// parseClientPayload parses the payload received from the client and returns a DataCsvReviews or DataCsvGames
func parseClientPayload(msgId message.ID, payload []byte, payloadLen uint64) (any, error, []byte) {
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
		return nil, err, payload
	}

	return data, nil, payload
}

func sendMsgToChunkSender(g *Gateway, msgId message.ID, newMsg any) {
	g.ChunkChan <- ChunkItem{msgId, newMsg}
}

func isEndOfFile(payloadLen uint64) bool {
	return payloadLen == 0
}
