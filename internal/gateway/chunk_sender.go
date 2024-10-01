package gateway

import (
	"tp1/pkg/broker"
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
	Msg   any //DataCSVGames or DataCSVReviews
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
			if item.Msg == nil {
				chunkSender.updateReviewsChunk(message.DataCSVReviews{}, true)
			} else {
				chunkSender.updateReviewsChunk(item.Msg.(message.DataCSVReviews), false)
			}
		case message.GameIdMsg:
			if item.Msg == nil {
				chunkSender.updateGamesChunk(message.DataCSVGames{}, true)
			} else {
				chunkSender.updateGamesChunk(item.Msg.(message.DataCSVGames), false)
			}
		default:
		}
	}
}

func (s ChunkSender) updateReviewsChunk(reviews message.DataCSVReviews, eof bool) {
	if !eof {
		s.reviewsChunk[s.reviewsCount] = reviews
		s.reviewsCount++
	}
	s.sendChunk(uint8(message.ReviewIdMsg), s.routingKeys[message.ReviewIdMsg], &s.reviewsCount, s.reviewsChunk, wrapReviewFromClientReview, eof)
}

func (s ChunkSender) updateGamesChunk(games message.DataCSVGames, eof bool) {
	if !eof {
		s.gamesChunk[s.gamesCount] = games
		s.gamesCount++
	}
	s.sendChunk(uint8(message.GameIdMsg), s.routingKeys[message.GameIdMsg], &s.gamesCount, s.gamesChunk, wrapGamesFromClientGames, eof)
}

// sendChunk sends a chunk of data to the broker if the chunk is full or the eof flag is true
// In case eof is true, it sends an EOF message to the broker
func (s ChunkSender) sendChunk(key uint8, routingKey string, count *uint8, chunk any, mapper ToBytes, eof bool) {
	if (*count == s.maxChunkSize) || (eof && *count > 0) {
		auxChunk := make([]byte, *count)
		auxChunk = append(auxChunk, key)
		auxChunk = append(auxChunk, *count)
		bytes, err := mapper(chunk)
		if err != nil { //TODO: handle error with a log
			return
		}
		auxChunk = append(auxChunk, bytes...)

		err = s.broker.Publish(s.exchange, string(key), key, auxChunk)
		if err != nil { //TODO: handle error
			return
		} //TODO: handle error

		*count = 0
		chunk = make([]any, s.maxChunkSize)
	} else if eof {
		err := s.broker.Publish(s.exchange, routingKey, key, []byte{uint8(message.EofMsg)})
		if err != nil { //TODO: handle error
			return
		}
	}
}
