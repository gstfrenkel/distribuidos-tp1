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
	//return message.GameFromClientGames(data.([]message.DataCSVGames)) TODO: implement
	return nil, nil
}

func newChunkSender(channel <-chan ChunkItem, broker broker.MessageBroker, exchange string, chunkMaxSize uint8) *ChunkSender {
	return &ChunkSender{
		channel:      channel,
		broker:       broker,
		exchange:     exchange,
		reviewsCount: 0,
		gamesCount:   0,
		reviewsChunk: make([]message.DataCSVReviews, chunkMaxSize),
		gamesChunk:   make([]message.DataCSVGames, chunkMaxSize),
		maxChunkSize: chunkMaxSize,
	}
}

func startChunkSender(channel <-chan ChunkItem, broker broker.MessageBroker, exchange string, chunkMaxSize uint8) {
	chunkSender := newChunkSender(channel, broker, exchange, chunkMaxSize)
	for {
		item := <-channel
		switch item.MsgId {
		case message.ReviewIdMsg:
			chunkSender.updateReviewsChunk(item.Msg.(message.DataCSVReviews))
		case message.GameIdMsg:
			chunkSender.updateGamesChunk(item.Msg.(message.DataCSVGames))
		case message.EofMsg: //TODO
			//chunkSender.sendChunk(uint8(message.EofMsg))
		default:
			//TODO: handle error
		}
	}

}

func (s ChunkSender) updateReviewsChunk(reviews message.DataCSVReviews) {
	s.reviewsChunk[s.reviewsCount] = reviews
	s.reviewsCount++
	s.sendChunk(uint8(message.ReviewIdMsg), &s.reviewsCount, s.reviewsChunk, wrapReviewFromClientReview)
}

func (s ChunkSender) updateGamesChunk(games message.DataCSVGames) {
	s.gamesChunk[s.gamesCount] = games
	s.gamesCount++
	s.sendChunk(uint8(message.GameIdMsg), &s.gamesCount, s.gamesChunk, wrapGamesFromClientGames)
}

func (s ChunkSender) sendChunk(key uint8, count *uint8, chunk any, mapper ToBytes) {
	if *count == s.maxChunkSize || key == uint8(message.EofMsg) { //TODO: handle EOF
		auxChunk := make([]byte, *count)
		auxChunk = append(auxChunk, key)
		auxChunk = append(auxChunk, *count)
		bytes, err := mapper(chunk)
		if err != nil { //TODO: handle error with a log
			return
		}
		auxChunk = append(auxChunk, bytes...)

		err = s.broker.Publish(s.exchange, string(key), key, auxChunk)
		*count = 0
		chunk = make([]any, s.maxChunkSize)
		if err != nil { //TODO: handle error
			return
		}
	}
}
