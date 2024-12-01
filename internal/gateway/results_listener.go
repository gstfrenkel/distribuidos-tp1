package gateway

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"tp1/pkg/recovery"

	"tp1/pkg/amqp"
	"tp1/pkg/ioutils"
	"tp1/pkg/logs"
	"tp1/pkg/message"
)

const (
	ack      = "ACK"
	idPos    = 1
	Query1Id = '1'
	Query2Id = '2'
	Query3Id = '3'
	Query4Id = '4'
	Query5Id = '5'
)

// ListenResultsRequests waits until a client connects to the results listener and sends the results to the client
func (g *Gateway) listenResultsRequests() error {
	return g.listenForConnections(ResultsListener, g.SendResults)
}

// SendResults gets reports from the result chan and sends them to the client
func (g *Gateway) SendResults(cliConn net.Conn) {
	clientId := g.readClientId(cliConn)
	clientChanI, _ := g.clientChannels.LoadOrStore(clientId, make(chan []byte))
	clientChan := clientChanI.(chan []byte)

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

		readAck(cliConn)
		originId := getOriginID(rabbitMsg)
		header := amqp.Header{ClientId: clientId, OriginId: originId}
		g.recoveryMu.Lock()
		if err := g.recovery.Log(recovery.NewRecord(header, nil, []byte(ack))); err != nil {
			logs.Logger.Errorf("Failed to Log: %s", err)
		}
		g.recoveryMu.Unlock()
	}
}

