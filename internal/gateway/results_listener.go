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
	clientId := g.readClientId(cliConn)

	clientChan := make(chan []byte)
	g.clientChannels.Store(clientId, clientChan)

	defer func() {
		g.clientChannels.Delete(clientId)
		close(clientChan)
		cliConn.Close()
	}()

	for {
		rabbitMsg := <-clientChan
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

// ListenResults listens for results from the "reports" queue and sends them to the results channel
func (g *Gateway) ListenResults() {
	messages, err := g.broker.Consume(g.queues[len(g.queues)-1].Name, "", true, false)
	if err != nil {
		logs.Logger.Errorf("Failed to start consuming messages from reports_queue: %s", err.Error())
		return
	}

	// Accumulated results for queries 4 and 5 for each clientID
	clientAccumulatedResults := make(map[string]map[uint8]string)

	for m := range messages {
		g.handleMessage(m, clientAccumulatedResults)
	}
}

func (g *Gateway) handleMessage(m amqp.Delivery, clientAccumulatedResults map[string]map[uint8]string) {
	clientID := m.Headers[amqp.ClientIdHeader].(string)
	originID, ok := m.Headers[amqp.OriginIdHeader] //not all workers send this header
	if !ok {
		logs.Logger.Errorf("message missing origin ID")
		return
	}

	originIDUint8 := originID.(uint8)
	messageId, ok := m.Headers[amqp.MessageIdHeader]
	// Handle EOF or message content
	if originIDUint8 == amqp.Query4originId || originIDUint8 == amqp.Query5originId {
		if bytes.Equal(m.Body, amqp.EmptyEof) {
			g.handleEof(clientID, clientAccumulatedResults[clientID], originIDUint8)
		} else {
			initializeAccumulatedResultsForClient(clientAccumulatedResults, clientID)
			handleAppendMsg(originIDUint8, m, clientAccumulatedResults[clientID])
		}
	} else if !ok || message.ID(messageId.(uint8)) != message.EofMsg {
		result, err := parseMessageBody(originIDUint8, m.Body)
		if err != nil {
			logs.Logger.Errorf("Failed to parse message body: %v", err)
			return
		}
		g.handleResultMsg(clientID, originIDUint8, result)
	}
}

func initializeAccumulatedResultsForClient(clientAccumulatedResults map[string]map[uint8]string, clientID string) {
	if _, exists := clientAccumulatedResults[clientID]; !exists {
		clientAccumulatedResults[clientID] = map[uint8]string{
			amqp.Query4originId: "",
			amqp.Query5originId: "",
		}
	}
}

func (g *Gateway) handleResultMsg(clientID string, originIDUint8 uint8, result interface{}) {
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
		return
	}

	sendResultThroughChannel(g, clientID, resultStr)
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

func (g *Gateway) handleEof(clientID string, accumulatedResults map[uint8]string, originIDUint8 uint8) {
	result := accumulatedResults[originIDUint8]
	var resultStr string

	if originIDUint8 == amqp.Query4originId {
		resultStr = message.ToQ4ResultString(result)
	} else if originIDUint8 == amqp.Query5originId {
		resultStr = message.ToQ5ResultString(result)
	}

	sendResultThroughChannel(g, clientID, resultStr)
}

func sendResultThroughChannel(g *Gateway, clientID string, resultStr string) {
	if clientChanI, exists := g.clientChannels.Load(clientID); exists {
		clientChan := clientChanI.(chan []byte)
		clientChan <- []byte(resultStr)
	} else {
		logs.Logger.Errorf("No client channel found for clientID %v", clientID)
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
	case amqp.Query4originId, amqp.Query5originId:
		return nil, fmt.Errorf("parseMessageBody should not be called for queries 4 and 5")
	default:
		return nil, fmt.Errorf("unknown origin ID: %v", originID)
	}
}
