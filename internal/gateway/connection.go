package gateway

import (
	"net"
	"tp1/pkg/ioutils"
	"tp1/pkg/message"
)

const MsgIdSize = 1
const LenFieldSize = 8
const TransportProtocol = "tcp"
const maxEofs = 2

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
	buffer := make([]byte, bufferSize)
	var auxBuffer []byte

	for eofs < maxEofs {
		n, err := conn.Read(buffer[read:])
		if err != nil {
			logger.Errorf("Error reading from client: %s", err.Error())
			return
		}

		read += n
		if hasReadId(read, msgId) {
			buffer = readId(&msgId, buffer, &read)
			logger.Infof("Received message ID: %d", msgId)
		}

		if hasReadPayloadSize(read, payloadSize) {
			buffer = readPayloadSize(&payloadSize, buffer, &read)
			logger.Infof("Payload size: %d", payloadSize)
		}

		auxBuffer = append(auxBuffer, buffer[:read]...)
		read = 0
		if hasReadCompletePayload(auxBuffer, payloadSize) {
			err = processPayload(g, message.ID(msgId), auxBuffer[:payloadSize], payloadSize, &eofs)
			if err != nil {
				logger.Errorf("Error processing payload: %s", err.Error())
				return
			}

			auxBuffer = processRemaining(auxBuffer[payloadSize:], &msgId, &payloadSize)
		}
	}
}

func processRemaining(rem []byte, msgId *uint8, payloadSize *uint64) []byte {
	*msgId, *payloadSize = 0, 0
	if hasReadId(len(rem), *msgId) { //todo como se mueve el buf aca?
		rem = readId(msgId, rem, nil)
	}

	if hasReadPayloadSize(len(rem), *payloadSize) {
		rem = readPayloadSize(payloadSize, rem, nil)
	}

	return rem
}

func readPayloadSize(payloadSize *uint64, data []byte, read *int) []byte {
	*payloadSize = ioutils.ReadU64FromSlice(data)
	if read != nil && *read >= LenFieldSize {
		*read -= LenFieldSize
	}

	return moveBuff(data, LenFieldSize)
}

// moveBuff moves the buffer n positions to the left keeping the original capacity
func moveBuff(data []byte, n int) []byte {
	copy(data, data[n:])
	return data[:len(data)-n]
}

func readId(msgId *uint8, data []byte, read *int) []byte {
	*msgId = ioutils.ReadU8FromSlice(data)
	if read != nil && *read >= MsgIdSize {
		*read -= MsgIdSize
	}

	return moveBuff(data, MsgIdSize)
}

func hasReadCompletePayload(buf []byte, payloadSize uint64) bool {
	return len(buf) >= int(payloadSize)
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
func processPayload(g *Gateway, msgId message.ID, payload []byte, payloadLen uint64, eofs *uint8) error {
	if isEndOfFile(payloadLen) {
		logger.Infof("End of file received for message ID: %d", msgId)
		sendMsgToChunkSender(g, msgId, nil)
		*eofs++ //TODO podria validar q sea de diferente tipo de ID el eof... hashmap?
		return nil
	}

	newMsg, err := parseClientPayload(msgId, payload)
	if err != nil {
		return err
	}
	logger.Infof("Parsed message ID: %d", msgId)

	sendMsgToChunkSender(g, msgId, newMsg)

	return nil
}

// parseClientPayload parses the payload received from the client and returns a DataCsvReviews or DataCsvGames
func parseClientPayload(msgId message.ID, payload []byte) (any, error) {
	var (
		data any
		err  error
	)

	switch msgId {
	case message.ReviewIdMsg:
		data, err = message.DataCSVReviewsFromBytes(payload)
	case message.GameIdMsg:
		data, err = message.DataCSVGamesFromBytes(payload)
	}

	if err != nil {
		return nil, err
	}

	return data, nil
}

func sendMsgToChunkSender(g *Gateway, msgId message.ID, newMsg any) {
	g.ChunkChan <- ChunkItem{msgId, newMsg}
	logger.Infof("Sent message: %d to chunk sender", msgId)
}

func isEndOfFile(payloadLen uint64) bool {
	return payloadLen == 0
}
