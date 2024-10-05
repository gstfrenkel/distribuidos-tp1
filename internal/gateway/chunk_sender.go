package gateway

import (
	"tp1/pkg/broker"
	"tp1/pkg/logs"
	"tp1/pkg/message"
)

type ChunkSender struct {
	channel      <-chan ChunkItem
	broker       broker.MessageBroker
	exchange     string
	reviewsCount uint8
	gamesCount   uint8
	reviewsChunk []message.DataCSVReviews
	gamesChunk   []message.DataCSVGames
	maxChunkSize uint8
	routingKeys  map[message.ID]string
}

type ChunkItem struct {
	MsgId message.ID
	Msg   []byte //DataCSVGames or DataCSVReviews bytes
}

type ToBytes func(any) ([]byte, error)

func wrapReviewFromClientReview(data any) ([]byte, error) {
	return message.ReviewFromClientReview(data.([]message.DataCSVReviews))
}

func wrapGamesFromClientGames(data any) ([]byte, error) {
	return message.GameFromClientGame(data.([]message.DataCSVGames))
}

func newChunkSender(channel <-chan ChunkItem, broker broker.MessageBroker, exchange string, chunkMaxSize uint8, rRoutingKey string, gRoutingKey string) *ChunkSender {
	return &ChunkSender{
		channel:      channel,
		broker:       broker,
		exchange:     exchange,
		reviewsCount: 0,
		gamesCount:   0,
		reviewsChunk: make([]message.DataCSVReviews, chunkMaxSize),
		gamesChunk:   make([]message.DataCSVGames, chunkMaxSize),
		maxChunkSize: chunkMaxSize,
		routingKeys: map[message.ID]string{
			message.ReviewIdMsg: rRoutingKey,
			message.GameIdMsg:   gRoutingKey,
		},
	}
}

func startChunkSender(channel <-chan ChunkItem, broker broker.MessageBroker, exchange string, chunkMaxSize uint8, rRoutingKey string, gRoutingKey string) {
	chunkSender := newChunkSender(channel, broker, exchange, chunkMaxSize, rRoutingKey, gRoutingKey)
	for {
		item := <-channel
		switch item.MsgId {
		case message.ReviewIdMsg:
			if item.Msg == nil { //eof
				chunkSender.updateReviewsChunk(message.DataCSVReviews{}, true)
			} else {
				chunkSender.updateReviewsChunk(parseClientPayload(message.ReviewIdMsg, item.Msg).(message.DataCSVReviews), false)
			}
		case message.GameIdMsg:
			if item.Msg == nil { //eof
				chunkSender.updateGamesChunk(message.DataCSVGames{}, true)
			} else {
				chunkSender.updateGamesChunk(parseClientPayload(message.GameIdMsg, item.Msg).(message.DataCSVGames), false)
			}
		default:
			logs.Logger.Errorf("Unsupported message ID: %d", item.MsgId)
		}
	}
}

// parseClientPayload parses the payload received from the client and returns a DataCsvReviews or DataCsvGames
func parseClientPayload(msgId message.ID, payload []byte) any {
	var data any

	switch msgId {
	case message.ReviewIdMsg:
		data, _ = message.DataCSVReviewsFromBytes(payload)
	case message.GameIdMsg:
		data, _ = message.DataCSVGamesFromBytes(payload)
	default:
		logs.Logger.Errorf("Unsupported message ID: %d", msgId)
	}

	return data
}

func (s *ChunkSender) updateReviewsChunk(reviews message.DataCSVReviews, eof bool) {
	if !eof {
		s.reviewsChunk[s.reviewsCount] = reviews
		s.reviewsCount++
		s.sendChunk(uint8(message.ReviewIdMsg), s.routingKeys[message.ReviewIdMsg], &s.reviewsCount, s.reviewsChunk, wrapReviewFromClientReview, eof)
	} else {
		s.reviewsCount = 0
		s.sendChunk(uint8(message.EofMsg), s.routingKeys[message.ReviewIdMsg], &s.reviewsCount, nil, wrapReviewFromClientReview, eof)
	}
}

func (s *ChunkSender) updateGamesChunk(games message.DataCSVGames, eof bool) {
	if !eof {
		s.gamesChunk[s.gamesCount] = games
		s.gamesCount++
		s.sendChunk(uint8(message.GameIdMsg), s.routingKeys[message.GameIdMsg], &s.gamesCount, s.gamesChunk, wrapGamesFromClientGames, eof)
	} else {
		s.gamesCount = 0
		s.sendChunk(uint8(message.EofMsg), s.routingKeys[message.GameIdMsg], &s.gamesCount, nil, wrapGamesFromClientGames, eof)
	}
}

// sendChunk sends a chunk of data to the broker if the chunk is full or the eof flag is true
// In case eof is true, it sends an EOF message to the broker
func (s *ChunkSender) sendChunk(msgId uint8, routingKey string, count *uint8, chunk any, mapper ToBytes, eof bool) {
	if (*count == s.maxChunkSize) || (eof && *count > 0) {
		bytes, err := mapper(chunk)
		if err != nil {
			logs.Logger.Errorf("Error converting chunk to bytes: %s", err.Error())
			return
		}

		err = s.broker.Publish(s.exchange, routingKey, msgId, bytes)
		if err != nil {
			logs.Logger.Errorf("Error publishing chunk: %s", err.Error())
			return
		}

		*count = 0
		chunk = make([]any, s.maxChunkSize)
		logs.Logger.Infof("Sent chunk with key %v", routingKey)
	} else if eof {
		err := s.broker.Publish(s.exchange, routingKey, msgId, nil)
		if err != nil {
			logs.Logger.Errorf("Error publishing EOF: %s", err.Error())
			return
		}
		logs.Logger.Infof("Sent eof with key %v", routingKey)
	}
}
