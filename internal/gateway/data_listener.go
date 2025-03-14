package gateway

import (
	"encoding/binary"
	"net"
	"sync"
	"tp1/internal/gateway/chunk"
	"tp1/internal/gateway/utils"
	"tp1/pkg/logs"
	"tp1/pkg/message"
	"tp1/pkg/utils/id"
	"tp1/pkg/utils/io"
)

const (
	bufferSizeKey     = "gateway.buffer_size"
	defaultBufferSize = 1024
	eofPayloadSize    = 0
)

// listenForData listens for incoming games or reviews
func (g *Gateway) listenForData(listener int) error {
	return g.listenForConnections(listener, func(c net.Conn) {
		g.handleDataConnection(c, utils.MatchMessageId(listener))
	})
}

func (g *Gateway) handleDataConnection(c net.Conn, msgId message.Id) {
	clientId := g.readClientId(c)
	sends := 0
	auxBuf := make([]byte, g.Config.Int(bufferSizeKey, defaultBufferSize))
	buf := make([]byte, 0, g.Config.Int(bufferSizeKey, defaultBufferSize))
	finished := false

	clientChan := make(chan []byte)

	if msgId == message.ReviewId {
		g.clientReviewsAckChannels.Store(clientId, clientChan)
		go sendAcksToClient(&g.clientReviewsAckChannels, clientId, c)
	} else {
		g.clientGamesAckChannels.Store(clientId, clientChan)
		go sendAcksToClient(&g.clientGamesAckChannels, clientId, c)
	}

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

		for len(buf) >= LenFieldSize*2 {
			g.finishedMu.Lock()
			if g.finished {
				g.finishedMu.Unlock()
				break
			}
			g.finishedMu.Unlock()

			batchNumBytes := buf[:LenFieldSize]
			batchNum := binary.BigEndian.Uint32(batchNumBytes)

			buf = buf[LenFieldSize:]

			if len(buf) < LenFieldSize {
				buf = append(batchNumBytes, buf...)
				break
			}

			payloadSizeBytes := buf[:LenFieldSize]
			payloadSize := binary.BigEndian.Uint32(payloadSizeBytes)

			buf = buf[LenFieldSize:]

			if uint32(len(buf)) < payloadSize {
				buf = append(batchNumBytes, append(payloadSizeBytes, buf...)...)
				break
			}

			sends++
			finished = g.processPayload(msgId, buf[:payloadSize], payloadSize, clientId, batchNum)

			buf = buf[payloadSize:]
		}
	}
	logs.Logger.Infof("%d - Received %d messages", msgId, sends)
}

func (g *Gateway) readClientId(c net.Conn) string {
	clientId := make([]byte, id.ClientIdLen)
	if err := io.ReadFull(c, clientId, id.ClientIdLen); err != nil {
		logs.Logger.Errorf("Error reading client id from client: %s", err)
	}
	return id.DecodeClientId(clientId)
}

// processPayload parses the data received from the client and appends it to the corresponding chunks
// Returns true if the end of the file was reached
func (g *Gateway) processPayload(msgId message.Id, payload []byte, payloadSize uint32, clientId string, batchNum uint32) bool {
	if isEndOfFile(payloadSize) {
		logs.Logger.Infof("End of file received for message Id: %d", msgId)
		g.sendMsgToChunkSender(msgId, nil, clientId, batchNum)
		return true
	}

	g.sendMsgToChunkSender(msgId, payload, clientId, batchNum)
	return false
}

func (g *Gateway) sendMsgToChunkSender(msgId message.Id, payload []byte, clientId string, batchNum uint32) {
	var data any
	if payload != nil {
		if msgId == message.ReviewId {
			data, _ = message.DataCSVReviewsFromBytes(payload)
		} else {
			data, _ = message.DataCSVGamesFromBytes(payload)
		}
	} else {
		data = nil
	}

	g.ChunkChans[utils.MatchListenerId(msgId)] <- chunk.Item{Msg: data, ClientId: clientId, BatchNum: batchNum}
}

func isEndOfFile(payloadSize uint32) bool {
	return payloadSize == eofPayloadSize
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
