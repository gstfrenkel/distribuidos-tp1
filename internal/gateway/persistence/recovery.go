package persistence

import (
	"bytes"
	"tp1/internal/gateway/utils"
	"tp1/pkg/amqp"
	"tp1/pkg/message"
	"tp1/pkg/recovery"
)

// HandleSimpleQueryRecovery Handle recovery for Q1, Q2, and Q3
func HandleSimpleQueryRecovery(
	recoveredMsg recovery.Record,
	recoveredMessages map[string]map[uint8]string,
) {
	clientId := recoveredMsg.Header().ClientId
	body := recoveredMsg.Message()
	originId := recoveredMsg.Header().OriginId

	if string(body) == utils.Ack {
		removeProcessedMessage(recoveredMessages, clientId, originId)
	} else {
		parsedBody, _ := utils.ParseMessageBody(originId, body)
		bodyStr, _ := utils.ResultBodyToString(originId, parsedBody)

		if _, exists := recoveredMessages[clientId]; !exists {
			recoveredMessages[clientId] = make(map[uint8]string)
		}
		recoveredMessages[clientId][originId] = bodyStr
	}
}

// HandleAccumulatingQueryRecovery Handle recovery for Q4 and Q5
func HandleAccumulatingQueryRecovery(
	recoveredMsg recovery.Record,
	clientAccumulatedResults map[string]map[uint8]string,
	recoveredMessages map[string]map[uint8]string,
) {
	clientId := recoveredMsg.Header().ClientId
	body := recoveredMsg.Message()
	originId := recoveredMsg.Header().OriginId

	if string(body) == utils.Ack {
		removeProcessedMessage(recoveredMessages, clientId, originId)
		removeProcessedMessage(clientAccumulatedResults, clientId, originId)
		return
	}

	if bytes.Equal(body, amqp.EmptyEof) {
		handleEofCase(clientId, originId, clientAccumulatedResults, recoveredMessages)
		return
	}

	accumulateResults(clientId, originId, body, clientAccumulatedResults)
}

func removeProcessedMessage(
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

func handleEofCase(
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

func accumulateResults(
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
