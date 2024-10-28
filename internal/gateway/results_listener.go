package gateway

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"tp1/pkg/amqp"
	"tp1/pkg/ioutils"
	"tp1/pkg/logs"
	"tp1/pkg/message"
)

// ListenResultsRequests waits until a client connects to the results listener and sends the results to the client
func (g *Gateway) listenResultsRequests() error {
	return g.listenForConnections(ResultsListener, g.SendResults)
}

// SendResults gets reports from the result chan and sends them to the client
func (g *Gateway) SendResults(cliConn net.Conn) {

	defer cliConn.Close()

	for {
		select {
		case rabbitMsg := <-g.resultsChan:
			clientMsg := message.ClientMessage{
				DataLen: uint32(len(rabbitMsg)),
				Data:    rabbitMsg,
			}

			data := make([]byte, LenFieldSize+len(clientMsg.Data))
			binary.BigEndian.PutUint32(data[:LenFieldSize], clientMsg.DataLen)
			copy(data[LenFieldSize:], clientMsg.Data)

			if err := ioutils.SendAll(cliConn, data); err != nil {
				logs.Logger.Errorf("Error sending message to client: %s", err)
				return
			}
		}
	}
}

// ListenResults listens for results from the "reports" queue and sends them to the results channel
func (g *Gateway) ListenResults() {
	reportsQueue := g.Config.String("rabbit_q.reports_q", "reports")
	messages, err := g.broker.Consume(reportsQueue, "", true, false)
	if err != nil {
		logs.Logger.Errorf("Failed to start consuming messages from reports_queue: %s", err.Error())
		return
	}

	accumulatedResults := map[uint8]string{
		amqp.Query4originId: "",
		amqp.Query5originId: "",
	}

	for m := range messages {
		g.handleMessage(m, accumulatedResults)
	}
}

func (g *Gateway) handleMessage(m amqp.Delivery, accumulatedResults map[uint8]string) {
	if originID, ok := m.Headers["x-origin-id"]; ok {
		if originIDUint8, ok := originID.(uint8); ok {
			// Eof msg for queries 4 & 5
			if bytes.Equal(m.Body, amqp.EmptyEof) && (originIDUint8 == amqp.Query4originId || originIDUint8 == amqp.Query5originId) {
				g.handleEof(accumulatedResults, originIDUint8)
			} else {
				if originIDUint8 == amqp.Query4originId || originIDUint8 == amqp.Query5originId {
					// Append msg for queries 4 & 5
					handleAppendMsg(originIDUint8, m, accumulatedResults)
				} else {
					// Result msg for queries 1, 2 & 3
					result, err := parseMessageBody(originIDUint8, m.Body)
					if err != nil {
						logs.Logger.Errorf("Failed to parse message body into Platform struct: %v", err)
					}
					g.handleResultMsg(originIDUint8, result)
				}
			}
		} else {
			logs.Logger.Errorf("Header x-origin-id is not a valid uint8 value, got: %v", originID)
		}
	}
}

func (g *Gateway) handleResultMsg(originIDUint8 uint8, result interface{}) {
	var resultStr string
	switch originIDUint8 {
	case amqp.Query1originId:
		resultStr = result.(message.Platform).ToResultString()
	case amqp.Query2originId:
		resultStr = result.(message.DateFilteredReleases).ToResultString()
	case amqp.Query3originId:
		resultStr = result.(message.ScoredReviews).ToQ3ResultString()
	default:
		logs.Logger.Infof("Header x-origin-id does not match any known origin IDs, got: %v", originIDUint8)
	}
	g.resultsChan <- []byte(resultStr)
}

func handleAppendMsg(originIDUint8 uint8, m amqp.Delivery, accumulatedResults map[uint8]string) {
	switch originIDUint8 {
	case amqp.Query4originId:
		parsedBody, err := message.GameNamesFromBytes(m.Body)
		if err != nil {
			logs.Logger.Errorf("Failed to parse scored reviews: %v", err)

		}
		accumulatedResults[originIDUint8] = accumulatedResults[originIDUint8] + parsedBody.ToStringAux()
	case amqp.Query5originId:
		parsedBody, err := message.ScoredReviewsFromBytes(m.Body)
		if err != nil {
			logs.Logger.Errorf("Failed to parse games names: %v", err)

		}
		accumulatedResults[originIDUint8] = accumulatedResults[originIDUint8] + parsedBody.ToStringAux()
	}
}

func (g *Gateway) handleEof(accumulatedResults map[uint8]string, originIDUint8 uint8) {
	result := accumulatedResults[originIDUint8]

	var resultStr string

	if originIDUint8 == amqp.Query4originId {
		resultStr = message.ToQ4ResultString(result)
	} else if originIDUint8 == amqp.Query5originId {
		resultStr = message.ToQ5ResultString(result)
	}

	g.resultsChan <- []byte(resultStr)
}

func parseMessageBody(originID uint8, body []byte) (interface{}, error) {
	switch originID {
	case amqp.Query1originId:
		return message.PlatfromFromBytes(body)
	case amqp.Query2originId:
		return message.DateFilteredReleasesFromBytes(body)
	case amqp.Query3originId:
		return message.ScoredReviewsFromBytes(body)
	case amqp.Query4originId, amqp.Query5originId:
		return nil, fmt.Errorf("parseMessageBody should not be called for queries 4 and 5")
	default:
		return nil, fmt.Errorf("unknown origin ID: %v", originID)
	}
}
