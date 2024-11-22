package gateway

import (
	"net"
	"sync"
	"tp1/internal/gateway/id_generator"
	"tp1/pkg/ioutils"
	"tp1/pkg/logs"
	"tp1/pkg/message"
)

// listenForData listens for incoming games or reviews
func (g *Gateway) listenForData(listener int) error {
	return g.listenForConnections(listener, func(c net.Conn) {
		g.handleDataConnection(c, matchMessageId(listener))
	})
}

func (g *Gateway) handleDataConnection(c net.Conn, msgId message.ID) {
	clientId := g.readClientId(c)
	sends := 0

	clientChan := make(chan []byte)

	if msgId == message.ReviewIdMsg {
		g.clientReviewsAckChannels.Store(clientId, clientChan)
		go sendAcksToClient(&g.clientReviewsAckChannels, clientId, c)
	} else {
		g.clientGamesAckChannels.Store(clientId, clientChan)
		go sendAcksToClient(&g.clientGamesAckChannels, clientId, c)
	}

	auxBuf := make([]byte, g.Config.Int("gateway.buffer_size", 1024))
	buf := make([]byte, 0, g.Config.Int("gateway.buffer_size", 1024))
	finished := false

	for !finished {
		g.finishedMu.Lock()
		if g.finished {
			g.finishedMu.Unlock()
			break
		}
		g.finishedMu.Unlock()
		n, err := c.Read(auxBuf)
		if err != nil {
			logs.Logger.Errorf("Error reading from listener: %s", err)
			return
		}

		buf = append(buf, auxBuf[:n]...)

		for !finished {
			g.finishedMu.Lock()
			if g.finished {
				g.finishedMu.Unlock()
				break
			}
			g.finishedMu.Unlock()

			if len(buf) < LenFieldSize {
				break
			}

			payloadSize := ioutils.ReadU32FromSlice(buf)

			if uint32(len(buf)) < payloadSize+LenFieldSize {
				break
			}

			_, buf = readPayloadSize(buf)

			sends += 1
			finished = g.processPayload(msgId, buf[:payloadSize], payloadSize, clientId)
			buf = ioutils.MoveBuff(buf, int(payloadSize))
		}
	}

	logs.Logger.Infof("%d - Received %d messages", msgId, sends)
	sendConfirmationToClient(c)
}

func (g *Gateway) readClientId(c net.Conn) string {
	clientId := make([]byte, id_generator.ClientIdLen)
	if err := ioutils.ReadFull(c, clientId, id_generator.ClientIdLen); err != nil {
		logs.Logger.Errorf("Error reading client id from client: %s", err)
	}
	return id_generator.DecodeClientId(clientId)
}

func readPayloadSize(data []byte) (uint32, []byte) {
	return ioutils.ReadU32FromSlice(data), ioutils.MoveBuff(data, LenFieldSize)
}

// processPayload parses the data received from the client and appends it to the corresponding chunks
// Returns true if the end of the file was reached
func (g *Gateway) processPayload(msgId message.ID, payload []byte, payloadSize uint32, clientId string) bool {
	if isEndOfFile(payloadSize) {
		logs.Logger.Infof("End of file received for message ID: %d", msgId)
		g.sendMsgToChunkSender(msgId, nil, clientId)
		return true
	}

	g.sendMsgToChunkSender(msgId, payload, clientId)
	return false
}

func (g *Gateway) sendMsgToChunkSender(msgId message.ID, payload []byte, clientId string) {
	var data any
	if payload != nil {
		if msgId == message.ReviewIdMsg {
			data, _ = message.DataCSVReviewsFromBytes(payload)
		} else {
			data, _ = message.DataCSVGamesFromBytes(payload)
		}
	} else {
		data = nil
	}

	g.ChunkChans[matchListenerId(msgId)] <- ChunkItem{data, clientId}
}

func isEndOfFile(payloadSize uint32) bool {
	return payloadSize == 0
}

func sendConfirmationToClient(conn net.Conn) {
	eofMsg := message.ClientMessage{
		DataLen: 0,
		Data:    nil,
	}
	if err := message.SendMessage(conn, eofMsg); err != nil {
		logs.Logger.Error("Error sending confirmation message to client")
	}
}

func sendAcksToClient(clientAckChannels *sync.Map, clientId string, conn net.Conn) {
	if ch, ok := clientAckChannels.Load(clientId); ok {
		ackChannel := ch.(chan []byte)
		for ack := range ackChannel {
			// Send the ack byte to the client
			_, err := conn.Write(ack)
			if err != nil {
				logs.Logger.Errorf("Error sending ack to client %s: %v\n", clientId, err)
				break
			}
		}
	}
}
