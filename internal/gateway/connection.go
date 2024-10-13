package gateway

import (
	"fmt"
	"net"
	"tp1/pkg/amqp"
	"tp1/pkg/ioutils"
	"tp1/pkg/logs"
	"tp1/pkg/message"
)

const LenFieldSize = 4
const TransportProtocol = "tcp"

func (g *Gateway) createGatewaySockets() error {
	gamesListener, err := g.newListener("gateway.games-address")
	if err != nil {
		return err
	}
	reviewsListener, err := g.newListener("gateway.reviews-address")
	if err != nil {
		return err
	}

	g.Listeners[GamesListener] = gamesListener
	g.Listeners[ReviewsListener] = reviewsListener
	logs.Logger.Infof("Gateway listening games on %s", gamesListener.Addr().String())
	logs.Logger.Infof("Gateway listening reviews on %s", reviewsListener.Addr().String())
	return nil
}

func (g *Gateway) newListener(configKey string) (net.Listener, error) {
	addr := g.Config.String(configKey, "")
	return net.Listen(TransportProtocol, addr)
}

func (g *Gateway) listenForNewClients(listener int) error {
	logs.Logger.Infof("Waiting for new client connections, listener %d...", listener)
	for {
		g.finishedMu.Lock()
		if g.finished {
			g.finishedMu.Unlock()
			break
		}
		g.finishedMu.Unlock()
		c, err := g.Listeners[listener].Accept()
		logs.Logger.Infof("Successfully established new connection! Listener %d...", listener)
		if err != nil {
			return err
		}

		if listener == ReviewsListener {
			go func() {
				g.HandleResults()
			}()
		}

		go g.handleConnection(c, matchMessageId(listener))
	}
	return nil
}

func (g *Gateway) handleConnection(c net.Conn, msgId message.ID) {
	sends := 0

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
			finished = g.processPayload(msgId, buf[:payloadSize], payloadSize)
			buf = ioutils.MoveBuff(buf, int(payloadSize))
		}
	}

	logs.Logger.Infof("%d - Received %d messages", msgId, sends)
	sendConfirmationToClient(c)
}

func readPayloadSize(data []byte) (uint32, []byte) {
	return ioutils.ReadU32FromSlice(data), ioutils.MoveBuff(data, LenFieldSize)
}

// processPayload parses the data received from the client and appends it to the corresponding chunk
// Returns true if the end of the file was reached
func (g *Gateway) processPayload(msgId message.ID, payload []byte, payloadSize uint32) bool {
	if isEndOfFile(payloadSize) {
		logs.Logger.Infof("End of file received for message ID: %d", msgId)
		g.sendMsgToChunkSender(msgId, nil)
		return true
	}

	g.sendMsgToChunkSender(msgId, payload)
	return false
}

func (g *Gateway) sendMsgToChunkSender(msgId message.ID, payload []byte) {
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

	g.ChunkChans[matchListenerId(msgId)] <- ChunkItem{msgId, data}
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

func ListenResults(g *Gateway) {
	reportsQueue := g.Config.String("rabbit_q.reports_q", "reports")
	messages, err := g.broker.Consume(reportsQueue, "", true, false)
	if err != nil {
		logs.Logger.Errorf("Failed to start consuming messages from reports_queue: %s", err.Error())
		return
	}

	for m := range messages {
		if originID, ok := m.Headers["x-origin-id"]; ok {
			if originIDUint8, ok := originID.(uint8); ok {
				result, err := parseMessageBody(originIDUint8, m.Body)
				if err != nil {
					logs.Logger.Errorf("Failed to parse message body into Platform struct: %v", err)
					continue
				}

				var resultStr string
				switch originIDUint8 {
				case amqp.Query1originId:
					resultStr = result.(message.Platform).ToResultString()
				case amqp.Query2originId:
					resultStr = result.(message.DateFilteredReleases).ToResultString()
				case amqp.Query3originId:
					resultStr = result.(message.ScoredReviews).ToQ3ResultString()
				case amqp.Query4originId:
					resultStr = result.(message.GameNames).ToResultString()
				case amqp.Query5originId:
					resultStr = result.(message.ScoredReviews).ToQ5ResultString()
				default:
					logs.Logger.Infof("Header x-origin-id does not match any known origin IDs, got: %v", originIDUint8)
				}

				g.resultsChan <- []byte(resultStr)
			} else {
				logs.Logger.Errorf("Header x-origin-id is not a valid uint8 value, got: %v", originID)
			}
		}
	}
}

func parseMessageBody(originID uint8, body []byte) (interface{}, error) {
	switch originID {
	case amqp.Query1originId:
		return message.PlatfromFromBytes(body)
	case amqp.Query2originId:
		return message.DateFilteredReleasesFromBytes(body)
	case amqp.Query3originId:
		return message.ScoredReviewsFromBytes(body)
	case amqp.Query4originId:
		return message.GameNamesFromBytes(body)
	case amqp.Query5originId:
		return message.ScoredReviewsFromBytes(body)
	default:
		return nil, fmt.Errorf("unknown origin ID: %v", originID)
	}
}
