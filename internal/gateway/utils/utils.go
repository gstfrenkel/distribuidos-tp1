package utils

import (
	"fmt"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"
)

const (
	ReviewsListener  = 1
	ResultsListener  = 2
	GamesListener    = 0
	ClientIdListener = 3
	Ack              = "ACK"
)

func ResultBodyToString(originIDUint8 uint8, result interface{}) (string, bool) {
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

func ParseMessageBody(originID uint8, body []byte) (interface{}, error) {
	switch originID {
	case amqp.Query1originId:
		return message.PlatfromFromBytes(body)
	case amqp.Query2originId:
		return message.DateFilteredReleasesFromBytes(body)
	case amqp.Query3originId:
		return message.ScoredReviewsFromBytes(body)
	case amqp.Query4originId, amqp.Query5originId:
		return nil, fmt.Errorf("ParseMessageBody should not be called for queries 4 and 5")
	default:
		return nil, fmt.Errorf("unknown origin ID: %v", originID)
	}
}

func MatchMessageId(listener int) message.ID {
	if listener == ReviewsListener {
		return message.ReviewIdMsg
	}
	return message.GameIdMsg
}

func MatchListenerId(msgId message.ID) int {
	if msgId == message.ReviewIdMsg {
		return ReviewsListener
	}
	return GamesListener
}
