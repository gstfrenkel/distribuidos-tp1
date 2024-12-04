package gateway

import (
	"bytes"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"
	"tp1/pkg/recovery"
	"tp1/pkg/sequence"
)

// Handle recovery for Q1, Q2, and Q3
func (g *Gateway) handleSimpleQueryRecovery(
	recoveredMsg recovery.Record,
	recoveredMessages map[string]map[uint8]string,
) {
	clientId := recoveredMsg.Header().ClientId
	body := recoveredMsg.Message()
	originId := recoveredMsg.Header().OriginId

	if string(body) == ack {
		g.removeProcessedMessage(recoveredMessages, clientId, originId)
	} else {
		parsedBody, _ := parseMessageBody(originId, body)
		bodyStr, _ := resultBodyToString(originId, parsedBody)

		if _, exists := recoveredMessages[clientId]; !exists {
			recoveredMessages[clientId] = make(map[uint8]string)
		}
		recoveredMessages[clientId][originId] = bodyStr
	}
}

// Handle recovery for Q4 and Q5
func (g *Gateway) handleAccumulatingQueryRecovery(
	recoveredMsg recovery.Record,
	clientAccumulatedResults map[string]map[uint8]string,
	recoveredMessages map[string]map[uint8]string,
) {
	clientId := recoveredMsg.Header().ClientId
	body := recoveredMsg.Message()
	originId := recoveredMsg.Header().OriginId

	if string(body) == ack {
		g.removeProcessedMessage(recoveredMessages, clientId, originId)
		g.removeProcessedMessage(clientAccumulatedResults, clientId, originId)
		return
	}

	if bytes.Equal(body, amqp.EmptyEof) {
		g.handleEofCase(clientId, originId, clientAccumulatedResults, recoveredMessages)
		return
	}

	g.accumulateResults(clientId, originId, body, clientAccumulatedResults)
}

func (g *Gateway) removeProcessedMessage(
	messageMap map[string]map[uint8]string,
	clientId string,
	originId uint8,
) {
	if _, exists := messageMap[clientId]; exists {
		delete(messageMap[clientId], originId)
		if len(messageMap[clientId]) == 0 {
			delete(messageMap, clientId)
		}
	}
}

func (g *Gateway) handleEofCase(
	clientId string,
	originId uint8,
	clientAccumulatedResults map[string]map[uint8]string,
	recoveredMessages map[string]map[uint8]string,
) {

	if _, ok := clientAccumulatedResults[clientId]; !ok {
		clientAccumulatedResults[clientId] = make(map[uint8]string)
	}

	if _, ok := recoveredMessages[clientId]; !ok {
		recoveredMessages[clientId] = make(map[uint8]string)
	}

	result := clientAccumulatedResults[clientId][originId]

	if originId == amqp.Query4originId {
		resultStr := message.ToQ4ResultString(result)
		recoveredMessages[clientId][originId] = resultStr
	} else if originId == amqp.Query5originId {
		resultStr := message.ToQ5ResultString(result)
		recoveredMessages[clientId][originId] = resultStr
	}
}

func (g *Gateway) accumulateResults(
	clientId string,
	originId uint8,
	body []byte,
	clientAccumulatedResults map[string]map[uint8]string,
) {
	if _, ok := clientAccumulatedResults[clientId]; !ok {
		clientAccumulatedResults[clientId] = make(map[uint8]string)
	}
	bodyStr, _ := message.GameNamesFromBytes(body)
	clientAccumulatedResults[clientId][originId] += bodyStr.ToStringAux()
}

func (g *Gateway) recoverResults(
	ch chan recovery.Record,
	clientAccumulatedResults map[string]map[uint8]string,
	recoveredMessages map[string]map[uint8]string,
) {
	go g.recovery.Recover(ch)

	for recoveredMsg := range ch {
		originId := recoveredMsg.Header().OriginId
		sequenceId := recoveredMsg.Header().SequenceId

		if string(recoveredMsg.Message()) != ack {
			seqSource, err := sequence.SrcFromString(sequenceId)
			if err != nil {
				logs.Logger.Errorf("Failed to parse sequence source: %v", err)
				continue
			}

			g.dup.Add(*seqSource)
		}

		switch originId {
		case amqp.Query1originId, amqp.Query2originId, amqp.Query3originId:
			g.handleSimpleQueryRecovery(recoveredMsg, recoveredMessages)
		case amqp.Query4originId, amqp.Query5originId:
			g.handleAccumulatingQueryRecovery(
				recoveredMsg,
				clientAccumulatedResults,
				recoveredMessages,
			)
		default:
			logs.Logger.Infof("Header x-origin-id does not match any known origin IDs, got: %v", originId)
		}
	}
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

func (g *Gateway) logResults() {
	for record := range g.logChannel {
		if err := g.recovery.Log(record); err != nil {
			logs.Logger.Errorf("Failed to log record: %s", err)
		}
	}
}