// ListenResults listens for results from the "reports" queue and sends them to the results channel
func (g *Gateway) ListenResults() {
	messages, err := g.broker.Consume(g.queues[len(g.queues)-1].Name, "", false, false)
	if err != nil {
		logs.Logger.Errorf("Failed to start consuming messages from reports_queue: %s", err.Error())
		return
	}

	ch := make(chan recovery.Record)

	g.recoveryMu.Lock()
	go g.recovery.Recover(ch)

	// Accumulated results for queries 4 and 5 for each clientID
	clientAccumulatedResults := make(map[string]map[uint8]string)
	recoveredMessages := make(map[string]map[uint8]string)

	for recoveredMsg := range ch {
		switch recoveredMsg.Header().OriginId {

		case amqp.Query1originId:
			clientId := recoveredMsg.Header().ClientId
			body := recoveredMsg.Message()
			if string(body) == ack {
				if _, exists := recoveredMessages[clientId]; exists {
					delete(recoveredMessages[clientId], amqp.Query1originId)
					if len(recoveredMessages[clientId]) == 0 {
						delete(recoveredMessages, clientId)
					}
				}
			} else {

				parsedBody, _ := parseMessageBody(amqp.Query1originId, body)
				bodyStr, _ := resultBodyToString(amqp.Query1originId, parsedBody)
				recoveredMessages[clientId] = map[uint8]string{
					amqp.Query1originId: bodyStr,
				}
			}

		case amqp.Query2originId:
			clientId := recoveredMsg.Header().ClientId
			body := recoveredMsg.Message()
			if string(body) == ack {
				if _, exists := recoveredMessages[clientId]; exists {
					delete(recoveredMessages[clientId], amqp.Query2originId)
					if len(recoveredMessages[clientId]) == 0 {
						delete(recoveredMessages, clientId)
					}
				}
			} else {
				parsedBody, _ := parseMessageBody(amqp.Query2originId, body)
				bodyStr, _ := resultBodyToString(amqp.Query2originId, parsedBody)
				recoveredMessages[clientId] = map[uint8]string{
					amqp.Query2originId: bodyStr,
				}
			}
		case amqp.Query3originId:
			clientId := recoveredMsg.Header().ClientId
			body := recoveredMsg.Message()
			if string(body) == ack {
				if _, exists := recoveredMessages[clientId]; exists {
					delete(recoveredMessages[clientId], amqp.Query3originId)
					if len(recoveredMessages[clientId]) == 0 {
						delete(recoveredMessages, clientId)
					}
				}

			} else {
				parsedBody, _ := parseMessageBody(amqp.Query3originId, body)
				bodyStr, _ := resultBodyToString(amqp.Query3originId, parsedBody)
				recoveredMessages[clientId] = map[uint8]string{
					amqp.Query3originId: bodyStr,
				}
			}
		case amqp.Query4originId:

			clientId := recoveredMsg.Header().ClientId
			body := recoveredMsg.Message()

			if string(body) == ack {

				// borrar de recovered messages y de accumulated results
				if _, exists := recoveredMessages[clientId]; exists {
					delete(recoveredMessages[clientId], amqp.Query4originId)
					if len(recoveredMessages[clientId]) == 0 {
						delete(recoveredMessages, clientId)
					}
				}

				if _, exists := clientAccumulatedResults[clientId]; exists {
					delete(clientAccumulatedResults[clientId], amqp.Query4originId)
					if len(clientAccumulatedResults[clientId]) == 0 {
						delete(clientAccumulatedResults, clientId)
					}
				}

			} else if bytes.Equal(body, amqp.EmptyEof) {

				// agarrar lo de acum results y poner para mandar

				accumulatedResults, ok := clientAccumulatedResults[clientId]
				if !ok {
					accumulatedResults = make(map[uint8]string)
					clientAccumulatedResults[clientId] = accumulatedResults
				}

				clientRecoveredMessages, ok := recoveredMessages[clientId]
				if !ok {
					clientRecoveredMessages = make(map[uint8]string)
					recoveredMessages[clientId] = clientRecoveredMessages
				}

				clientRecoveredMessages[amqp.Query4originId] = accumulatedResults[amqp.Query4originId]

			} else {
				// guardar resultsaod en accumulated results
				// guardar result en accumulated results
				bodyStr, _ := message.GameNamesFromBytes(body)

				// Ensure the map exists and initialize it if it doesn't
				if accumulatedResults, ok := clientAccumulatedResults[clientId]; !ok {
					accumulatedResults = make(map[uint8]string)
					clientAccumulatedResults[clientId] = accumulatedResults
				}

				// Now you can safely modify the map
				accumulatedResults := clientAccumulatedResults[clientId]
				accumulatedResults[amqp.Query4originId] = accumulatedResults[amqp.Query4originId] + bodyStr.ToStringAux()
			}

		case amqp.Query5originId:
			clientId := recoveredMsg.Header().ClientId
			body := recoveredMsg.Message()

			if string(body) == ack {

				// borrar de recovered messages y de accumulated results
				if _, exists := recoveredMessages[clientId]; exists {
					delete(recoveredMessages[clientId], amqp.Query5originId)
					if len(recoveredMessages[clientId]) == 0 {
						delete(recoveredMessages, clientId)
					}
				}

				if _, exists := clientAccumulatedResults[clientId]; exists {
					delete(clientAccumulatedResults[clientId], amqp.Query5originId)
					if len(clientAccumulatedResults[clientId]) == 0 {
						delete(clientAccumulatedResults, clientId)
					}
				}

			} else if bytes.Equal(body, amqp.EmptyEof) {

				accumulatedResults, ok := clientAccumulatedResults[clientId]
				if !ok {
					accumulatedResults = make(map[uint8]string)
					clientAccumulatedResults[clientId] = accumulatedResults
				}

				clientRecoveredMessages, ok := recoveredMessages[clientId]
				if !ok {
					clientRecoveredMessages = make(map[uint8]string)
					recoveredMessages[clientId] = clientRecoveredMessages
				}

				clientRecoveredMessages[amqp.Query5originId] = accumulatedResults[amqp.Query5originId]

			} else {
				// guardar result en accumulated results
				bodyStr, _ := message.GameNamesFromBytes(body)

				// Ensure the map exists and initialize it if it doesn't
				if accumulatedResults, ok := clientAccumulatedResults[clientId]; !ok {
					accumulatedResults = make(map[uint8]string)
					clientAccumulatedResults[clientId] = accumulatedResults
				}

				// Now you can safely modify the map
				accumulatedResults := clientAccumulatedResults[clientId]
				accumulatedResults[amqp.Query5originId] = accumulatedResults[amqp.Query5originId] + bodyStr.ToStringAux()
			}
		default:
			//
		}
	}

	g.recoveryMu.Unlock()

	for clientID, innerMap := range recoveredMessages {
		for _, resultStr := range innerMap {
			sendResultThroughChannel(g, clientID, resultStr)
		}
	}

	for m := range messages {
		g.handleMessage(m, clientAccumulatedResults)

		err := m.Ack(false)
		if err != nil {
			logs.Logger.Errorf("Failed to acknowledge message: %s", err.Error())
		}
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

	g.recoveryMu.Lock()
	header := amqp.Header{ClientId: clientID, OriginId: originIDUint8}
	if err := g.recovery.Log(recovery.NewRecord(header, nil, m.Body)); err != nil {
		logs.Logger.Errorf("Failed to Log: %s", err)
	}
	g.recoveryMu.Unlock()

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
	resultStr, shouldReturn := resultBodyToString(originIDUint8, result)
	if shouldReturn {
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
	clientChanI, _ := g.clientChannels.LoadOrStore(clientID, make(chan []byte))
	clientChan := clientChanI.(chan []byte)
	clientChan <- []byte(resultStr)
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

func readAck(conn net.Conn) error {
	ackBuf := make([]byte, 1)
	_, err := conn.Read(ackBuf)
	if err != nil {
		return fmt.Errorf("error reading ACK: %w", err)
	}

	return nil
}

func getOriginID(rabbitMsg []byte) uint8 {
	var originId uint8

	switch rabbitMsg[idPos] {
	case Query1Id:
		originId = amqp.Query1originId
	case Query2Id:
		originId = amqp.Query2originId
	case Query3Id:
		originId = amqp.Query3originId
	case Query4Id:
		originId = amqp.Query4originId
	case Query5Id:
		originId = amqp.Query5originId
	}

	return originId
}

func resultBodyToString(originIDUint8 uint8, result interface{}) (string, bool) {
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
		return "", true
	}
	return resultStr, false
}
