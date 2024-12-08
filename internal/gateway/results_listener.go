package gateway

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"tp1/internal/gateway/utils"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"
	"tp1/pkg/recovery"
	"tp1/pkg/sequence"
	"tp1/pkg/utils/io"
)

const (
	idPos    = 1
	zeroChar = '0'
)

// ListenResultsRequests waits until a client connects to the results listener and sends the results to the client
func (g *Gateway) listenResultsRequests() error {
	return g.listenForConnections(utils.ResultsListener, g.SendResults)
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

		if err := io.SendAll(cliConn, data); err != nil {
			logs.Logger.Errorf("Error sending message to client: %s", err)
			return
		}

		readAck(cliConn)
		originId := uint8(rabbitMsg[idPos]-zeroChar) + 1
		g.logChannel <- recovery.NewRecord(amqp.Header{ClientId: clientId, OriginId: originId}, nil, []byte(utils.Ack))
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
	clientAccumulatedResults := make(map[string]map[uint8]string)
	recoveredMessages := make(map[string]map[uint8]string)
	g.recoverResults(ch, clientAccumulatedResults, recoveredMessages)

	for clientID, innerMap := range recoveredMessages {
		for _, resultStr := range innerMap {
			sendResultThroughChannel(g, clientID, resultStr)
		}
	}

	for m := range messages {
		g.handleMessage(m, clientAccumulatedResults)
		err = m.Ack(false)
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
	sequenceId := m.Headers[amqp.SequenceIdHeader].(string)

	seqSource, err := sequence.SrcFromString(sequenceId)
	if err != nil {
		logs.Logger.Errorf("Failed to parse sequence source: %v", err)
		return
	}

	if g.dup.IsDuplicate(*seqSource, clientID) {
		return
	} else {
		g.dup.RecoverSequenceId(*seqSource, clientID)
	}

	g.logChannel <- recovery.NewRecord(amqp.HeadersFromDelivery(m), nil, m.Body)

	// Handle EOF or message content
	if originIDUint8 == amqp.Query4OriginId || originIDUint8 == amqp.Query5OriginId {
		if bytes.Equal(m.Body, amqp.EmptyEof) {
			g.handleEof(clientID, clientAccumulatedResults[clientID], originIDUint8)
		} else {
			initializeAccumulatedResultsForClient(clientAccumulatedResults, clientID)
			handleAppendMsg(originIDUint8, m, clientAccumulatedResults[clientID])
		}
	} else if !ok || message.Id(messageId.(uint8)) != message.EofId {
		result, err := utils.ParseMessageBody(originIDUint8, m.Body)
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
			amqp.Query4OriginId: "",
			amqp.Query5OriginId: "",
		}
	}
}

func (g *Gateway) handleResultMsg(clientID string, originIDUint8 uint8, result interface{}) {
	resultStr, shouldReturn := utils.ResultBodyToString(originIDUint8, result)
	if shouldReturn {
		return
	}

	sendResultThroughChannel(g, clientID, resultStr)
}

func handleAppendMsg(originIDUint8 uint8, m amqp.Delivery, accumulatedResults map[uint8]string) {
	var body string

	switch originIDUint8 {
	case amqp.Query4OriginId:
		parsedBody, err := message.GameNamesFromBytes(m.Body)
		if err != nil {
			logs.Logger.Errorf("Failed to parse scored reviews: %v", err)
		}
		body = parsedBody.ToStringAux()
	case amqp.Query5OriginId:
		parsedBody, err := message.ScoredReviewsFromBytes(m.Body)
		if err != nil {
			logs.Logger.Errorf("Failed to parse games names: %v", err)
		}
		body = parsedBody.ToStringAux()
	}

	accumulatedResults[originIDUint8] = accumulatedResults[originIDUint8] + body
}

func (g *Gateway) handleEof(clientID string, accumulatedResults map[uint8]string, originIDUint8 uint8) {
	result := accumulatedResults[originIDUint8]
	var resultStr string

	if originIDUint8 == amqp.Query4OriginId {
		resultStr = message.ToQ4ResultString(result)
	} else if originIDUint8 == amqp.Query5OriginId {
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
	case amqp.Query1OriginId:
		return message.PlatfromFromBytes(body)
	case amqp.Query2OriginId:
		return message.DateFilteredReleasesFromBytes(body)
	case amqp.Query3OriginId:
		return message.ScoredReviewsFromBytes(body)
	case amqp.Query4OriginId, amqp.Query5OriginId:
		return nil, fmt.Errorf("parseMessageBody should not be called for queries 4 and 5")
	default:
		return nil, fmt.Errorf("unknown origin Id: %v", originID)
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
