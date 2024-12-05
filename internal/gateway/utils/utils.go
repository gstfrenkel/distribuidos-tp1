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
	case amqp.Query1OriginId:
		resultStr = result.(message.Platform).ToResultString()
	case amqp.Query2OriginId:
		resultStr = result.(message.DateFilteredReleases).ToResultString()
	case amqp.Query3OriginId:
		resultStr = result.(message.ScoredReviews).ToQ3ResultString()
	default:
		logs.Logger.Infof("Header x-origin-id does not match any known origin IDs, got: %v", originIDUint8)
		return "", true
	}
	return resultStr, false
}

func ParseMessageBody(originID uint8, body []byte) (interface{}, error) {
	switch originID {
	case amqp.Query1OriginId:
		return message.PlatfromFromBytes(body)
	case amqp.Query2OriginId:
		return message.DateFilteredReleasesFromBytes(body)
	case amqp.Query3OriginId:
		return message.ScoredReviewsFromBytes(body)
	case amqp.Query4OriginId, amqp.Query5OriginId:
		return nil, fmt.Errorf("ParseMessageBody should not be called for queries 4 and 5")
	default:
		return nil, fmt.Errorf("unknown origin ID: %v", originID)
	}
}

func MatchMessageId(listener int) message.Id {
	if listener == ReviewsListener {
		return message.ReviewId
	}
	return message.GameId
}

func MatchListenerId(msgId message.Id) int {
	if msgId == message.ReviewId {
		return ReviewsListener
	}
	return GamesListener
}
